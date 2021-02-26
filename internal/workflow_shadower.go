// Copyright (c) 2017-2021 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package internal

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/facebookgo/clock"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/internal/common/util"
	"go.uber.org/zap"
)

const (
	statusInitialized int32 = 0
	statusStarted     int32 = 1
	statusStopped     int32 = 2
)

type (
	// WorkflowShadowerOptions is used to configure a WorkflowShadower.
	WorkflowShadowerOptions struct {
		Domain string

		WorkflowQuery           string
		WorkflowTypes           []string
		WorkflowStatus          []string
		WorkflowStartTimeFilter *TimeFilter
		SamplingRate            float64

		ExitCondition *WorkflowShadowerExitCondition

		Logger *zap.Logger
	}

	// TimeFilter represents a time range through the min and max timestamp
	TimeFilter struct {
		MinTimestamp time.Time
		MaxTimestamp time.Time
	}

	// WorkflowShadowerExitCondition configures when the workflow shadower should exit.
	// If not specified shadower will exit after replaying all workflows satisfying the visibility query.
	WorkflowShadowerExitCondition struct {
		ExpirationTime time.Duration
		ShadowingCount int
	}

	// WorkflowShadower retrieves and replays workflow history from Cadence service to determine if there's any nondeterministic changes in the workflow definition
	WorkflowShadower struct {
		service  workflowserviceclient.Interface
		options  *WorkflowShadowerOptions
		replayer *WorkflowReplayer

		status     int32
		shutdownCh chan struct{}
		shutdownWG sync.WaitGroup

		clock clock.Clock
	}
)

// NewWorkflowShadower creates an instance of the WorkflowShadower
func NewWorkflowShadower(
	service workflowserviceclient.Interface,
	options *WorkflowShadowerOptions,
) (*WorkflowShadower, error) {
	if err := options.validateAndPopulateFields(); err != nil {
		return nil, err
	}
	return &WorkflowShadower{
		service:  service,
		options:  options,
		replayer: NewWorkflowReplayer(),

		status:     statusInitialized,
		shutdownCh: make(chan struct{}),

		clock: clock.New(),
	}, nil
}

// RegisterWorkflow registers workflow function to replay
func (s *WorkflowShadower) RegisterWorkflow(w interface{}) {
	s.replayer.RegisterWorkflow(w)
}

// RegisterWorkflowWithOptions registers workflow function with custom workflow name to replay
func (s *WorkflowShadower) RegisterWorkflowWithOptions(w interface{}, options RegisterWorkflowOptions) {
	s.replayer.RegisterWorkflowWithOptions(w, options)
}

// Run starts WorkflowShadower in a blocking fashion
func (s *WorkflowShadower) Run() error {
	if !atomic.CompareAndSwapInt32(&s.status, statusInitialized, statusStarted) {
		return errors.New("Workflow shadower already started")
	}

	return s.shadowWorker()
}

// Stop stops WorkflowShadower and wait up to one miniute for all goroutines to finish before returning
func (s *WorkflowShadower) Stop() {
	if !atomic.CompareAndSwapInt32(&s.status, statusStarted, statusStopped) {
		return
	}

	close(s.shutdownCh)

	if success := util.AwaitWaitGroup(&s.shutdownWG, time.Minute); !success {
		s.options.Logger.Warn("Workflow Shadower timedout on shutdown")
	}
}

func (s *WorkflowShadower) shadowWorker() error {
	s.shutdownWG.Add(1)
	defer s.shutdownWG.Done()

	scanRequest := scanWorkflowActivityParams{
		Domain:        s.options.Domain,
		WorkflowQuery: s.options.WorkflowQuery,
		SamplingRate:  s.options.SamplingRate,
	}
	s.options.Logger.Info("Shadow workflow query",
		zap.String(tagVisibilityQuery, s.options.WorkflowQuery),
	)

	ctx := context.Background()
	expirationTime := time.Unix(0, math.MaxInt64)
	if s.options.ExitCondition != nil && s.options.ExitCondition.ExpirationTime != 0 {
		expirationTime = s.clock.Now().Add(s.options.ExitCondition.ExpirationTime)
	}

	replayCount := 0
	maxReplayCount := math.MaxInt64
	if s.options.ExitCondition != nil && s.options.ExitCondition.ShadowingCount != 0 {
		maxReplayCount = s.options.ExitCondition.ShadowingCount
	}
	rand.Seed(s.clock.Now().UnixNano())
	for {
		scanResult, err := scanWorkflowExecutionsHelper(ctx, s.service, scanRequest, s.options.Logger)
		if err != nil {
			return err
		}

		for _, execution := range scanResult.Executions {
			if s.clock.Now().After(expirationTime) {
				return nil
			}

			success, err := replayWorkflowExecutionHelper(
				ctx,
				s.replayer,
				s.service,
				s.options.Logger,
				s.options.Domain,
				execution,
			)
			if err != nil {
				return err
			}
			if success {
				replayCount++
			}

			if replayCount == maxReplayCount {
				return nil
			}
		}

		if len(scanResult.NextPageToken) == 0 {
			return nil
		}

		scanRequest.NextPageToken = scanResult.NextPageToken
	}

}

func (o *WorkflowShadowerOptions) validateAndPopulateFields() error {
	// validate domain
	if len(o.Domain) == 0 {
		return fmt.Errorf("Domain is not set in options")
	}

	// validate workflow status
	if len(o.WorkflowQuery) == 0 {
		queryBuilder := NewQueryBuilder().WorkflowTypes(o.WorkflowTypes)

		statuses := make([]WorkflowStatus, 0, len(o.WorkflowStatus))
		for _, statusString := range o.WorkflowStatus {
			status, err := ToWorkflowStatus(statusString)
			if err != nil {
				return err
			}
			statuses = append(statuses, status)
		}
		queryBuilder.WorkflowStatus(statuses)

		if o.WorkflowStartTimeFilter != nil {
			if err := o.WorkflowStartTimeFilter.validateAndPopulateFields(); err != nil {
				return fmt.Errorf("invalid start time filter, error: %v", err)
			}
			queryBuilder.StartTime(o.WorkflowStartTimeFilter.MinTimestamp, o.WorkflowStartTimeFilter.MaxTimestamp)
		}

		o.WorkflowQuery = queryBuilder.Build()
	}

	if o.SamplingRate == 0 {
		// if not set, defaults to replay all workflows
		o.SamplingRate = 1
	}

	// use no-op logger if not specified
	if o.Logger == nil {
		o.Logger = zap.NewNop()
	}

	return nil
}

func (t *TimeFilter) validateAndPopulateFields() error {
	if t.MaxTimestamp.IsZero() {
		t.MaxTimestamp = maxTimestamp
	}

	return nil
}
