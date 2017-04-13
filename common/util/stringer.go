package util

import (
	"bytes"
	"fmt"
	"reflect"

	s "github.com/uber-go/cadence-client/.gen/go/shared"
)

func anyToString(d interface{}) string {
	var buf bytes.Buffer
	v := reflect.ValueOf(d)
	t := reflect.TypeOf(d)
	switch v.Kind() {
	case reflect.Ptr:
		buf.WriteString(anyToString(v.Elem().Interface()))
	case reflect.Struct:
		buf.WriteString("(")
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if f.Kind() == reflect.Invalid {
				continue
			}
			fieldValue := valueToString(f)
			if len(fieldValue) == 0 {
				continue
			}
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(fmt.Sprintf("%s:%s", t.Field(i).Name, fieldValue))
		}
		buf.WriteString(")")
	default:
		buf.WriteString(fmt.Sprint(d))
	}
	return buf.String()
}

func valueToString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Ptr:
		return valueToString(v.Elem())
	case reflect.Struct:
		return anyToString(v.Interface())
	case reflect.Invalid:
		return ""
	case reflect.Slice:
		// TODO: find a better way to handle this.
		return fmt.Sprintf("len=%d", v.Len())
	default:
		return fmt.Sprint(v.Interface())
	}
}

// HistoryEventToString convert HistoryEvent to string
func HistoryEventToString(e *s.HistoryEvent) string {
	var attr interface{}
	switch e.GetEventType() {
	case s.EventType_WorkflowExecutionStarted:
		attr = e.GetWorkflowExecutionStartedEventAttributes()

	case s.EventType_WorkflowExecutionCompleted:
		attr = e.GetWorkflowExecutionCompletedEventAttributes()

	case s.EventType_WorkflowExecutionFailed:
		attr = e.GetWorkflowExecutionFailedEventAttributes()

	case s.EventType_WorkflowExecutionTimedOut:
		attr = e.GetWorkflowExecutionTimedOutEventAttributes()

	case s.EventType_DecisionTaskScheduled:
		attr = e.GetDecisionTaskScheduledEventAttributes()

	case s.EventType_DecisionTaskStarted:
		attr = e.GetDecisionTaskStartedEventAttributes()

	case s.EventType_DecisionTaskCompleted:
		attr = e.GetDecisionTaskCompletedEventAttributes()

	case s.EventType_DecisionTaskTimedOut:
		attr = e.GetDecisionTaskTimedOutEventAttributes()

	case s.EventType_ActivityTaskScheduled:
		attr = e.GetActivityTaskScheduledEventAttributes()

	case s.EventType_ActivityTaskStarted:
		attr = e.GetActivityTaskStartedEventAttributes()

	case s.EventType_ActivityTaskCompleted:
		attr = e.GetActivityTaskCompletedEventAttributes()

	case s.EventType_ActivityTaskFailed:
		attr = e.GetActivityTaskFailedEventAttributes()

	case s.EventType_ActivityTaskTimedOut:
		attr = e.GetActivityTaskTimedOutEventAttributes()

	case s.EventType_ActivityTaskCancelRequested:
		attr = e.GetActivityTaskCancelRequestedEventAttributes()

	case s.EventType_RequestCancelActivityTaskFailed:
		attr = e.GetRequestCancelActivityTaskFailedEventAttributes()

	case s.EventType_ActivityTaskCanceled:
		attr = e.GetActivityTaskCanceledEventAttributes()

	case s.EventType_TimerStarted:
		attr = e.GetTimerStartedEventAttributes()

	case s.EventType_TimerFired:
		attr = e.GetTimerFiredEventAttributes()

	case s.EventType_CompleteWorkflowExecutionFailed:
		attr = e.GetCompleteWorkflowExecutionFailedEventAttributes()

	case s.EventType_CancelTimerFailed:
		attr = e.GetCancelTimerFailedEventAttributes()

	case s.EventType_TimerCanceled:
		attr = e.GetTimerCanceledEventAttributes()

	case s.EventType_MarkerRecorded:
		attr = e.GetMarkerRecordedEventAttributes()

	case s.EventType_WorkflowExecutionTerminated:
		attr = e.GetWorkflowExecutionTerminatedEventAttributes()
	}

	return e.GetEventType().String() + ": " + anyToString(attr)
}

// DecisionToString convert Decision to string
func DecisionToString(d *s.Decision) string {
	var attr interface{}
	switch d.GetDecisionType() {
	case s.DecisionType_ScheduleActivityTask:
		attr = d.GetScheduleActivityTaskDecisionAttributes()

	case s.DecisionType_RequestCancelActivityTask:
		attr = d.GetRequestCancelActivityTaskDecisionAttributes()

	case s.DecisionType_StartTimer:
		attr = d.GetStartTimerDecisionAttributes()

	case s.DecisionType_CancelTimer:
		attr = d.GetCancelTimerDecisionAttributes()

	case s.DecisionType_CompleteWorkflowExecution:
		attr = d.GetCompleteWorkflowExecutionDecisionAttributes()

	case s.DecisionType_FailWorkflowExecution:
		attr = d.GetFailWorkflowExecutionDecisionAttributes()

	case s.DecisionType_RecordMarker:
		attr = d.GetRecordMarkerDecisionAttributes()
	}

	return d.GetDecisionType().String() + ": " + anyToString(attr)
}
