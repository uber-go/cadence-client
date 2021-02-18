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
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/backoff"
	"go.uber.org/cadence/internal/common/util"
	"go.uber.org/zap"
)

const (
	statusInitialized int32 = 0
	statusStarted     int32 = 1
	statusStopped     int32 = 2
)

type (
	// WorkflowShadowerOptions configs WorkflowShadower
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

	// TimeFilter represents a time range filter
	TimeFilter struct {
		MinTimestamp time.Time
		MaxTimestamp time.Time
	}

	// WorkflowShadowerExitCondition specifies when Shadower should stop shadowing and exit
	WorkflowShadowerExitCondition struct {
		ExpirationTime    time.Duration
		MaxShadowingCount int
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

	scanWorkflowReq := &shared.ListWorkflowExecutionsRequest{
		Domain: common.StringPtr(s.options.Domain),
		Query:  common.StringPtr(s.options.WorkflowQuery),
	}
	s.options.Logger.Info("Shadow workflow query",
		zap.String("Query", s.options.WorkflowQuery),
	)

	ctx := context.Background()
	expirationTime := time.Unix(0, math.MaxInt64)
	if s.options.ExitCondition != nil && s.options.ExitCondition.ExpirationTime != 0 {
		expirationTime = s.clock.Now().Add(s.options.ExitCondition.ExpirationTime)
	}

	replayCount := 0
	maxReplayCount := math.MaxInt64
	if s.options.ExitCondition != nil && s.options.ExitCondition.MaxShadowingCount != 0 {
		maxReplayCount = s.options.ExitCondition.MaxShadowingCount
	}
	rand.Seed(s.clock.Now().UnixNano())
	for {
		scanWorkflowResp, err := s.scanWorkflowExecutionsWithRetry(ctx, scanWorkflowReq)
		if err != nil {
			return err
		}

		for _, execution := range scanWorkflowResp.Executions {
			if s.clock.Now().After(expirationTime) {
				return nil
			}

			if rand.Float64() >= s.options.SamplingRate {
				continue
			}

			if err := s.replayer.ReplayWorkflowExecution(
				ctx,
				s.service,
				s.options.Logger,
				s.options.Domain,
				WorkflowExecution{
					ID:    execution.Execution.GetWorkflowId(),
					RunID: execution.Execution.GetRunId(),
				},
			); err != nil {
				if err == errReplayHistoryTooShort {
					// less than 3 history events, potentially cron workflow
					continue
				}

				switch err.(type) {
				case *shared.EntityNotExistsError:
					continue
				case *shared.InternalServiceError:
					// workflow not exist, or potentially corrupted workflow
					s.options.Logger.Warn("Skipped replaying workflow",
						zap.String("WorkflowID", execution.Execution.GetWorkflowId()),
						zap.String("RunID", execution.Execution.GetRunId()),
						zap.Error(err),
					)
					continue
				}
				return err
			}

			s.options.Logger.Info("Successfully replayed workflow",
				zap.String("WorkflowID", execution.Execution.GetWorkflowId()),
				zap.String("RunID", execution.Execution.GetRunId()),
			)

			replayCount++
			if replayCount == maxReplayCount {
				return nil
			}
		}

		if len(scanWorkflowResp.NextPageToken) == 0 {
			return nil
		}

		scanWorkflowReq.NextPageToken = scanWorkflowResp.NextPageToken
	}

}

func (s *WorkflowShadower) scanWorkflowExecutionsWithRetry(
	ctx context.Context,
	request *shared.ListWorkflowExecutionsRequest,
) (*shared.ListWorkflowExecutionsResponse, error) {
	var resp *shared.ListWorkflowExecutionsResponse
	if err := backoff.Retry(ctx,
		func() error {
			tchCtx, cancel, opt := newChannelContext(context.Background())

			var err error
			resp, err = s.service.ScanWorkflowExecutions(tchCtx, request, opt...)
			cancel()

			return err
		},
		createDynamicServiceRetryPolicy(ctx),
		isServiceTransientError,
	); err != nil {
		return nil, err
	}

	return resp, nil
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
