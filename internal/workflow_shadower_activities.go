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
	"go.uber.org/cadence/.gen/go/shadower"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/backoff"
	"go.uber.org/cadence/internal/common/metrics"
	"go.uber.org/zap"
)

type (
	replayWorkflowActivityProgress struct {
		Result           shadower.ReplayWorkflowActivityResult
		NextExecutionIdx int
	}
)

const (
	serviceClientContextKey    contextKey = "serviceClient"
	workflowReplayerContextKey contextKey = "workflowReplayer"
)

func registerShadowerActivities(w *aggregatedWorker) {
	w.RegisterActivityWithOptions(scanWorkflowActivity, RegisterActivityOptions{
		Name: shadower.ScanWorkflowActivityName,
	})
	w.RegisterActivityWithOptions(replayWorkflowActivity, RegisterActivityOptions{
		Name: shadower.ReplayWorkflowActivityName,
	})
}

func scanWorkflowActivity(
	ctx context.Context,
	params shadower.ScanWorkflowActivityParams,
) (shadower.ScanWorkflowActivityResult, error) {
	logger := GetActivityLogger(ctx)
	service := ctx.Value(serviceClientContextKey).(workflowserviceclient.Interface)

	scanResult, err := scanWorkflowExecutionsHelper(ctx, service, params, logger)
	switch err.(type) {
	case *shared.EntityNotExistsError:
		err = NewCustomError(shadower.ErrReasonDomainNotExists, err.Error())
	case *shared.BadRequestError:
		err = NewCustomError(shadower.ErrReasonInvalidQuery, err.Error())
	}
	return scanResult, err
}

func scanWorkflowExecutionsHelper(
	ctx context.Context,
	service workflowserviceclient.Interface,
	params shadower.ScanWorkflowActivityParams,
	logger *zap.Logger,
) (shadower.ScanWorkflowActivityResult, error) {

	request := &shared.ListWorkflowExecutionsRequest{
		Domain:        params.Domain,
		Query:         params.WorkflowQuery,
		NextPageToken: params.NextPageToken,
		PageSize:      params.PageSize,
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
			zap.String(tagDomain, params.GetDomain()),
			zap.String(tagVisibilityQuery, params.GetWorkflowQuery()),
			zap.Error(err),
		)
		return shadower.ScanWorkflowActivityResult{}, err
	}

	executions := make([]*shared.WorkflowExecution, 0, len(resp.Executions))
	for _, execution := range resp.Executions {
		if shouldReplay(params.GetSamplingRate()) {
			executions = append(executions, execution.Execution)
		}
	}

	return shadower.ScanWorkflowActivityResult{
		Executions:    executions,
		NextPageToken: resp.NextPageToken,
	}, nil
}

func shouldReplay(probability float64) bool {
	if probability == 0 {
		return true
	}

	return rand.Float64() <= probability
}

func replayWorkflowActivity(
	ctx context.Context,
	params shadower.ReplayWorkflowActivityParams,
) (shadower.ReplayWorkflowActivityResult, error) {
	logger := GetActivityLogger(ctx)
	scope := tagScope(GetActivityMetricsScope(ctx), tagDomain, params.GetDomain(), tagTaskList, GetActivityInfo(ctx).TaskList)
	service := ctx.Value(serviceClientContextKey).(workflowserviceclient.Interface)
	replayer := ctx.Value(workflowReplayerContextKey).(*WorkflowReplayer)

	var progress replayWorkflowActivityProgress
	if err := GetHeartbeatDetails(ctx, &progress); err != nil {
		progress = replayWorkflowActivityProgress{
			NextExecutionIdx: 0,
			Result: shadower.ReplayWorkflowActivityResult{
				Succeeded: common.Int32Ptr(0),
				Skipped:   common.Int32Ptr(0),
				Failed:    common.Int32Ptr(0),
			},
		}
	}

	// following code assumes all pointers in progress.Result are not nil, this is ensured by:
	//   1. if not previous progress, init to pointer to 0
	//   2. if has previous progress, the progress uploaded during heartbeat has non nil pointers

	for _, execution := range params.Executions[progress.NextExecutionIdx:] {
		if execution == nil {
			continue
		}

		sw := scope.Timer(metrics.ReplayLatency).Start()
		success, err := replayWorkflowExecutionHelper(ctx, replayer, service, logger, params.GetDomain(), WorkflowExecution{
			ID:    execution.GetWorkflowId(),
			RunID: execution.GetRunId(),
		})
		if err != nil {
			scope.Counter(metrics.ReplayFailedCounter).Inc(1)
			*progress.Result.Failed++
			if isWorkflowTypeNotRegisteredError(err) {
				// this should fail the replay workflow as it requires worker deployment to fix the workflow registration.
				return progress.Result, NewCustomError(shadower.ErrReasonWorkflowTypeNotRegistered, err.Error())
			}
		} else if success {
			scope.Counter(metrics.ReplaySucceedCounter).Inc(1)
			*progress.Result.Succeeded++
		} else {
			scope.Counter(metrics.ReplaySkippedCounter).Inc(1)
			*progress.Result.Skipped++
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
