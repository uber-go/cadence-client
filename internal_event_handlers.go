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

package cadence

// All code in this file is private to the package.

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	m "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/common"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Assert that structs do indeed implement the interfaces
var _ workflowEnvironment = (*workflowEnvironmentImpl)(nil)
var _ workflowExecutionEventHandler = (*workflowExecutionEventHandlerImpl)(nil)

type (
	// completionHandler Handler to indicate completion result
	completionHandler func(result []byte, err error)

	// workflowExecutionEventHandlerImpl handler to handle workflowExecutionEventHandler
	workflowExecutionEventHandlerImpl struct {
		*workflowEnvironmentImpl
		workflowDefinition workflowDefinition
	}

	childWorkflowHandle struct {
		resultCallback      resultHandler
		startedCallback     func(r WorkflowExecution, e error)
		waitForCancellation bool
		workflowExecution   *WorkflowExecution
	}

	// workflowEnvironmentImpl an implementation of workflowEnvironment represents a environment for workflow execution.
	workflowEnvironmentImpl struct {
		workflowInfo              *WorkflowInfo
		workflowDefinitionFactory workflowDefinitionFactory

		scheduledActivities            map[string]resultHandler // Map of Activities(activity ID ->) and their response handlers
		waitForCancelRequestActivities map[string]bool          // Map of activity ID to whether to wait for cancelation.
		scheduledEventIDToActivityID   map[int64]string         // Mapping from scheduled event ID to activity ID
		scheduledTimers                map[string]resultHandler // Map of scheduledTimers(timer ID ->) and their response handlers
		sideEffectResult               map[int32][]byte
		scheduledChildWorkflows        map[string]*childWorkflowHandle // Map of scheduledChildWorkflows
		counterID                      int32                           // To generate activity IDs
		executeDecisions               []*m.Decision                   // Decisions made during the execute of the workflow
		completeHandler                completionHandler               // events completion handler
		currentReplayTime              time.Time                       // Indicates current replay time of the decision.
		postEventHooks                 []func()                        // postEvent hooks that need to be executed at the end of the event.
		cancelHandler                  func()                          // A cancel handler to be invoked on a cancel notification
		signalHandler                  func(name string, input []byte) // A signal handler to be invoked on a signal event
		logger                         *zap.Logger
		isReplay                       bool // flag to indicate if workflow is in replay mode
		enableLoggingInReplay          bool // flag to indicate if workflow should enable logging in replay mode
	}

	// wrapper around zapcore.Core that will be aware of replay
	replayAwareZapCore struct {
		zapcore.Core
		isReplay              *bool // pointer to bool that indicate if it is in replay mode
		enableLoggingInReplay *bool // pointer to bool that indicate if logging is enabled in replay mode
	}
)

var sideEffectMarkerName = "SideEffect"

func wrapLogger(isReplay *bool, enableLoggingInReplay *bool) func(zapcore.Core) zapcore.Core {
	return func(c zapcore.Core) zapcore.Core {
		return &replayAwareZapCore{c, isReplay, enableLoggingInReplay}
	}
}

func (c *replayAwareZapCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if *c.isReplay && !*c.enableLoggingInReplay {
		return checkedEntry
	}
	return c.Core.Check(entry, checkedEntry)
}

func (c *replayAwareZapCore) With(fields []zapcore.Field) zapcore.Core {
	coreWithFields := c.Core.With(fields)
	return &replayAwareZapCore{coreWithFields, c.isReplay, c.enableLoggingInReplay}
}

func newWorkflowExecutionEventHandler(workflowInfo *WorkflowInfo, workflowDefinitionFactory workflowDefinitionFactory,
	completeHandler completionHandler, logger *zap.Logger, enableLoggingInReplay bool) workflowExecutionEventHandler {
	context := &workflowEnvironmentImpl{
		workflowInfo:                   workflowInfo,
		workflowDefinitionFactory:      workflowDefinitionFactory,
		scheduledActivities:            make(map[string]resultHandler),
		waitForCancelRequestActivities: make(map[string]bool),
		scheduledEventIDToActivityID:   make(map[int64]string),
		scheduledTimers:                make(map[string]resultHandler),
		executeDecisions:               make([]*m.Decision, 0),
		sideEffectResult:               make(map[int32][]byte),
		scheduledChildWorkflows:        make(map[string]*childWorkflowHandle),
		completeHandler:                completeHandler,
		postEventHooks:                 []func(){},
		enableLoggingInReplay:          enableLoggingInReplay,
	}
	context.logger = logger.With(
		zapcore.Field{Key: tagWorkflowType, Type: zapcore.StringType, String: workflowInfo.WorkflowType.Name},
		zapcore.Field{Key: tagWorkflowID, Type: zapcore.StringType, String: workflowInfo.WorkflowExecution.ID},
		zapcore.Field{Key: tagRunID, Type: zapcore.StringType, String: workflowInfo.WorkflowExecution.RunID},
	).WithOptions(zap.WrapCore(wrapLogger(&context.isReplay, &context.enableLoggingInReplay)))

	return &workflowExecutionEventHandlerImpl{context, nil}
}

func (wc *workflowEnvironmentImpl) WorkflowInfo() *WorkflowInfo {
	return wc.workflowInfo
}

func (wc *workflowEnvironmentImpl) Complete(result []byte, err error) {
	wc.completeHandler(result, err)
}

func (wc *workflowEnvironmentImpl) RequestCancelWorkflow(domainName, workflowID, runID string) error {
	if domainName == "" {
		return errors.New("need a valid domain, provided empty")
	}
	if workflowID == "" {
		return errors.New("need a valid workflow ID, provided empty")
	}

	attributes := m.NewRequestCancelExternalWorkflowExecutionDecisionAttributes()
	attributes.Domain = common.StringPtr(domainName)
	attributes.WorkflowId = common.StringPtr(workflowID)
	attributes.RunId = common.StringPtr(runID)

	decision := wc.CreateNewDecision(m.DecisionType_RequestCancelExternalWorkflowExecution)
	decision.RequestCancelExternalWorkflowExecutionDecisionAttributes = attributes

	wc.executeDecisions = append(wc.executeDecisions, decision)

	childWorkflowHandle, ok := wc.scheduledChildWorkflows[workflowID]
	if ok && childWorkflowHandle.workflowExecution != nil &&
		childWorkflowHandle.workflowExecution.RunID == runID &&
		!childWorkflowHandle.waitForCancellation {
		delete(wc.scheduledChildWorkflows, workflowID)
		wc.addPostEventHooks(func() {
			childWorkflowHandle.resultCallback(nil, NewCanceledError())
		})
	}
	return nil
}

func (wc *workflowEnvironmentImpl) RegisterCancelHandler(handler func()) {
	wc.cancelHandler = handler
}

func (wc *workflowEnvironmentImpl) ExecuteChildWorkflow(
	options workflowOptions, callback resultHandler, startedHandler func(r WorkflowExecution, e error)) error {
	if options.workflowID == "" {
		options.workflowID = wc.workflowInfo.WorkflowExecution.RunID + "_" + wc.GenerateSequenceID()
	}

	attributes := m.NewStartChildWorkflowExecutionDecisionAttributes()

	attributes.Domain = options.domain
	attributes.TaskList = &m.TaskList{Name: options.taskListName}
	attributes.WorkflowId = common.StringPtr(options.workflowID)
	attributes.ExecutionStartToCloseTimeoutSeconds = options.executionStartToCloseTimeoutSeconds
	attributes.TaskStartToCloseTimeoutSeconds = options.taskStartToCloseTimeoutSeconds
	attributes.Input = options.input
	attributes.WorkflowType = workflowTypePtr(*options.workflowType)
	attributes.ChildPolicy = options.childPolicy.toThriftChildPolicyPtr()

	decision := wc.CreateNewDecision(m.DecisionType_StartChildWorkflowExecution)
	decision.StartChildWorkflowExecutionDecisionAttributes = attributes

	wc.scheduledChildWorkflows[options.workflowID] = &childWorkflowHandle{
		resultCallback:      callback,
		startedCallback:     startedHandler,
		waitForCancellation: options.waitForCancellation,
	}
	wc.executeDecisions = append(wc.executeDecisions, decision)

	return nil
}

func (wc *workflowEnvironmentImpl) RegisterSignalHandler(handler func(name string, input []byte)) {
	wc.signalHandler = handler
}

func (wc *workflowEnvironmentImpl) GetLogger() *zap.Logger {
	return wc.logger
}

func (wc *workflowEnvironmentImpl) GenerateSequenceID() string {
	return fmt.Sprintf("%d", wc.GenerateSequence())
}

func (wc *workflowEnvironmentImpl) GenerateSequence() int32 {
	result := wc.counterID
	wc.counterID++
	return result
}

func (wc *workflowEnvironmentImpl) SwapExecuteDecisions(decisions []*m.Decision) []*m.Decision {
	oldDecisions := wc.executeDecisions
	wc.executeDecisions = decisions
	return oldDecisions
}

func (wc *workflowEnvironmentImpl) CreateNewDecision(decisionType m.DecisionType) *m.Decision {
	return &m.Decision{
		DecisionType: common.DecisionTypePtr(decisionType),
	}
}

func (wc *workflowEnvironmentImpl) ExecuteActivity(parameters executeActivityParameters, callback resultHandler) *activityInfo {

	scheduleTaskAttr := &m.ScheduleActivityTaskDecisionAttributes{}
	if parameters.ActivityID == nil || *parameters.ActivityID == "" {
		scheduleTaskAttr.ActivityId = common.StringPtr(wc.GenerateSequenceID())
	} else {
		scheduleTaskAttr.ActivityId = parameters.ActivityID
	}
	scheduleTaskAttr.ActivityType = activityTypePtr(parameters.ActivityType)
	scheduleTaskAttr.TaskList = common.TaskListPtr(m.TaskList{Name: common.StringPtr(parameters.TaskListName)})
	scheduleTaskAttr.Input = parameters.Input
	scheduleTaskAttr.ScheduleToCloseTimeoutSeconds = common.Int32Ptr(parameters.ScheduleToCloseTimeoutSeconds)
	scheduleTaskAttr.StartToCloseTimeoutSeconds = common.Int32Ptr(parameters.StartToCloseTimeoutSeconds)
	scheduleTaskAttr.ScheduleToStartTimeoutSeconds = common.Int32Ptr(parameters.ScheduleToStartTimeoutSeconds)
	scheduleTaskAttr.HeartbeatTimeoutSeconds = common.Int32Ptr(parameters.HeartbeatTimeoutSeconds)

	decision := wc.CreateNewDecision(m.DecisionType_ScheduleActivityTask)
	decision.ScheduleActivityTaskDecisionAttributes = scheduleTaskAttr

	wc.executeDecisions = append(wc.executeDecisions, decision)
	wc.scheduledActivities[scheduleTaskAttr.GetActivityId()] = callback
	wc.waitForCancelRequestActivities[scheduleTaskAttr.GetActivityId()] = parameters.WaitForCancellation

	wc.logger.Debug("ExectueActivity",
		zap.String(tagActivityID, scheduleTaskAttr.GetActivityId()),
		zap.String(tagActivityType, scheduleTaskAttr.GetActivityType().GetName()))

	return &activityInfo{activityID: scheduleTaskAttr.GetActivityId()}
}

func (wc *workflowEnvironmentImpl) RequestCancelActivity(activityID string) {
	handler, ok := wc.scheduledActivities[activityID]
	if !ok {
		wc.logger.Debug("RequestCancelActivity failed because the activity ID doesn't exist.",
			zap.String(tagActivityID, activityID))
		return
	}

	requestCancelAttr := &m.RequestCancelActivityTaskDecisionAttributes{
		ActivityId: common.StringPtr(activityID)}

	decision := wc.CreateNewDecision(m.DecisionType_RequestCancelActivityTask)
	decision.RequestCancelActivityTaskDecisionAttributes = requestCancelAttr
	wc.executeDecisions = append(wc.executeDecisions, decision)

	if wait, ok := wc.waitForCancelRequestActivities[activityID]; ok && !wait {
		delete(wc.scheduledActivities, activityID)
		wc.addPostEventHooks(func() {
			handler(nil, NewCanceledError())
		})
	}
	wc.logger.Debug("RequestCancelActivity", zap.String(tagActivityID, requestCancelAttr.GetActivityId()))
}

func (wc *workflowEnvironmentImpl) SetCurrentReplayTime(replayTime time.Time) {
	wc.currentReplayTime = replayTime
}

func (wc *workflowEnvironmentImpl) Now() time.Time {
	return wc.currentReplayTime
}

func (wc *workflowEnvironmentImpl) NewTimer(d time.Duration, callback resultHandler) *timerInfo {
	if d < 0 {
		callback(nil, errors.New("Invalid delayInSeconds provided"))
		return nil
	}
	if d == 0 {
		callback(nil, nil)
		return nil
	}

	timerID := wc.GenerateSequenceID()
	startTimerAttr := &m.StartTimerDecisionAttributes{}
	startTimerAttr.TimerId = common.StringPtr(timerID)
	startTimerAttr.StartToFireTimeoutSeconds = common.Int64Ptr(int64(d.Seconds()))
	decision := wc.CreateNewDecision(m.DecisionType_StartTimer)
	decision.StartTimerDecisionAttributes = startTimerAttr

	wc.executeDecisions = append(wc.executeDecisions, decision)
	wc.scheduledTimers[startTimerAttr.GetTimerId()] = callback
	wc.logger.Debug("NewTimer",
		zap.String(tagTimerID, startTimerAttr.GetTimerId()),
		zap.Duration("Duration", d))

	return &timerInfo{timerID: timerID}
}

func (wc *workflowEnvironmentImpl) RequestCancelTimer(timerID string) {
	handler, ok := wc.scheduledTimers[timerID]
	if !ok {
		wc.logger.Debug("RequestCancelTimer failed, TimerID not exists.", zap.String(tagTimerID, timerID))
		return
	}
	cancelTimerAttr := &m.CancelTimerDecisionAttributes{TimerId: common.StringPtr(timerID)}
	decision := wc.CreateNewDecision(m.DecisionType_CancelTimer)
	decision.CancelTimerDecisionAttributes = cancelTimerAttr

	wc.executeDecisions = append(wc.executeDecisions, decision)

	wc.addPostEventHooks(func() {
		handler(nil, NewCanceledError())
	})
	delete(wc.scheduledTimers, timerID)

	wc.logger.Debug("RequestCancelTimer", zap.String(tagTimerID, timerID))
}

func (wc *workflowEnvironmentImpl) addPostEventHooks(hook func()) {
	wc.postEventHooks = append(wc.postEventHooks, hook)
}

func (wc *workflowEnvironmentImpl) SideEffect(f func() ([]byte, error), callback resultHandler) {
	sideEffectID := wc.GenerateSequence()
	var details []byte
	var result []byte
	if wc.isReplay {
		var ok bool
		result, ok = wc.sideEffectResult[sideEffectID]
		if !ok {
			keys := make([]int32, 0, len(wc.sideEffectResult))
			for k := range wc.sideEffectResult {
				keys = append(keys, k)
			}
			panic(fmt.Sprintf("No cached result found for side effectID=%v. KnownSideEffects=%v",
				sideEffectID, keys))
		}
		wc.logger.Debug("SideEffect returning already caclulated result.",
			zap.Int32(tagSideEffectID, sideEffectID))
		details = result
	} else {
		var err error
		result, err = f()
		if err != nil {
			callback(result, err)
			return
		}
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(sideEffectID); err != nil {
			callback(nil, fmt.Errorf("failure encoding sideEffectID: %v", err))
			return
		}
		if err := enc.Encode(result); err != nil {
			callback(nil, fmt.Errorf("failure encoding side effect result: %v", err))
			return
		}
		details = buf.Bytes()
	}
	recordMarker := &m.RecordMarkerDecisionAttributes{
		MarkerName: common.StringPtr(sideEffectMarkerName),
		Details:    details, // Keep
	}
	decision := wc.CreateNewDecision(m.DecisionType_RecordMarker)
	decision.RecordMarkerDecisionAttributes = recordMarker
	wc.executeDecisions = append(wc.executeDecisions, decision)

	callback(result, nil)
	wc.logger.Debug("SideEffect Marker added", zap.Int32(tagSideEffectID, sideEffectID))
}

func (weh *workflowExecutionEventHandlerImpl) ProcessEvent(
	event *m.HistoryEvent,
	isReplay bool,
	isLast bool,
) ([]*m.Decision, error) {
	if event == nil {
		return nil, errors.New("nil event provided")
	}

	weh.isReplay = isReplay
	if enableVerboseLogging {
		weh.logger.Debug("ProcessEvent",
			zap.Int64(tagEventID, event.GetEventId()),
			zap.String(tagEventType, event.GetEventType().String()))
	}

	switch event.GetEventType() {
	case m.EventType_WorkflowExecutionStarted:
		err := weh.handleWorkflowExecutionStarted(event.WorkflowExecutionStartedEventAttributes)
		if err != nil {
			return nil, err
		}

	case m.EventType_WorkflowExecutionCompleted:
		// No Operation
	case m.EventType_WorkflowExecutionFailed:
		// No Operation
	case m.EventType_WorkflowExecutionTimedOut:
		// No Operation
	case m.EventType_DecisionTaskScheduled:
		// No Operation
	case m.EventType_DecisionTaskStarted:
		weh.workflowDefinition.OnDecisionTaskStarted()
	case m.EventType_DecisionTaskTimedOut:
		// No Operation
	case m.EventType_DecisionTaskFailed:
		// No Operation
	case m.EventType_DecisionTaskCompleted:
		// No Operation
	case m.EventType_ActivityTaskScheduled:
		attributes := event.ActivityTaskScheduledEventAttributes
		weh.scheduledEventIDToActivityID[event.GetEventId()] = attributes.GetActivityId()

	case m.EventType_ActivityTaskStarted:
	// No Operation
	case m.EventType_ActivityTaskCompleted:
		err := weh.handleActivityTaskCompleted(event.ActivityTaskCompletedEventAttributes)
		if err != nil {
			return nil, err
		}

	case m.EventType_ActivityTaskFailed:
		err := weh.handleActivityTaskFailed(event.ActivityTaskFailedEventAttributes)
		if err != nil {
			return nil, err
		}

	case m.EventType_ActivityTaskTimedOut:
		err := weh.handleActivityTaskTimedOut(event.ActivityTaskTimedOutEventAttributes)
		if err != nil {
			return nil, err
		}

	case m.EventType_ActivityTaskCancelRequested:
		// No Operation.
	case m.EventType_RequestCancelActivityTaskFailed:
		// No operation.

	case m.EventType_ActivityTaskCanceled:
		err := weh.handleActivityTaskCanceled(event.ActivityTaskCanceledEventAttributes)
		if err != nil {
			return nil, err
		}

	case m.EventType_TimerStarted:
		// No Operation
	case m.EventType_TimerFired:
		err := weh.handleTimerFired(event.TimerFiredEventAttributes)
		if err != nil {
			return nil, err
		}

	case m.EventType_TimerCanceled:
		// No Operation:
		// As we always cancel the timer immediately if asked, we don't wait for it.
	case m.EventType_CancelTimerFailed:
		// No Operation.

	case m.EventType_WorkflowExecutionCancelRequested:
		weh.handleWorkflowExecutionCancelRequested(event.WorkflowExecutionCancelRequestedEventAttributes)

	case m.EventType_WorkflowExecutionCanceled:
		// No Operation.

	case m.EventType_RequestCancelExternalWorkflowExecutionInitiated:
		// No Operation.
	case m.EventType_ExternalWorkflowExecutionCancelRequested:
		// No Operation.
	case m.EventType_RequestCancelExternalWorkflowExecutionFailed:
		// No Operation.
	case m.EventType_WorkflowExecutionContinuedAsNew:
		// No Operation.

	case m.EventType_WorkflowExecutionSignaled:
		weh.handleWorkflowExecutionSignaled(event.WorkflowExecutionSignaledEventAttributes)

	case m.EventType_MarkerRecorded:
		err := weh.handleMarkerRecorded(event.GetEventId(), event.MarkerRecordedEventAttributes)
		if err != nil {
			return nil, err
		}
	case m.EventType_StartChildWorkflowExecutionInitiated:
		// No Operation.
	case m.EventType_StartChildWorkflowExecutionFailed:
		err := weh.handleChildWorkflowExecutionFailed(event.ChildWorkflowExecutionFailedEventAttributes)
		if err != nil {
			return nil, err
		}
	case m.EventType_ChildWorkflowExecutionStarted:
		err := weh.handleChildWorkflowExecutionStarted(event.ChildWorkflowExecutionStartedEventAttributes)
		if err != nil {
			return nil, err
		}
	case m.EventType_ChildWorkflowExecutionCompleted:
		err := weh.handleChildWorkflowExecutionCompleted(event.ChildWorkflowExecutionCompletedEventAttributes)
		if err != nil {
			return nil, err
		}
	case m.EventType_ChildWorkflowExecutionFailed:
		err := weh.handleChildWorkflowExecutionFailed(event.ChildWorkflowExecutionFailedEventAttributes)
		if err != nil {
			return nil, err
		}
	case m.EventType_ChildWorkflowExecutionCanceled:
		err := weh.handleChildWorkflowExecutionCanceled(event.ChildWorkflowExecutionCanceledEventAttributes)
		if err != nil {
			return nil, err
		}
	case m.EventType_ChildWorkflowExecutionTimedOut:
		err := weh.handleChildWorkflowExecutionTimedOut(event.ChildWorkflowExecutionTimedOutEventAttributes)
		if err != nil {
			return nil, err
		}
	case m.EventType_ChildWorkflowExecutionTerminated:
		err := weh.handleChildWorkflowExecutionTerminated(event.ChildWorkflowExecutionTerminatedEventAttributes)
		if err != nil {
			return nil, err
		}
	default:
		weh.logger.Error("unknown event type",
			zap.Int64(tagEventID, event.GetEventId()),
			zap.String(tagEventType, event.GetEventType().String()))
		// Do not fail to be forward compatible with new events
	}

	// Invoke any pending post event hooks that have been added while processing the event.
	if len(weh.postEventHooks) > 0 {
		for _, c := range weh.postEventHooks {
			c()
		}
		weh.postEventHooks = []func(){}
	}

	// When replaying histories to get stack trace or current state the last event might be not
	// decision started. So always call OnDecisionTaskStarted on the last event.
	// Don't call for EventType_DecisionTaskStarted as it was already called when handling it.
	if isLast && event.GetEventType() != m.EventType_DecisionTaskStarted {
		weh.workflowDefinition.OnDecisionTaskStarted()
	}

	return weh.SwapExecuteDecisions([]*m.Decision{}), nil
}

func (weh *workflowExecutionEventHandlerImpl) StackTrace() string {
	return weh.workflowDefinition.StackTrace()
}

func (weh *workflowExecutionEventHandlerImpl) Close() {
	if weh.workflowDefinition != nil {
		weh.workflowDefinition.Close()
	}
}

func (weh *workflowExecutionEventHandlerImpl) handleWorkflowExecutionStarted(
	attributes *m.WorkflowExecutionStartedEventAttributes) (err error) {
	weh.workflowDefinition, err = weh.workflowDefinitionFactory(weh.workflowInfo.WorkflowType)
	if err != nil {
		return err
	}

	// Invoke the workflow.
	weh.workflowDefinition.Execute(weh, attributes.Input)
	return nil
}

func (weh *workflowExecutionEventHandlerImpl) handleActivityTaskCompleted(
	attributes *m.ActivityTaskCompletedEventAttributes) error {

	activityID, ok := weh.scheduledEventIDToActivityID[attributes.GetScheduledEventId()]
	if !ok {
		return fmt.Errorf("unable to find activity ID for the event: %v", attributes)
	}
	handler, ok := weh.scheduledActivities[activityID]
	if !ok {
		if wait, exist := weh.waitForCancelRequestActivities[activityID]; exist && !wait {
			return nil
		}
		return fmt.Errorf("unable to find callback handler for the event: %v, with activity ID: %v, ok: %v",
			attributes, activityID, ok)
	}

	// Clear this so we don't have a recursive call that while executing might call the cancel one.
	delete(weh.scheduledActivities, activityID)

	// Invoke the callback
	handler(attributes.GetResult_(), nil)

	return nil
}

func (weh *workflowExecutionEventHandlerImpl) handleActivityTaskFailed(
	attributes *m.ActivityTaskFailedEventAttributes) error {

	activityID, ok := weh.scheduledEventIDToActivityID[attributes.GetScheduledEventId()]
	if !ok {
		return fmt.Errorf("unable to find activity ID for the event: %v", attributes)
	}
	handler, ok := weh.scheduledActivities[activityID]
	if !ok {
		if wait, exist := weh.waitForCancelRequestActivities[activityID]; exist && !wait {
			return nil
		}
		return fmt.Errorf("unable to find callback handler for the event: %v, with activity ID: %v, ok: %v",
			attributes, activityID, ok)
	}

	// Clear this so we don't have a recursive call that while executing might call the cancel one.
	delete(weh.scheduledActivities, activityID)

	err := NewErrorWithDetails(*attributes.Reason, attributes.Details)
	// Invoke the callback
	handler(nil, err)
	return nil
}

func (weh *workflowExecutionEventHandlerImpl) handleActivityTaskTimedOut(
	attributes *m.ActivityTaskTimedOutEventAttributes) error {

	activityID, ok := weh.scheduledEventIDToActivityID[attributes.GetScheduledEventId()]
	if !ok {
		return fmt.Errorf("unable to find activity ID for the event: %v", attributes)
	}
	handler, ok := weh.scheduledActivities[activityID]
	if !ok {
		if wait, exist := weh.waitForCancelRequestActivities[activityID]; exist && !wait {
			return nil
		}
		return fmt.Errorf("unable to find callback handler for the event: %v, with activity ID: %v, ok: %v",
			attributes, activityID, ok)
	}

	// Clear this so we don't have a recursive call that while executing might call the cancel one.
	delete(weh.scheduledActivities, activityID)

	var err error
	tt := attributes.GetTimeoutType()
	if tt == m.TimeoutType_HEARTBEAT {
		err = NewHeartbeatTimeoutError(attributes.GetDetails())
	} else {
		err = NewTimeoutError(attributes.GetTimeoutType())
	}
	// Invoke the callback
	handler(nil, err)
	return nil
}

func (weh *workflowExecutionEventHandlerImpl) handleActivityTaskCanceled(
	attributes *m.ActivityTaskCanceledEventAttributes) error {

	activityID, ok := weh.scheduledEventIDToActivityID[attributes.GetScheduledEventId()]
	if !ok {
		return fmt.Errorf("unable to find activity ID for the event: %v", attributes)
	}
	handler, ok := weh.scheduledActivities[activityID]
	if !ok {
		if wait, exist := weh.waitForCancelRequestActivities[activityID]; exist && !wait {
			return nil
		}
		return fmt.Errorf("unable to find callback handler for the event: %v, ok: %v", attributes, ok)
	}

	// Clear this so we don't have a recursive call that while executing might call the cancel one.
	delete(weh.scheduledActivities, activityID)

	err := NewCanceledError(attributes.GetDetails())
	// Invoke the callback
	handler(nil, err)
	return nil
}

func (weh *workflowExecutionEventHandlerImpl) handleTimerFired(
	attributes *m.TimerFiredEventAttributes) error {
	handler, ok := weh.scheduledTimers[attributes.GetTimerId()]
	if !ok {
		weh.logger.Debug("Unable to find the timer callback when it is fired.", zap.String(tagTimerID, attributes.GetTimerId()))
		return nil
	}

	// Clear this so we don't have a recursive call that while invoking might call the cancel one.
	delete(weh.scheduledTimers, attributes.GetTimerId())

	// Invoke the callback
	handler(nil, nil)
	return nil
}

func (weh *workflowExecutionEventHandlerImpl) handleWorkflowExecutionCancelRequested(
	attributes *m.WorkflowExecutionCancelRequestedEventAttributes) {
	weh.cancelHandler()
}

// Currently handles only side effect markers
func (weh *workflowExecutionEventHandlerImpl) handleMarkerRecorded(
	eventID int64,
	attributes *m.MarkerRecordedEventAttributes,
) error {
	if attributes.GetMarkerName() != sideEffectMarkerName {
		return fmt.Errorf("unknown marker name \"%v\" for eventID \"%v\"",
			attributes.GetMarkerName(), eventID)
	}
	dec := gob.NewDecoder(bytes.NewBuffer(attributes.GetDetails()))
	var sideEffectID int32

	if err := dec.Decode(&sideEffectID); err != nil {
		return fmt.Errorf("failure decodeing sideEffectID: %v", err)
	}
	var result []byte
	if err := dec.Decode(&result); err != nil {
		return fmt.Errorf("failure decoding side effect result: %v", err)
	}
	weh.sideEffectResult[sideEffectID] = result
	return nil
}

func (weh *workflowExecutionEventHandlerImpl) handleWorkflowExecutionSignaled(
	attributes *m.WorkflowExecutionSignaledEventAttributes) {
	weh.signalHandler(attributes.GetSignalName(), attributes.GetInput())
}

func (weh *workflowExecutionEventHandlerImpl) handleStartChildWorkflowExecutionInitiated(
	attributes *m.StartChildWorkflowExecutionInitiatedEventAttributes) error {
	return nil
}

func (weh *workflowExecutionEventHandlerImpl) handleStartChildWorkflowExecutionFailed(
	attributes *m.StartChildWorkflowExecutionFailedEventAttributes) error {
	childWorkflowID := attributes.GetWorkflowId()
	childWorkflowHandle, ok := weh.scheduledChildWorkflows[childWorkflowID]
	if !ok {
		return fmt.Errorf("unable to find child workflow callback: %v", attributes)
	}
	delete(weh.scheduledChildWorkflows, childWorkflowID)
	err := fmt.Errorf("ChildWorkflowFailed: %v", attributes.GetCause())
	childWorkflowHandle.resultCallback(nil, err)

	return nil
}

func (weh *workflowExecutionEventHandlerImpl) handleChildWorkflowExecutionStarted(
	attributes *m.ChildWorkflowExecutionStartedEventAttributes) error {
	childWorkflowID := attributes.WorkflowExecution.GetWorkflowId()
	childWorkflowHandle, ok := weh.scheduledChildWorkflows[childWorkflowID]
	if !ok {
		return fmt.Errorf("unable to find child workflow callback: %v", attributes)
	}

	childWorkflowExecution := WorkflowExecution{
		ID:    childWorkflowID,
		RunID: attributes.WorkflowExecution.GetRunId(),
	}

	childWorkflowHandle.workflowExecution = &childWorkflowExecution
	childWorkflowHandle.startedCallback(childWorkflowExecution, nil)

	return nil
}

func (weh *workflowExecutionEventHandlerImpl) handleChildWorkflowExecutionCompleted(
	attributes *m.ChildWorkflowExecutionCompletedEventAttributes) error {
	childWorkflowID := attributes.WorkflowExecution.GetWorkflowId()
	childWorkflowHandle, ok := weh.scheduledChildWorkflows[childWorkflowID]
	if !ok {
		return fmt.Errorf("unable to find child workflow callback: %v", attributes)
	}
	delete(weh.scheduledChildWorkflows, childWorkflowID)

	childWorkflowHandle.resultCallback(attributes.GetResult_(), nil)

	return nil
}

func (weh *workflowExecutionEventHandlerImpl) handleChildWorkflowExecutionFailed(
	attributes *m.ChildWorkflowExecutionFailedEventAttributes) error {

	childWorkflowID := attributes.WorkflowExecution.GetWorkflowId()
	childWorkflowHandle, ok := weh.scheduledChildWorkflows[childWorkflowID]
	if !ok {
		return fmt.Errorf("unable to find child workflow callback: %v", attributes)
	}
	delete(weh.scheduledChildWorkflows, childWorkflowID)

	err := NewErrorWithDetails(attributes.GetReason(), attributes.GetDetails())
	childWorkflowHandle.resultCallback(nil, err)

	return nil
}

func (weh *workflowExecutionEventHandlerImpl) handleChildWorkflowExecutionCanceled(
	attributes *m.ChildWorkflowExecutionCanceledEventAttributes) error {
	childWorkflowID := attributes.WorkflowExecution.GetWorkflowId()
	childWorkflowHandle, ok := weh.scheduledChildWorkflows[childWorkflowID]
	if !ok {
		// this could happen if waitForCancellation is false
		return nil
	}
	delete(weh.scheduledChildWorkflows, childWorkflowID)
	err := NewCanceledError(attributes.GetDetails())

	childWorkflowHandle.resultCallback(nil, err)
	return nil
}

func (weh *workflowExecutionEventHandlerImpl) handleChildWorkflowExecutionTimedOut(
	attributes *m.ChildWorkflowExecutionTimedOutEventAttributes) error {
	childWorkflowID := attributes.WorkflowExecution.GetWorkflowId()
	childWorkflowHandle, ok := weh.scheduledChildWorkflows[childWorkflowID]
	if !ok {
		return fmt.Errorf("unable to find child workflow callback: %v", attributes)
	}
	delete(weh.scheduledChildWorkflows, childWorkflowID)
	err := NewTimeoutError(attributes.GetTimeoutType())

	childWorkflowHandle.resultCallback(nil, err)

	return nil
}

func (weh *workflowExecutionEventHandlerImpl) handleChildWorkflowExecutionTerminated(
	attributes *m.ChildWorkflowExecutionTerminatedEventAttributes) error {
	childWorkflowID := attributes.WorkflowExecution.GetWorkflowId()
	childWorkflowHandle, ok := weh.scheduledChildWorkflows[childWorkflowID]
	if !ok {
		return fmt.Errorf("unable to find child workflow callback: %v", attributes)
	}
	delete(weh.scheduledChildWorkflows, childWorkflowID)
	err := errors.New("terminated")
	childWorkflowHandle.resultCallback(nil, err)

	return nil
}
