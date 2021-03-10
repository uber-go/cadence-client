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
	"go.uber.org/cadence/.gen/go/shadower"
	"go.uber.org/cadence/internal/common"
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
		WorkflowQuery           string
		WorkflowTypes           []string
		WorkflowStatus          []string
		WorkflowStartTimeFilter *TimeFilter
		SamplingRate            float64

		Mode          ShadowMode
		ExitCondition *ShadowExitCondition

		Concurrency int
	}

	// TimeFilter represents a time range through the min and max timestamp
	TimeFilter struct {
		MinTimestamp time.Time
		MaxTimestamp time.Time
	}

	ShadowMode int

	// ShadowExitCondition configures when the workflow shadower should exit.
	// If not specified shadower will exit after replaying all workflows satisfying the visibility query.
	ShadowExitCondition struct {
		ExpirationInterval time.Duration
		ShadowingCount     int
	}

	// WorkflowShadower retrieves and replays workflow history from Cadence service to determine if there's any nondeterministic changes in the workflow definition
	WorkflowShadower struct {
		service  workflowserviceclient.Interface
		domain   string
		options  *WorkflowShadowerOptions
		logger   *zap.Logger
		replayer *WorkflowReplayer

		status     int32
		shutdownCh chan struct{}
		shutdownWG sync.WaitGroup

		clock clock.Clock
	}
)

const (
	ShadowModeNormal ShadowMode = iota
	ShadowModeContinuous
)

// NewWorkflowShadower creates an instance of the WorkflowShadower
func NewWorkflowShadower(
	service workflowserviceclient.Interface,
	domain string,
	options *WorkflowShadowerOptions,
	logger *zap.Logger,
) (*WorkflowShadower, error) {
	if len(domain) == 0 {
		return nil, errors.New("domain is not set")
	}

	if err := options.validateAndPopulateFields(); err != nil {
		return nil, err
	}

	// use no-op logger if not specified
	if logger == nil {
		logger = zap.NewNop()
	}

	return &WorkflowShadower{
		service:  service,
		domain:   domain,
		options:  options,
		logger:   logger,
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
		s.logger.Warn("Workflow Shadower timedout on shutdown")
	}
}

func (s *WorkflowShadower) shadowWorker() error {
	s.shutdownWG.Add(1)
	defer s.shutdownWG.Done()

	scanRequest := shadower.ScanWorkflowActivityParams{
		Domain:        common.StringPtr(s.domain),
		WorkflowQuery: common.StringPtr(s.options.WorkflowQuery),
		SamplingRate:  common.Float64Ptr(s.options.SamplingRate),
	}
	s.logger.Info("Shadow workflow query",
		zap.String(tagVisibilityQuery, s.options.WorkflowQuery),
	)

	ctx := context.Background()
	expirationTime := time.Unix(0, math.MaxInt64)
	if s.options.ExitCondition != nil && s.options.ExitCondition.ExpirationInterval != 0 {
		expirationTime = s.clock.Now().Add(s.options.ExitCondition.ExpirationInterval)
	}

	replayCount := 0
	maxReplayCount := math.MaxInt64
	if s.options.ExitCondition != nil && s.options.ExitCondition.ShadowingCount != 0 {
		maxReplayCount = s.options.ExitCondition.ShadowingCount
	}
	rand.Seed(s.clock.Now().UnixNano())
	for {
		scanResult, err := scanWorkflowExecutionsHelper(ctx, s.service, scanRequest, s.logger)
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
				s.logger,
				s.domain,
				WorkflowExecution{
					ID:    execution.GetWorkflowId(),
					RunID: execution.GetRunId(),
				},
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

		if len(scanResult.NextPageToken) == 0 && s.options.Mode == ShadowModeNormal {
			return nil
		}

		scanRequest.NextPageToken = scanResult.NextPageToken
	}

}

func (o *WorkflowShadowerOptions) validateAndPopulateFields() error {
	exitConditionSpecified := o.ExitCondition != nil && (o.ExitCondition.ExpirationInterval > 0 || o.ExitCondition.ShadowingCount > 0)
	if o.Mode == ShadowModeContinuous && !exitConditionSpecified {
		return errors.New("exit condition must be specified if shadow mode is set to continuous")
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

	return nil
}

func (t *TimeFilter) validateAndPopulateFields() error {
	if t.MaxTimestamp.IsZero() {
		t.MaxTimestamp = maxTimestamp
	}

	return nil
}

func (m ShadowMode) toThriftPtr() *shadower.Mode {
	switch m {
	case ShadowModeNormal:
		return shadower.ModeNormal.Ptr()
	case ShadowModeContinuous:
		return shadower.ModeContinuous.Ptr()
	default:
		panic(fmt.Sprintf("unknown shadow mode %v", m))
	}
}

func (e *ShadowExitCondition) toThriftPtr() *shadower.ExitCondition {
	if e == nil {
		return nil
	}

	return &shadower.ExitCondition{
		ShadowCount:                 common.Int32Ptr(int32(e.ShadowingCount)),
		ExpirationIntervalInSeconds: common.Int32Ptr(int32(e.ExpirationInterval.Seconds())),
	}
}
