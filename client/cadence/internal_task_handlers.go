package cadence

// All code in this file is private to the package.

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/uber-common/bark"
	m "github.com/uber-go/cadence-client/.gen/go/cadence"
	s "github.com/uber-go/cadence-client/.gen/go/shared"
	"github.com/uber-go/cadence-client/common"
	"github.com/uber-go/cadence-client/common/backoff"
	"github.com/uber-go/cadence-client/common/metrics"
	"github.com/uber-go/tally"
	"golang.org/x/net/context"
)

// interfaces
type (
	// workflowExecutionEventHandler process a single event.
	workflowExecutionEventHandler interface {
		// Process a single event and return the assosciated decisions.
		// Return List of decisions made, whether a decision is unhandled, any error.
		ProcessEvent(event *s.HistoryEvent) ([]*s.Decision, bool, error)
		StackTrace() string
		// Close for cleaning up resources on this event handler
		Close()
	}

	// workflowTask wraps a decision task.
	workflowTask struct {
		task *s.PollForDecisionTaskResponse
	}

	// activityTask wraps a activity task.
	activityTask struct {
		task *s.PollForActivityTaskResponse
	}
)

type (
	// workflowTaskHandlerImpl is the implementation of WorkflowTaskHandler
	workflowTaskHandlerImpl struct {
		workflowDefFactory workflowDefinitionFactory
		metricsScope       tally.Scope
		ppMgr              pressurePointMgr
		logger             bark.Logger
		identity           string
	}

	// activityTaskHandlerImpl is the implementation of ActivityTaskHandler
	activityTaskHandlerImpl struct {
		taskListName    string
		identity        string
		implementations map[ActivityType]activity
		service         m.TChanWorkflowService
		metricsScope    tally.Scope
		logger          bark.Logger
	}

	// history wrapper method to help information about events.
	history struct {
		workflowTask      *workflowTask
		eventsHandler     *workflowExecutionEventHandlerImpl
		currentIndex      int
		historyEventsSize int
	}
)

func newHistory(task *workflowTask, eventsHandler *workflowExecutionEventHandlerImpl) *history {
	return &history{
		workflowTask:      task,
		eventsHandler:     eventsHandler,
		currentIndex:      0,
		historyEventsSize: len(task.task.History.Events),
	}
}

// Get last non replayed event ID.
func (eh *history) LastNonReplayedID() int64 {
	if eh.workflowTask.task.PreviousStartedEventId == nil {
		return 0
	}
	return *eh.workflowTask.task.PreviousStartedEventId
}

func (eh *history) IsNextDecisionTimedOut(startIndex int) bool {
	events := eh.workflowTask.task.History.Events
	eventsSize := len(events)
	for i := startIndex; i < eventsSize; i++ {
		switch events[i].GetEventType() {
		case s.EventType_DecisionTaskCompleted:
			return false
		case s.EventType_DecisionTaskTimedOut:
			return true
		}
	}
	return false
}

func (eh *history) IsDecisionEvent(eventType s.EventType) bool {
	switch eventType {
	case s.EventType_WorkflowExecutionCompleted, s.EventType_WorkflowExecutionFailed, s.EventType_WorkflowExecutionTimedOut:
		return true
	case s.EventType_ActivityTaskScheduled, s.EventType_TimerStarted:
		return true
	default:
		return false
	}
}

func (eh *history) NextEvents() []*s.HistoryEvent {
	return eh.getNextEvents()
}

func (eh *history) getNextEvents() []*s.HistoryEvent {

	if eh.currentIndex == eh.historyEventsSize {
		return []*s.HistoryEvent{}
	}

	// Process events
	reorderedEvents := []*s.HistoryEvent{}
	history := eh.workflowTask.task.History

	// We need to re-order the events so the decider always sees in the same order.
	// For Ex: (pseudo code)
	//   ResultA := Schedule_Activity_A
	//   ResultB := Schedule_Activity_B
	//   if ResultB.IsReady() { panic error }
	//   ResultC := Schedule_Activity_C(ResultA)
	// If both A and B activities complete then we could have two different paths, Either Scheduling C (or) Panic'ing.
	// Workflow events:
	// 	Workflow_Start, DecisionStart1, DecisionComplete1, A_Schedule, B_Schedule, A_Complete,
	//      DecisionStart2, B_Complete, DecisionComplete2, C_Schedule.
	// B_Complete happened concurrent to execution of the decision(2), where C_Schedule is a result made by execution of decision(2).
	// One way to address is: Move all concurrent decisions to one after the decisions made by current decision.

	decisionStartToCompletionEvents := []*s.HistoryEvent{}
	decisionCompletionToStartEvents := []*s.HistoryEvent{}
	concurrentToDecision := true
	lastDecisionIndex := -1

OrderEvents:
	for ; eh.currentIndex < eh.historyEventsSize; eh.currentIndex++ {
		event := history.Events[eh.currentIndex]
		switch event.GetEventType() {
		case s.EventType_DecisionTaskStarted:
			if !eh.IsNextDecisionTimedOut(eh.currentIndex) {
				// Set replay clock.
				ts := time.Unix(0, event.GetTimestamp())
				eh.eventsHandler.workflowEnvironmentImpl.SetCurrentReplayTime(ts)
				eh.currentIndex++ // Sine we already processed the current event
				break OrderEvents
			}

		case s.EventType_DecisionTaskCompleted:
			concurrentToDecision = false

		case s.EventType_DecisionTaskScheduled, s.EventType_DecisionTaskTimedOut:
		// Skip

		default:
			if concurrentToDecision {
				decisionStartToCompletionEvents = append(decisionStartToCompletionEvents, event)
			} else {
				if eh.IsDecisionEvent(event.GetEventType()) {
					lastDecisionIndex = len(decisionCompletionToStartEvents)
				}
				decisionCompletionToStartEvents = append(decisionCompletionToStartEvents, event)
			}
		}
	}

	// Reorder events to correspond to the order that decider sees them.
	// The main difference is that events that were added during decision task execution
	// should be processed after events that correspond to the decisions.
	// Otherwise the replay is going to break.

	// First are events that correspond to the previous task decisions
	if lastDecisionIndex >= 0 {
		reorderedEvents = decisionCompletionToStartEvents[:lastDecisionIndex+1]
	}
	// Second are events that were added during previous task execution
	reorderedEvents = append(reorderedEvents, decisionStartToCompletionEvents...)
	// The last are events that were added after previous task completion
	if lastDecisionIndex+1 < len(decisionCompletionToStartEvents) {
		reorderedEvents = append(reorderedEvents, decisionCompletionToStartEvents[lastDecisionIndex+1:]...)
	}

	return reorderedEvents
}

// newWorkflowTaskHandler returns an implementation of workflow task handler.
func newWorkflowTaskHandler(factory workflowDefinitionFactory,
	params workerExecutionParameters, ppMgr pressurePointMgr) WorkflowTaskHandler {
	return &workflowTaskHandlerImpl{
		workflowDefFactory: factory,
		logger:             params.Logger,
		ppMgr:              ppMgr,
		metricsScope:       params.MetricsScope,
		identity:           params.Identity}
}

// ProcessWorkflowTask processes each all the events of the workflow task.
func (wth *workflowTaskHandlerImpl) ProcessWorkflowTask(
	task *s.PollForDecisionTaskResponse,
	emitStack bool,
) (result *s.RespondDecisionTaskCompletedRequest, stackTrace string, err error) {
	if task == nil {
		return nil, "", errors.New("nil workflowtask provided")
	}
	h := task.GetHistory()
	if h == nil || len(h.Events) == 0 {
		return nil, "", errors.New("nil or empty history")
	}
	event := h.Events[0]
	if h == nil {
		return nil, "", errors.New("nil first history event")
	}
	attributes := event.GetWorkflowExecutionStartedEventAttributes()
	if attributes == nil {
		return nil, "", errors.New("first history event is not WorkflowExecutionStarted")
	}
	taskList := attributes.GetTaskList()
	if taskList == nil {
		return nil, "", errors.New("nil TaskList in WorkflowExecutionStarted event")
	}

	wth.logger.Debugf("Processing New Workflow Task: Type=%s, PreviousStartedEventId=%d",
		task.GetWorkflowType().GetName(), task.GetPreviousStartedEventId())

	// Setup workflow Info
	workflowInfo := &WorkflowInfo{
		WorkflowType: flowWorkflowTypeFrom(*task.WorkflowType),
		TaskListName: taskList.GetName(),
		// workflowExecution
	}

	isWorkflowCompleted := false
	var completionResult []byte
	var failure error

	completeHandler := func(result []byte, err error) {
		completionResult = result
		failure = err
		isWorkflowCompleted = true
	}

	eventHandler := newWorkflowExecutionEventHandler(
		workflowInfo, wth.workflowDefFactory, completeHandler, wth.logger)
	defer eventHandler.Close()
	reorderedHistory := newHistory(&workflowTask{task: task}, eventHandler.(*workflowExecutionEventHandlerImpl))
	decisions := []*s.Decision{}
	replayDecisions := []*s.Decision{}
	respondEvents := []*s.HistoryEvent{}
	unhandledDecision := false

	startTime := time.Now()

	// Process events
ProcessEvents:
	for {
		reorderedEvents := reorderedHistory.NextEvents()
		if len(reorderedEvents) == 0 {
			break ProcessEvents
		}

		for _, event := range reorderedEvents {
			wth.logger.Debugf("ProcessEvent: Id=%d, EventType=%v", event.GetEventId(), event.GetEventType())

			isInReplay := event.GetEventId() < reorderedHistory.LastNonReplayedID()
			if isEventTypeRespondToDecision(event.GetEventType()) {
				respondEvents = append(respondEvents, event)
			}

			// Any metrics.
			wth.reportAnyMetrics(event, isInReplay)

			// Any pressure points.
			err := wth.executeAnyPressurePoints(event, isInReplay)
			if err != nil {
				return nil, "", err
			}

			eventDecisions, unhandled, err := eventHandler.ProcessEvent(event)
			if err != nil {
				return nil, "", err
			}
			if unhandled {
				unhandledDecision = unhandled
			}

			if eventDecisions != nil {
				if !isInReplay {
					decisions = append(decisions, eventDecisions...)
				} else {
					replayDecisions = append(replayDecisions, eventDecisions...)
				}
			}

			if isWorkflowCompleted {
				// If workflow is already completed then we can break from processing
				// further decisions.
				break ProcessEvents
			}
		}
	}

	// check if decisions from reply matches to the history events
	if err := matchReplayWithHistory(replayDecisions, respondEvents); err != nil {
		wth.logger.Warnf("replay and history match failed: %s", err)
		return nil, "", err
	}

	eventDecisions := wth.completeWorkflow(isWorkflowCompleted, unhandledDecision, completionResult, failure)
	if len(eventDecisions) > 0 {
		decisions = append(decisions, eventDecisions...)
		if wth.metricsScope != nil {
			wth.metricsScope.Counter(metrics.WorkflowsCompletionTotalCounter).Inc(1)
			elapsed := time.Now().Sub(startTime)
			wth.metricsScope.Timer(metrics.WorkflowEndToEndLatency).Record(elapsed)
		}
	}

	// Fill the response.
	taskCompletionRequest := &s.RespondDecisionTaskCompletedRequest{
		TaskToken: task.TaskToken,
		Decisions: decisions,
		Identity:  common.StringPtr(wth.identity),
		// ExecutionContext:
	}
	if emitStack {
		stackTrace = eventHandler.StackTrace()
	}
	return taskCompletionRequest, stackTrace, nil
}

// for every decision, there is one EventType respond to that decision.
func isEventTypeRespondToDecision(t s.EventType) bool {
	switch t {
	case s.EventType_ActivityTaskScheduled:
		return true
	case s.EventType_ActivityTaskCancelRequested:
		return true
	case s.EventType_TimerStarted:
		return true
	case s.EventType_TimerCanceled:
		return true
	case s.EventType_WorkflowExecutionCompleted:
		return true
	case s.EventType_WorkflowExecutionFailed:
		return true
	default:
		return false
	}
}

func matchReplayWithHistory(replayDecisions []*s.Decision, historyEvents []*s.HistoryEvent) error {
	for i := 0; i < len(historyEvents); i++ {
		e := historyEvents[i]
		if i >= len(replayDecisions) {
			return fmt.Errorf("nondeterministic workflow: missing replay decision for %s", historyEventToString(e))
		}
		d := replayDecisions[i]
		if !isDecisionMatchEvent(d, e, false) {
			return fmt.Errorf("nondeterministic workflow: history event is %s, replay decision is %s",
				historyEventToString(e), decisionToString(d))
		}
	}
	return nil
}

// TODO: find a better place for these toString methods. also consider making them public.
func historyEventToString(e *s.HistoryEvent) string {
	switch e.GetEventType() {
	case s.EventType_ActivityTaskScheduled:
		attr := e.GetActivityTaskScheduledEventAttributes()
		return fmt.Sprintf("%s, ActivityType: %s, ActivityId: %s, TaskList: %s, InputSize: %d",
			e.GetEventType(), attr.GetActivityType().GetName(), attr.GetActivityId(), attr.GetTaskList().GetName(), len(attr.GetInput()))

	case s.EventType_ActivityTaskCancelRequested:
		attr := e.GetActivityTaskCancelRequestedEventAttributes()
		return fmt.Sprintf("%s, ActivityId: %s", e.GetEventType(), attr.GetActivityId())

	case s.EventType_TimerStarted:
		attr := e.GetTimerStartedEventAttributes()
		return fmt.Sprintf("%s, TimerId: %s, StartToFireTimeoutSeconds: %d",
			e.GetEventType(), attr.GetTimerId(), attr.GetStartToFireTimeoutSeconds())

	case s.EventType_TimerCanceled:
		attr := e.GetTimerCanceledEventAttributes()
		return fmt.Sprintf("%s, TimerId: %s", e.GetEventType(), attr.GetTimerId())

	case s.EventType_WorkflowExecutionCompleted:
		attr := e.GetWorkflowExecutionCompletedEventAttributes()
		return fmt.Sprintf("%s, ResultSize: %d", e.GetEventType(), len(attr.GetResult_()))

	case s.EventType_WorkflowExecutionFailed:
		attr := e.GetWorkflowExecutionFailedEventAttributes()
		return fmt.Sprintf("%s, Reason: %d", e.GetEventType(), attr.GetReason())
	}

	// TODO: add cases for other event type. for now this should work.
	return e.GetEventType().String()
}

func decisionToString(d *s.Decision) string {
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
	}

	return d.GetDecisionType().String()
}

func isDecisionMatchEvent(d *s.Decision, e *s.HistoryEvent, strictMode bool) bool {
	switch d.GetDecisionType() {
	case s.DecisionType_ScheduleActivityTask:
		if e.GetEventType() != s.EventType_ActivityTaskScheduled {
			return false
		}
		eventAttributes := e.GetActivityTaskScheduledEventAttributes()
		decisionAttributes := d.GetScheduleActivityTaskDecisionAttributes()

		if eventAttributes.GetActivityId() != decisionAttributes.GetActivityId() ||
			eventAttributes.GetActivityType().GetName() != decisionAttributes.GetActivityType().GetName() ||
			(strictMode && eventAttributes.GetTaskList().GetName() != decisionAttributes.GetTaskList().GetName()) ||
			(strictMode && bytes.Compare(eventAttributes.GetInput(), decisionAttributes.GetInput()) != 0) {
			return false
		}

		return true

	case s.DecisionType_RequestCancelActivityTask:
		if e.GetEventType() != s.EventType_ActivityTaskCancelRequested {
			return false
		}
		eventAttributes := e.GetActivityTaskCancelRequestedEventAttributes()
		decisionAttributes := d.GetRequestCancelActivityTaskDecisionAttributes()

		if eventAttributes.GetActivityId() != decisionAttributes.GetActivityId() {
			return false
		}

		return true

	case s.DecisionType_StartTimer:
		if e.GetEventType() != s.EventType_TimerStarted {
			return false
		}
		eventAttributes := e.GetTimerStartedEventAttributes()
		decisionAttributes := d.GetStartTimerDecisionAttributes()

		if eventAttributes.GetTimerId() != decisionAttributes.GetTimerId() ||
			eventAttributes.GetStartToFireTimeoutSeconds() != decisionAttributes.GetStartToFireTimeoutSeconds() {
			return false
		}

		return true

	case s.DecisionType_CancelTimer:
		if e.GetEventType() != s.EventType_TimerCanceled {
			return false
		}
		eventAttributes := e.GetTimerCanceledEventAttributes()
		decisionAttributes := d.GetCancelTimerDecisionAttributes()

		if eventAttributes.GetTimerId() != decisionAttributes.GetTimerId() {
			return false
		}

		return true

	case s.DecisionType_CompleteWorkflowExecution:
		if e.GetEventType() != s.EventType_WorkflowExecutionCompleted {
			return false
		}
		if strictMode {
			eventAttributes := e.GetWorkflowExecutionCompletedEventAttributes()
			decisionAttributes := d.GetCompleteWorkflowExecutionDecisionAttributes()

			if bytes.Compare(eventAttributes.GetResult_(), decisionAttributes.GetResult_()) != 0 {
				return false
			}
		}

		return true

	case s.DecisionType_FailWorkflowExecution:
		if e.GetEventType() != s.EventType_WorkflowExecutionFailed {
			return false
		}
		if strictMode {
			eventAttributes := e.GetWorkflowExecutionFailedEventAttributes()
			decisionAttributes := d.GetFailWorkflowExecutionDecisionAttributes()

			if eventAttributes.GetReason() != decisionAttributes.GetReason() ||
				bytes.Compare(eventAttributes.GetDetails(), decisionAttributes.GetDetails()) != 0 {
				return false
			}
		}

		return true
	}

	return false
}

func (wth *workflowTaskHandlerImpl) completeWorkflow(isWorkflowCompleted bool, unhandledDecision bool, completionResult []byte,
	err error) []*s.Decision {
	decisions := []*s.Decision{}
	if !unhandledDecision {
		if err != nil {
			// Workflow failures
			failDecision := createNewDecision(s.DecisionType_FailWorkflowExecution)
			reason, details := getErrorDetails(err)
			failDecision.FailWorkflowExecutionDecisionAttributes = &s.FailWorkflowExecutionDecisionAttributes{
				Reason:  common.StringPtr(reason),
				Details: details,
			}
			decisions = append(decisions, failDecision)
		} else if isWorkflowCompleted {
			// Workflow completion
			completeDecision := createNewDecision(s.DecisionType_CompleteWorkflowExecution)
			completeDecision.CompleteWorkflowExecutionDecisionAttributes = &s.CompleteWorkflowExecutionDecisionAttributes{
				Result_: completionResult,
			}
			decisions = append(decisions, completeDecision)
		}
	}
	return decisions
}

func (wth *workflowTaskHandlerImpl) executeAnyPressurePoints(event *s.HistoryEvent, isInReplay bool) error {
	if wth.ppMgr != nil && !reflect.ValueOf(wth.ppMgr).IsNil() && !isInReplay {
		switch event.GetEventType() {
		case s.EventType_DecisionTaskStarted:
			return wth.ppMgr.Execute(pressurePointTypeDecisionTaskStartTimeout)
		case s.EventType_ActivityTaskScheduled:
			return wth.ppMgr.Execute(pressurePointTypeActivityTaskScheduleTimeout)
		case s.EventType_ActivityTaskStarted:
			return wth.ppMgr.Execute(pressurePointTypeActivityTaskStartTimeout)
		case s.EventType_DecisionTaskCompleted:
			return wth.ppMgr.Execute(pressurePointTypeDecisionTaskCompleted)
		}
	}
	return nil
}

func (wth *workflowTaskHandlerImpl) reportAnyMetrics(event *s.HistoryEvent, isInReplay bool) {
	if wth.metricsScope != nil && !isInReplay {
		switch event.GetEventType() {
		case s.EventType_DecisionTaskTimedOut:
			wth.metricsScope.Counter(metrics.DecisionsTimeoutCounter).Inc(1)
		}
	}
}

func newActivityTaskHandler(activities []activity,
	service m.TChanWorkflowService, params workerExecutionParameters) ActivityTaskHandler {
	implementations := make(map[ActivityType]activity)
	for _, a := range activities {
		implementations[a.ActivityType()] = a
	}
	return &activityTaskHandlerImpl{
		taskListName:    params.TaskList,
		identity:        params.Identity,
		implementations: implementations,
		service:         service,
		logger:          params.Logger,
		metricsScope:    params.MetricsScope}
}

type cadenceInvoker struct {
	identity  string
	service   m.TChanWorkflowService
	taskToken []byte
}

func (i *cadenceInvoker) Heartbeat(details []byte) error {
	return recordActivityHeartbeat(i.service, i.identity, i.taskToken, details)
}

func newServiceInvoker(taskToken []byte, identity string, service m.TChanWorkflowService) ServiceInvoker {
	return &cadenceInvoker{
		taskToken: taskToken,
		identity:  identity,
		service:   service,
	}
}

// Execute executes an implementation of the activity.
func (ath *activityTaskHandlerImpl) Execute(t *s.PollForActivityTaskResponse) (interface{}, error) {
	ath.logger.Debugf("[WorkflowID: %s] Execute activity: %s",
		t.GetWorkflowExecution().GetWorkflowId(), t.GetActivityType().GetName())

	invoker := newServiceInvoker(t.TaskToken, ath.identity, ath.service)
	ctx := WithActivityTask(context.Background(), t, invoker)
	activityType := *t.GetActivityType()
	activityImplementation, ok := ath.implementations[flowActivityTypeFrom(activityType)]
	if !ok {
		// Couldn't find the activity implementation.
		return nil, fmt.Errorf("No implementation for activityType=%v", activityType.GetName())
	}

	output, err := activityImplementation.Execute(ctx, t.GetInput())
	return convertActivityResultToRespondRequest(ath.identity, t.TaskToken, output, err), nil
}

func createNewDecision(decisionType s.DecisionType) *s.Decision {
	return &s.Decision{
		DecisionType: common.DecisionTypePtr(decisionType),
	}
}

func recordActivityHeartbeat(service m.TChanWorkflowService, identity string, taskToken, details []byte) error {
	request := &s.RecordActivityTaskHeartbeatRequest{
		TaskToken: taskToken,
		Details:   details,
		Identity:  common.StringPtr(identity)}

	var heartbeatResponse *s.RecordActivityTaskHeartbeatResponse
	heartbeatErr := backoff.Retry(
		func() error {
			ctx, cancel := common.NewTChannelContext(respondTaskServiceTimeOut, common.RetryDefaultOptions)
			defer cancel()

			var err error
			heartbeatResponse, err = service.RecordActivityTaskHeartbeat(ctx, request)
			return err
		}, serviceOperationRetryPolicy, isServiceTransientError)

	if heartbeatErr == nil && heartbeatResponse != nil && heartbeatResponse.GetCancelRequested() {
		return NewCanceledError()
	}

	return heartbeatErr
}
