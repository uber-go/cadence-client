package util

import (
	"fmt"

	s "github.com/uber-go/cadence-client/.gen/go/shared"
)

// TODO: find a better way to present input/result []byte

// HistoryEventToString convert HistoryEvent to string
func HistoryEventToString(e *s.HistoryEvent) string {

	switch e.GetEventType() {

	case s.EventType_WorkflowExecutionStarted:
		attr := e.GetWorkflowExecutionStartedEventAttributes()
		return fmt.Sprintf("%s, Identity: %s, WorkflowType: %s, TaskList: %s, InputSize: %d",
			e.GetEventType(), attr.GetIdentity(), attr.GetWorkflowType().GetName(), attr.GetTaskList().GetName(), len(attr.GetInput()))

	case s.EventType_WorkflowExecutionCompleted:
		attr := e.GetWorkflowExecutionCompletedEventAttributes()
		return fmt.Sprintf("%s, ResultSize: %d", e.GetEventType(), len(attr.GetResult_()))

	case s.EventType_WorkflowExecutionFailed:
		attr := e.GetWorkflowExecutionFailedEventAttributes()
		return fmt.Sprintf("%s, Reason: %d", e.GetEventType(), attr.GetReason())

	case s.EventType_WorkflowExecutionTimedOut:
		attr := e.GetWorkflowExecutionTimedOutEventAttributes()
		return fmt.Sprintf("%s, TimeoutType: %s", e.GetEventType(), attr.GetTimeoutType())

	case s.EventType_DecisionTaskScheduled:
		attr := e.GetDecisionTaskScheduledEventAttributes()
		return fmt.Sprintf("%s, TaskList: %s, StartToCloseTimeoutSeconds: %d", e.GetEventType(), attr.GetTaskList().GetName(), attr.GetStartToCloseTimeoutSeconds())

	case s.EventType_DecisionTaskStarted:
		attr := e.GetDecisionTaskStartedEventAttributes()
		return fmt.Sprintf("%s, Identity: %s, RequestId: %s, ScheduledEventId: %d", e.GetEventType(), attr.GetIdentity(), attr.GetRequestId(), attr.GetScheduledEventId())

	case s.EventType_DecisionTaskCompleted:
		attr := e.GetDecisionTaskCompletedEventAttributes()
		return fmt.Sprintf("%s, Identity: %s, ScheduledEventId: %d, StartedEventId: %d", e.GetEventType(), attr.GetIdentity(), attr.GetScheduledEventId(), attr.GetStartedEventId())

	case s.EventType_DecisionTaskTimedOut:
		attr := e.GetDecisionTaskTimedOutEventAttributes()
		return fmt.Sprintf("%s, TimeoutType: %s, ScheduledEventId: %d, StartedEventId: %d", e.GetEventType(), attr.GetTimeoutType(), attr.GetScheduledEventId(), attr.GetStartedEventId())

	case s.EventType_ActivityTaskScheduled:
		attr := e.GetActivityTaskScheduledEventAttributes()
		return fmt.Sprintf("%s, ActivityType: %s, ActivityId: %s, TaskList: %s, InputSize: %d",
			e.GetEventType(), attr.GetActivityType().GetName(), attr.GetActivityId(), attr.GetTaskList().GetName(), len(attr.GetInput()))

	case s.EventType_ActivityTaskStarted:
		attr := e.GetActivityTaskStartedEventAttributes()
		return fmt.Sprintf("%s, Identity: %s, RequestId: %s, ScheduledEventId: %d", e.GetEventType(), attr.GetIdentity(), attr.GetRequestId(), attr.GetScheduledEventId())

	case s.EventType_ActivityTaskCompleted:
		attr := e.GetActivityTaskCompletedEventAttributes()
		return fmt.Sprintf("%s, Identity: %s, ScheduledEventId: %d, StartedEventId: %d, ResultSize: %d", e.GetEventType(), attr.GetIdentity(), attr.GetStartedEventId(), attr.GetScheduledEventId(), len(attr.GetResult_()))

	case s.EventType_ActivityTaskFailed:
		attr := e.GetActivityTaskFailedEventAttributes()
		return fmt.Sprintf("%s, Identity: %s, ScheduledEventId: %d, StartedEventId: %d, Reason: %s", e.GetEventType(), attr.GetIdentity(), attr.GetStartedEventId(), attr.GetScheduledEventId(), attr.GetReason())

	case s.EventType_ActivityTaskTimedOut:
		attr := e.GetActivityTaskTimedOutEventAttributes()
		return fmt.Sprintf("%s, TimeoutType: %s, ScheduledEventId: %d, StartedEventId: %d", e.GetEventType(), attr.GetTimeoutType(), attr.GetStartedEventId(), attr.GetScheduledEventId())

	case s.EventType_ActivityTaskCancelRequested:
		attr := e.GetActivityTaskCancelRequestedEventAttributes()
		return fmt.Sprintf("%s, ActivityId: %s", e.GetEventType(), attr.GetActivityId())

	case s.EventType_RequestCancelActivityTaskFailed:
		attr := e.GetRequestCancelActivityTaskFailedEventAttributes()
		return fmt.Sprintf("%s, ActivityId: %s, Cause: %s", e.GetEventType(), attr.GetActivityId(), attr.GetCause())

	case s.EventType_ActivityTaskCanceled:
		attr := e.GetActivityTaskCanceledEventAttributes()
		return fmt.Sprintf("%s, Identity: %s, ScheduledEventId: %d, LatestCancelRequestedEventId %d", e.GetEventType(), attr.GetIdentity(), attr.GetScheduledEventId(), attr.GetLatestCancelRequestedEventId())

	case s.EventType_TimerStarted:
		attr := e.GetTimerStartedEventAttributes()
		return fmt.Sprintf("%s, TimerId: %s, StartToFireTimeoutSeconds: %d",
			e.GetEventType(), attr.GetTimerId(), attr.GetStartToFireTimeoutSeconds())

	case s.EventType_TimerFired:
		attr := e.GetTimerFiredEventAttributes()
		return fmt.Sprintf("%s, TimerId: %s", e.GetEventType(), attr.GetTimerId())

	case s.EventType_CompleteWorkflowExecutionFailed:
		attr := e.GetCompleteWorkflowExecutionFailedEventAttributes()
		return fmt.Sprintf("%s, Cause: %s", e.GetEventType(), attr.GetCause())

	case s.EventType_CancelTimerFailed:
		attr := e.GetCancelTimerFailedEventAttributes()
		return fmt.Sprintf("%s, Identity: %s, TimerId: %s, Cause: %s", e.GetEventType(), attr.GetIdentity(), attr.GetTimerId(), attr.GetCause())

	case s.EventType_TimerCanceled:
		attr := e.GetTimerCanceledEventAttributes()
		return fmt.Sprintf("%s, TimerId: %s", e.GetEventType(), attr.GetTimerId())

	case s.EventType_MarkerRecorded:
		attr := e.GetMarkerRecordedEventAttributes()
		return fmt.Sprintf("%s, Marker: %s", e.GetEventType(), attr.GetMarkerName())

	case s.EventType_WorkflowExecutionTerminated:
		attr := e.GetWorkflowExecutionTerminatedEventAttributes()
		return fmt.Sprintf("%s, Identity :%s, Reason: %s", e.GetEventType(), attr.GetIdentity(), attr.GetReason())
	}

	return e.GetEventType().String()
}

// DecisionToString convert Decision to string
func DecisionToString(d *s.Decision) string {
	switch d.GetDecisionType() {
	case s.DecisionType_ScheduleActivityTask:
		attr := d.GetScheduleActivityTaskDecisionAttributes()
		return fmt.Sprintf("%s, ActivityType: %s, ActivityId: %s, TaskList: %s, InputSize: %d",
			d.GetDecisionType(), attr.GetActivityType().GetName(), attr.GetActivityId(), attr.GetTaskList().GetName(), len(attr.GetInput()))

	case s.DecisionType_RequestCancelActivityTask:
		attr := d.GetRequestCancelActivityTaskDecisionAttributes()
		return fmt.Sprintf("%s, ActivityId: %s", d.GetDecisionType(), attr.GetActivityId())

	case s.DecisionType_StartTimer:
		attr := d.GetStartTimerDecisionAttributes()
		return fmt.Sprintf("%s, TimerId: %s, StartToFireTimeoutSeconds: %d",
			d.GetDecisionType(), attr.GetTimerId(), attr.GetStartToFireTimeoutSeconds())

	case s.DecisionType_CancelTimer:
		attr := d.GetCancelTimerDecisionAttributes()
		return fmt.Sprintf("%s, TimerId: %s", d.GetDecisionType(), attr.GetTimerId())

	case s.DecisionType_CompleteWorkflowExecution:
		attr := d.GetCompleteWorkflowExecutionDecisionAttributes()
		return fmt.Sprintf("%s, ResultSize: %d", d.GetDecisionType(), len(attr.GetResult_()))

	case s.DecisionType_FailWorkflowExecution:
		attr := d.GetFailWorkflowExecutionDecisionAttributes()
		return fmt.Sprintf("%s, Reason: %d", d.GetDecisionType(), attr.GetReason())

	case s.DecisionType_RecordMarker:
		attr := d.GetRecordMarkerDecisionAttributes()
		return fmt.Sprintf("%s, MarkerName: %d", d.GetDecisionType(), attr.GetMarkerName())
	}

	return d.GetDecisionType().String()
}
