// Copyright (c) 2021 Uber Technologies Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package testdata

import apiv1 "go.uber.org/cadence/.gen/proto/api/v1"

var (
	ArchivalStatus                             = apiv1.ArchivalStatus_ARCHIVAL_STATUS_ENABLED
	CancelExternalWorkflowExecutionFailedCause = apiv1.CancelExternalWorkflowExecutionFailedCause_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_UNKNOWN_EXTERNAL_WORKFLOW_EXECUTION
	ChildWorkflowExecutionFailedCause          = apiv1.ChildWorkflowExecutionFailedCause_CHILD_WORKFLOW_EXECUTION_FAILED_CAUSE_WORKFLOW_ALREADY_RUNNING
	ContinueAsNewInitiator                     = apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_RETRY_POLICY
	DecisionTaskFailedCause                    = apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CANCEL_WORKFLOW_EXECUTION_ATTRIBUTES
	DecisionTaskTimedOutCause                  = apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_RESET
	DomainStatus                               = apiv1.DomainStatus_DOMAIN_STATUS_REGISTERED
	EncodingType                               = apiv1.EncodingType_ENCODING_TYPE_JSON
	HistoryEventFilterType                     = apiv1.EventFilterType_EVENT_FILTER_TYPE_CLOSE_EVENT
	IndexedValueType                           = apiv1.IndexedValueType_INDEXED_VALUE_TYPE_INT
	ParentClosePolicy                          = apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_TERMINATE
	PendingActivityState                       = apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_CANCEL_REQUESTED
	PendingDecisionState                       = apiv1.PendingDecisionState_PENDING_DECISION_STATE_STARTED
	QueryConsistencyLevel                      = apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_STRONG
	QueryRejectCondition                       = apiv1.QueryRejectCondition_QUERY_REJECT_CONDITION_NOT_COMPLETED_CLEANLY
	QueryResultType                            = apiv1.QueryResultType_QUERY_RESULT_TYPE_FAILED
	SignalExternalWorkflowExecutionFailedCause = apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_UNKNOWN_EXTERNAL_WORKFLOW_EXECUTION
	TaskListKind                               = apiv1.TaskListKind_TASK_LIST_KIND_STICKY
	TaskListType                               = apiv1.TaskListType_TASK_LIST_TYPE_ACTIVITY
	TimeoutType                                = apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_START
	WorkflowExecutionCloseStatus               = apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CONTINUED_AS_NEW
	WorkflowIDReusePolicy                      = apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING
)
