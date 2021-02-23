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

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/backoff"
	"go.uber.org/cadence/internal/common/metrics"
	"go.uber.org/zap"
)

type (
	// workflowExecution struct {
	// 	WorkflowID string
	// 	RunID      string
	// }

	scanWorkflowActivityParams struct {
		Domain        string
		WorkflowQuery string
		NextPageToken []byte
		SamplingRate  float64
	}

	scanWorkflowActivityResult struct {
		Executions    []WorkflowExecution
		NextPageToken []byte
	}

	replayWorkflowActivityParams struct {
		Domain     string
		Executions []WorkflowExecution
	}

	replayWorkflowActivityProgress struct {
		Result           replayWorkflowActivityResult
		NextExecutionIdx int
	}

	replayWorkflowActivityResult struct {
		Succeed int
		Failed  int // not needed if return once a failed workflow is encountered
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

	return scanWorkflowExecutionsHelper(ctx, service, params, logger)
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

	executions := make([]WorkflowExecution, 0, len(resp.Executions))
	for _, execution := range resp.Executions {
		if shouldReplay(params.SamplingRate) {
			executions = append(executions, WorkflowExecution{
				ID:    execution.Execution.GetWorkflowId(),
				RunID: execution.Execution.GetRunId(),
			})
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
		success, err := replayWorkflowExecutionHelper(ctx, replayer, service, logger, params.Domain, execution)
		if err != nil {
			scope.Counter(metrics.ReplayFailedCounter).Inc(1)
			progress.Result.Failed++
		} else if success {
			scope.Counter(metrics.ReplaySucceedCounter).Inc(1)
			progress.Result.Succeed++
		} else {
			scope.Counter(metrics.ReplaySkippedCounter).Inc(1)
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

	if isExpectedReplayErr(err) {
		taggedLogger.Info("Skipped replaying workflow", zap.Error(err))
		return false, nil
	}

	taggedLogger.Error("Replay workflow failed", zap.Error(err))
	return false, err
}

func isExpectedReplayErr(err error) bool {
	if err == errReplayHistoryTooShort {
		// less than 3 history events, potentially cron workflow
		return true
	}

	switch err.(type) {
	case *shared.EntityNotExistsError, *shared.InternalServiceError:
		// workflow not exist, or potentially corrupted workflow
		return true
	}

	return false
}
