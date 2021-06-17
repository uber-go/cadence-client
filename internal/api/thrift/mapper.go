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

package thrift

import (
	"go.uber.org/cadence/v2/internal/common"
	"time"

	"github.com/gogo/protobuf/types"
	"go.uber.org/cadence/v2/.gen/go/shared"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
)

func ToHistoryEvents(t []*shared.HistoryEvent) []*apiv1.HistoryEvent {
	if t == nil {
		return nil
	}
	v := make([]*apiv1.HistoryEvent, len(t))
	for i := range t {
		v[i] = toHistoryEvent(t[i])
	}
	return v
}

func toHistoryEvent(e *shared.HistoryEvent) *apiv1.HistoryEvent {
	if e == nil {
		return nil
	}
	event := apiv1.HistoryEvent{
		EventId:   e.GetEventId(),
		EventTime: unixNanoToTime(e.Timestamp),
		Version:   e.GetVersion(),
		TaskId:    e.GetTaskId(),
	}
	switch *e.EventType {
	case shared.EventTypeWorkflowExecutionStarted:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionStartedEventAttributes{
			WorkflowExecutionStartedEventAttributes: toWorkflowExecutionStartedEventAttributes(e.WorkflowExecutionStartedEventAttributes),
		}
	case shared.EventTypeWorkflowExecutionCompleted:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes{
			WorkflowExecutionCompletedEventAttributes: toWorkflowExecutionCompletedEventAttributes(e.WorkflowExecutionCompletedEventAttributes),
		}
	case shared.EventTypeWorkflowExecutionFailed:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionFailedEventAttributes{
			WorkflowExecutionFailedEventAttributes: toWorkflowExecutionFailedEventAttributes(e.WorkflowExecutionFailedEventAttributes),
		}
	case shared.EventTypeWorkflowExecutionTimedOut:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionTimedOutEventAttributes{
			WorkflowExecutionTimedOutEventAttributes: toWorkflowExecutionTimedOutEventAttributes(e.WorkflowExecutionTimedOutEventAttributes),
		}
	case shared.EventTypeDecisionTaskScheduled:
		event.Attributes = &apiv1.HistoryEvent_DecisionTaskScheduledEventAttributes{
			DecisionTaskScheduledEventAttributes: toDecisionTaskScheduledEventAttributes(e.DecisionTaskScheduledEventAttributes),
		}
	case shared.EventTypeDecisionTaskStarted:
		event.Attributes = &apiv1.HistoryEvent_DecisionTaskStartedEventAttributes{
			DecisionTaskStartedEventAttributes: toDecisionTaskStartedEventAttributes(e.DecisionTaskStartedEventAttributes),
		}
	case shared.EventTypeDecisionTaskCompleted:
		event.Attributes = &apiv1.HistoryEvent_DecisionTaskCompletedEventAttributes{
			DecisionTaskCompletedEventAttributes: toDecisionTaskCompletedEventAttributes(e.DecisionTaskCompletedEventAttributes),
		}
	case shared.EventTypeDecisionTaskTimedOut:
		event.Attributes = &apiv1.HistoryEvent_DecisionTaskTimedOutEventAttributes{
			DecisionTaskTimedOutEventAttributes: toDecisionTaskTimedOutEventAttributes(e.DecisionTaskTimedOutEventAttributes),
		}
	case shared.EventTypeDecisionTaskFailed:
		event.Attributes = &apiv1.HistoryEvent_DecisionTaskFailedEventAttributes{
			DecisionTaskFailedEventAttributes: toDecisionTaskFailedEventAttributes(e.DecisionTaskFailedEventAttributes),
		}
	case shared.EventTypeActivityTaskScheduled:
		event.Attributes = &apiv1.HistoryEvent_ActivityTaskScheduledEventAttributes{
			ActivityTaskScheduledEventAttributes: toActivityTaskScheduledEventAttributes(e.ActivityTaskScheduledEventAttributes),
		}
	case shared.EventTypeActivityTaskStarted:
		event.Attributes = &apiv1.HistoryEvent_ActivityTaskStartedEventAttributes{
			ActivityTaskStartedEventAttributes: toActivityTaskStartedEventAttributes(e.ActivityTaskStartedEventAttributes),
		}
	case shared.EventTypeActivityTaskCompleted:
		event.Attributes = &apiv1.HistoryEvent_ActivityTaskCompletedEventAttributes{
			ActivityTaskCompletedEventAttributes: toActivityTaskCompletedEventAttributes(e.ActivityTaskCompletedEventAttributes),
		}
	case shared.EventTypeActivityTaskFailed:
		event.Attributes = &apiv1.HistoryEvent_ActivityTaskFailedEventAttributes{
			ActivityTaskFailedEventAttributes: toActivityTaskFailedEventAttributes(e.ActivityTaskFailedEventAttributes),
		}
	case shared.EventTypeActivityTaskTimedOut:
		event.Attributes = &apiv1.HistoryEvent_ActivityTaskTimedOutEventAttributes{
			ActivityTaskTimedOutEventAttributes: toActivityTaskTimedOutEventAttributes(e.ActivityTaskTimedOutEventAttributes),
		}
	case shared.EventTypeActivityTaskCancelRequested:
		event.Attributes = &apiv1.HistoryEvent_ActivityTaskCancelRequestedEventAttributes{
			ActivityTaskCancelRequestedEventAttributes: toActivityTaskCancelRequestedEventAttributes(e.ActivityTaskCancelRequestedEventAttributes),
		}
	case shared.EventTypeRequestCancelActivityTaskFailed:
		event.Attributes = &apiv1.HistoryEvent_RequestCancelActivityTaskFailedEventAttributes{
			RequestCancelActivityTaskFailedEventAttributes: toRequestCancelActivityTaskFailedEventAttributes(e.RequestCancelActivityTaskFailedEventAttributes),
		}
	case shared.EventTypeActivityTaskCanceled:
		event.Attributes = &apiv1.HistoryEvent_ActivityTaskCanceledEventAttributes{
			ActivityTaskCanceledEventAttributes: toActivityTaskCanceledEventAttributes(e.ActivityTaskCanceledEventAttributes),
		}
	case shared.EventTypeTimerStarted:
		event.Attributes = &apiv1.HistoryEvent_TimerStartedEventAttributes{
			TimerStartedEventAttributes: toTimerStartedEventAttributes(e.TimerStartedEventAttributes),
		}
	case shared.EventTypeTimerFired:
		event.Attributes = &apiv1.HistoryEvent_TimerFiredEventAttributes{
			TimerFiredEventAttributes: toTimerFiredEventAttributes(e.TimerFiredEventAttributes),
		}
	case shared.EventTypeTimerCanceled:
		event.Attributes = &apiv1.HistoryEvent_TimerCanceledEventAttributes{
			TimerCanceledEventAttributes: toTimerCanceledEventAttributes(e.TimerCanceledEventAttributes),
		}
	case shared.EventTypeCancelTimerFailed:
		event.Attributes = &apiv1.HistoryEvent_CancelTimerFailedEventAttributes{
			CancelTimerFailedEventAttributes: toCancelTimerFailedEventAttributes(e.CancelTimerFailedEventAttributes),
		}
	case shared.EventTypeWorkflowExecutionCancelRequested:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionCancelRequestedEventAttributes{
			WorkflowExecutionCancelRequestedEventAttributes: toWorkflowExecutionCancelRequestedEventAttributes(e.WorkflowExecutionCancelRequestedEventAttributes),
		}
	case shared.EventTypeWorkflowExecutionCanceled:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionCanceledEventAttributes{
			WorkflowExecutionCanceledEventAttributes: toWorkflowExecutionCanceledEventAttributes(e.WorkflowExecutionCanceledEventAttributes),
		}
	case shared.EventTypeRequestCancelExternalWorkflowExecutionInitiated:
		event.Attributes = &apiv1.HistoryEvent_RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{
			RequestCancelExternalWorkflowExecutionInitiatedEventAttributes: toRequestCancelExternalWorkflowExecutionInitiatedEventAttributes(e.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes),
		}
	case shared.EventTypeRequestCancelExternalWorkflowExecutionFailed:
		event.Attributes = &apiv1.HistoryEvent_RequestCancelExternalWorkflowExecutionFailedEventAttributes{
			RequestCancelExternalWorkflowExecutionFailedEventAttributes: toRequestCancelExternalWorkflowExecutionFailedEventAttributes(e.RequestCancelExternalWorkflowExecutionFailedEventAttributes),
		}
	case shared.EventTypeExternalWorkflowExecutionCancelRequested:
		event.Attributes = &apiv1.HistoryEvent_ExternalWorkflowExecutionCancelRequestedEventAttributes{
			ExternalWorkflowExecutionCancelRequestedEventAttributes: toExternalWorkflowExecutionCancelRequestedEventAttributes(e.ExternalWorkflowExecutionCancelRequestedEventAttributes),
		}
	case shared.EventTypeMarkerRecorded:
		event.Attributes = &apiv1.HistoryEvent_MarkerRecordedEventAttributes{
			MarkerRecordedEventAttributes: toMarkerRecordedEventAttributes(e.MarkerRecordedEventAttributes),
		}
	case shared.EventTypeWorkflowExecutionSignaled:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionSignaledEventAttributes{
			WorkflowExecutionSignaledEventAttributes: toWorkflowExecutionSignaledEventAttributes(e.WorkflowExecutionSignaledEventAttributes),
		}
	case shared.EventTypeWorkflowExecutionTerminated:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionTerminatedEventAttributes{
			WorkflowExecutionTerminatedEventAttributes: toWorkflowExecutionTerminatedEventAttributes(e.WorkflowExecutionTerminatedEventAttributes),
		}
	case shared.EventTypeWorkflowExecutionContinuedAsNew:
		event.Attributes = &apiv1.HistoryEvent_WorkflowExecutionContinuedAsNewEventAttributes{
			WorkflowExecutionContinuedAsNewEventAttributes: toWorkflowExecutionContinuedAsNewEventAttributes(e.WorkflowExecutionContinuedAsNewEventAttributes),
		}
	case shared.EventTypeStartChildWorkflowExecutionInitiated:
		event.Attributes = &apiv1.HistoryEvent_StartChildWorkflowExecutionInitiatedEventAttributes{
			StartChildWorkflowExecutionInitiatedEventAttributes: toStartChildWorkflowExecutionInitiatedEventAttributes(e.StartChildWorkflowExecutionInitiatedEventAttributes),
		}
	case shared.EventTypeStartChildWorkflowExecutionFailed:
		event.Attributes = &apiv1.HistoryEvent_StartChildWorkflowExecutionFailedEventAttributes{
			StartChildWorkflowExecutionFailedEventAttributes: toStartChildWorkflowExecutionFailedEventAttributes(e.StartChildWorkflowExecutionFailedEventAttributes),
		}
	case shared.EventTypeChildWorkflowExecutionStarted:
		event.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionStartedEventAttributes{
			ChildWorkflowExecutionStartedEventAttributes: toChildWorkflowExecutionStartedEventAttributes(e.ChildWorkflowExecutionStartedEventAttributes),
		}
	case shared.EventTypeChildWorkflowExecutionCompleted:
		event.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionCompletedEventAttributes{
			ChildWorkflowExecutionCompletedEventAttributes: toChildWorkflowExecutionCompletedEventAttributes(e.ChildWorkflowExecutionCompletedEventAttributes),
		}
	case shared.EventTypeChildWorkflowExecutionFailed:
		event.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionFailedEventAttributes{
			ChildWorkflowExecutionFailedEventAttributes: toChildWorkflowExecutionFailedEventAttributes(e.ChildWorkflowExecutionFailedEventAttributes),
		}
	case shared.EventTypeChildWorkflowExecutionCanceled:
		event.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionCanceledEventAttributes{
			ChildWorkflowExecutionCanceledEventAttributes: toChildWorkflowExecutionCanceledEventAttributes(e.ChildWorkflowExecutionCanceledEventAttributes),
		}
	case shared.EventTypeChildWorkflowExecutionTimedOut:
		event.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionTimedOutEventAttributes{
			ChildWorkflowExecutionTimedOutEventAttributes: toChildWorkflowExecutionTimedOutEventAttributes(e.ChildWorkflowExecutionTimedOutEventAttributes),
		}
	case shared.EventTypeChildWorkflowExecutionTerminated:
		event.Attributes = &apiv1.HistoryEvent_ChildWorkflowExecutionTerminatedEventAttributes{
			ChildWorkflowExecutionTerminatedEventAttributes: toChildWorkflowExecutionTerminatedEventAttributes(e.ChildWorkflowExecutionTerminatedEventAttributes),
		}
	case shared.EventTypeSignalExternalWorkflowExecutionInitiated:
		event.Attributes = &apiv1.HistoryEvent_SignalExternalWorkflowExecutionInitiatedEventAttributes{
			SignalExternalWorkflowExecutionInitiatedEventAttributes: toSignalExternalWorkflowExecutionInitiatedEventAttributes(e.SignalExternalWorkflowExecutionInitiatedEventAttributes),
		}
	case shared.EventTypeSignalExternalWorkflowExecutionFailed:
		event.Attributes = &apiv1.HistoryEvent_SignalExternalWorkflowExecutionFailedEventAttributes{
			SignalExternalWorkflowExecutionFailedEventAttributes: toSignalExternalWorkflowExecutionFailedEventAttributes(e.SignalExternalWorkflowExecutionFailedEventAttributes),
		}
	case shared.EventTypeExternalWorkflowExecutionSignaled:
		event.Attributes = &apiv1.HistoryEvent_ExternalWorkflowExecutionSignaledEventAttributes{
			ExternalWorkflowExecutionSignaledEventAttributes: toExternalWorkflowExecutionSignaledEventAttributes(e.ExternalWorkflowExecutionSignaledEventAttributes),
		}
	case shared.EventTypeUpsertWorkflowSearchAttributes:
		event.Attributes = &apiv1.HistoryEvent_UpsertWorkflowSearchAttributesEventAttributes{
			UpsertWorkflowSearchAttributesEventAttributes: toUpsertWorkflowSearchAttributesEventAttributes(e.UpsertWorkflowSearchAttributesEventAttributes),
		}
	}
	return &event
}

func toActivityTaskCancelRequestedEventAttributes(t *shared.ActivityTaskCancelRequestedEventAttributes) *apiv1.ActivityTaskCancelRequestedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityTaskCancelRequestedEventAttributes{
		ActivityId:                   t.GetActivityId(),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
	}
}

func toActivityTaskCanceledEventAttributes(t *shared.ActivityTaskCanceledEventAttributes) *apiv1.ActivityTaskCanceledEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityTaskCanceledEventAttributes{
		Details:                      toPayload(t.Details),
		LatestCancelRequestedEventId: t.GetLatestCancelRequestedEventId(),
		ScheduledEventId:             t.GetScheduledEventId(),
		StartedEventId:               t.GetStartedEventId(),
		Identity:                     t.GetIdentity(),
	}
}

func toActivityTaskCompletedEventAttributes(t *shared.ActivityTaskCompletedEventAttributes) *apiv1.ActivityTaskCompletedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityTaskCompletedEventAttributes{
		Result:           toPayload(t.Result),
		ScheduledEventId: t.GetScheduledEventId(),
		StartedEventId:   t.GetStartedEventId(),
		Identity:         t.GetIdentity(),
	}
}

func toActivityTaskFailedEventAttributes(t *shared.ActivityTaskFailedEventAttributes) *apiv1.ActivityTaskFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityTaskFailedEventAttributes{
		Failure:          toFailure(t.Reason, t.Details),
		ScheduledEventId: t.GetScheduledEventId(),
		StartedEventId:   t.GetStartedEventId(),
		Identity:         t.GetIdentity(),
	}
}

func toActivityTaskScheduledEventAttributes(t *shared.ActivityTaskScheduledEventAttributes) *apiv1.ActivityTaskScheduledEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityTaskScheduledEventAttributes{
		ActivityId:                   t.GetActivityId(),
		ActivityType:                 toActivityType(t.ActivityType),
		Domain:                       t.GetDomain(),
		TaskList:                     toTaskList(t.TaskList),
		Input:                        toPayload(t.Input),
		ScheduleToCloseTimeout:       secondsToDuration(t.ScheduleToCloseTimeoutSeconds),
		ScheduleToStartTimeout:       secondsToDuration(t.ScheduleToStartTimeoutSeconds),
		StartToCloseTimeout:          secondsToDuration(t.StartToCloseTimeoutSeconds),
		HeartbeatTimeout:             secondsToDuration(t.HeartbeatTimeoutSeconds),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		RetryPolicy:                  toRetryPolicy(t.RetryPolicy),
		Header:                       toHeader(t.Header),
	}
}

func toActivityTaskStartedEventAttributes(t *shared.ActivityTaskStartedEventAttributes) *apiv1.ActivityTaskStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityTaskStartedEventAttributes{
		ScheduledEventId: t.GetScheduledEventId(),
		Identity:         t.GetIdentity(),
		RequestId:        t.GetRequestId(),
		Attempt:          t.GetAttempt(),
		LastFailure:      toFailure(t.LastFailureReason, t.LastFailureDetails),
	}
}

func toActivityTaskTimedOutEventAttributes(t *shared.ActivityTaskTimedOutEventAttributes) *apiv1.ActivityTaskTimedOutEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityTaskTimedOutEventAttributes{
		Details:          toPayload(t.Details),
		ScheduledEventId: t.GetScheduledEventId(),
		StartedEventId:   t.GetStartedEventId(),
		TimeoutType:      toTimeoutType(t.TimeoutType),
		LastFailure:      toFailure(t.LastFailureReason, t.LastFailureDetails),
	}
}

func toCancelTimerFailedEventAttributes(t *shared.CancelTimerFailedEventAttributes) *apiv1.CancelTimerFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.CancelTimerFailedEventAttributes{
		TimerId:                      t.GetTimerId(),
		Cause:                        t.GetCause(),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		Identity:                     t.GetIdentity(),
	}
}

func toChildWorkflowExecutionCanceledEventAttributes(t *shared.ChildWorkflowExecutionCanceledEventAttributes) *apiv1.ChildWorkflowExecutionCanceledEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ChildWorkflowExecutionCanceledEventAttributes{
		Domain:            t.GetDomain(),
		WorkflowExecution: toWorkflowExecution(t.WorkflowExecution),
		WorkflowType:      toWorkflowType(t.WorkflowType),
		InitiatedEventId:  t.GetInitiatedEventId(),
		StartedEventId:    t.GetStartedEventId(),
		Details:           toPayload(t.Details),
	}
}

func toChildWorkflowExecutionCompletedEventAttributes(t *shared.ChildWorkflowExecutionCompletedEventAttributes) *apiv1.ChildWorkflowExecutionCompletedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ChildWorkflowExecutionCompletedEventAttributes{
		Domain:            t.GetDomain(),
		WorkflowExecution: toWorkflowExecution(t.WorkflowExecution),
		WorkflowType:      toWorkflowType(t.WorkflowType),
		InitiatedEventId:  t.GetInitiatedEventId(),
		StartedEventId:    t.GetStartedEventId(),
		Result:            toPayload(t.Result),
	}
}

func toWorkflowExecutionStartedEventAttributes(t *shared.WorkflowExecutionStartedEventAttributes) *apiv1.WorkflowExecutionStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionStartedEventAttributes{
		WorkflowType:                 toWorkflowType(t.WorkflowType),
		ParentExecutionInfo:          toParentExecutionInfoFields(t.ParentWorkflowDomain, t.ParentWorkflowExecution, t.ParentInitiatedEventId),
		TaskList:                     toTaskList(t.TaskList),
		Input:                        toPayload(t.Input),
		ExecutionStartToCloseTimeout: secondsToDuration(t.ExecutionStartToCloseTimeoutSeconds),
		TaskStartToCloseTimeout:      secondsToDuration(t.TaskStartToCloseTimeoutSeconds),
		ContinuedExecutionRunId:      t.GetContinuedExecutionRunId(),
		Initiator:                    toContinueAsNewInitiator(t.Initiator),
		ContinuedFailure:             toFailure(t.ContinuedFailureReason, t.ContinuedFailureDetails),
		LastCompletionResult:         toPayload(t.LastCompletionResult),
		OriginalExecutionRunId:       t.GetOriginalExecutionRunId(),
		Identity:                     t.GetIdentity(),
		FirstExecutionRunId:          t.GetFirstExecutionRunId(),
		RetryPolicy:                  toRetryPolicy(t.RetryPolicy),
		Attempt:                      t.GetAttempt(),
		ExpirationTime:               unixNanoToTime(t.ExpirationTimestamp),
		CronSchedule:                 t.GetCronSchedule(),
		FirstDecisionTaskBackoff:     secondsToDuration(t.FirstDecisionTaskBackoffSeconds),
		Memo:                         toMemo(t.Memo),
		SearchAttributes:             toSearchAttributes(t.SearchAttributes),
		PrevAutoResetPoints:          toResetPoints(t.PrevAutoResetPoints),
		Header:                       toHeader(t.Header),
	}
}

func toWorkflowExecutionTerminatedEventAttributes(t *shared.WorkflowExecutionTerminatedEventAttributes) *apiv1.WorkflowExecutionTerminatedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionTerminatedEventAttributes{
		Reason:   t.GetReason(),
		Details:  toPayload(t.Details),
		Identity: t.GetIdentity(),
	}
}

func toWorkflowExecutionTimedOutEventAttributes(t *shared.WorkflowExecutionTimedOutEventAttributes) *apiv1.WorkflowExecutionTimedOutEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionTimedOutEventAttributes{
		TimeoutType: toTimeoutType(t.TimeoutType),
	}
}

func toWorkflowExecutionCompletedEventAttributes(t *shared.WorkflowExecutionCompletedEventAttributes) *apiv1.WorkflowExecutionCompletedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionCompletedEventAttributes{
		Result:                       toPayload(t.Result),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
	}
}

func toWorkflowExecutionFailedEventAttributes(t *shared.WorkflowExecutionFailedEventAttributes) *apiv1.WorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionFailedEventAttributes{
		Failure:                      toFailure(t.Reason, t.Details),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
	}
}

func toDecisionTaskFailedEventAttributes(t *shared.DecisionTaskFailedEventAttributes) *apiv1.DecisionTaskFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.DecisionTaskFailedEventAttributes{
		ScheduledEventId: t.GetScheduledEventId(),
		StartedEventId:   t.GetStartedEventId(),
		Cause:            toDecisionTaskFailedCause(t.Cause),
		Failure:          toFailure(t.Reason, t.Details),
		Identity:         t.GetIdentity(),
		BaseRunId:        t.GetBaseRunId(),
		NewRunId:         t.GetNewRunId(),
		ForkEventVersion: t.GetForkEventVersion(),
		BinaryChecksum:   t.GetBinaryChecksum(),
	}
}

func toDecisionTaskScheduledEventAttributes(t *shared.DecisionTaskScheduledEventAttributes) *apiv1.DecisionTaskScheduledEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.DecisionTaskScheduledEventAttributes{
		TaskList:            toTaskList(t.TaskList),
		StartToCloseTimeout: secondsToDuration(t.StartToCloseTimeoutSeconds),
		Attempt:             int32(t.GetAttempt()),
	}
}

func toDecisionTaskStartedEventAttributes(t *shared.DecisionTaskStartedEventAttributes) *apiv1.DecisionTaskStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.DecisionTaskStartedEventAttributes{
		ScheduledEventId: t.GetScheduledEventId(),
		Identity:         t.GetIdentity(),
		RequestId:        t.GetRequestId(),
	}
}

func toDecisionTaskTimedOutEventAttributes(t *shared.DecisionTaskTimedOutEventAttributes) *apiv1.DecisionTaskTimedOutEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.DecisionTaskTimedOutEventAttributes{
		ScheduledEventId: t.GetScheduledEventId(),
		StartedEventId:   t.GetStartedEventId(),
		TimeoutType:      toTimeoutType(t.TimeoutType),
		BaseRunId:        t.GetBaseRunId(),
		NewRunId:         t.GetNewRunId(),
		ForkEventVersion: t.GetForkEventVersion(),
		Reason:           t.GetReason(),
		Cause:            toDecisionTaskTimedOutCause(t.Cause),
	}
}

func toRequestCancelActivityTaskFailedEventAttributes(t *shared.RequestCancelActivityTaskFailedEventAttributes) *apiv1.RequestCancelActivityTaskFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.RequestCancelActivityTaskFailedEventAttributes{
		ActivityId:                   t.GetActivityId(),
		Cause:                        t.GetCause(),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
	}
}

func toTimerCanceledEventAttributes(t *shared.TimerCanceledEventAttributes) *apiv1.TimerCanceledEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.TimerCanceledEventAttributes{
		TimerId:                      t.GetTimerId(),
		StartedEventId:               t.GetStartedEventId(),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		Identity:                     t.GetIdentity(),
	}
}

func toTimerFiredEventAttributes(t *shared.TimerFiredEventAttributes) *apiv1.TimerFiredEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.TimerFiredEventAttributes{
		TimerId:        t.GetTimerId(),
		StartedEventId: t.GetStartedEventId(),
	}
}

func toTimerStartedEventAttributes(t *shared.TimerStartedEventAttributes) *apiv1.TimerStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.TimerStartedEventAttributes{
		TimerId:                      t.GetTimerId(),
		StartToFireTimeout:           secondsToDuration(int64To32(t.StartToFireTimeoutSeconds)),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
	}
}

func toWorkflowExecutionCancelRequestedEventAttributes(t *shared.WorkflowExecutionCancelRequestedEventAttributes) *apiv1.WorkflowExecutionCancelRequestedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionCancelRequestedEventAttributes{
		Cause:                 t.GetCause(),
		ExternalExecutionInfo: toExternalExecutionInfoFields(t.ExternalWorkflowExecution, t.ExternalInitiatedEventId),
		Identity:              t.GetIdentity(),
	}
}

func toWorkflowExecutionCanceledEventAttributes(t *shared.WorkflowExecutionCanceledEventAttributes) *apiv1.WorkflowExecutionCanceledEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionCanceledEventAttributes{
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		Details:                      toPayload(t.Details),
	}
}

func toRequestCancelExternalWorkflowExecutionFailedEventAttributes(t *shared.RequestCancelExternalWorkflowExecutionFailedEventAttributes) *apiv1.RequestCancelExternalWorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.RequestCancelExternalWorkflowExecutionFailedEventAttributes{
		Cause:                        toCancelExternalWorkflowExecutionFailedCause(t.Cause),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		Domain:                       t.GetDomain(),
		WorkflowExecution:            toWorkflowExecution(t.WorkflowExecution),
		InitiatedEventId:             t.GetInitiatedEventId(),
		Control:                      t.Control,
	}
}

func toRequestCancelExternalWorkflowExecutionInitiatedEventAttributes(t *shared.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes) *apiv1.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		Domain:                       t.GetDomain(),
		WorkflowExecution:            toWorkflowExecution(t.WorkflowExecution),
		Control:                      t.Control,
		ChildWorkflowOnly:            t.GetChildWorkflowOnly(),
	}
}


func toDecisionTaskCompletedEventAttributes(t *shared.DecisionTaskCompletedEventAttributes) *apiv1.DecisionTaskCompletedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.DecisionTaskCompletedEventAttributes{
		ScheduledEventId: t.GetScheduledEventId(),
		StartedEventId:   t.GetStartedEventId(),
		Identity:         t.GetIdentity(),
		BinaryChecksum:   t.GetBinaryChecksum(),
		ExecutionContext: t.ExecutionContext,
	}
}

func toExternalWorkflowExecutionCancelRequestedEventAttributes(t *shared.ExternalWorkflowExecutionCancelRequestedEventAttributes) *apiv1.ExternalWorkflowExecutionCancelRequestedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ExternalWorkflowExecutionCancelRequestedEventAttributes{
		InitiatedEventId:  t.GetInitiatedEventId(),
		Domain:            t.GetDomain(),
		WorkflowExecution: toWorkflowExecution(t.WorkflowExecution),
	}
}

func toMarkerRecordedEventAttributes(t *shared.MarkerRecordedEventAttributes) *apiv1.MarkerRecordedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.MarkerRecordedEventAttributes{
		MarkerName:                   t.GetMarkerName(),
		Details:                      toPayload(t.Details),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		Header:                       toHeader(t.Header),
	}
}

func toWorkflowExecutionSignaledEventAttributes(t *shared.WorkflowExecutionSignaledEventAttributes) *apiv1.WorkflowExecutionSignaledEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionSignaledEventAttributes{
		SignalName: t.GetSignalName(),
		Input:      toPayload(t.Input),
		Identity:   t.GetIdentity(),
	}
}

func toWorkflowExecutionContinuedAsNewEventAttributes(t *shared.WorkflowExecutionContinuedAsNewEventAttributes) *apiv1.WorkflowExecutionContinuedAsNewEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionContinuedAsNewEventAttributes{
		NewExecutionRunId:            t.GetNewExecutionRunId(),
		WorkflowType:                 toWorkflowType(t.WorkflowType),
		TaskList:                     toTaskList(t.TaskList),
		Input:                        toPayload(t.Input),
		ExecutionStartToCloseTimeout: secondsToDuration(t.ExecutionStartToCloseTimeoutSeconds),
		TaskStartToCloseTimeout:      secondsToDuration(t.TaskStartToCloseTimeoutSeconds),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		BackoffStartInterval:         secondsToDuration(t.BackoffStartIntervalInSeconds),
		Initiator:                    toContinueAsNewInitiator(t.Initiator),
		Failure:                      toFailure(t.FailureReason, t.FailureDetails),
		LastCompletionResult:         toPayload(t.LastCompletionResult),
		Header:                       toHeader(t.Header),
		Memo:                         toMemo(t.Memo),
		SearchAttributes:             toSearchAttributes(t.SearchAttributes),
	}
}

func toStartChildWorkflowExecutionInitiatedEventAttributes(t *shared.StartChildWorkflowExecutionInitiatedEventAttributes) *apiv1.StartChildWorkflowExecutionInitiatedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.StartChildWorkflowExecutionInitiatedEventAttributes{
		Domain:                       t.GetDomain(),
		WorkflowId:                   t.GetWorkflowId(),
		WorkflowType:                 toWorkflowType(t.WorkflowType),
		TaskList:                     toTaskList(t.TaskList),
		Input:                        toPayload(t.Input),
		ExecutionStartToCloseTimeout: secondsToDuration(t.ExecutionStartToCloseTimeoutSeconds),
		TaskStartToCloseTimeout:      secondsToDuration(t.TaskStartToCloseTimeoutSeconds),
		ParentClosePolicy:            toParentClosePolicy(t.ParentClosePolicy),
		Control:                      t.Control,
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		WorkflowIdReusePolicy:        toWorkflowIDReusePolicy(t.WorkflowIdReusePolicy),
		RetryPolicy:                  toRetryPolicy(t.RetryPolicy),
		CronSchedule:                 t.GetCronSchedule(),
		Header:                       toHeader(t.Header),
		Memo:                         toMemo(t.Memo),
		SearchAttributes:             toSearchAttributes(t.SearchAttributes),
		DelayStart:                   secondsToDuration(t.DelayStartSeconds),
	}
}

func toStartChildWorkflowExecutionFailedEventAttributes(t *shared.StartChildWorkflowExecutionFailedEventAttributes) *apiv1.StartChildWorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.StartChildWorkflowExecutionFailedEventAttributes{
		Domain:                       t.GetDomain(),
		WorkflowId:                   t.GetWorkflowId(),
		WorkflowType:                 toWorkflowType(t.WorkflowType),
		Cause:                        toChildWorkflowExecutionFailedCause(t.Cause),
		Control:                      t.Control,
		InitiatedEventId:             t.GetInitiatedEventId(),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
	}
}

func toChildWorkflowExecutionStartedEventAttributes(t *shared.ChildWorkflowExecutionStartedEventAttributes) *apiv1.ChildWorkflowExecutionStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ChildWorkflowExecutionStartedEventAttributes{
		Domain:            t.GetDomain(),
		WorkflowExecution: toWorkflowExecution(t.WorkflowExecution),
		WorkflowType:      toWorkflowType(t.WorkflowType),
		InitiatedEventId:  t.GetInitiatedEventId(),
		Header:            toHeader(t.Header),
	}
}

func toChildWorkflowExecutionFailedEventAttributes(t *shared.ChildWorkflowExecutionFailedEventAttributes) *apiv1.ChildWorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ChildWorkflowExecutionFailedEventAttributes{
		Domain:            t.GetDomain(),
		WorkflowExecution: toWorkflowExecution(t.WorkflowExecution),
		WorkflowType:      toWorkflowType(t.WorkflowType),
		InitiatedEventId:  t.GetInitiatedEventId(),
		StartedEventId:    t.GetStartedEventId(),
		Failure:           toFailure(t.Reason, t.Details),
	}
}

func toChildWorkflowExecutionTimedOutEventAttributes(t *shared.ChildWorkflowExecutionTimedOutEventAttributes) *apiv1.ChildWorkflowExecutionTimedOutEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ChildWorkflowExecutionTimedOutEventAttributes{
		Domain:            t.GetDomain(),
		WorkflowExecution: toWorkflowExecution(t.WorkflowExecution),
		WorkflowType:      toWorkflowType(t.WorkflowType),
		InitiatedEventId:  t.GetInitiatedEventId(),
		StartedEventId:    t.GetStartedEventId(),
		TimeoutType:       toTimeoutType(t.TimeoutType),
	}
}


func toChildWorkflowExecutionTerminatedEventAttributes(t *shared.ChildWorkflowExecutionTerminatedEventAttributes) *apiv1.ChildWorkflowExecutionTerminatedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ChildWorkflowExecutionTerminatedEventAttributes{
		Domain:            t.GetDomain(),
		WorkflowExecution: toWorkflowExecution(t.WorkflowExecution),
		WorkflowType:      toWorkflowType(t.WorkflowType),
		InitiatedEventId:  t.GetInitiatedEventId(),
		StartedEventId:    t.GetStartedEventId(),
	}
}

func toSignalExternalWorkflowExecutionInitiatedEventAttributes(t *shared.SignalExternalWorkflowExecutionInitiatedEventAttributes) *apiv1.SignalExternalWorkflowExecutionInitiatedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.SignalExternalWorkflowExecutionInitiatedEventAttributes{
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		Domain:                       t.GetDomain(),
		WorkflowExecution:            toWorkflowExecution(t.WorkflowExecution),
		SignalName:                   t.GetSignalName(),
		Input:                        toPayload(t.Input),
		Control:                      t.Control,
		ChildWorkflowOnly:            t.GetChildWorkflowOnly(),
	}
}

func toSignalExternalWorkflowExecutionFailedEventAttributes(t *shared.SignalExternalWorkflowExecutionFailedEventAttributes) *apiv1.SignalExternalWorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.SignalExternalWorkflowExecutionFailedEventAttributes{
		Cause:                        toSignalExternalWorkflowExecutionFailedCause(t.Cause),
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		Domain:                       t.GetDomain(),
		WorkflowExecution:            toWorkflowExecution(t.WorkflowExecution),
		InitiatedEventId:             t.GetInitiatedEventId(),
		Control:                      t.Control,
	}
}

func toExternalWorkflowExecutionSignaledEventAttributes(t *shared.ExternalWorkflowExecutionSignaledEventAttributes) *apiv1.ExternalWorkflowExecutionSignaledEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.ExternalWorkflowExecutionSignaledEventAttributes{
		InitiatedEventId:  t.GetInitiatedEventId(),
		Domain:            t.GetDomain(),
		WorkflowExecution: toWorkflowExecution(t.WorkflowExecution),
		Control:           t.Control,
	}
}

func toUpsertWorkflowSearchAttributesEventAttributes(t *shared.UpsertWorkflowSearchAttributesEventAttributes) *apiv1.UpsertWorkflowSearchAttributesEventAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.UpsertWorkflowSearchAttributesEventAttributes{
		DecisionTaskCompletedEventId: t.GetDecisionTaskCompletedEventId(),
		SearchAttributes:             toSearchAttributes(t.SearchAttributes),
	}
}

func toPayload(data []byte) *apiv1.Payload {
	if data == nil {
		return nil
	}
	return &apiv1.Payload{
		Data: data,
	}
}

func toFailure(reason *string, details []byte) *apiv1.Failure {
	if reason == nil {
		return nil
	}
	return &apiv1.Failure{
		Reason:  *reason,
		Details: details,
	}
}

func toWorkflowExecution(t *shared.WorkflowExecution) *apiv1.WorkflowExecution {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecution{
		WorkflowId: t.GetWorkflowId(),
		RunId:      t.GetRunId(),
	}
}

func toExternalExecutionInfoFields(we *shared.WorkflowExecution, initiatedID *int64) *apiv1.ExternalExecutionInfo {
	if we == nil && initiatedID == nil {
		return nil
	}
	if we == nil || initiatedID == nil {
		panic("either all or none external execution info fields must be set")
	}
	return &apiv1.ExternalExecutionInfo{
		WorkflowExecution: toWorkflowExecution(we),
		InitiatedId:       *initiatedID,
	}
}

func toActivityType(t *shared.ActivityType) *apiv1.ActivityType {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityType{
		Name: t.GetName(),
	}
}

func toWorkflowType(t *shared.WorkflowType) *apiv1.WorkflowType {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowType{
		Name: t.GetName(),
	}
}

func toTaskList(t *shared.TaskList) *apiv1.TaskList {
	if t == nil {
		return nil
	}
	return &apiv1.TaskList{
		Name: t.GetName(),
		Kind: toTaskListKind(t.Kind),
	}
}

func toTaskListKind(t *shared.TaskListKind) apiv1.TaskListKind {
	if t == nil {
		return apiv1.TaskListKind_TASK_LIST_KIND_INVALID
	}
	switch *t {
	case shared.TaskListKindNormal:
		return apiv1.TaskListKind_TASK_LIST_KIND_NORMAL
	case shared.TaskListKindSticky:
		return apiv1.TaskListKind_TASK_LIST_KIND_STICKY
	}
	panic("unexpected enum value")
}

func toHeader(t *shared.Header) *apiv1.Header {
	if t == nil {
		return nil
	}
	return &apiv1.Header{
		Fields: toPayloadMap(t.Fields),
	}
}

func toMemo(t *shared.Memo) *apiv1.Memo {
	if t == nil {
		return nil
	}
	return &apiv1.Memo{
		Fields: toPayloadMap(t.Fields),
	}
}

func toSearchAttributes(t *shared.SearchAttributes) *apiv1.SearchAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.SearchAttributes{
		IndexedFields: toPayloadMap(t.IndexedFields),
	}
}

func toPayloadMap(t map[string][]byte) map[string]*apiv1.Payload {
	if t == nil {
		return nil
	}
	v := make(map[string]*apiv1.Payload, len(t))
	for key := range t {
		v[key] = toPayload(t[key])
	}
	return v
}

func toRetryPolicy(t *shared.RetryPolicy) *apiv1.RetryPolicy {
	if t == nil {
		return nil
	}
	return &apiv1.RetryPolicy{
		InitialInterval:          secondsToDuration(t.InitialIntervalInSeconds),
		BackoffCoefficient:       t.GetBackoffCoefficient(),
		MaximumInterval:          secondsToDuration(t.MaximumIntervalInSeconds),
		MaximumAttempts:          t.GetMaximumAttempts(),
		NonRetryableErrorReasons: t.NonRetriableErrorReasons,
		ExpirationInterval:       secondsToDuration(t.ExpirationIntervalInSeconds),
	}
}

func toResetPoints(t *shared.ResetPoints) *apiv1.ResetPoints {
	if t == nil {
		return nil
	}
	return &apiv1.ResetPoints{
		Points: toResetPointInfoArray(t.Points),
	}
}

func toResetPointInfoArray(t []*shared.ResetPointInfo) []*apiv1.ResetPointInfo {
	if t == nil {
		return nil
	}
	v := make([]*apiv1.ResetPointInfo, len(t))
	for i := range t {
		v[i] = toResetPointInfo(t[i])
	}
	return v
}

func toResetPointInfo(t *shared.ResetPointInfo) *apiv1.ResetPointInfo {
	if t == nil {
		return nil
	}
	return &apiv1.ResetPointInfo{
		BinaryChecksum:           t.GetBinaryChecksum(),
		RunId:                    t.GetRunId(),
		FirstDecisionCompletedId: t.GetFirstDecisionCompletedId(),
		CreatedTime:              unixNanoToTime(t.CreatedTimeNano),
		ExpiringTime:             unixNanoToTime(t.ExpiringTimeNano),
		Resettable:               t.GetResettable(),
	}
}

func toParentExecutionInfoFields(domainName *string, we *shared.WorkflowExecution, initiatedID *int64) *apiv1.ParentExecutionInfo {
	if domainName == nil && we == nil && initiatedID == nil {
		return nil
	}
	if domainName == nil || we == nil || initiatedID == nil {
		panic("either all or none parent execution info must be set")
	}

	return &apiv1.ParentExecutionInfo{
		DomainName:        *domainName,
		WorkflowExecution: toWorkflowExecution(we),
		InitiatedId:       *initiatedID,
	}
}

func secondsToDuration(d *int32) *types.Duration {
	if d == nil {
		return nil
	}
	return types.DurationProto(time.Duration(*d) * time.Second)
}

func unixNanoToTime(t *int64) *types.Timestamp {
	if t == nil {
		return nil
	}
	time, err := types.TimestampProto(time.Unix(0, *t))
	if err != nil {
		panic(err)
	}
	return time
}

func int64To32(v *int64) *int32 {
	if v == nil {
		return nil
	}
	return common.Int32Ptr(int32(*v))
}

func toTimeoutType(t *shared.TimeoutType) apiv1.TimeoutType {
	if t == nil {
		return apiv1.TimeoutType_TIMEOUT_TYPE_INVALID
	}
	switch *t {
	case shared.TimeoutTypeStartToClose:
		return apiv1.TimeoutType_TIMEOUT_TYPE_START_TO_CLOSE
	case shared.TimeoutTypeScheduleToStart:
		return apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_START
	case shared.TimeoutTypeScheduleToClose:
		return apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_CLOSE
	case shared.TimeoutTypeHeartbeat:
		return apiv1.TimeoutType_TIMEOUT_TYPE_HEARTBEAT
	}
	panic("unexpected enum value")
}

func toContinueAsNewInitiator(t *shared.ContinueAsNewInitiator) apiv1.ContinueAsNewInitiator {
	if t == nil {
		return apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_INVALID
	}
	switch *t {
	case shared.ContinueAsNewInitiatorDecider:
		return apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_DECIDER
	case shared.ContinueAsNewInitiatorRetryPolicy:
		return apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_RETRY_POLICY
	case shared.ContinueAsNewInitiatorCronSchedule:
		return apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_CRON_SCHEDULE
	}
	panic("unexpected enum value")
}

func toDecisionTaskFailedCause(t *shared.DecisionTaskFailedCause) apiv1.DecisionTaskFailedCause {
	if t == nil {
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_INVALID
	}
	switch *t {
	case shared.DecisionTaskFailedCauseUnhandledDecision:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_UNHANDLED_DECISION
	case shared.DecisionTaskFailedCauseBadScheduleActivityAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SCHEDULE_ACTIVITY_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadRequestCancelActivityAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_REQUEST_CANCEL_ACTIVITY_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadStartTimerAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_START_TIMER_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadCancelTimerAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CANCEL_TIMER_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadRecordMarkerAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_RECORD_MARKER_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadCompleteWorkflowExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_COMPLETE_WORKFLOW_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadFailWorkflowExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_FAIL_WORKFLOW_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadCancelWorkflowExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CANCEL_WORKFLOW_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadRequestCancelExternalWorkflowExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_REQUEST_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadContinueAsNewAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CONTINUE_AS_NEW_ATTRIBUTES
	case shared.DecisionTaskFailedCauseStartTimerDuplicateID:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_START_TIMER_DUPLICATE_ID
	case shared.DecisionTaskFailedCauseResetStickyTasklist:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_RESET_STICKY_TASK_LIST
	case shared.DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_WORKFLOW_WORKER_UNHANDLED_FAILURE
	case shared.DecisionTaskFailedCauseBadSignalWorkflowExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SIGNAL_WORKFLOW_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadStartChildExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_START_CHILD_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseForceCloseDecision:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_FORCE_CLOSE_DECISION
	case shared.DecisionTaskFailedCauseFailoverCloseDecision:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_FAILOVER_CLOSE_DECISION
	case shared.DecisionTaskFailedCauseBadSignalInputSize:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SIGNAL_INPUT_SIZE
	case shared.DecisionTaskFailedCauseResetWorkflow:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_RESET_WORKFLOW
	case shared.DecisionTaskFailedCauseBadBinary:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_BINARY
	case shared.DecisionTaskFailedCauseScheduleActivityDuplicateID:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_SCHEDULE_ACTIVITY_DUPLICATE_ID
	case shared.DecisionTaskFailedCauseBadSearchAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SEARCH_ATTRIBUTES
	}
	panic("unexpected enum value")
}

func toDecisionTaskTimedOutCause(t *shared.DecisionTaskTimedOutCause) apiv1.DecisionTaskTimedOutCause {
	if t == nil {
		return apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_INVALID
	}
	switch *t {
	case shared.DecisionTaskTimedOutCauseTimeout:
		return apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_TIMEOUT
	case shared.DecisionTaskTimedOutCauseReset:
		return apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_RESET
	}
	panic("unexpected enum value")
}

func toCancelExternalWorkflowExecutionFailedCause(t *shared.CancelExternalWorkflowExecutionFailedCause) apiv1.CancelExternalWorkflowExecutionFailedCause {
	if t == nil {
		return apiv1.CancelExternalWorkflowExecutionFailedCause_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID
	}
	switch *t {
	case shared.CancelExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution:
		return apiv1.CancelExternalWorkflowExecutionFailedCause_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_UNKNOWN_EXTERNAL_WORKFLOW_EXECUTION
	}
	panic("unexpected enum value")
}

func toParentClosePolicy(t *shared.ParentClosePolicy) apiv1.ParentClosePolicy {
	if t == nil {
		return apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_INVALID
	}
	switch *t {
	case shared.ParentClosePolicyAbandon:
		return apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_ABANDON
	case shared.ParentClosePolicyRequestCancel:
		return apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_REQUEST_CANCEL
	case shared.ParentClosePolicyTerminate:
		return apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_TERMINATE
	}
	panic("unexpected enum value")
}

func toWorkflowIDReusePolicy(t *shared.WorkflowIdReusePolicy) apiv1.WorkflowIdReusePolicy {
	if t == nil {
		return apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_INVALID
	}
	switch *t {
	case shared.WorkflowIdReusePolicyAllowDuplicateFailedOnly:
		return apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY
	case shared.WorkflowIdReusePolicyAllowDuplicate:
		return apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE
	case shared.WorkflowIdReusePolicyRejectDuplicate:
		return apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE
	case shared.WorkflowIdReusePolicyTerminateIfRunning:
		return apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING
	}
	panic("unexpected enum value")
}

func toChildWorkflowExecutionFailedCause(t *shared.ChildWorkflowExecutionFailedCause) apiv1.ChildWorkflowExecutionFailedCause {
	if t == nil {
		return apiv1.ChildWorkflowExecutionFailedCause_CHILD_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID
	}
	switch *t {
	case shared.ChildWorkflowExecutionFailedCauseWorkflowAlreadyRunning:
		return apiv1.ChildWorkflowExecutionFailedCause_CHILD_WORKFLOW_EXECUTION_FAILED_CAUSE_WORKFLOW_ALREADY_RUNNING
	}
	panic("unexpected enum value")
}

func toSignalExternalWorkflowExecutionFailedCause(t *shared.SignalExternalWorkflowExecutionFailedCause) apiv1.SignalExternalWorkflowExecutionFailedCause {
	if t == nil {
		return apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID
	}
	switch *t {
	case shared.SignalExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution:
		return apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_UNKNOWN_EXTERNAL_WORKFLOW_EXECUTION
	}
	panic("unexpected enum value")
}