// Copyright (c) 2021 Uber Technologies, Inc.
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

package client

import (
	"time"

	cadence "go.uber.org/cadence/v2"
)

type (
	// WorkflowStartOption allows passing optional parameters when starting a workflow.
	WorkflowStartOption interface {
		workflowStartOption()
	}

	// WorkflowSignalOption allows passing optional parameters when signaling a workflow.
	WorkflowSignalOption interface {
		workflowSignalOption()
	}

	// WorkflowQueryOption allows passing optional parameters when querying a workflow.
	WorkflowQueryOption interface {
		workflowQueryOption()
	}

	// WorkflowDescribeOption allows passing optional parameters when describing a workflow.
	WorkflowDescribeOption interface {
		workflowDescribeOption()
	}

	// WorkflowCancelOption allows passing optional parameters when canceling a workflow.
	WorkflowCancelOption interface {
		workflowCancelOption()
	}

	// WorkflowTerminateOption allows passing optional parameters when terminating a workflow.
	WorkflowTerminateOption interface {
		workflowTerminateOption()
	}

	// WorkflowResetOption allows passing optional parameters when resetting a workflow.
	WorkflowResetOption interface {
		workflowResetOption()
	}

	// WorkflowGetResultOption allows passing optional parameters when retrieving workflow result.
	WorkflowGetResultOption interface {
		workflowGetResultOption()
	}

	// WorkflowCountOption allows passing optional parameters when counting workflows.
	WorkflowCountOption interface {
		workflowCountOption()
	}

	// WorkflowListOption allows passing optional parameters when listing workflows.
	WorkflowListOption interface {
		workflowListOption()
	}

	// ActivityCompleteOption allows passing optional parameters when asynchronously completing an activity.
	ActivityCompleteOption interface {
		activityCompleteOption()
	}

	// ActivityHeartbeatOption allows passing optional parameters when heartbeating an activity.
	ActivityHeartbeatOption interface {
		activityHeartbeatOption()
	}

	// WorkflowResetPoint specifies the point in history for workflow reset.
	WorkflowResetPoint interface {
		workflowResetPoint()
	}
)

// WithWorkflowID sets the business identifier of the workflow execution.
// Default: generated UUID
func WithWorkflowID(workflowID string) WorkflowStartOption {
	panic("not implemented")
}

// WithDecisionTaskStartToCloseTimeout sets the timeout for processing decision task from the time the worker
// pulled this task. If a decision task is lost, it is retried after this timeout.
// The resolution is seconds.
// Default: 10 seconds
func WithDecisionTaskStartToCloseTimeout(timeout time.Duration) WorkflowStartOption {
	panic("not implemented")
}

// WithWorkflowIDReusePolicy defines whether server allow reuse of workflow ID.
// Can be useful for dedup logic if set to WorkflowIdReusePolicyRejectDuplicate.
// Default: WorkflowIDReusePolicyAllowDuplicateFailedOnly
func WithWorkflowIDReusePolicy(policy cadence.WorkflowIDReusePolicy) WorkflowStartOption {
	panic("not implemented")
}

// WithRetryPolicy sets the retry policy for the workflow.
// If provided, in case of workflow failure server will start new workflow execution if needed based on the retry policy.
func WithRetryPolicy(policy cadence.RetryPolicy) WorkflowStartOption {
	panic("not implemented")
}

// WithCronSchedule sets the cron schedule for workflow. If a cron schedule is specified, the workflow will run
// as a cron based on the schedule. The scheduling will be based on UTC time. Schedule for next run only happen
// after the current run is completed/failed/timeout. If a RetryPolicy is also supplied, and the workflow failed
// or timeout, the workflow will be retried based on the retry policy. While the workflow is retrying, it won't
// schedule its next run. If next schedule is due while workflow is running (or retrying), then it will skip that
// schedule. Cron workflow will not stop until it is terminated or cancelled (by returning cadence.CanceledError).
// The cron spec is as following:
// ┌───────────── minute (0 - 59)
// │ ┌───────────── hour (0 - 23)
// │ │ ┌───────────── day of the month (1 - 31)
// │ │ │ ┌───────────── month (1 - 12)
// │ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday)
// │ │ │ │ │
// │ │ │ │ │
// * * * * *
func WithCronSchedule(cron string) WorkflowStartOption {
	panic("not implemented")
}

// WithMemo sets non-indexed info that will be shown in list workflow.
func WithMemo(memo map[string]interface{}) WorkflowStartOption {
	panic("not implemented")
}

// WithSearchAttributes sets indexed info that can be used in query of List/Count workflow APIs (only
// supported when Cadence server is using ElasticSearch). The key and value type must be registered on Cadence server side.
// Use client.SearchAttributes().List() to get valid keys and corresponding value types.
func WithSearchAttributes(searchAttributes map[string]interface{}) WorkflowStartOption {
	panic("not implemented")
}

// WithDelayedStart sets the delay of the workflow start.
// The resolution is seconds.
// Default: 0 seconds
func WithDelayedStart(delay time.Duration) WorkflowStartOption {
	panic("not implemented")
}

// WithStart will start a workflow if he workflow is not running or not found, it starts the workflow and then sends the signal in transaction.
func WithStart(wfFunc interface{}, args []interface{}, taskList string, timeout time.Duration, opts ...WorkflowStartOption) WorkflowSignalOption {
	panic("not implemented")
}

// WithQueryRejectCondition sets the query behaviour based on workflow state.
func WithQueryRejectCondition(condition cadence.QueryRejectCondition) WorkflowQueryOption {
	panic("not implemented")
}

// WithQueryConsistencyLevel sets the consistency level on query.
// Default: QueryConsistencyLevelEventual
func WithQueryConsistencyLevel(level cadence.QueryConsistencyLevel) WorkflowQueryOption {
	panic("not implemented")
}

// WithTerminateReason provides a details why workflows is being terminated.
func WithTerminateDetails(details []byte) WorkflowTerminateOption {
	panic("not implemented")
}

// WithNoSignalReapply will not apply signals after the reset point.
func WithNoSignalReapply() WorkflowResetOption {
	panic("not implemented")
}

// ResetToLastCompletedDecision will reset the workfow to the last completed decision.
func ResetToLastCompletedDecision() WorkflowResetPoint {
	panic("not implemented")
}

// ResetToLastScheduledDecision will reset the workflow to the last scheduled decision.
func ResetToLastScheduledDecision() WorkflowResetPoint {
	panic("not implemented")
}

// ResetToLastContinueAsNew will reset the workflow to the last continue-as-new.
func ResetToLastContinueAsNew() WorkflowResetPoint {
	panic("not implemented")
}

// ResetToFirstCompletedDecision will reset the workflow to the first completed decision.
func ResetToFirstCompletedDecision() WorkflowResetPoint {
	panic("not implemented")
}

// ResetToFirstScheduledDecision will reset the workflow to the first scheduled decision.
func ResetToFirstScheduledDecision() WorkflowResetPoint {
	panic("not implemented")
}

// ResetToBadBinary will reset the workflow to the given bad binary checksum.
func ResetToBadBinary(checksum string) WorkflowResetPoint {
	panic("not implemented")
}

// ResetToEarliestDecisionCompletedAfter will reset the workflow to the earliest completed decition after the given timestamp.
func ResetToEarliestDecisionCompletedAfter(timestamp time.Time) WorkflowResetPoint {
	panic("not implemented")
}

// ResetToEventId will reset the workflow to exact given event ID.
func ResetToEventId(eventId int64) WorkflowResetPoint {
	panic("not implemented")
}
