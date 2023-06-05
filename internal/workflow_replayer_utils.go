package internal

import (
	"bytes"
	"fmt"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common/util"
	"reflect"
	"strings"
)

func matchReplayWithHistory(replayDecisions []*s.Decision, historyEvents []*s.HistoryEvent) error {
	di := 0
	hi := 0
	hSize := len(historyEvents)
	dSize := len(replayDecisions)
matchLoop:
	for hi < hSize || di < dSize {
		var e *s.HistoryEvent
		if hi < hSize {
			e = historyEvents[hi]
			if skipDeterministicCheckForUpsertChangeVersion(historyEvents, hi) {
				hi += 2
				continue matchLoop
			}
			if skipDeterministicCheckForEvent(e) {
				hi++
				continue matchLoop
			}
		}

		var d *s.Decision
		if di < dSize {
			d = replayDecisions[di]
			if skipDeterministicCheckForDecision(d) {
				di++
				continue matchLoop
			}
		}

		if d == nil {
			return fmt.Errorf("nondeterministic workflow: missing replay decision for %s", util.HistoryEventToString(e))
		}

		if e == nil {
			return fmt.Errorf("nondeterministic workflow: extra replay decision for %s", util.DecisionToString(d))
		}

		if !isDecisionMatchEvent(d, e, false) {
			return fmt.Errorf("nondeterministic workflow: history event is %s, replay decision is %s",
				util.HistoryEventToString(e), util.DecisionToString(d))
		}

		di++
		hi++
	}
	return nil
}

func lastPartOfName(name string) string {
	name = strings.TrimSuffix(name, "-fm")
	lastDotIdx := strings.LastIndex(name, ".")
	if lastDotIdx < 0 || lastDotIdx == len(name)-1 {
		return name
	}
	return name[lastDotIdx+1:]
}

func skipDeterministicCheckForEvent(e *s.HistoryEvent) bool {
	if e.GetEventType() == s.EventTypeMarkerRecorded {
		markerName := e.MarkerRecordedEventAttributes.GetMarkerName()
		if markerName == versionMarkerName || markerName == mutableSideEffectMarkerName {
			return true
		}
	}
	return false
}

// special check for upsert change version event
func skipDeterministicCheckForUpsertChangeVersion(events []*s.HistoryEvent, idx int) bool {
	e := events[idx]
	if e.GetEventType() == s.EventTypeMarkerRecorded &&
		e.MarkerRecordedEventAttributes.GetMarkerName() == versionMarkerName &&
		idx < len(events)-1 &&
		events[idx+1].GetEventType() == s.EventTypeUpsertWorkflowSearchAttributes {
		if _, ok := events[idx+1].UpsertWorkflowSearchAttributesEventAttributes.SearchAttributes.IndexedFields[CadenceChangeVersion]; ok {
			return true
		}
	}
	return false
}

func skipDeterministicCheckForDecision(d *s.Decision) bool {
	if d.GetDecisionType() == s.DecisionTypeRecordMarker {
		markerName := d.RecordMarkerDecisionAttributes.GetMarkerName()
		if markerName == versionMarkerName || markerName == mutableSideEffectMarkerName {
			return true
		}
	}
	return false
}

func isDecisionMatchEvent(d *s.Decision, e *s.HistoryEvent, strictMode bool) bool {
	switch d.GetDecisionType() {
	case s.DecisionTypeScheduleActivityTask:
		if e.GetEventType() != s.EventTypeActivityTaskScheduled {
			return false
		}
		eventAttributes := e.ActivityTaskScheduledEventAttributes
		decisionAttributes := d.ScheduleActivityTaskDecisionAttributes

		if eventAttributes.GetActivityId() != decisionAttributes.GetActivityId() ||
			lastPartOfName(eventAttributes.ActivityType.GetName()) != lastPartOfName(decisionAttributes.ActivityType.GetName()) ||
			(strictMode && eventAttributes.TaskList.GetName() != decisionAttributes.TaskList.GetName()) ||
			(strictMode && bytes.Compare(eventAttributes.Input, decisionAttributes.Input) != 0) {
			return false
		}

		return true

	case s.DecisionTypeRequestCancelActivityTask:
		if e.GetEventType() != s.EventTypeActivityTaskCancelRequested {
			return false
		}
		decisionAttributes := d.RequestCancelActivityTaskDecisionAttributes
		eventAttributes := e.ActivityTaskCancelRequestedEventAttributes
		if eventAttributes.GetActivityId() != decisionAttributes.GetActivityId() {
			return false
		}

		return true

	case s.DecisionTypeStartTimer:
		if e.GetEventType() != s.EventTypeTimerStarted {
			return false
		}
		eventAttributes := e.TimerStartedEventAttributes
		decisionAttributes := d.StartTimerDecisionAttributes

		if eventAttributes.GetTimerId() != decisionAttributes.GetTimerId() ||
			(strictMode && eventAttributes.GetStartToFireTimeoutSeconds() != decisionAttributes.GetStartToFireTimeoutSeconds()) {
			return false
		}

		return true

	case s.DecisionTypeCancelTimer:
		if e.GetEventType() != s.EventTypeTimerCanceled && e.GetEventType() != s.EventTypeCancelTimerFailed {
			return false
		}
		decisionAttributes := d.CancelTimerDecisionAttributes
		if e.GetEventType() == s.EventTypeTimerCanceled {
			eventAttributes := e.TimerCanceledEventAttributes
			if eventAttributes.GetTimerId() != decisionAttributes.GetTimerId() {
				return false
			}
		} else if e.GetEventType() == s.EventTypeCancelTimerFailed {
			eventAttributes := e.CancelTimerFailedEventAttributes
			if eventAttributes.GetTimerId() != decisionAttributes.GetTimerId() {
				return false
			}
		}

		return true

	case s.DecisionTypeCompleteWorkflowExecution:
		if e.GetEventType() != s.EventTypeWorkflowExecutionCompleted {
			return false
		}
		if strictMode {
			eventAttributes := e.WorkflowExecutionCompletedEventAttributes
			decisionAttributes := d.CompleteWorkflowExecutionDecisionAttributes

			if bytes.Compare(eventAttributes.Result, decisionAttributes.Result) != 0 {
				return false
			}
		}

		return true

	case s.DecisionTypeFailWorkflowExecution:
		if e.GetEventType() != s.EventTypeWorkflowExecutionFailed {
			return false
		}
		if strictMode {
			eventAttributes := e.WorkflowExecutionFailedEventAttributes
			decisionAttributes := d.FailWorkflowExecutionDecisionAttributes

			if eventAttributes.GetReason() != decisionAttributes.GetReason() ||
				bytes.Compare(eventAttributes.Details, decisionAttributes.Details) != 0 {
				return false
			}
		}

		return true

	case s.DecisionTypeRecordMarker:
		if e.GetEventType() != s.EventTypeMarkerRecorded {
			return false
		}
		eventAttributes := e.MarkerRecordedEventAttributes
		decisionAttributes := d.RecordMarkerDecisionAttributes
		if eventAttributes.GetMarkerName() != decisionAttributes.GetMarkerName() {
			return false
		}

		return true

	case s.DecisionTypeRequestCancelExternalWorkflowExecution:
		if e.GetEventType() != s.EventTypeRequestCancelExternalWorkflowExecutionInitiated {
			return false
		}
		eventAttributes := e.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes
		decisionAttributes := d.RequestCancelExternalWorkflowExecutionDecisionAttributes
		if checkDomainsInDecisionAndEvent(eventAttributes.GetDomain(), decisionAttributes.GetDomain()) ||
			eventAttributes.WorkflowExecution.GetWorkflowId() != decisionAttributes.GetWorkflowId() {
			return false
		}

		return true

	case s.DecisionTypeSignalExternalWorkflowExecution:
		if e.GetEventType() != s.EventTypeSignalExternalWorkflowExecutionInitiated {
			return false
		}
		eventAttributes := e.SignalExternalWorkflowExecutionInitiatedEventAttributes
		decisionAttributes := d.SignalExternalWorkflowExecutionDecisionAttributes
		if checkDomainsInDecisionAndEvent(eventAttributes.GetDomain(), decisionAttributes.GetDomain()) ||
			eventAttributes.GetSignalName() != decisionAttributes.GetSignalName() ||
			eventAttributes.WorkflowExecution.GetWorkflowId() != decisionAttributes.Execution.GetWorkflowId() {
			return false
		}

		return true

	case s.DecisionTypeCancelWorkflowExecution:
		if e.GetEventType() != s.EventTypeWorkflowExecutionCanceled {
			return false
		}
		if strictMode {
			eventAttributes := e.WorkflowExecutionCanceledEventAttributes
			decisionAttributes := d.CancelWorkflowExecutionDecisionAttributes
			if bytes.Compare(eventAttributes.Details, decisionAttributes.Details) != 0 {
				return false
			}
		}
		return true

	case s.DecisionTypeContinueAsNewWorkflowExecution:
		if e.GetEventType() != s.EventTypeWorkflowExecutionContinuedAsNew {
			return false
		}

		return true

	case s.DecisionTypeStartChildWorkflowExecution:
		if e.GetEventType() != s.EventTypeStartChildWorkflowExecutionInitiated {
			return false
		}
		eventAttributes := e.StartChildWorkflowExecutionInitiatedEventAttributes
		decisionAttributes := d.StartChildWorkflowExecutionDecisionAttributes
		if lastPartOfName(eventAttributes.WorkflowType.GetName()) != lastPartOfName(decisionAttributes.WorkflowType.GetName()) ||
			(strictMode && checkDomainsInDecisionAndEvent(eventAttributes.GetDomain(), decisionAttributes.GetDomain())) ||
			(strictMode && eventAttributes.TaskList.GetName() != decisionAttributes.TaskList.GetName()) {
			return false
		}

		return true

	case s.DecisionTypeUpsertWorkflowSearchAttributes:
		if e.GetEventType() != s.EventTypeUpsertWorkflowSearchAttributes {
			return false
		}
		eventAttributes := e.UpsertWorkflowSearchAttributesEventAttributes
		decisionAttributes := d.UpsertWorkflowSearchAttributesDecisionAttributes
		if strictMode && !isSearchAttributesMatched(eventAttributes.SearchAttributes, decisionAttributes.SearchAttributes) {
			return false
		}
		return true
	}

	return false
}

func isSearchAttributesMatched(attrFromEvent, attrFromDecision *s.SearchAttributes) bool {
	if attrFromEvent != nil && attrFromDecision != nil {
		return reflect.DeepEqual(attrFromEvent.IndexedFields, attrFromDecision.IndexedFields)
	}
	return attrFromEvent == nil && attrFromDecision == nil
}

// return true if the check fails:
//
//	domain is not empty in decision
//	and domain is not replayDomain
//	and domains unmatch in decision and events
func checkDomainsInDecisionAndEvent(eventDomainName, decisionDomainName string) bool {
	if decisionDomainName == "" || IsReplayDomain(decisionDomainName) {
		return false
	}
	return eventDomainName != decisionDomainName
}
