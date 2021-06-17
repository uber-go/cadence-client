// Copyright (c) 2017-2020 Uber Technologies Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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

// All code in this file is private to the package.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/opentracing/opentracing-go"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
	"go.uber.org/cadence/v2/internal/api"
	"go.uber.org/cadence/v2/internal/common"
	"go.uber.org/cadence/v2/internal/common/backoff"
	"go.uber.org/cadence/v2/internal/common/cache"
	"go.uber.org/cadence/v2/internal/common/metrics"
	"go.uber.org/cadence/v2/internal/common/util"
	"go.uber.org/zap"
)

const (
	defaultHeartBeatIntervalInSec = 10 * 60

	defaultStickyCacheSize = 10000

	noRetryBackoff = time.Duration(-1)
)

type (
	// workflowExecutionEventHandler process a single event.
	workflowExecutionEventHandler interface {
		// Process a single event and return the assosciated decisions.
		// Return List of decisions made, any error.
		ProcessEvent(event *apiv1.HistoryEvent, isReplay bool, isLast bool) error
		// ProcessQuery process a query request.
		ProcessQuery(queryType string, queryArgs []byte) ([]byte, error)
		StackTrace() string
		// Close for cleaning up resources on this event handler
		Close()
	}

	// workflowTask wraps a decision task.
	workflowTask struct {
		task            *apiv1.PollForDecisionTaskResponse
		historyIterator HistoryIterator
		doneCh          chan struct{}
		laResultCh      chan *localActivityResult
	}

	// activityTask wraps a activity task.
	activityTask struct {
		task          *apiv1.PollForActivityTaskResponse
		pollStartTime time.Time
	}

	// resetStickinessTask wraps a ResetStickyTaskListRequest.
	resetStickinessTask struct {
		task *apiv1.ResetStickyTaskListRequest
	}

	// workflowExecutionContextImpl is the cached workflow state for sticky execution
	workflowExecutionContextImpl struct {
		mutex             sync.Mutex
		workflowStartTime time.Time
		workflowInfo      *WorkflowInfo
		wth               *workflowTaskHandlerImpl

		// eventHandler is changed to a atomic.Value as a temporally bug fix for local activity
		// retry issue (github issue #915). Therefore, when accessing/modifying this field, the
		// mutex should still be held.
		eventHandler atomic.Value

		isWorkflowCompleted bool
		result              []byte
		err                 error

		previousStartedEventID int64

		newDecisions        []*apiv1.Decision
		currentDecisionTask *apiv1.PollForDecisionTaskResponse
		laTunnel            *localActivityTunnel
		decisionStartTime   time.Time
	}

	// workflowTaskHandlerImpl is the implementation of WorkflowTaskHandler
	workflowTaskHandlerImpl struct {
		domain                         string
		metricsScope                   *metrics.TaggedScope
		ppMgr                          pressurePointMgr
		logger                         *zap.Logger
		identity                       string
		enableLoggingInReplay          bool
		disableStickyExecution         bool
		registry                       *registry
		laTunnel                       *localActivityTunnel
		nonDeterministicWorkflowPolicy NonDeterministicWorkflowPolicy
		dataConverter                  DataConverter
		contextPropagators             []ContextPropagator
		tracer                         opentracing.Tracer
		workflowInterceptors           []WorkflowInterceptorFactory
	}

	activityProvider func(name string) activity

	// activityTaskHandlerImpl is the implementation of ActivityTaskHandler
	activityTaskHandlerImpl struct {
		taskListName       string
		identity           string
		service            api.Interface
		metricsScope       *metrics.TaggedScope
		logger             *zap.Logger
		userContext        context.Context
		registry           *registry
		activityProvider   activityProvider
		dataConverter      DataConverter
		workerStopCh       <-chan struct{}
		contextPropagators []ContextPropagator
		tracer             opentracing.Tracer
		featureFlags       FeatureFlags
	}

	// history wrapper method to help information about events.
	history struct {
		workflowTask   *workflowTask
		eventsHandler  *workflowExecutionEventHandlerImpl
		loadedEvents   []*apiv1.HistoryEvent
		currentIndex   int
		nextEventID    int64 // next expected eventID for sanity
		lastEventID    int64 // last expected eventID, zero indicates read until end of stream
		next           []*apiv1.HistoryEvent
		binaryChecksum *string
	}

	decisionHeartbeatError struct {
		Message string
	}
)

func newHistory(task *workflowTask, eventsHandler *workflowExecutionEventHandlerImpl) *history {
	result := &history{
		workflowTask:  task,
		eventsHandler: eventsHandler,
		loadedEvents:  task.task.History.Events,
		currentIndex:  0,
		// don't set lastEventID to task.GetNextEventId()
		// as for sticky query, the history in workflow task will be empty
		// and query will be run based on existing workflow state.
		// so the sanity check in verifyAllEventsProcessed will fail
		lastEventID: task.task.GetStartedEventId(),
	}
	if len(result.loadedEvents) > 0 {
		result.nextEventID = result.loadedEvents[0].GetEventId()
	}
	return result
}

func (e decisionHeartbeatError) Error() string {
	return e.Message
}

// Get workflow start event.
func (eh *history) GetWorkflowStartedEvent() (*apiv1.HistoryEvent, error) {
	events := eh.workflowTask.task.History.Events
	if len(events) == 0 || events[0].GetWorkflowExecutionStartedEventAttributes() == nil{
		return nil, errors.New("unable to find WorkflowExecutionStartedEventAttributes in the history")
	}
	return events[0], nil
}

func (eh *history) IsReplayEvent(event *apiv1.HistoryEvent) bool {
	return event.GetEventId() <= eh.workflowTask.task.GetPreviousStartedEventId().GetValue() || isDecisionEvent(event)
}

func (eh *history) IsNextDecisionFailed() (isFailed bool, binaryChecksum *string, err error) {

	nextIndex := eh.currentIndex + 1
	if nextIndex >= len(eh.loadedEvents) && eh.hasMoreEvents() { // current page ends and there is more pages
		if err := eh.loadMoreEvents(); err != nil {
			return false, nil, err
		}
	}

	if nextIndex < len(eh.loadedEvents) {
		nextEvent := eh.loadedEvents[nextIndex]
		isFailed := false
		var binaryChecksum string
		switch attr := nextEvent.Attributes.(type) {
		case *apiv1.HistoryEvent_DecisionTaskTimedOutEventAttributes:
		case *apiv1.HistoryEvent_DecisionTaskFailedEventAttributes:
			isFailed = true
		case *apiv1.HistoryEvent_DecisionTaskCompletedEventAttributes:
			binaryChecksum = attr.DecisionTaskCompletedEventAttributes.BinaryChecksum
		}
		return isFailed, &binaryChecksum, nil
	}
	return false, nil, nil
}

func (eh *history) loadMoreEvents() error {
	historyPage, err := eh.getMoreEvents()
	if err != nil {
		return err
	}
	eh.loadedEvents = append(eh.loadedEvents, historyPage.Events...)
	if eh.nextEventID == 0 && len(eh.loadedEvents) > 0 {
		eh.nextEventID = eh.loadedEvents[0].GetEventId()
	}
	return nil
}

func isDecisionEventAttributes(attr interface{}) bool {
	switch attr.(type) {
	case *apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes,
		*apiv1.HistoryEvent_WorkflowExecutionFailedEventAttributes,
		*apiv1.HistoryEvent_WorkflowExecutionCanceledEventAttributes,
		*apiv1.HistoryEvent_WorkflowExecutionContinuedAsNewEventAttributes,
		*apiv1.HistoryEvent_ActivityTaskScheduledEventAttributes,
		*apiv1.HistoryEvent_ActivityTaskCancelRequestedEventAttributes,
		*apiv1.HistoryEvent_TimerStartedEventAttributes,
		*apiv1.HistoryEvent_TimerCanceledEventAttributes,
		*apiv1.HistoryEvent_CancelTimerFailedEventAttributes,
		*apiv1.HistoryEvent_MarkerRecordedEventAttributes,
		*apiv1.HistoryEvent_StartChildWorkflowExecutionInitiatedEventAttributes,
		*apiv1.HistoryEvent_RequestCancelExternalWorkflowExecutionInitiatedEventAttributes,
		*apiv1.HistoryEvent_SignalExternalWorkflowExecutionInitiatedEventAttributes,
		*apiv1.HistoryEvent_UpsertWorkflowSearchAttributesEventAttributes:
		return true
	default:
		return false
	}
}

func isDecisionEvent(event *apiv1.HistoryEvent) bool {
	return isDecisionEventAttributes(event.Attributes)
}

// NextDecisionEvents returns events that there processed as new by the next decision.
// TODO(maxim): Refactor to return a struct instead of multiple parameters
func (eh *history) NextDecisionEvents() (result []*apiv1.HistoryEvent, markers []*apiv1.HistoryEvent, binaryChecksum *string, err error) {
	if eh.next == nil {
		eh.next, _, err = eh.nextDecisionEvents()
		if err != nil {
			return result, markers, eh.binaryChecksum, err
		}
	}

	result = eh.next
	checksum := eh.binaryChecksum
	if len(result) > 0 {
		eh.next, markers, err = eh.nextDecisionEvents()
	}
	return result, markers, checksum, err
}

func (eh *history) HasNextDecisionEvents() bool {
	return len(eh.next) != 0 || eh.currentIndex != len(eh.loadedEvents) || eh.hasMoreEvents()
}

func (eh *history) hasMoreEvents() bool {
	historyIterator := eh.workflowTask.historyIterator
	return historyIterator != nil && historyIterator.HasNextPage()
}

func (eh *history) getMoreEvents() (*apiv1.History, error) {
	return eh.workflowTask.historyIterator.GetNextPage()
}

func (eh *history) verifyAllEventsProcessed() error {
	if eh.lastEventID > 0 && eh.nextEventID <= eh.lastEventID {
		return fmt.Errorf(
			"history_events: premature end of stream, expectedLastEventID=%v but no more events after eventID=%v",
			eh.lastEventID,
			eh.nextEventID-1)
	}
	if eh.lastEventID > 0 && eh.nextEventID != (eh.lastEventID+1) {
		eh.eventsHandler.logger.Warn(
			"history_events: processed events past the expected lastEventID",
			zap.Int64("expectedLastEventID", eh.lastEventID),
			zap.Int64("processedLastEventID", eh.nextEventID-1))
	}
	return nil
}

func (eh *history) nextDecisionEvents() (nextEvents []*apiv1.HistoryEvent, markers []*apiv1.HistoryEvent, err error) {
	if eh.currentIndex == len(eh.loadedEvents) && !eh.hasMoreEvents() {
		if err := eh.verifyAllEventsProcessed(); err != nil {
			return nil, nil, err
		}
		return []*apiv1.HistoryEvent{}, []*apiv1.HistoryEvent{}, nil
	}

	// Process events

OrderEvents:
	for {
		// load more history events if needed
		for eh.currentIndex == len(eh.loadedEvents) {
			if !eh.hasMoreEvents() {
				if err = eh.verifyAllEventsProcessed(); err != nil {
					return
				}
				break OrderEvents
			}
			if err = eh.loadMoreEvents(); err != nil {
				return
			}
		}

		event := eh.loadedEvents[eh.currentIndex]
		eventID := event.GetEventId()
		if eventID != eh.nextEventID {
			err = fmt.Errorf(
				"missing history events, expectedNextEventID=%v but receivedNextEventID=%v",
				eh.nextEventID, eventID)
			return
		}

		eh.nextEventID++

		switch event.Attributes.(type) {
		case *apiv1.HistoryEvent_DecisionTaskStartedEventAttributes:
			isFailed, binaryChecksum, err1 := eh.IsNextDecisionFailed()
			if err1 != nil {
				err = err1
				return
			}
			if !isFailed {
				eh.binaryChecksum = binaryChecksum
				eh.currentIndex++
				nextEvents = append(nextEvents, event)
				break OrderEvents
			}
		case *apiv1.HistoryEvent_DecisionTaskScheduledEventAttributes:
		case *apiv1.HistoryEvent_DecisionTaskTimedOutEventAttributes:
		case *apiv1.HistoryEvent_DecisionTaskFailedEventAttributes:
			// Skip
		default:
			if isPreloadMarkerEvent(event) {
				markers = append(markers, event)
			}
			nextEvents = append(nextEvents, event)
		}
		eh.currentIndex++
	}

	// shrink loaded events so it can be GCed
	eh.loadedEvents = eh.loadedEvents[eh.currentIndex:]
	eh.currentIndex = 0

	return nextEvents, markers, nil
}

func isPreloadMarkerEvent(event *apiv1.HistoryEvent) bool {
	return event.GetMarkerRecordedEventAttributes() != nil
}

// newWorkflowTaskHandler returns an implementation of workflow task handler.
func newWorkflowTaskHandler(
	domain string,
	params workerExecutionParameters,
	ppMgr pressurePointMgr,
	registry *registry,
) WorkflowTaskHandler {
	ensureRequiredParams(&params)
	return &workflowTaskHandlerImpl{
		domain:                         domain,
		logger:                         params.Logger,
		ppMgr:                          ppMgr,
		metricsScope:                   metrics.NewTaggedScope(params.MetricsScope),
		identity:                       params.Identity,
		enableLoggingInReplay:          params.EnableLoggingInReplay,
		disableStickyExecution:         params.DisableStickyExecution,
		registry:                       registry,
		nonDeterministicWorkflowPolicy: params.NonDeterministicWorkflowPolicy,
		dataConverter:                  params.DataConverter,
		contextPropagators:             params.ContextPropagators,
		tracer:                         params.Tracer,
		workflowInterceptors:           params.WorkflowInterceptors,
	}
}

// TODO: need a better eviction policy based on memory usage
var workflowCache cache.Cache
var stickyCacheSize = defaultStickyCacheSize
var initCacheOnce sync.Once
var stickyCacheLock sync.Mutex

// SetStickyWorkflowCacheSize sets the cache size for sticky workflow cache. Sticky workflow execution is the affinity
// between decision tasks of a specific workflow execution to a specific worker. The affinity is set if sticky execution
// is enabled via Worker.Options (It is enabled by default unless disabled explicitly). The benefit of sticky execution
// is that workflow does not have to reconstruct the state by replaying from beginning of history events. But the cost
// is it consumes more memory as it rely on caching workflow execution's running state on the worker. The cache is shared
// between workers running within same process. This must be called before any worker is started. If not called, the
// default size of 10K (might change in future) will be used.
func SetStickyWorkflowCacheSize(cacheSize int) {
	stickyCacheLock.Lock()
	defer stickyCacheLock.Unlock()
	if workflowCache != nil {
		panic("cache already created, please set cache size before worker starts.")
	}
	stickyCacheSize = cacheSize
}

func getWorkflowCache() cache.Cache {
	initCacheOnce.Do(func() {
		stickyCacheLock.Lock()
		defer stickyCacheLock.Unlock()
		workflowCache = cache.New(stickyCacheSize, &cache.Options{
			RemovedFunc: func(cachedEntity interface{}) {
				wc := cachedEntity.(*workflowExecutionContextImpl)
				wc.onEviction()
			},
		})
	})
	return workflowCache
}

func getWorkflowContext(runID string) *workflowExecutionContextImpl {
	o := getWorkflowCache().Get(runID)
	if o == nil {
		return nil
	}
	wc := o.(*workflowExecutionContextImpl)
	return wc
}

func putWorkflowContext(runID string, wc *workflowExecutionContextImpl) (*workflowExecutionContextImpl, error) {
	existing, err := getWorkflowCache().PutIfNotExist(runID, wc)
	if err != nil {
		return nil, err
	}
	return existing.(*workflowExecutionContextImpl), nil
}

func removeWorkflowContext(runID string) {
	getWorkflowCache().Delete(runID)
}

func newWorkflowExecutionContext(
	startTime time.Time,
	workflowInfo *WorkflowInfo,
	taskHandler *workflowTaskHandlerImpl,
) *workflowExecutionContextImpl {
	workflowContext := &workflowExecutionContextImpl{
		workflowStartTime: startTime,
		workflowInfo:      workflowInfo,
		wth:               taskHandler,
	}
	workflowContext.createEventHandler()
	return workflowContext
}

func (w *workflowExecutionContextImpl) Lock() {
	w.mutex.Lock()
}

func (w *workflowExecutionContextImpl) Unlock(err error) {
	if err != nil || w.err != nil || w.isWorkflowCompleted || (w.wth.disableStickyExecution && !w.hasPendingLocalActivityWork()) {
		// TODO: in case of closed, it assumes the close decision always succeed. need server side change to return
		// error to indicate the close failure case. This should be rear case. For now, always remove the cache, and
		// if the close decision failed, the next decision will have to rebuild the state.
		if getWorkflowCache().Exist(w.workflowInfo.WorkflowExecution.RunID) {
			removeWorkflowContext(w.workflowInfo.WorkflowExecution.RunID)
		} else {
			// sticky is disabled, manually clear the workflow state.
			w.clearState()
		}
	}

	w.mutex.Unlock()
}

func (w *workflowExecutionContextImpl) getEventHandler() *workflowExecutionEventHandlerImpl {
	eventHandler := w.eventHandler.Load()
	if eventHandler == nil {
		return nil
	}
	eventHandlerImpl, ok := eventHandler.(*workflowExecutionEventHandlerImpl)
	if !ok {
		panic("unknown type for workflow execution event handler")
	}
	return eventHandlerImpl
}

func (w *workflowExecutionContextImpl) completeWorkflow(result []byte, err error) {
	w.isWorkflowCompleted = true
	w.result = result
	w.err = err
}

func (w *workflowExecutionContextImpl) shouldResetStickyOnEviction() bool {
	// Not all evictions from the cache warrant a call to the server
	// to reset stickiness.
	// Cases when this is redundant or unnecessary include
	// when an error was encountered during execution
	// or workflow simply completed successfully.
	return w.err == nil && !w.isWorkflowCompleted
}

func (w *workflowExecutionContextImpl) onEviction() {
	// onEviction is run by LRU cache's removeFunc in separate goroutinue
	w.mutex.Lock()

	// Queue a ResetStickiness request *BEFORE* calling clearState
	// because once destroyed, no sensible information
	// may be ascertained about the execution context's state,
	// nor should any of its methods be invoked.
	if w.shouldResetStickyOnEviction() {
		w.queueResetStickinessTask()
	}

	w.clearState()
	w.mutex.Unlock()
}

func (w *workflowExecutionContextImpl) IsDestroyed() bool {
	return w.getEventHandler() == nil
}

func (w *workflowExecutionContextImpl) queueResetStickinessTask() {
	var task resetStickinessTask
	task.task = &apiv1.ResetStickyTaskListRequest{
		Domain: w.workflowInfo.Domain,
		WorkflowExecution: &apiv1.WorkflowExecution{
			WorkflowId: w.workflowInfo.WorkflowExecution.ID,
			RunId:      w.workflowInfo.WorkflowExecution.RunID,
		},
	}
	// w.laTunnel could be nil for worker.ReplayHistory() because there is no worker started, in that case we don't
	// care about resetStickinessTask.
	if w.laTunnel != nil && w.laTunnel.resultCh != nil {
		w.laTunnel.resultCh <- &task
	}
}

func (w *workflowExecutionContextImpl) clearState() {
	w.clearCurrentTask()
	w.isWorkflowCompleted = false
	w.result = nil
	w.err = nil
	w.previousStartedEventID = 0
	w.newDecisions = nil

	eventHandler := w.getEventHandler()
	if eventHandler != nil {
		// Set isReplay to true to prevent user code in defer guarded by !isReplaying() from running
		eventHandler.isReplay = true
		eventHandler.Close()
		w.eventHandler.Store((*workflowExecutionEventHandlerImpl)(nil))
	}
}

func (w *workflowExecutionContextImpl) createEventHandler() {
	w.clearState()
	eventHandler := newWorkflowExecutionEventHandler(
		w.workflowInfo,
		w.completeWorkflow,
		w.wth.logger,
		w.wth.enableLoggingInReplay,
		w.wth.metricsScope,
		w.wth.registry,
		w.wth.dataConverter,
		w.wth.contextPropagators,
		w.wth.tracer,
		w.wth.workflowInterceptors,
	)
	w.eventHandler.Store(eventHandler)
}

func resetHistory(task *apiv1.PollForDecisionTaskResponse, historyIterator HistoryIterator) (*apiv1.History, error) {
	historyIterator.Reset()
	firstPageHistory, err := historyIterator.GetNextPage()
	if err != nil {
		return nil, err
	}
	task.History = firstPageHistory
	return firstPageHistory, nil
}

func (wth *workflowTaskHandlerImpl) createWorkflowContext(task *apiv1.PollForDecisionTaskResponse) (*workflowExecutionContextImpl, error) {
	h := task.History
	attributes := h.Events[0].GetWorkflowExecutionStartedEventAttributes()
	if attributes == nil {
		return nil, errors.New("first history event is not WorkflowExecutionStarted")
	}
	taskList := attributes.TaskList
	if taskList == nil {
		return nil, errors.New("nil TaskList in WorkflowExecutionStarted event")
	}

	runID := task.WorkflowExecution.GetRunId()
	workflowID := task.WorkflowExecution.GetWorkflowId()

	// Setup workflow Info
	var parentWorkflowExecution *WorkflowExecution
	var parentDomain *string
	if attributes.ParentExecutionInfo != nil {
		parentDomain = &attributes.ParentExecutionInfo.DomainName
		parentWorkflowExecution = &WorkflowExecution{
			ID:    attributes.ParentExecutionInfo.WorkflowExecution.GetWorkflowId(),
			RunID: attributes.ParentExecutionInfo.WorkflowExecution.GetRunId(),
		}
	}
	workflowInfo := &WorkflowInfo{
		WorkflowExecution: WorkflowExecution{
			ID:    workflowID,
			RunID: runID,
		},
		WorkflowType:                        WorkflowType{Name: task.WorkflowType.Name},
		TaskListName:                        taskList.GetName(),
		ExecutionStartToCloseTimeoutSeconds: api.SecondsFromProto(attributes.ExecutionStartToCloseTimeout),
		TaskStartToCloseTimeoutSeconds:      api.SecondsFromProto(attributes.TaskStartToCloseTimeout),
		Domain:                              wth.domain,
		Attempt:                             attributes.GetAttempt(),
		lastCompletionResult:                attributes.LastCompletionResult.GetData(),
		CronSchedule:                        &attributes.CronSchedule,
		ContinuedExecutionRunID:             &attributes.ContinuedExecutionRunId,
		ParentWorkflowDomain:                parentDomain,
		ParentWorkflowExecution:             parentWorkflowExecution,
		Memo:                                attributes.Memo,
		SearchAttributes:                    attributes.SearchAttributes,
		RetryPolicy:                         attributes.RetryPolicy,
	}

	wfStartTime := api.TimeFromProto(h.Events[0].EventTime)
	return newWorkflowExecutionContext(wfStartTime, workflowInfo, wth), nil
}

func (wth *workflowTaskHandlerImpl) getOrCreateWorkflowContext(
	task *apiv1.PollForDecisionTaskResponse,
	historyIterator HistoryIterator,
) (workflowContext *workflowExecutionContextImpl, err error) {
	metricsScope := wth.metricsScope.GetTaggedScope(tagWorkflowType, task.WorkflowType.GetName())
	defer func() {
		if err == nil && workflowContext != nil && workflowContext.laTunnel == nil {
			workflowContext.laTunnel = wth.laTunnel
		}
		metricsScope.Gauge(metrics.StickyCacheSize).Update(float64(getWorkflowCache().Size()))
	}()

	runID := task.WorkflowExecution.GetRunId()

	history := task.History
	isFullHistory := isFullHistory(history)

	workflowContext = nil
	if task.Query == nil || (task.Query != nil && !isFullHistory) {
		workflowContext = getWorkflowContext(runID)
	}

	if workflowContext != nil {
		workflowContext.Lock()
		if task.Query != nil && !isFullHistory {
			// query task and we have a valid cached state
			metricsScope.Counter(metrics.StickyCacheHit).Inc(1)
		} else if history.Events[0].GetEventId() == workflowContext.previousStartedEventID+1 {
			// non query task and we have a valid cached state
			metricsScope.Counter(metrics.StickyCacheHit).Inc(1)
		} else {
			// non query task and cached state is missing events, we need to discard the cached state and rebuild one.
			workflowContext.ResetIfStale(task, historyIterator)
		}
	} else {
		if !isFullHistory {
			// we are getting partial history task, but cached state was already evicted.
			// we need to reset history so we get events from beginning to replay/rebuild the state
			metricsScope.Counter(metrics.StickyCacheMiss).Inc(1)
			if history, err = resetHistory(task, historyIterator); err != nil {
				return
			}
		}

		if workflowContext, err = wth.createWorkflowContext(task); err != nil {
			return
		}

		if !wth.disableStickyExecution && task.Query == nil {
			workflowContext, _ = putWorkflowContext(runID, workflowContext)
		}
		workflowContext.Lock()
	}

	err = workflowContext.resetStateIfDestroyed(task, historyIterator)
	if err != nil {
		workflowContext.Unlock(err)
	}

	return
}

func isFullHistory(history *apiv1.History) bool {
	if len(history.Events) == 0 || history.Events[0].GetWorkflowExecutionStartedEventAttributes() == nil {
		return false
	}
	return true
}

func (w *workflowExecutionContextImpl) resetStateIfDestroyed(task *apiv1.PollForDecisionTaskResponse, historyIterator HistoryIterator) error {
	// It is possible that 2 threads (one for decision task and one for query task) that both are getting this same
	// cached workflowContext. If one task finished with err, it would destroy the cached state. In that case, the
	// second task needs to reset the cache state and start from beginning of the history.
	if w.IsDestroyed() {
		w.createEventHandler()
		// reset history events if necessary
		if !isFullHistory(task.History) {
			if _, err := resetHistory(task, historyIterator); err != nil {
				return err
			}
		}
	}
	return nil
}

// ProcessWorkflowTask processes all the events of the workflow task.
func (wth *workflowTaskHandlerImpl) ProcessWorkflowTask(
	workflowTask *workflowTask,
	heartbeatFunc decisionHeartbeatFunc,
) (completeRequest interface{}, errRet error) {
	if workflowTask == nil || workflowTask.task == nil {
		return nil, errors.New("nil workflow task provided")
	}
	task := workflowTask.task
	if task.History == nil || len(task.History.Events) == 0 {
		task.History = &apiv1.History{
			Events: []*apiv1.HistoryEvent{},
		}
	}
	if task.Query == nil && len(task.History.Events) == 0 {
		return nil, errors.New("nil or empty history")
	}

	if task.Query != nil && len(task.Queries) != 0 {
		return nil, errors.New("invalid query decision task")
	}

	runID := task.WorkflowExecution.GetRunId()
	workflowID := task.WorkflowExecution.GetWorkflowId()
	traceLog(func() {
		wth.logger.Debug("Processing new workflow task.",
			zap.String(tagWorkflowType, task.WorkflowType.GetName()),
			zap.String(tagWorkflowID, workflowID),
			zap.String(tagRunID, runID),
			zap.Int64("PreviousStartedEventId", task.GetPreviousStartedEventId().GetValue()))
	})

	workflowContext, err := wth.getOrCreateWorkflowContext(task, workflowTask.historyIterator)
	if err != nil {
		return nil, err
	}

	defer func() {
		workflowContext.Unlock(errRet)
	}()

	var response interface{}
process_Workflow_Loop:
	for {
		startTime := time.Now()
		response, err = workflowContext.ProcessWorkflowTask(workflowTask)
		if err == nil && response == nil {
		wait_LocalActivity_Loop:
			for {
				deadlineToTrigger := time.Duration(float32(ratioToForceCompleteDecisionTaskComplete) * float32(workflowContext.GetDecisionTimeout()))
				delayDuration := startTime.Add(deadlineToTrigger).Sub(time.Now())
				select {
				case <-time.After(delayDuration):
					// force complete, call the decision heartbeat function
					workflowTask, err = heartbeatFunc(
						workflowContext.CompleteDecisionTask(workflowTask, false),
						startTime,
					)
					if err != nil {
						return nil, &decisionHeartbeatError{Message: fmt.Sprintf("error sending decision heartbeat %v", err)}
					}
					if workflowTask == nil {
						return nil, nil
					}
					continue process_Workflow_Loop

				case lar := <-workflowTask.laResultCh:
					// local activity result ready
					response, err = workflowContext.ProcessLocalActivityResult(workflowTask, lar)
					if err == nil && response == nil {
						// decision task is not done yet, still waiting for more local activities
						continue wait_LocalActivity_Loop
					}
					break process_Workflow_Loop
				}
			}
		} else {
			break process_Workflow_Loop
		}
	}
	return response, err
}

func (w *workflowExecutionContextImpl) ProcessWorkflowTask(workflowTask *workflowTask) (interface{}, error) {
	task := workflowTask.task
	historyIterator := workflowTask.historyIterator
	if err := w.ResetIfStale(task, historyIterator); err != nil {
		return nil, err
	}
	w.SetCurrentTask(task)

	eventHandler := w.getEventHandler()
	reorderedHistory := newHistory(workflowTask, eventHandler)
	var replayDecisions []*apiv1.Decision
	var respondEvents []*apiv1.HistoryEvent

	skipReplayCheck := w.skipReplayCheck()
	isReplayTest := task.GetPreviousStartedEventId().GetValue() == replayPreviousStartedEventID
	if isReplayTest {
		w.wth.logger.Info("Processing workflow task in replay test mode",
			zap.String(tagWorkflowType, task.WorkflowType.GetName()),
			zap.String(tagWorkflowID, task.WorkflowExecution.GetWorkflowId()),
			zap.String(tagRunID, task.WorkflowExecution.GetRunId()),
		)
	}
	// Process events
ProcessEvents:
	for {
		reorderedEvents, markers, binaryChecksum, err := reorderedHistory.NextDecisionEvents()
		if err != nil {
			return nil, err
		}

		if len(reorderedEvents) == 0 {
			break ProcessEvents
		}
		if binaryChecksum == nil {
			w.workflowInfo.BinaryChecksum = common.StringPtr(getBinaryChecksum())
		} else {
			w.workflowInfo.BinaryChecksum = binaryChecksum
		}
		// Markers are from the events that are produced from the current decision
		for _, m := range markers {
			if m.GetMarkerRecordedEventAttributes().GetMarkerName() != localActivityMarkerName {
				// local activity marker needs to be applied after decision task started event
				err := eventHandler.ProcessEvent(m, true, false)
				if err != nil {
					return nil, err
				}
				if w.isWorkflowCompleted {
					break ProcessEvents
				}
			}
		}

		for i, event := range reorderedEvents {
			isInReplay := reorderedHistory.IsReplayEvent(event)
			isLast := !isInReplay && i == len(reorderedEvents)-1
			if !skipReplayCheck && isDecisionEvent(event) {
				respondEvents = append(respondEvents, event)
			}

			if isPreloadMarkerEvent(event) {
				// marker events are processed separately
				continue
			}

			// Any pressure points.
			err := w.wth.executeAnyPressurePoints(event, isInReplay)
			if err != nil {
				return nil, err
			}

			err = eventHandler.ProcessEvent(event, isInReplay, isLast)
			if err != nil {
				return nil, err
			}
			if w.isWorkflowCompleted {
				break ProcessEvents
			}
		}

		// now apply local activity markers
		for _, m := range markers {
			if m.GetMarkerRecordedEventAttributes().GetMarkerName() == localActivityMarkerName {
				err := eventHandler.ProcessEvent(m, true, false)
				if err != nil {
					return nil, err
				}
				if w.isWorkflowCompleted {
					break ProcessEvents
				}
			}
		}
		isReplay := len(reorderedEvents) > 0 && reorderedHistory.IsReplayEvent(reorderedEvents[len(reorderedEvents)-1])
		lastDecisionEventsForReplayTest := isReplayTest && !reorderedHistory.HasNextDecisionEvents()
		if isReplay && !lastDecisionEventsForReplayTest {
			eventDecisions := eventHandler.decisionsHelper.getDecisions(true)
			if len(eventDecisions) > 0 && !skipReplayCheck {
				replayDecisions = append(replayDecisions, eventDecisions...)
			}
		}
	}

	// Non-deterministic error could happen in 2 different places:
	//   1) the replay decisions does not match to history events. This is usually due to non backwards compatible code
	// change to decider logic. For example, change calling one activity to a different activity.
	//   2) the decision state machine is trying to make illegal state transition while replay a history event (like
	// activity task completed), but the corresponding decider code that start the event has been removed. In that case
	// the replay of that event will panic on the decision state machine and the workflow will be marked as completed
	// with the panic error.
	var nonDeterministicErr error
	if !skipReplayCheck && !w.isWorkflowCompleted || isReplayTest {
		// check if decisions from reply matches to the history events
		if err := matchReplayWithHistory(replayDecisions, respondEvents); err != nil {
			nonDeterministicErr = err
		}
	}
	if nonDeterministicErr == nil && w.err != nil {
		if panicErr, ok := w.err.(*workflowPanicError); ok && panicErr.value != nil {
			if _, isStateMachinePanic := panicErr.value.(stateMachineIllegalStatePanic); isStateMachinePanic {
				if isReplayTest {
					// NOTE: we should do this regardless if it's in replay test or not
					// but since previously we checked the wrong error type, it may break existing customers workflow
					nonDeterministicErr = panicErr
				} else {
					w.wth.logger.Warn("Ignored workflow panic error",
						zap.String(tagWorkflowType, task.WorkflowType.GetName()),
						zap.String(tagWorkflowID, task.WorkflowExecution.GetWorkflowId()),
						zap.String(tagRunID, task.WorkflowExecution.GetRunId()),
						zap.Error(nonDeterministicErr),
					)
				}
			}
		}
	}

	if nonDeterministicErr != nil {

		w.wth.metricsScope.GetTaggedScope(tagWorkflowType, task.WorkflowType.GetName()).Counter(metrics.NonDeterministicError).Inc(1)
		w.wth.logger.Error("non-deterministic-error",
			zap.String(tagWorkflowType, task.WorkflowType.GetName()),
			zap.String(tagWorkflowID, task.WorkflowExecution.GetWorkflowId()),
			zap.String(tagRunID, task.WorkflowExecution.GetRunId()),
			zap.Error(nonDeterministicErr))

		switch w.wth.nonDeterministicWorkflowPolicy {
		case NonDeterministicWorkflowPolicyFailWorkflow:
			// complete workflow with custom error will fail the workflow
			eventHandler.Complete(nil, NewCustomError("NonDeterministicWorkflowPolicyFailWorkflow", nonDeterministicErr.Error()))
		case NonDeterministicWorkflowPolicyBlockWorkflow:
			// return error here will be convert to DecisionTaskFailed for the first time, and ignored for subsequent
			// attempts which will cause DecisionTaskTimeout and server will retry forever until issue got fixed or
			// workflow timeout.
			return nil, nonDeterministicErr
		default:
			panic(fmt.Sprintf("unknown mismatched workflow history policy."))
		}
	}

	return w.CompleteDecisionTask(workflowTask, true), nil
}

func (w *workflowExecutionContextImpl) ProcessLocalActivityResult(workflowTask *workflowTask, lar *localActivityResult) (interface{}, error) {
	if lar.err != nil && w.retryLocalActivity(lar) {
		return nil, nil // nothing to do here as we are retrying...
	}

	err := w.getEventHandler().ProcessLocalActivityResult(lar)
	if err != nil {
		return nil, err
	}

	return w.CompleteDecisionTask(workflowTask, true), nil
}

func (w *workflowExecutionContextImpl) retryLocalActivity(lar *localActivityResult) bool {
	if lar.task.retryPolicy == nil || lar.err == nil || IsCanceledError(lar.err) {
		return false
	}

	backoff := getRetryBackoff(lar, time.Now())
	if backoff > 0 && backoff <= w.GetDecisionTimeout() {
		// we need a local retry
		time.AfterFunc(backoff, func() {
			// TODO: this should not be a separate goroutine as it introduces race condition when accessing eventHandler.
			// currently this is solved by changing eventHandler to an atomic.Value. Ideally, this retry timer should be
			// part of the event loop for processing the workflow task.
			eventHandler := w.getEventHandler()

			// if decision heartbeat failed, the workflow execution context will be cleared and eventHandler will be nil
			if eventHandler == nil {
				return
			}

			if _, ok := eventHandler.pendingLaTasks[lar.task.activityID]; !ok {
				return
			}

			lar.task.attempt++

			if !w.laTunnel.sendTask(lar.task) {
				lar.task.attempt--
			}
		})
		return true
	}
	// Backoff could be large and potentially much larger than DecisionTaskTimeout. We cannot just sleep locally for
	// retry. Because it will delay the local activity from complete which keeps the decision task open. In order to
	// keep decision task open, we have to keep "heartbeating" current decision task.
	// In that case, it is more efficient to create a server timer with backoff duration and retry when that backoff
	// timer fires. So here we will return false to indicate we don't need local retry anymore. However, we have to
	// store the current attempt and backoff to the same LocalActivityResultMarker so the replay can do the right thing.
	// The backoff timer will be created by workflow.ExecuteLocalActivity().
	lar.backoff = backoff

	return false
}

func getRetryBackoff(lar *localActivityResult, now time.Time) time.Duration {
	p := lar.task.retryPolicy
	var errReason string
	if len(p.NonRetriableErrorReasons) > 0 {
		if lar.err == ErrDeadlineExceeded {
			errReason = "timeout: " + timeoutTypeString(apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_START)
		} else {
			errReason, _ = getErrorDetails(lar.err, nil)
		}
	}
	return getRetryBackoffWithNowTime(p, lar.task.attempt, errReason, now, lar.task.expireTime)
}

func getRetryBackoffWithNowTime(p *RetryPolicy, attempt int32, errReason string, now, expireTime time.Time) time.Duration {
	if p.MaximumAttempts == 0 && p.ExpirationInterval == 0 {
		return noRetryBackoff
	}

	if p.MaximumAttempts > 0 && attempt > p.MaximumAttempts-1 {
		return noRetryBackoff // max attempt reached
	}

	backoffInterval := time.Duration(float64(p.InitialInterval) * math.Pow(p.BackoffCoefficient, float64(attempt)))
	if backoffInterval <= 0 {
		// math.Pow() could overflow
		if p.MaximumInterval > 0 {
			backoffInterval = p.MaximumInterval
		} else {
			return noRetryBackoff
		}
	}

	if p.MaximumInterval > 0 && backoffInterval > p.MaximumInterval {
		// cap next interval to MaxInterval
		backoffInterval = p.MaximumInterval
	}

	nextScheduleTime := now.Add(backoffInterval)
	if !expireTime.IsZero() && nextScheduleTime.After(expireTime) {
		return noRetryBackoff
	}

	// check if error is non-retriable
	for _, er := range p.NonRetriableErrorReasons {
		if er == errReason {
			return noRetryBackoff
		}
	}

	return backoffInterval
}

func (w *workflowExecutionContextImpl) CompleteDecisionTask(workflowTask *workflowTask, waitLocalActivities bool) interface{} {
	if w.currentDecisionTask == nil {
		return nil
	}
	eventHandler := w.getEventHandler()

	// w.laTunnel could be nil for worker.ReplayHistory() because there is no worker started, in that case we don't
	// care about the pending local activities, and just return because the result is ignored anyway by the caller.
	if w.hasPendingLocalActivityWork() && w.laTunnel != nil {
		if len(eventHandler.unstartedLaTasks) > 0 {
			// start new local activity tasks
			unstartedLaTasks := make(map[string]struct{})
			for activityID := range eventHandler.unstartedLaTasks {
				task := eventHandler.pendingLaTasks[activityID]
				task.wc = w
				task.workflowTask = workflowTask
				if !w.laTunnel.sendTask(task) {
					unstartedLaTasks[activityID] = struct{}{}
					task.wc = nil
					task.workflowTask = nil
				}
			}
			eventHandler.unstartedLaTasks = unstartedLaTasks
		}
		// cannot complete decision task as there are pending local activities
		if waitLocalActivities {
			return nil
		}
	}

	eventDecisions := eventHandler.decisionsHelper.getDecisions(true)
	if len(eventDecisions) > 0 {
		w.newDecisions = append(w.newDecisions, eventDecisions...)
	}

	completeRequest := w.wth.completeWorkflow(eventHandler, w.currentDecisionTask, w, w.newDecisions, !waitLocalActivities)
	w.clearCurrentTask()

	return completeRequest
}

func (w *workflowExecutionContextImpl) hasPendingLocalActivityWork() bool {
	eventHandler := w.getEventHandler()
	return !w.isWorkflowCompleted &&
		w.currentDecisionTask != nil &&
		w.currentDecisionTask.Query == nil && // don't run local activity for query task
		eventHandler != nil &&
		len(eventHandler.pendingLaTasks) > 0
}

func (w *workflowExecutionContextImpl) clearCurrentTask() {
	w.newDecisions = nil
	w.currentDecisionTask = nil
}

func (w *workflowExecutionContextImpl) skipReplayCheck() bool {
	return w.currentDecisionTask.Query != nil || !isFullHistory(w.currentDecisionTask.History)
}

func (w *workflowExecutionContextImpl) SetCurrentTask(task *apiv1.PollForDecisionTaskResponse) {
	w.currentDecisionTask = task
	// do not update the previousStartedEventID for query task
	if task.Query == nil {
		w.previousStartedEventID = task.GetStartedEventId()
	}
	w.decisionStartTime = time.Now()
}

func (w *workflowExecutionContextImpl) ResetIfStale(task *apiv1.PollForDecisionTaskResponse, historyIterator HistoryIterator) error {
	if len(task.History.Events) > 0 && task.History.Events[0].GetEventId() != w.previousStartedEventID+1 {
		w.wth.logger.Debug("Cached state staled, new task has unexpected events",
			zap.String(tagWorkflowID, task.WorkflowExecution.GetWorkflowId()),
			zap.String(tagRunID, task.WorkflowExecution.GetRunId()),
			zap.Int64("CachedPreviousStartedEventID", w.previousStartedEventID),
			zap.Int64("TaskFirstEventID", task.History.Events[0].GetEventId()),
			zap.Int64("TaskStartedEventID", task.GetStartedEventId()),
			zap.Int64("TaskPreviousStartedEventID", task.GetPreviousStartedEventId().GetValue()))

		w.wth.metricsScope.
			GetTaggedScope(tagWorkflowType, task.WorkflowType.GetName()).
			Counter(metrics.StickyCacheStall).Inc(1)

		w.clearState()
		return w.resetStateIfDestroyed(task, historyIterator)
	}
	return nil
}

func (w *workflowExecutionContextImpl) GetDecisionTimeout() time.Duration {
	return time.Second * time.Duration(w.workflowInfo.TaskStartToCloseTimeoutSeconds)
}

func skipDeterministicCheckForDecision(d *apiv1.Decision) bool {
	if attr := d.GetRecordMarkerDecisionAttributes(); attr != nil {
		markerName := attr.GetMarkerName()
		if markerName == versionMarkerName || markerName == mutableSideEffectMarkerName {
			return true
		}
	}
	return false
}

func skipDeterministicCheckForEvent(e *apiv1.HistoryEvent) bool {
	if attr := e.GetMarkerRecordedEventAttributes(); attr != nil {
		markerName := attr.GetMarkerName()
		if markerName == versionMarkerName || markerName == mutableSideEffectMarkerName {
			return true
		}
	}
	return false
}

// special check for upsert change version event
func skipDeterministicCheckForUpsertChangeVersion(events []*apiv1.HistoryEvent, idx int) bool {
	e := events[idx]
	if attr := e.GetMarkerRecordedEventAttributes(); attr != nil &&
		attr.GetMarkerName() == versionMarkerName &&
		idx < len(events)-1 &&
		events[idx+1].GetUpsertWorkflowSearchAttributesEventAttributes() != nil {
		if _, ok := events[idx+1].GetUpsertWorkflowSearchAttributesEventAttributes().SearchAttributes.IndexedFields[CadenceChangeVersion]; ok {
			return true
		}
	}
	return false
}

func matchReplayWithHistory(replayDecisions []*apiv1.Decision, historyEvents []*apiv1.HistoryEvent) error {
	di := 0
	hi := 0
	hSize := len(historyEvents)
	dSize := len(replayDecisions)
matchLoop:
	for hi < hSize || di < dSize {
		var e *apiv1.HistoryEvent
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

		var d *apiv1.Decision
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

func isDecisionMatchEvent(d *apiv1.Decision, e *apiv1.HistoryEvent, strictMode bool) bool {
	switch attr := d.Attributes.(type) {
	case *apiv1.Decision_ScheduleActivityTaskDecisionAttributes:
		eventAttributes := e.GetActivityTaskScheduledEventAttributes()
		if eventAttributes == nil {
			return false
		}
		decisionAttributes := attr.ScheduleActivityTaskDecisionAttributes

		if eventAttributes.GetActivityId() != decisionAttributes.GetActivityId() ||
			lastPartOfName(eventAttributes.ActivityType.GetName()) != lastPartOfName(decisionAttributes.ActivityType.GetName()) ||
			(strictMode && eventAttributes.TaskList.GetName() != decisionAttributes.TaskList.GetName()) ||
			(strictMode && bytes.Compare(eventAttributes.Input.GetData(), decisionAttributes.Input.GetData()) != 0) {
			return false
		}

		return true

	case *apiv1.Decision_RequestCancelActivityTaskDecisionAttributes:
		eventAttributes := e.GetActivityTaskCancelRequestedEventAttributes()
		if eventAttributes == nil {
			return false
		}
		decisionAttributes := attr.RequestCancelActivityTaskDecisionAttributes

		if eventAttributes.GetActivityId() != decisionAttributes.GetActivityId() {
			return false
		}

		return true

	case *apiv1.Decision_StartTimerDecisionAttributes:
		eventAttributes := e.GetTimerStartedEventAttributes()
		if eventAttributes == nil {
			return false
		}
		decisionAttributes := attr.StartTimerDecisionAttributes

		if eventAttributes.GetTimerId() != decisionAttributes.GetTimerId() ||
			(strictMode && api.DurationFromProto(eventAttributes.GetStartToFireTimeout()) != api.DurationFromProto(decisionAttributes.GetStartToFireTimeout())) {
			return false
		}

		return true

	case *apiv1.Decision_CancelTimerDecisionAttributes:
		decisionAttributes := attr.CancelTimerDecisionAttributes
		switch eAttr := e.Attributes.(type) {
		case *apiv1.HistoryEvent_TimerCanceledEventAttributes:
			eventAttributes := eAttr.TimerCanceledEventAttributes
			if eventAttributes.GetTimerId() != decisionAttributes.GetTimerId() {
				return false
			}
			return true
		case *apiv1.HistoryEvent_CancelTimerFailedEventAttributes:
			eventAttributes := eAttr.CancelTimerFailedEventAttributes
			if eventAttributes.GetTimerId() != decisionAttributes.GetTimerId() {
				return false
			}
			return true
		default:
			return false
		}

	case *apiv1.Decision_CompleteWorkflowExecutionDecisionAttributes:
		eventAttributes := e.GetWorkflowExecutionCompletedEventAttributes()
		if eventAttributes == nil {
			return false
		}
		if strictMode {
			decisionAttributes := attr.CompleteWorkflowExecutionDecisionAttributes

			if bytes.Compare(eventAttributes.Result.GetData(), decisionAttributes.Result.GetData()) != 0 {
				return false
			}
		}

		return true

	case *apiv1.Decision_FailWorkflowExecutionDecisionAttributes:
		eventAttributes := e.GetWorkflowExecutionFailedEventAttributes()
		if eventAttributes == nil {
			return false
		}
		if strictMode {
			decisionAttributes := attr.FailWorkflowExecutionDecisionAttributes

			if eventAttributes.Failure.GetReason() != decisionAttributes.Failure.GetReason() ||
				bytes.Compare(eventAttributes.Failure.GetDetails(), decisionAttributes.Failure.GetDetails()) != 0 {
				return false
			}
		}

		return true

	case *apiv1.Decision_RecordMarkerDecisionAttributes:
		eventAttributes := e.GetMarkerRecordedEventAttributes()
		if eventAttributes == nil {
			return false
		}

		decisionAttributes := attr.RecordMarkerDecisionAttributes
		if eventAttributes.GetMarkerName() != decisionAttributes.GetMarkerName() {
			return false
		}

		return true

	case *apiv1.Decision_RequestCancelExternalWorkflowExecutionDecisionAttributes:
		eventAttributes := e.GetRequestCancelExternalWorkflowExecutionInitiatedEventAttributes()
		if eventAttributes == nil {
			return false
		}

		decisionAttributes := attr.RequestCancelExternalWorkflowExecutionDecisionAttributes
		if checkDomainsInDecisionAndEvent(eventAttributes.GetDomain(), decisionAttributes.GetDomain()) ||
			eventAttributes.WorkflowExecution.GetWorkflowId() != decisionAttributes.WorkflowExecution.GetWorkflowId() {
			return false
		}

		return true

	case *apiv1.Decision_SignalExternalWorkflowExecutionDecisionAttributes:
		eventAttributes := e.GetSignalExternalWorkflowExecutionInitiatedEventAttributes()
		if eventAttributes == nil {
			return false
		}

		decisionAttributes := attr.SignalExternalWorkflowExecutionDecisionAttributes
		if checkDomainsInDecisionAndEvent(eventAttributes.GetDomain(), decisionAttributes.GetDomain()) ||
			eventAttributes.GetSignalName() != decisionAttributes.GetSignalName() ||
			eventAttributes.WorkflowExecution.GetWorkflowId() != decisionAttributes.WorkflowExecution.GetWorkflowId() {
			return false
		}

		return true

	case *apiv1.Decision_CancelWorkflowExecutionDecisionAttributes:
		eventAttributes := e.GetWorkflowExecutionCanceledEventAttributes()
		if eventAttributes == nil {
			return false
		}
		if strictMode {
			decisionAttributes := attr.CancelWorkflowExecutionDecisionAttributes
			if bytes.Compare(eventAttributes.Details.GetData(), decisionAttributes.Details.GetData()) != 0 {
				return false
			}
		}
		return true

	case *apiv1.Decision_ContinueAsNewWorkflowExecutionDecisionAttributes:
		if e.GetWorkflowExecutionContinuedAsNewEventAttributes() == nil {
			return false
		}

		return true

	case *apiv1.Decision_StartChildWorkflowExecutionDecisionAttributes:
		eventAttributes := e.GetStartChildWorkflowExecutionInitiatedEventAttributes()
		if eventAttributes == nil {
			return false
		}

		decisionAttributes := attr.StartChildWorkflowExecutionDecisionAttributes
		if lastPartOfName(eventAttributes.WorkflowType.GetName()) != lastPartOfName(decisionAttributes.WorkflowType.GetName()) ||
			(strictMode && checkDomainsInDecisionAndEvent(eventAttributes.GetDomain(), decisionAttributes.GetDomain())) ||
			(strictMode && eventAttributes.TaskList.GetName() != decisionAttributes.TaskList.GetName()) {
			return false
		}

		return true

	case *apiv1.Decision_UpsertWorkflowSearchAttributesDecisionAttributes:
		eventAttributes := e.GetUpsertWorkflowSearchAttributesEventAttributes()
		if eventAttributes == nil {
			return false
		}
		decisionAttributes := attr.UpsertWorkflowSearchAttributesDecisionAttributes
		if strictMode && !isSearchAttributesMatched(eventAttributes.SearchAttributes, decisionAttributes.SearchAttributes) {
			return false
		}
		return true
	}

	return false
}

func isSearchAttributesMatched(attrFromEvent, attrFromDecision *apiv1.SearchAttributes) bool {
	if attrFromEvent != nil && attrFromDecision != nil {
		return reflect.DeepEqual(attrFromEvent.IndexedFields, attrFromDecision.IndexedFields)
	}
	return attrFromEvent == nil && attrFromDecision == nil
}

// return true if the check fails:
//    domain is not empty in decision
//    and domain is not replayDomain
//    and domains unmatch in decision and events
func checkDomainsInDecisionAndEvent(eventDomainName, decisionDomainName string) bool {
	if decisionDomainName == "" || IsReplayDomain(decisionDomainName) {
		return false
	}
	return eventDomainName != decisionDomainName
}

func (wth *workflowTaskHandlerImpl) completeWorkflow(
	eventHandler *workflowExecutionEventHandlerImpl,
	task *apiv1.PollForDecisionTaskResponse,
	workflowContext *workflowExecutionContextImpl,
	decisions []*apiv1.Decision,
	forceNewDecision bool) interface{} {

	// for query task
	if task.Query != nil {
		queryCompletedRequest := &apiv1.RespondQueryTaskCompletedRequest{TaskToken: task.TaskToken}
		if panicErr, ok := workflowContext.err.(*PanicError); ok {
			// NOTE: this code path should never be executed, we should check for workflowPanicError instead of PanicError
			wth.logger.Warn("Encountered PanicError in workflow query task",
				zap.String(tagWorkflowID, task.WorkflowExecution.GetWorkflowId()),
				zap.String(tagRunID, task.WorkflowExecution.GetRunId()),
				zap.String(tagPanicError, panicErr.Error()),
				zap.String(tagPanicStack, panicErr.StackTrace()),
			)

			queryCompletedRequest.Result = &apiv1.WorkflowQueryResult{
				ResultType: apiv1.QueryResultType_QUERY_RESULT_TYPE_FAILED,
				ErrorMessage: "Workflow panic: " + panicErr.Error(),
			}
			return queryCompletedRequest
		}

		if workflowPanicErr, ok := workflowContext.err.(*workflowPanicError); ok {
			// NOTE: in this case we should return complete query task with CompletedTypeFailed
			// but we didn't check for the right error type before, this may break existing customer
			wth.logger.Warn("Ignored workflow panic error for query, query result may be partial",
				zap.String(tagWorkflowID, task.WorkflowExecution.GetWorkflowId()),
				zap.String(tagRunID, task.WorkflowExecution.GetRunId()),
				zap.String(tagPanicError, workflowPanicErr.Error()),
				zap.String(tagPanicStack, workflowPanicErr.StackTrace()),
				zap.Int64("PreviousStartedEventID", task.GetPreviousStartedEventId().GetValue()),
			)
		}

		result, err := eventHandler.ProcessQuery(task.Query.GetQueryType(), task.Query.QueryArgs.GetData())
		if err != nil {
			queryCompletedRequest.Result = &apiv1.WorkflowQueryResult{
				ResultType: apiv1.QueryResultType_QUERY_RESULT_TYPE_FAILED,
				ErrorMessage: err.Error(),
			}
		} else {
			queryCompletedRequest.Result = &apiv1.WorkflowQueryResult{
				ResultType: apiv1.QueryResultType_QUERY_RESULT_TYPE_ANSWERED,
				Answer: &apiv1.Payload{Data: result},
			}
		}
		return queryCompletedRequest
	}

	metricsScope := wth.metricsScope.GetTaggedScope(tagWorkflowType, eventHandler.workflowEnvironmentImpl.workflowInfo.WorkflowType.Name)

	// fail decision task on decider panic
	if panicErr, ok := workflowContext.err.(*workflowPanicError); ok {
		// Workflow panic
		metricsScope.Counter(metrics.DecisionTaskPanicCounter).Inc(1)
		wth.logger.Error("Workflow panic.",
			zap.String(tagWorkflowID, task.WorkflowExecution.GetWorkflowId()),
			zap.String(tagRunID, task.WorkflowExecution.GetRunId()),
			zap.String(tagPanicError, panicErr.Error()),
			zap.String(tagPanicStack, panicErr.StackTrace()))
		return errorToFailDecisionTask(task.TaskToken, panicErr, wth.identity)
	}

	// complete decision task
	var closeDecision *apiv1.Decision
	if canceledErr, ok := workflowContext.err.(*CanceledError); ok {
		// Workflow cancelled
		metricsScope.Counter(metrics.WorkflowCanceledCounter).Inc(1)
		_, details := getErrorDetails(canceledErr, wth.dataConverter)
		closeDecision = &apiv1.Decision{
			Attributes: &apiv1.Decision_CancelWorkflowExecutionDecisionAttributes{
				CancelWorkflowExecutionDecisionAttributes: &apiv1.CancelWorkflowExecutionDecisionAttributes{
					Details: &apiv1.Payload{Data: details},
				},
			},
		}
	} else if contErr, ok := workflowContext.err.(*ContinueAsNewError); ok {
		// Continue as new error.
		metricsScope.Counter(metrics.WorkflowContinueAsNewCounter).Inc(1)
		closeDecision = &apiv1.Decision{
			Attributes: &apiv1.Decision_ContinueAsNewWorkflowExecutionDecisionAttributes{
				ContinueAsNewWorkflowExecutionDecisionAttributes: &apiv1.ContinueAsNewWorkflowExecutionDecisionAttributes{
					WorkflowType:                 &apiv1.WorkflowType{Name: contErr.params.workflowType.Name},
					Input:                        &apiv1.Payload{Data: contErr.params.input},
					TaskList:                     &apiv1.TaskList{Name: *contErr.params.taskListName},
					ExecutionStartToCloseTimeout: api.SecondsPtrToProto(contErr.params.executionStartToCloseTimeoutSeconds),
					TaskStartToCloseTimeout:      api.SecondsPtrToProto(contErr.params.taskStartToCloseTimeoutSeconds),
					Header:                       contErr.params.header,
					Memo:                         workflowContext.workflowInfo.Memo,
					SearchAttributes:             workflowContext.workflowInfo.SearchAttributes,
					RetryPolicy:                  workflowContext.workflowInfo.RetryPolicy,
				},
			},
		}
	} else if workflowContext.err != nil {
		// Workflow failures
		metricsScope.Counter(metrics.WorkflowFailedCounter).Inc(1)
		reason, details := getErrorDetails(workflowContext.err, wth.dataConverter)
		closeDecision = &apiv1.Decision{
			Attributes: &apiv1.Decision_FailWorkflowExecutionDecisionAttributes{
				FailWorkflowExecutionDecisionAttributes: &apiv1.FailWorkflowExecutionDecisionAttributes{
					Failure: &apiv1.Failure{
						Reason:  reason,
						Details: details,
					},
				},
			},
		}
	} else if workflowContext.isWorkflowCompleted {
		// Workflow completion
		metricsScope.Counter(metrics.WorkflowCompletedCounter).Inc(1)
		closeDecision = &apiv1.Decision{
			Attributes: &apiv1.Decision_CompleteWorkflowExecutionDecisionAttributes{
				CompleteWorkflowExecutionDecisionAttributes: &apiv1.CompleteWorkflowExecutionDecisionAttributes{
					Result: &apiv1.Payload{Data: workflowContext.result},
				},
			},
		}
	}

	if closeDecision != nil {
		decisions = append(decisions, closeDecision)
		elapsed := time.Now().Sub(workflowContext.workflowStartTime)
		metricsScope.Timer(metrics.WorkflowEndToEndLatency).Record(elapsed)
		forceNewDecision = false
	}

	var queryResults map[string]*apiv1.WorkflowQueryResult
	if len(task.Queries) != 0 {
		queryResults = make(map[string]*apiv1.WorkflowQueryResult)
		for queryID, query := range task.Queries {
			result, err := eventHandler.ProcessQuery(query.GetQueryType(), query.QueryArgs.GetData())
			if err != nil {
				queryResults[queryID] = &apiv1.WorkflowQueryResult{
					ResultType:   apiv1.QueryResultType_QUERY_RESULT_TYPE_FAILED,
					ErrorMessage: err.Error(),
				}
			} else {
				queryResults[queryID] = &apiv1.WorkflowQueryResult{
					ResultType: apiv1.QueryResultType_QUERY_RESULT_TYPE_ANSWERED,
					Answer:     &apiv1.Payload{Data: result},
				}
			}
		}
	}

	return &apiv1.RespondDecisionTaskCompletedRequest{
		TaskToken:                  task.TaskToken,
		Decisions:                  decisions,
		Identity:                   wth.identity,
		ReturnNewDecisionTask:      true,
		ForceCreateNewDecisionTask: forceNewDecision,
		BinaryChecksum:             getBinaryChecksum(),
		QueryResults:               queryResults,
	}
}

func errorToFailDecisionTask(taskToken []byte, err error, identity string) *apiv1.RespondDecisionTaskFailedRequest {
	_, details := getErrorDetails(err, nil)
	return &apiv1.RespondDecisionTaskFailedRequest{
		TaskToken:      taskToken,
		Cause:          apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_WORKFLOW_WORKER_UNHANDLED_FAILURE,
		Details:        &apiv1.Payload{Data: details},
		Identity:       identity,
		BinaryChecksum: getBinaryChecksum(),
	}
}

func (wth *workflowTaskHandlerImpl) executeAnyPressurePoints(event *apiv1.HistoryEvent, isInReplay bool) error {
	if wth.ppMgr != nil && !reflect.ValueOf(wth.ppMgr).IsNil() && !isInReplay {
		switch event.Attributes.(type) {
		case *apiv1.HistoryEvent_DecisionTaskStartedEventAttributes:
			return wth.ppMgr.Execute(pressurePointTypeDecisionTaskStartTimeout)
		case *apiv1.HistoryEvent_ActivityTaskScheduledEventAttributes:
			return wth.ppMgr.Execute(pressurePointTypeActivityTaskScheduleTimeout)
		case *apiv1.HistoryEvent_ActivityTaskStartedEventAttributes:
			return wth.ppMgr.Execute(pressurePointTypeActivityTaskStartTimeout)
		case *apiv1.HistoryEvent_DecisionTaskCompletedEventAttributes:
			return wth.ppMgr.Execute(pressurePointTypeDecisionTaskCompleted)
		}
	}
	return nil
}

func newActivityTaskHandler(
	service api.Interface,
	params workerExecutionParameters,
	registry *registry,
) ActivityTaskHandler {
	return newActivityTaskHandlerWithCustomProvider(service, params, registry, nil)
}

func newActivityTaskHandlerWithCustomProvider(
	service api.Interface,
	params workerExecutionParameters,
	registry *registry,
	activityProvider activityProvider,
) ActivityTaskHandler {
	return &activityTaskHandlerImpl{
		taskListName:       params.TaskList,
		identity:           params.Identity,
		service:            service,
		logger:             params.Logger,
		metricsScope:       metrics.NewTaggedScope(params.MetricsScope),
		userContext:        params.UserContext,
		registry:           registry,
		activityProvider:   activityProvider,
		dataConverter:      params.DataConverter,
		workerStopCh:       params.WorkerStopChannel,
		contextPropagators: params.ContextPropagators,
		tracer:             params.Tracer,
		featureFlags:       params.FeatureFlags,
	}
}

type cadenceInvoker struct {
	sync.Mutex
	identity              string
	service               api.Interface
	taskToken             []byte
	cancelHandler         func()
	heartBeatTimeoutInSec int32       // The heart beat interval configured for this activity.
	hbBatchEndTimer       *time.Timer // Whether we started a batch of operations that need to be reported in the cycle. This gets started on a user call.
	detailsToReport       *[]byte     // Details to be reported in the next reporting interval.
	lastDetailsReported   *[]byte     // Details that were reported in the last reporting interval.
	closeCh               chan struct{}
	workerStopChannel     <-chan struct{}
	featureFlags          FeatureFlags
}

func (i *cadenceInvoker) Heartbeat(details []byte) error {
	i.Lock()
	defer i.Unlock()

	_, err := i.internalHeartBeat(details)
	return err
}

func (i *cadenceInvoker) BackgroundHeartbeat() error {
	i.Lock()
	defer i.Unlock()

	if i.hbBatchEndTimer != nil {
		if i.detailsToReport == nil {
			i.detailsToReport = i.lastDetailsReported
		}

		return nil
	}

	var details []byte
	if i.detailsToReport != nil {
		details = *i.detailsToReport
	} else if i.lastDetailsReported != nil {
		details = *i.lastDetailsReported
	}

	return i.heartbeatAndScheduleNextRun(details)
}

func (i *cadenceInvoker) BatchHeartbeat(details []byte) error {
	i.Lock()
	defer i.Unlock()

	if i.hbBatchEndTimer != nil {
		// If we have started batching window, keep track of last reported progress.
		i.detailsToReport = &details
		return nil
	}

	return i.heartbeatAndScheduleNextRun(details)
}

func (i *cadenceInvoker) heartbeatAndScheduleNextRun(details []byte) error {
	isActivityCancelled, err := i.internalHeartBeat(details)

	// If the activity is cancelled, the activity can ignore the cancellation and do its work
	// and complete. Our cancellation is co-operative, so we will try to heartbeat.
	if err == nil || isActivityCancelled {
		// We have successfully sent heartbeat, start next batching window.
		i.lastDetailsReported = &details
		i.detailsToReport = nil

		// Create timer to fire before the threshold to report.
		deadlineToTrigger := i.heartBeatTimeoutInSec
		if deadlineToTrigger <= 0 {
			// If we don't have any heartbeat timeout configured.
			deadlineToTrigger = defaultHeartBeatIntervalInSec
		}

		// We set a deadline at 80% of the timeout.
		duration := time.Duration(0.8*float32(deadlineToTrigger)) * time.Second
		i.hbBatchEndTimer = time.NewTimer(duration)

		go func() {
			select {
			case <-i.hbBatchEndTimer.C:
				// We are close to deadline.
			case <-i.workerStopChannel:
				// Activity worker is close to stop. This does the same steps as batch timer ends.
			case <-i.closeCh:
				// We got closed.
				return
			}

			// We close the batch and report the progress.
			var detailsToReport *[]byte

			i.Lock()
			detailsToReport = i.detailsToReport
			i.hbBatchEndTimer.Stop()
			i.hbBatchEndTimer = nil

			if detailsToReport != nil {
				i.heartbeatAndScheduleNextRun(*detailsToReport)
			}
			i.Unlock()
		}()
	}

	return err
}

func (i *cadenceInvoker) internalHeartBeat(details []byte) (bool, error) {
	isActivityCancelled := false
	timeout := time.Duration(i.heartBeatTimeoutInSec) * time.Second
	if timeout <= 0 {
		timeout = time.Duration(defaultHeartBeatIntervalInSec) * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := recordActivityHeartbeat(ctx, i.service, i.identity, i.taskToken, details, i.featureFlags)

	switch err.(type) {
	case *CanceledError:
		// We are asked to cancel. inform the activity about cancellation through context.
		i.cancelHandler()
		isActivityCancelled = true

	case *api.EntityNotExistsError, *api.WorkflowExecutionAlreadyCompletedError, *api.DomainNotActiveError:
		// We will pass these through as cancellation for now but something we can change
		// later when we have setter on cancel handler.
		i.cancelHandler()
		isActivityCancelled = true
	}

	// We don't want to bubble temporary errors to the user.
	// This error won't be return to user check RecordActivityHeartbeat().
	return isActivityCancelled, err
}

func (i *cadenceInvoker) Close(flushBufferedHeartbeat bool) {
	i.Lock()
	defer i.Unlock()
	close(i.closeCh)
	if i.hbBatchEndTimer != nil {
		i.hbBatchEndTimer.Stop()
		if flushBufferedHeartbeat && i.detailsToReport != nil {
			i.internalHeartBeat(*i.detailsToReport)
			i.lastDetailsReported = i.detailsToReport
			i.detailsToReport = nil
		}
	}
}

func (i *cadenceInvoker) SignalWorkflow(ctx context.Context, domain, workflowID, runID, signalName string, signalInput []byte) error {
	return signalWorkflow(ctx, i.service, i.identity, domain, workflowID, runID, signalName, signalInput, i.featureFlags)
}

func newServiceInvoker(
	taskToken []byte,
	identity string,
	service api.Interface,
	cancelHandler func(),
	heartBeatTimeoutInSec int32,
	workerStopChannel <-chan struct{},
	featureFlags FeatureFlags,
) ServiceInvoker {
	return &cadenceInvoker{
		taskToken:             taskToken,
		identity:              identity,
		service:               service,
		cancelHandler:         cancelHandler,
		heartBeatTimeoutInSec: heartBeatTimeoutInSec,
		closeCh:               make(chan struct{}),
		workerStopChannel:     workerStopChannel,
		featureFlags:          featureFlags,
	}
}

// Execute executes an implementation of the activity.
func (ath *activityTaskHandlerImpl) Execute(taskList string, t *apiv1.PollForActivityTaskResponse) (result interface{}, err error) {
	traceLog(func() {
		ath.logger.Debug("Processing new activity task",
			zap.String(tagWorkflowID, t.WorkflowExecution.GetWorkflowId()),
			zap.String(tagRunID, t.WorkflowExecution.GetRunId()),
			zap.String(tagActivityType, t.ActivityType.GetName()))
	})

	rootCtx := ath.userContext
	if rootCtx == nil {
		rootCtx = context.Background()
	}
	canCtx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	invoker := newServiceInvoker(t.TaskToken, ath.identity, ath.service, cancel, api.SecondsFromProto(t.HeartbeatTimeout), ath.workerStopCh, ath.featureFlags)
	defer func() {
		_, activityCompleted := result.(*apiv1.RespondActivityTaskCompletedRequest)
		invoker.Close(!activityCompleted) // flush buffered heartbeat if activity was not successfully completed.
	}()

	workflowType := t.WorkflowType.GetName()
	activityType := t.ActivityType.GetName()
	metricsScope := getMetricsScopeForActivity(ath.metricsScope, workflowType, activityType)
	ctx := WithActivityTask(canCtx, t, taskList, invoker, ath.logger, metricsScope, ath.dataConverter, ath.workerStopCh, ath.contextPropagators, ath.tracer)

	activityImplementation := ath.getActivity(activityType)
	if activityImplementation == nil {
		// Couldn't find the activity implementation.
		supported := strings.Join(ath.getRegisteredActivityNames(), ", ")
		return nil, fmt.Errorf("unable to find activityType=%v. Supported types: [%v]", activityType, supported)
	}

	// panic handler
	defer func() {
		if p := recover(); p != nil {
			topLine := fmt.Sprintf("activity for %s [panic]:", ath.taskListName)
			st := getStackTraceRaw(topLine, 7, 0)
			ath.logger.Error("Activity panic.",
				zap.String(tagWorkflowID, t.WorkflowExecution.GetWorkflowId()),
				zap.String(tagRunID, t.WorkflowExecution.GetRunId()),
				zap.String(tagActivityType, activityType),
				zap.String(tagPanicError, fmt.Sprintf("%v", p)),
				zap.String(tagPanicStack, st))
			metricsScope.Counter(metrics.ActivityTaskPanicCounter).Inc(1)
			panicErr := newPanicError(p, st)
			result, err = convertActivityResultToRespondRequest(ath.identity, t.TaskToken, nil, panicErr, ath.dataConverter), nil
		}
	}()

	// propagate context information into the activity context from the headers
	for _, ctxProp := range ath.contextPropagators {
		var err error
		if ctx, err = ctxProp.Extract(ctx, NewHeaderReader(t.Header)); err != nil {
			return nil, fmt.Errorf("unable to propagate context %v", err)
		}
	}

	info := ctx.Value(activityEnvContextKey).(*activityEnvironment)
	ctx, dlCancelFunc := context.WithDeadline(ctx, info.deadline)
	defer dlCancelFunc()

	ctx, span := createOpenTracingActivitySpan(ctx, ath.tracer, time.Now(), activityType, t.WorkflowExecution.GetWorkflowId(), t.WorkflowExecution.GetRunId())
	defer span.Finish()

	heartbeatTimeoutSeconds := api.SecondsFromProto(t.HeartbeatTimeout)
	if activityImplementation.GetOptions().EnableAutoHeartbeat && heartbeatTimeoutSeconds > 0 {
		go func() {
			autoHbInterval := time.Duration(heartbeatTimeoutSeconds) * time.Second / 2
			ticker := time.NewTicker(autoHbInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ath.workerStopCh:
					return
				case <-ctx.Done():
					return
				case <-ticker.C:
					hbErr := invoker.BackgroundHeartbeat()
					if hbErr != nil && !IsCanceledError(hbErr) {
						ath.logger.Error("Activity auto heartbeat error.",
							zap.String(tagWorkflowID, t.WorkflowExecution.GetWorkflowId()),
							zap.String(tagRunID, t.WorkflowExecution.GetRunId()),
							zap.String(tagActivityType, activityType),
							zap.Error(err),
						)
					}
				}
			}
		}()
	}

	output, err := activityImplementation.Execute(ctx, t.Input.GetData())

	dlCancelFunc()
	if <-ctx.Done(); ctx.Err() == context.DeadlineExceeded {
		return nil, ctx.Err()
	}
	if err != nil && err != ErrActivityResultPending {
		ath.logger.Error("Activity error.",
			zap.String(tagWorkflowID, t.WorkflowExecution.GetWorkflowId()),
			zap.String(tagRunID, t.WorkflowExecution.GetRunId()),
			zap.String(tagActivityType, activityType),
			zap.Error(err),
		)
	}
	return convertActivityResultToRespondRequest(ath.identity, t.TaskToken, output, err, ath.dataConverter), nil
}

func (ath *activityTaskHandlerImpl) getActivity(name string) activity {
	if ath.activityProvider != nil {
		return ath.activityProvider(name)
	}

	if a, ok := ath.registry.GetActivity(name); ok {
		return a
	}

	return nil
}

func (ath *activityTaskHandlerImpl) getRegisteredActivityNames() (activityNames []string) {
	for _, a := range ath.registry.getRegisteredActivities() {
		activityNames = append(activityNames, a.ActivityType().Name)
	}
	return
}

func signalWorkflow(
	ctx context.Context,
	service api.Interface,
	identity string,
	domain string,
	workflowID string,
	runID string,
	signalName string,
	signalInput []byte,
	featureFlags FeatureFlags,
) error {
	request := &apiv1.SignalWorkflowExecutionRequest{
		Domain: domain,
		WorkflowExecution: &apiv1.WorkflowExecution{
			WorkflowId: workflowID,
			RunId:      runID,
		},
		SignalName: signalName,
		SignalInput: &apiv1.Payload{Data: signalInput},
		Identity:   identity,
	}

	return backoff.Retry(ctx,
		func() error {
			tchCtx, cancel, opt := newChannelContext(ctx, featureFlags)
			defer cancel()
			_, err := service.SignalWorkflowExecution(tchCtx, request, opt...)
			return api.ConvertError(err)
		}, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)
}

func recordActivityHeartbeat(
	ctx context.Context,
	service api.Interface,
	identity string,
	taskToken, details []byte,
	featureFlags FeatureFlags,
) error {
	request := &apiv1.RecordActivityTaskHeartbeatRequest{
		TaskToken: taskToken,
		Details:   &apiv1.Payload{Data: details},
		Identity:  identity}

	var heartbeatResponse *apiv1.RecordActivityTaskHeartbeatResponse
	heartbeatErr := backoff.Retry(ctx,
		func() error {
			tchCtx, cancel, opt := newChannelContext(ctx, featureFlags)
			defer cancel()

			var err error
			heartbeatResponse, err = service.RecordActivityTaskHeartbeat(tchCtx, request, opt...)
			return api.ConvertError(err)
		}, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)

	if heartbeatErr == nil && heartbeatResponse != nil && heartbeatResponse.GetCancelRequested() {
		return NewCanceledError()
	}

	return heartbeatErr
}

func recordActivityHeartbeatByID(
	ctx context.Context,
	service api.Interface,
	identity string,
	domain, workflowID, runID, activityID string,
	details []byte,
	featureFlags FeatureFlags,
) error {
	request := &apiv1.RecordActivityTaskHeartbeatByIDRequest{
		Domain:     domain,
		WorkflowExecution: &apiv1.WorkflowExecution{
			WorkflowId: workflowID,
			RunId: runID,
		},
		ActivityId: activityID,
		Details:    &apiv1.Payload{Data: details},
		Identity:   identity}

	var heartbeatResponse *apiv1.RecordActivityTaskHeartbeatByIDResponse
	heartbeatErr := backoff.Retry(ctx,
		func() error {
			tchCtx, cancel, opt := newChannelContext(ctx, featureFlags)
			defer cancel()

			var err error
			heartbeatResponse, err = service.RecordActivityTaskHeartbeatByID(tchCtx, request, opt...)
			return api.ConvertError(err)
		}, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)

	if heartbeatErr == nil && heartbeatResponse != nil && heartbeatResponse.GetCancelRequested() {
		return NewCanceledError()
	}

	return heartbeatErr
}

// This enables verbose logging in the client library.
// check worker.EnableVerboseLogging()
func traceLog(fn func()) {
	if enableVerboseLogging {
		fn()
	}
}
