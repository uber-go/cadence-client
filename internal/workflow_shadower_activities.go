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
	"math/rand"
	"strings"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/backoff"
	"go.uber.org/cadence/internal/common/metrics"
	"go.uber.org/zap"
)

const (
	scanWorkflowActivityName            = "scanWorkflowActivity"
	replayWorkflowExecutionActivityName = "replayWorkflowExecutionActivity"
)

const (
	errMsgDomainNotExists           = "domain not exists"
	errMsgInvalidQuery              = "invalid visibility query"
	errMsgWorkflowTypeNotRegistered = "workflow type not registered"
)

type (
	scanWorkflowActivityParams struct {
		Domain        string
		WorkflowQuery string
		NextPageToken []byte
		PageSize      int
		SamplingRate  float64
	}

	scanWorkflowActivityResult struct {
		Executions    []shared.WorkflowExecution
		NextPageToken []byte
	}

	replayWorkflowActivityParams struct {
		Domain     string
		Executions []shared.WorkflowExecution
	}

	replayWorkflowActivityProgress struct {
		Result           replayWorkflowActivityResult
		NextExecutionIdx int
	}

	replayWorkflowActivityResult struct {
		Succeeded        int
		Skipped          int
		Failed           int
		FailedExecutions []shared.WorkflowExecution
	}
)

const (
	serviceClientContextKey    contextKey = "serviceClient"
	workflowReplayerContextKey contextKey = "workflowReplayer"
)

func scanWorkflowActivity(
	ctx context.Context,
	params scanWorkflowActivityParams,
) (scanWorkflowActivityResult, error) {
	logger := GetActivityLogger(ctx)
	service := ctx.Value(serviceClientContextKey).(workflowserviceclient.Interface)

	scanResult, err := scanWorkflowExecutionsHelper(ctx, service, params, logger)
	switch err.(type) {
	case *shared.EntityNotExistsError:
		err = NewCustomError(errMsgDomainNotExists, err.Error())
	case *shared.BadRequestError:
		err = NewCustomError(errMsgInvalidQuery, err.Error())
	}
	return scanResult, err
}

func scanWorkflowExecutionsHelper(
	ctx context.Context,
	service workflowserviceclient.Interface,
	params scanWorkflowActivityParams,
	logger *zap.Logger,
) (scanWorkflowActivityResult, error) {

	request := &shared.ListWorkflowExecutionsRequest{
		Domain:        common.StringPtr(params.Domain),
		Query:         common.StringPtr(params.WorkflowQuery),
		NextPageToken: params.NextPageToken,
	}
	if params.PageSize != 0 {
		request.PageSize = common.Int32Ptr(int32(params.PageSize))
	}

	var resp *shared.ListWorkflowExecutionsResponse
	if err := backoff.Retry(ctx,
		func() error {
			tchCtx, cancel, opt := newChannelContext(ctx)

			var err error
			resp, err = service.ScanWorkflowExecutions(tchCtx, request, opt...)
			cancel()

			return err
		},
		createDynamicServiceRetryPolicy(ctx),
		isServiceTransientError,
	); err != nil {
		logger.Error("Failed to scan workflow executions",
			zap.String(tagDomain, params.Domain),
			zap.String(tagVisibilityQuery, params.WorkflowQuery),
			zap.Error(err),
		)
		return scanWorkflowActivityResult{}, err
	}

	executions := make([]shared.WorkflowExecution, 0, len(resp.Executions))
	for _, execution := range resp.Executions {
		if shouldReplay(params.SamplingRate) && execution.Execution != nil {
			executions = append(executions, *execution.Execution)
		}
	}

	return scanWorkflowActivityResult{
		Executions:    executions,
		NextPageToken: resp.NextPageToken,
	}, nil
}

func shouldReplay(probability float64) bool {
	return rand.Float64() <= probability
}

func replayWorkflowExecutionActivity(
	ctx context.Context,
	params replayWorkflowActivityParams,
) (replayWorkflowActivityResult, error) {
	logger := GetActivityLogger(ctx)
	scope := tagScope(GetActivityMetricsScope(ctx), tagDomain, params.Domain, tagTaskList, GetActivityInfo(ctx).TaskList)
	service := ctx.Value(serviceClientContextKey).(workflowserviceclient.Interface)
	replayer := ctx.Value(workflowReplayerContextKey).(*WorkflowReplayer)

	var progress replayWorkflowActivityProgress
	if err := GetHeartbeatDetails(ctx, &progress); err != nil {
		progress = replayWorkflowActivityProgress{
			NextExecutionIdx: 0,
			Result:           replayWorkflowActivityResult{},
		}
	}

	for _, execution := range params.Executions[progress.NextExecutionIdx:] {
		sw := scope.Timer(metrics.ReplayLatency).Start()
		success, err := replayWorkflowExecutionHelper(ctx, replayer, service, logger, params.Domain, WorkflowExecution{
			ID:    execution.GetWorkflowId(),
			RunID: execution.GetRunId(),
		})
		if err != nil {
			scope.Counter(metrics.ReplayFailedCounter).Inc(1)
			progress.Result.Failed++
			progress.Result.FailedExecutions = append(progress.Result.FailedExecutions, execution)
			if isWorkflowTypeNotRegisteredError(err) {
				// this should fail the replay workflow as it requires worker deployment to fix the workflow registration.
				return progress.Result, NewCustomError(errMsgWorkflowTypeNotRegistered, err.Error())
			}
		} else if success {
			scope.Counter(metrics.ReplaySucceedCounter).Inc(1)
			progress.Result.Succeeded++
		} else {
			scope.Counter(metrics.ReplaySkippedCounter).Inc(1)
			progress.Result.Skipped++
		}
		sw.Stop()

		progress.NextExecutionIdx++
		RecordActivityHeartbeat(ctx, progress)
	}

	return progress.Result, nil
}

func replayWorkflowExecutionHelper(
	ctx context.Context,
	replayer *WorkflowReplayer,
	service workflowserviceclient.Interface,
	logger *zap.Logger,
	domain string,
	execution WorkflowExecution,
) (bool, error) {
	taggedLogger := logger.With(
		zap.String(tagWorkflowID, execution.ID),
		zap.String(tagRunID, execution.RunID),
	)

	err := replayer.ReplayWorkflowExecution(ctx, service, logger, domain, execution)
	if err == nil {
		taggedLogger.Info("Successfully replayed workflow")
		return true, nil
	}

	if isNondeterministicErr(err) || isWorkflowTypeNotRegisteredError(err) {
		taggedLogger.Error("Replay workflow failed", zap.Error(err))
		return false, err
	}

	taggedLogger.Info("Skipped replaying workflow", zap.Error(err))
	return false, nil
}

func isNondeterministicErr(err error) bool {
	// There're a few expected replay errors, for example:
	//   1. errReplayHistoryTooShort
	//   2. workflow not exist
	//   3. internal service error when reading workflow history
	// since we can't get an exhaustive list of expected errors, we only treat replay as failed
	// when we are sure the error is due to non-determinisim to make sure there's no false positive.
	// as shadowing doesn't guarantee to catch all nondeterministic errors.
	return strings.Contains(err.Error(), "nondeterministic")
}

func isWorkflowTypeNotRegisteredError(err error) bool {
	return strings.Contains(err.Error(), errMsgUnknownWorkflowType)
}
