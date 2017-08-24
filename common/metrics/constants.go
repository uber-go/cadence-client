// Copyright (c) 2017 Uber Technologies, Inc.
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

package metrics

// Workflow Creation metrics
const (
	WorkflowStartCounter             = "cadence-workflow-start"
	WorkflowCompletedCounter         = "cadence-workflow-completed"
	WorkflowCanceledCounter          = "cadence-workflow-canceled"
	WorkflowFailedCounter            = "cadence-workflow-failed"
	WorkflowContinueAsNewCounter     = "cadence-workflow-continue-as-new"
	WorkflowEndToEndLatency          = "cadence-workflow-endtoend-latency" // measure workflow execution from start to close
	WorkflowGetHistoryCounter        = "cadence-workflow-get-history-total"
	WorkflowGetHistoryFailedCounter  = "cadence-workflow-get-history-failed"
	WorkflowGetHistorySucceedCounter = "cadence-workflow-get-history-succeed"
	WorkflowGetHistoryLatency        = "cadence-workflow-get-history-latency"
	DecisionTimeoutCounter           = "cadence-decision-timeout"

	DecisionPollCounter            = "cadence-decision-poll-total"
	DecisionPollFailedCounter      = "cadence-decision-poll-failed"
	DecisionPollNoTaskCounter      = "cadence-decision-poll-no-task"
	DecisionPollSucceedCounter     = "cadence-decision-poll-succeed"
	DecisionPollLatency            = "cadence-decision-poll-latency" // measure succeed poll request latency
	DecisionExecutionFailedCounter = "cadence-decision-execution-failed"
	DecisionExecutionLatency       = "cadence-decision-execution-latency"
	DecisionResponseFailedCounter  = "cadence-decision-response-failed"
	DecisionResponseLatency        = "cadence-decision-response-latency"
	DecisionEndToEndLatency        = "cadence-decision-endtoend-latency" // measure from poll request start to response completed
	DecisionTaskPanicCounter       = "cadence-decision-task-panic"
	DecisionTaskCompletedCounter   = "cadence-decision-task-completed"

	ActivityPollCounter            = "cadence-activity-poll-total"
	ActivityPollFailedCounter      = "cadence-activity-poll-failed"
	ActivityPollNoTaskCounter      = "cadence-activity-poll-no-task"
	ActivityPollSucceedCounter     = "cadence-activity-poll-succeed"
	ActivityPollLatency            = "cadence-activity-poll-latency"
	ActivityExecutionFailedCounter = "cadence-activity-execution-failed"
	ActivityExecutionLatency       = "cadence-activity-execution-latency"
	ActivityResponseLatency        = "cadence-activity-response-latency"
	ActivityResponseFailedCounter  = "cadence-activity-response-failed"
	ActivityEndToEndLatency        = "cadence-activity-endtoend-latency"
	ActivityTaskPanicCounter       = "cadence-activity-task-panic"
	ActivityTaskCompletedCounter   = "cadence-activity-task-completed"
	ActivityTaskFailedCounter      = "cadence-activity-task-failed"
	ActivityTaskCanceledCounter    = "cadence-activity-task-canceled"

	UnhandledSignalsCounter = "cadence-unhandled-signals"

	WorkerStartCounter = "cadence-worker-start"
	PollerStartCounter = "cadence-poller-start"

	CadenceRequest        = "cadence-request"
	CadenceError          = "cadence-error"
	CadenceLatency        = "cadence-latency"
	CadenceInvalidRequest = "cadence-invalid-request"
)
