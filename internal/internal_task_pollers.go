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
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/opentracing/opentracing-go"
	"github.com/pborman/uuid"
	"github.com/uber-go/tally"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
	"go.uber.org/cadence/v2/internal/api"
	"go.uber.org/cadence/v2/internal/common/backoff"
	"go.uber.org/cadence/v2/internal/common/metrics"
	"go.uber.org/cadence/v2/internal/common/serializer"
	"go.uber.org/zap"
)

const (
	pollTaskServiceTimeOut = 150 * time.Second // Server long poll is 2 * Minutes + delta

	stickyDecisionScheduleToStartTimeoutSeconds = 5

	ratioToForceCompleteDecisionTaskComplete = 0.8
)

type (
	// taskPoller interface to poll and process for task
	taskPoller interface {
		// PollTask polls for one new task
		PollTask() (interface{}, error)
		// ProcessTask processes a task
		ProcessTask(interface{}) error
	}

	// basePoller is the base class for all poller implementations
	basePoller struct {
		shutdownC <-chan struct{}
	}

	// workflowTaskPoller implements polling/processing a workflow task
	workflowTaskPoller struct {
		basePoller
		domain       string
		taskListName string
		identity     string
		service      api.Interface
		taskHandler  WorkflowTaskHandler
		ldaTunnel    *locallyDispatchedActivityTunnel
		metricsScope *metrics.TaggedScope
		logger       *zap.Logger

		stickyUUID                   string
		disableStickyExecution       bool
		StickyScheduleToStartTimeout time.Duration

		pendingRegularPollCount int
		pendingStickyPollCount  int
		stickyBacklog           int64
		requestLock             sync.Mutex
		featureFlags            FeatureFlags
	}

	// activityTaskPoller implements polling/processing a workflow task
	activityTaskPoller struct {
		basePoller
		domain              string
		taskListName        string
		identity            string
		service             api.Interface
		taskHandler         ActivityTaskHandler
		metricsScope        *metrics.TaggedScope
		logger              *zap.Logger
		activitiesPerSecond float64
		featureFlags        FeatureFlags
	}

	// locallyDispatchedActivityTaskPoller implements polling/processing a locally dispatched activity task
	locallyDispatchedActivityTaskPoller struct {
		activityTaskPoller
		ldaTunnel *locallyDispatchedActivityTunnel
	}

	historyIteratorImpl struct {
		iteratorFunc   func(nextPageToken []byte) (*apiv1.History, []byte, error)
		execution      *apiv1.WorkflowExecution
		nextPageToken  []byte
		domain         string
		service        api.Interface
		metricsScope   tally.Scope
		startedEventID int64
		maxEventID     int64
		featureFlags   FeatureFlags
	}

	localActivityTaskPoller struct {
		basePoller
		handler      *localActivityTaskHandler
		metricsScope tally.Scope
		logger       *zap.Logger
		laTunnel     *localActivityTunnel
	}

	localActivityTaskHandler struct {
		userContext        context.Context
		metricsScope       *metrics.TaggedScope
		logger             *zap.Logger
		dataConverter      DataConverter
		contextPropagators []ContextPropagator
		tracer             opentracing.Tracer
	}

	localActivityResult struct {
		result  []byte
		err     error
		task    *localActivityTask
		backoff time.Duration
	}

	localActivityTunnel struct {
		taskCh   chan *localActivityTask
		resultCh chan interface{}
		stopCh   <-chan struct{}
	}

	locallyDispatchedActivityTunnel struct {
		taskCh       chan *locallyDispatchedActivityTask
		stopCh       <-chan struct{}
		metricsScope *metrics.TaggedScope
	}
)

func newLocalActivityTunnel(stopCh <-chan struct{}) *localActivityTunnel {
	return &localActivityTunnel{
		taskCh:   make(chan *localActivityTask, 1000),
		resultCh: make(chan interface{}),
		stopCh:   stopCh,
	}
}

func (lat *localActivityTunnel) getTask() *localActivityTask {
	select {
	case task := <-lat.taskCh:
		return task
	case <-lat.stopCh:
		return nil
	}
}

func (lat *localActivityTunnel) sendTask(task *localActivityTask) bool {
	select {
	case lat.taskCh <- task:
		return true
	case <-lat.stopCh:
		return false
	}
}

func newLocallyDispatchedActivityTunnel(stopCh <-chan struct{}) *locallyDispatchedActivityTunnel {
	return &locallyDispatchedActivityTunnel{
		taskCh: make(chan *locallyDispatchedActivityTask),
		stopCh: stopCh,
	}
}

func (ldat *locallyDispatchedActivityTunnel) getTask() *locallyDispatchedActivityTask {
	var task *locallyDispatchedActivityTask
	select {
	case task = <-ldat.taskCh:
	case <-ldat.stopCh:
		return nil
	}

	select {
	case ready := <-task.readyCh:
		if ready {
			return task
		} else {
			return nil
		}
	case <-ldat.stopCh:
		return nil
	}
}

func (ldat *locallyDispatchedActivityTunnel) sendTask(task *locallyDispatchedActivityTask) bool {
	select {
	case ldat.taskCh <- task:
		return true
	default:
		return false
	}
}

func isClientSideError(err error) bool {
	// If an activity execution exceeds deadline.
	if err == context.DeadlineExceeded {
		return true
	}

	return false
}

// shuttingDown returns true if worker is shutting down right now
func (bp *basePoller) shuttingDown() bool {
	select {
	case <-bp.shutdownC:
		return true
	default:
		return false
	}
}

// doPoll runs the given pollFunc in a separate go routine. Returns when either of the conditions are met:
// - poll succeeds, poll fails or worker is shutting down
func (bp *basePoller) doPoll(
	featureFlags FeatureFlags,
	pollFunc func(ctx context.Context) (interface{}, error),
) (interface{}, error) {
	if bp.shuttingDown() {
		return nil, errShutdown
	}

	var err error
	var result interface{}

	doneC := make(chan struct{})
	ctx, cancel, _ := newChannelContext(context.Background(), featureFlags, chanTimeout(pollTaskServiceTimeOut))

	go func() {
		result, err = pollFunc(ctx)
		cancel()
		close(doneC)
	}()

	select {
	case <-doneC:
		return result, err
	case <-bp.shutdownC:
		cancel()
		return nil, errShutdown
	}
}

// newWorkflowTaskPoller creates a new workflow task poller which must have a one to one relationship to workflow worker
func newWorkflowTaskPoller(
	taskHandler WorkflowTaskHandler,
	ldaTunnel *locallyDispatchedActivityTunnel,
	service api.Interface,
	domain string,
	params workerExecutionParameters,
) *workflowTaskPoller {
	return &workflowTaskPoller{
		basePoller:                   basePoller{shutdownC: params.WorkerStopChannel},
		service:                      service,
		domain:                       domain,
		taskListName:                 params.TaskList,
		identity:                     params.Identity,
		taskHandler:                  taskHandler,
		ldaTunnel:                    ldaTunnel,
		metricsScope:                 metrics.NewTaggedScope(params.MetricsScope),
		logger:                       params.Logger,
		stickyUUID:                   uuid.New(),
		disableStickyExecution:       params.DisableStickyExecution,
		StickyScheduleToStartTimeout: params.StickyScheduleToStartTimeout,
		featureFlags:                 params.FeatureFlags,
	}
}

// PollTask polls a new task
func (wtp *workflowTaskPoller) PollTask() (interface{}, error) {
	// Get the task.
	workflowTask, err := wtp.doPoll(wtp.featureFlags, wtp.poll)
	if err != nil {
		return nil, err
	}

	return workflowTask, nil
}

// ProcessTask processes a task which could be workflow task or local activity result
func (wtp *workflowTaskPoller) ProcessTask(task interface{}) error {
	if wtp.shuttingDown() {
		return errShutdown
	}

	switch task.(type) {
	case *workflowTask:
		return wtp.processWorkflowTask(task.(*workflowTask))
	case *resetStickinessTask:
		return wtp.processResetStickinessTask(task.(*resetStickinessTask))
	default:
		panic("unknown task type.")
	}
}

func (wtp *workflowTaskPoller) processWorkflowTask(task *workflowTask) error {
	if task.task == nil {
		// We didn't have task, poll might have timeout.
		traceLog(func() {
			wtp.logger.Debug("Workflow task unavailable")
		})
		return nil
	}

	doneCh := make(chan struct{})
	laResultCh := make(chan *localActivityResult)
	// close doneCh so local activity worker won't get blocked forever when trying to send back result to laResultCh.
	defer close(doneCh)

	for {
		var response *apiv1.RespondDecisionTaskCompletedResponse
		startTime := time.Now()
		task.doneCh = doneCh
		task.laResultCh = laResultCh
		completedRequest, err := wtp.taskHandler.ProcessWorkflowTask(
			task,
			func(response interface{}, startTime time.Time) (*workflowTask, error) {
				wtp.logger.Debug("Force RespondDecisionTaskCompleted.", zap.Int64("TaskStartedEventID", task.task.GetStartedEventId()))
				wtp.metricsScope.Counter(metrics.DecisionTaskForceCompleted).Inc(1)
				heartbeatResponse, err := wtp.RespondTaskCompletedWithMetrics(response, nil, task.task, startTime)
				if err != nil {
					return nil, err
				}
				if heartbeatResponse == nil || heartbeatResponse.DecisionTask == nil {
					return nil, nil
				}
				task := wtp.toWorkflowTask(heartbeatResponse.DecisionTask)
				task.doneCh = doneCh
				task.laResultCh = laResultCh
				return task, nil
			},
		)
		if completedRequest == nil && err == nil {
			return nil
		}
		if _, ok := err.(decisionHeartbeatError); ok {
			return err
		}
		response, err = wtp.RespondTaskCompletedWithMetrics(completedRequest, err, task.task, startTime)
		if err != nil {
			return err
		}

		if response == nil || response.DecisionTask == nil {
			return nil
		}

		// we are getting new decision task, so reset the workflowTask and continue process the new one
		task = wtp.toWorkflowTask(response.DecisionTask)
	}
}

func (wtp *workflowTaskPoller) processResetStickinessTask(rst *resetStickinessTask) error {
	tchCtx, cancel, opt := newChannelContext(context.Background(), wtp.featureFlags)
	defer cancel()
	wtp.metricsScope.Counter(metrics.StickyCacheEvict).Inc(1)
	if _, err := wtp.service.ResetStickyTaskList(tchCtx, rst.task, opt...); err != nil {
		wtp.logger.Warn("ResetStickyTaskList failed",
			zap.String(tagWorkflowID, rst.task.WorkflowExecution.GetWorkflowId()),
			zap.String(tagRunID, rst.task.WorkflowExecution.GetRunId()),
			zap.Error(err))
		return api.ConvertError(err)
	}

	return nil
}

func (wtp *workflowTaskPoller) RespondTaskCompletedWithMetrics(completedRequest interface{}, taskErr error, task *apiv1.PollForDecisionTaskResponse, startTime time.Time) (response *apiv1.RespondDecisionTaskCompletedResponse, err error) {

	metricsScope := wtp.metricsScope.GetTaggedScope(tagWorkflowType, task.WorkflowType.GetName())
	if taskErr != nil {
		metricsScope.Counter(metrics.DecisionExecutionFailedCounter).Inc(1)
		wtp.logger.Warn("Failed to process decision task.",
			zap.String(tagWorkflowType, task.WorkflowType.GetName()),
			zap.String(tagWorkflowID, task.WorkflowExecution.GetWorkflowId()),
			zap.String(tagRunID, task.WorkflowExecution.GetRunId()),
			zap.Error(taskErr))
		// convert err to DecisionTaskFailed
		completedRequest = errorToFailDecisionTask(task.TaskToken, taskErr, wtp.identity)
	} else {
		metricsScope.Counter(metrics.DecisionTaskCompletedCounter).Inc(1)
	}

	metricsScope.Timer(metrics.DecisionExecutionLatency).Record(time.Now().Sub(startTime))

	responseStartTime := time.Now()
	if response, err = wtp.RespondTaskCompleted(completedRequest, task); err != nil {
		metricsScope.Counter(metrics.DecisionResponseFailedCounter).Inc(1)
		return
	}
	metricsScope.Timer(metrics.DecisionResponseLatency).Record(time.Now().Sub(responseStartTime))

	return
}

func (wtp *workflowTaskPoller) RespondTaskCompleted(completedRequest interface{}, task *apiv1.PollForDecisionTaskResponse) (response *apiv1.RespondDecisionTaskCompletedResponse, err error) {
	ctx := context.Background()
	// Respond task completion.
	err = backoff.Retry(ctx,
		func() error {
			tchCtx, cancel, opt := newChannelContext(ctx, wtp.featureFlags)
			defer cancel()
			var err1 error
			switch request := completedRequest.(type) {
			case *apiv1.RespondDecisionTaskFailedRequest:
				// Only fail decision on first attempt, subsequent failure on the same decision task will timeout.
				// This is to avoid spin on the failed decision task. Checking Attempt not nil for older server.
				if task.GetAttempt() == 0 {
					_, err1 = wtp.service.RespondDecisionTaskFailed(tchCtx, request, opt...)
					err1 = api.ConvertError(err1)
					if err1 != nil {
						traceLog(func() {
							wtp.logger.Debug("RespondDecisionTaskFailed failed.", zap.Error(err1))
						})
					}
				}
			case *apiv1.RespondDecisionTaskCompletedRequest:
				if request.StickyAttributes == nil && !wtp.disableStickyExecution {
					request.StickyAttributes = &apiv1.StickyExecutionAttributes{
						WorkerTaskList:         &apiv1.TaskList{Name: getWorkerTaskList(wtp.stickyUUID)},
						ScheduleToStartTimeout: api.DurationToProto(wtp.StickyScheduleToStartTimeout),
					}
				} else {
					request.ReturnNewDecisionTask = false
				}
				var activityTasks []*locallyDispatchedActivityTask
				if wtp.ldaTunnel != nil {
					for _, decision := range request.Decisions {
						attr := decision.GetScheduleActivityTaskDecisionAttributes()
						if attr != nil && wtp.taskListName == attr.TaskList.GetName() {
							// assume the activity type is in registry otherwise the activity would be failed and retried from server
							activityTask := &locallyDispatchedActivityTask{
								readyCh:                make(chan bool, 1),
								ActivityId:             attr.ActivityId,
								ActivityType:           attr.ActivityType,
								Input:                  attr.Input,
								Header:                 attr.Header,
								WorkflowDomain:         wtp.domain,
								ScheduleToCloseTimeout: attr.ScheduleToCloseTimeout,
								StartToCloseTimeout:    attr.StartToCloseTimeout,
								HeartbeatTimeout:       attr.HeartbeatTimeout,
								WorkflowExecution:      task.WorkflowExecution,
								WorkflowType:           task.WorkflowType,
							}
							if wtp.ldaTunnel.sendTask(activityTask) {
								wtp.metricsScope.Counter(metrics.ActivityLocalDispatchSucceedCounter).Inc(1)
								decision.GetScheduleActivityTaskDecisionAttributes().RequestLocalDispatch = true
								activityTasks = append(activityTasks, activityTask)
							} else {
								// all pollers are busy - no room to optimize
								wtp.metricsScope.Counter(metrics.ActivityLocalDispatchFailedCounter).Inc(1)
							}
						}
					}
				}
				defer func() {
					for _, at := range activityTasks {
						started := false
						if response != nil && err1 == nil {
							if adl, ok := response.ActivitiesToDispatchLocally[at.ActivityId]; ok {
								at.ScheduledTime = adl.ScheduledTime
								at.StartedTime = adl.StartedTime
								at.ScheduledTimeOfThisAttempt = adl.ScheduledTimeOfThisAttempt
								at.TaskToken = adl.TaskToken
								started = true
							}
						}
						at.readyCh <- started
					}
				}()
				response, err1 = wtp.service.RespondDecisionTaskCompleted(tchCtx, request, opt...)
				err1 = api.ConvertError(err1)
				if err1 != nil {
					traceLog(func() {
						wtp.logger.Debug("RespondDecisionTaskCompleted failed.", zap.Error(err1))
					})
				}

			case *apiv1.RespondQueryTaskCompletedRequest:
				_, err1 = wtp.service.RespondQueryTaskCompleted(tchCtx, request, opt...)
				err1 = api.ConvertError(err1)
				if err1 != nil {
					traceLog(func() {
						wtp.logger.Debug("RespondQueryTaskCompleted failed.", zap.Error(err1))
					})
				}
			default:
				// should not happen
				panic("unknown request type from ProcessWorkflowTask()")
			}

			return err1
		}, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)

	return
}

func newLocalActivityPoller(params workerExecutionParameters, laTunnel *localActivityTunnel) *localActivityTaskPoller {
	handler := &localActivityTaskHandler{
		userContext:        params.UserContext,
		metricsScope:       metrics.NewTaggedScope(params.MetricsScope),
		logger:             params.Logger,
		dataConverter:      params.DataConverter,
		contextPropagators: params.ContextPropagators,
		tracer:             params.Tracer,
	}
	return &localActivityTaskPoller{
		basePoller:   basePoller{shutdownC: params.WorkerStopChannel},
		handler:      handler,
		metricsScope: params.MetricsScope,
		logger:       params.Logger,
		laTunnel:     laTunnel,
	}
}

func (latp *localActivityTaskPoller) PollTask() (interface{}, error) {
	return latp.laTunnel.getTask(), nil
}

func (latp *localActivityTaskPoller) ProcessTask(task interface{}) error {
	if latp.shuttingDown() {
		return errShutdown
	}

	result := latp.handler.executeLocalActivityTask(task.(*localActivityTask))
	// We need to send back the local activity result to unblock workflowTaskPoller.processWorkflowTask() which is
	// synchronously listening on the laResultCh. We also want to make sure we don't block here forever in case
	// processWorkflowTask() already returns and nobody is receiving from laResultCh. We guarantee that doneCh is closed
	// before returning from workflowTaskPoller.processWorkflowTask().
	select {
	case result.task.workflowTask.laResultCh <- result:
		return nil
	case <-result.task.workflowTask.doneCh:
		// processWorkflowTask() already returns, just drop this local activity result.
		return nil
	}
}

func (lath *localActivityTaskHandler) executeLocalActivityTask(task *localActivityTask) (result *localActivityResult) {
	workflowType := task.params.WorkflowInfo.WorkflowType.Name
	activityType := task.params.ActivityType
	metricsScope := getMetricsScopeForLocalActivity(lath.metricsScope, workflowType, activityType)

	metricsScope.Counter(metrics.LocalActivityTotalCounter).Inc(1)

	ae := activityExecutor{name: activityType, fn: task.params.ActivityFn}

	rootCtx := lath.userContext
	if rootCtx == nil {
		rootCtx = context.Background()
	}

	workflowTypeLocal := task.params.WorkflowInfo.WorkflowType

	ctx := context.WithValue(rootCtx, activityEnvContextKey, &activityEnvironment{
		workflowType:      &workflowTypeLocal,
		workflowDomain:    task.params.WorkflowInfo.Domain,
		taskList:          task.params.WorkflowInfo.TaskListName,
		activityType:      ActivityType{Name: activityType},
		activityID:        fmt.Sprintf("%v", task.activityID),
		workflowExecution: task.params.WorkflowInfo.WorkflowExecution,
		logger:            lath.logger,
		metricsScope:      metricsScope,
		isLocalActivity:   true,
		dataConverter:     lath.dataConverter,
		attempt:           task.attempt,
	})

	// propagate context information into the local activity activity context from the headers
	for _, ctxProp := range lath.contextPropagators {
		var err error
		if ctx, err = ctxProp.Extract(ctx, NewHeaderReader(task.header)); err != nil {
			result = &localActivityResult{
				task:   task,
				result: nil,
				err:    fmt.Errorf("unable to propagate context %v", err),
			}
			return result
		}
	}

	// panic handler
	defer func() {
		if p := recover(); p != nil {
			topLine := fmt.Sprintf("local activity for %s [panic]:", activityType)
			st := getStackTraceRaw(topLine, 7, 0)
			lath.logger.Error("LocalActivity panic.",
				zap.String(tagWorkflowID, task.params.WorkflowInfo.WorkflowExecution.ID),
				zap.String(tagRunID, task.params.WorkflowInfo.WorkflowExecution.RunID),
				zap.String(tagActivityType, activityType),
				zap.String(tagPanicError, fmt.Sprintf("%v", p)),
				zap.String(tagPanicStack, st))
			metricsScope.Counter(metrics.LocalActivityPanicCounter).Inc(1)
			panicErr := newPanicError(p, st)
			result = &localActivityResult{
				task:   task,
				result: nil,
				err:    panicErr,
			}
		}
		if result.err != nil {
			metricsScope.Counter(metrics.LocalActivityFailedCounter).Inc(1)
		}
	}()

	timeoutDuration := time.Duration(task.params.ScheduleToCloseTimeoutSeconds) * time.Second
	deadline := time.Now().Add(timeoutDuration)
	if task.attempt > 0 && !task.expireTime.IsZero() && task.expireTime.Before(deadline) {
		// this is attempt and expire time is before SCHEDULE_TO_CLOSE timeout
		deadline = task.expireTime
	}

	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	task.Lock()
	if task.canceled {
		task.Unlock()
		return &localActivityResult{err: ErrCanceled, task: task}
	}
	task.cancelFunc = cancel
	task.Unlock()

	var laResult []byte
	var err error
	doneCh := make(chan struct{})
	go func(ch chan struct{}) {
		laStartTime := time.Now()
		ctx, span := createOpenTracingActivitySpan(ctx, lath.tracer, time.Now(), task.params.ActivityType, task.params.WorkflowInfo.WorkflowExecution.ID, task.params.WorkflowInfo.WorkflowExecution.RunID)
		defer span.Finish()
		laResult, err = ae.ExecuteWithActualArgs(ctx, task.params.InputArgs)
		executionLatency := time.Now().Sub(laStartTime)
		close(ch)
		metricsScope.Timer(metrics.LocalActivityExecutionLatency).Record(executionLatency)
		if executionLatency > timeoutDuration {
			// If local activity takes longer than expected timeout, the context would already be DeadlineExceeded and
			// the result would be discarded. Print a warning in this case.
			lath.logger.Warn("LocalActivity takes too long to complete.",
				zap.String("LocalActivityID", task.activityID),
				zap.String("LocalActivityType", activityType),
				zap.Int32("ScheduleToCloseTimeoutSeconds", task.params.ScheduleToCloseTimeoutSeconds),
				zap.Duration("ActualExecutionDuration", executionLatency))
		}
	}(doneCh)

Wait_Result:
	select {
	case <-ctx.Done():
		select {
		case <-doneCh:
			// double check if result is ready.
			break Wait_Result
		default:
		}

		// context is done
		if ctx.Err() == context.Canceled {
			metricsScope.Counter(metrics.LocalActivityCanceledCounter).Inc(1)
			return &localActivityResult{err: ErrCanceled, task: task}
		} else if ctx.Err() == context.DeadlineExceeded {
			metricsScope.Counter(metrics.LocalActivityTimeoutCounter).Inc(1)
			return &localActivityResult{err: ErrDeadlineExceeded, task: task}
		} else {
			// should not happen
			return &localActivityResult{err: NewCustomError("unexpected context done"), task: task}
		}
	case <-doneCh:
		// local activity completed
	}

	return &localActivityResult{result: laResult, err: err, task: task}
}

func (wtp *workflowTaskPoller) release(kind apiv1.TaskListKind) {
	if wtp.disableStickyExecution {
		return
	}

	wtp.requestLock.Lock()
	if kind == apiv1.TaskListKind_TASK_LIST_KIND_STICKY {
		wtp.pendingStickyPollCount--
	} else {
		wtp.pendingRegularPollCount--
	}
	wtp.requestLock.Unlock()
}

func (wtp *workflowTaskPoller) updateBacklog(taskListKind apiv1.TaskListKind, backlogCountHint int64) {
	if taskListKind == apiv1.TaskListKind_TASK_LIST_KIND_NORMAL || wtp.disableStickyExecution {
		// we only care about sticky backlog for now.
		return
	}
	wtp.requestLock.Lock()
	wtp.stickyBacklog = backlogCountHint
	wtp.requestLock.Unlock()
}

// getNextPollRequest returns appropriate next poll request based on poller configuration.
// Simple rules:
// 1) if sticky execution is disabled, always poll for regular task list
// 2) otherwise:
//   2.1) if sticky task list has backlog, always prefer to process sticky task first
//   2.2) poll from the task list that has less pending requests (prefer sticky when they are the same).
// TODO: make this more smart to auto adjust based on poll latency
func (wtp *workflowTaskPoller) getNextPollRequest() (request *apiv1.PollForDecisionTaskRequest) {
	taskListName := wtp.taskListName
	taskListKind := apiv1.TaskListKind_TASK_LIST_KIND_NORMAL
	if !wtp.disableStickyExecution {
		wtp.requestLock.Lock()
		if wtp.stickyBacklog > 0 || wtp.pendingStickyPollCount <= wtp.pendingRegularPollCount {
			wtp.pendingStickyPollCount++
			taskListName = getWorkerTaskList(wtp.stickyUUID)
			taskListKind = apiv1.TaskListKind_TASK_LIST_KIND_STICKY
		} else {
			wtp.pendingRegularPollCount++
		}
		wtp.requestLock.Unlock()
	}

	return &apiv1.PollForDecisionTaskRequest{
		Domain:         wtp.domain,
		TaskList:       &apiv1.TaskList{
			Name: taskListName,
			Kind: taskListKind,
		},
		Identity:       wtp.identity,
		BinaryChecksum: getBinaryChecksum(),
	}
}

// Poll for a single workflow task from the service
func (wtp *workflowTaskPoller) poll(ctx context.Context) (interface{}, error) {
	startTime := time.Now()
	wtp.metricsScope.Counter(metrics.DecisionPollCounter).Inc(1)

	traceLog(func() {
		wtp.logger.Debug("workflowTaskPoller::Poll")
	})

	request := wtp.getNextPollRequest()
	defer wtp.release(request.TaskList.GetKind())

	response, err := wtp.service.PollForDecisionTask(ctx, request, getYarpcCallOptions(wtp.featureFlags)...)
	err = api.ConvertError(err)
	if err != nil {
		if isServiceTransientError(err) {
			wtp.metricsScope.Counter(metrics.DecisionPollTransientFailedCounter).Inc(1)
		} else {
			wtp.metricsScope.Counter(metrics.DecisionPollFailedCounter).Inc(1)
		}
		wtp.updateBacklog(request.TaskList.GetKind(), 0)
		return nil, err
	}

	if response == nil || len(response.TaskToken) == 0 {
		wtp.metricsScope.Counter(metrics.DecisionPollNoTaskCounter).Inc(1)
		wtp.updateBacklog(request.TaskList.GetKind(), 0)
		return &workflowTask{}, nil
	}

	wtp.updateBacklog(request.TaskList.GetKind(), response.GetBacklogCountHint())

	task := wtp.toWorkflowTask(response)
	traceLog(func() {
		var firstEventID int64 = -1
		if response.History != nil && len(response.History.Events) > 0 {
			firstEventID = response.History.Events[0].GetEventId()
		}
		wtp.logger.Debug("workflowTaskPoller::Poll Succeed",
			zap.Int64("StartedEventID", response.GetStartedEventId()),
			zap.Int64("Attempt", response.GetAttempt()),
			zap.Int64("FirstEventID", firstEventID),
			zap.Bool("IsQueryTask", response.Query != nil))
	})

	metricsScope := wtp.metricsScope.GetTaggedScope(tagWorkflowType, response.WorkflowType.GetName())
	metricsScope.Counter(metrics.DecisionPollSucceedCounter).Inc(1)
	metricsScope.Timer(metrics.DecisionPollLatency).Record(time.Now().Sub(startTime))

	startedTime := api.TimeFromProto(response.StartedTime)
	scheduledTime := api.TimeFromProto(response.ScheduledTime)
	scheduledToStartLatency := startedTime.Sub(scheduledTime)
	metricsScope.Timer(metrics.DecisionScheduledToStartLatency).Record(scheduledToStartLatency)
	return task, nil
}

func (wtp *workflowTaskPoller) toWorkflowTask(response *apiv1.PollForDecisionTaskResponse) *workflowTask {
	startEventID := response.GetStartedEventId()
	nextEventID := response.GetNextEventId()
	if nextEventID != 0 &&
		startEventID != 0 &&
		nextEventID-1 != startEventID {
		wtp.logger.Warn("Invalid PollForDecisionTaskResponse, nextEventID doesn't match startedEventID",
			zap.Int64("StartedEventID", startEventID),
			zap.Int64("NextEventID", nextEventID),
		)
		wtp.metricsScope.Counter(metrics.DecisionPollInvalidCounter).Inc(1)
	}
	historyIterator := &historyIteratorImpl{
		nextPageToken:  response.NextPageToken,
		execution:      response.WorkflowExecution,
		domain:         wtp.domain,
		service:        wtp.service,
		metricsScope:   wtp.metricsScope,
		startedEventID: startEventID,
		maxEventID:     nextEventID - 1,
		featureFlags:   wtp.featureFlags,
	}
	task := &workflowTask{
		task:            response,
		historyIterator: historyIterator,
	}
	return task
}

func (h *historyIteratorImpl) GetNextPage() (*apiv1.History, error) {
	if h.iteratorFunc == nil {
		h.iteratorFunc = newGetHistoryPageFunc(
			context.Background(),
			h.service,
			h.domain,
			h.execution,
			h.startedEventID,
			h.maxEventID,
			h.metricsScope,
			h.featureFlags)
	}

	history, token, err := h.iteratorFunc(h.nextPageToken)
	if err != nil {
		return nil, err
	}
	h.nextPageToken = token
	return history, nil
}

func (h *historyIteratorImpl) Reset() {
	h.nextPageToken = nil
}

func (h *historyIteratorImpl) HasNextPage() bool {
	return h.nextPageToken != nil
}

func newGetHistoryPageFunc(
	ctx context.Context,
	service api.Interface,
	domain string,
	execution *apiv1.WorkflowExecution,
	atDecisionTaskCompletedEventID int64,
	maxEventID int64,
	metricsScope tally.Scope,
	featureFlags FeatureFlags,
) func(nextPageToken []byte) (*apiv1.History, []byte, error) {
	return func(nextPageToken []byte) (*apiv1.History, []byte, error) {
		metricsScope.Counter(metrics.WorkflowGetHistoryCounter).Inc(1)
		startTime := time.Now()
		var resp *apiv1.GetWorkflowExecutionHistoryResponse
		err := backoff.Retry(ctx,
			func() error {
				tchCtx, cancel, opt := newChannelContext(ctx, featureFlags)
				defer cancel()

				var err1 error
				resp, err1 = service.GetWorkflowExecutionHistory(tchCtx, &apiv1.GetWorkflowExecutionHistoryRequest{
					Domain:            domain,
					WorkflowExecution: execution,
					NextPageToken:     nextPageToken,
				}, opt...)
				return api.ConvertError(err1)
			}, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)
		if err != nil {
			metricsScope.Counter(metrics.WorkflowGetHistoryFailedCounter).Inc(1)
			return nil, nil, err
		}

		metricsScope.Counter(metrics.WorkflowGetHistorySucceedCounter).Inc(1)
		metricsScope.Timer(metrics.WorkflowGetHistoryLatency).Record(time.Now().Sub(startTime))

		var h *apiv1.History

		if resp.RawHistory != nil {
			var err1 error
			h, err1 = serializer.DeserializeBlobDataToHistoryEvents(resp.RawHistory, apiv1.EventFilterType_EVENT_FILTER_TYPE_ALL_EVENT)
			if err1 != nil {
				return nil, nil, err1
			}
		} else {
			h = resp.History
		}

		// TODO: is this check valid/useful? atDecisionTaskCompletedEventID is startedEventID in pollForDecisionTaskResponse and
		// - For decision tasks, since there's only one inflight decision task, there won't be any event after startEventID.
		//   Those events will be buffered. If there're too many buffer events, the current decision will be failed and events passed
		//   startEventID may be returned. In that case, the last event after truncation is still decision task started event not completed.
		// - For query tasks startEventID is not assigned so this check is never executed.
		if shouldTruncateHistory(h, atDecisionTaskCompletedEventID) {
			first := h.Events[0].GetEventId() // eventIds start from 1
			h.Events = h.Events[:atDecisionTaskCompletedEventID-first+1]
			if h.Events[len(h.Events)-1].GetDecisionTaskCompletedEventAttributes() == nil {
				return nil, nil, fmt.Errorf("newGetHistoryPageFunc: atDecisionTaskCompletedEventID(%v) "+
					"points to event that is not DecisionTaskCompleted", atDecisionTaskCompletedEventID)
			}
			return h, nil, nil
		}

		// TODO: Apply the check to decision tasks (remove the last condition)
		// after validating maxEventID always equal to atDecisionTaskCompletedEventID (startedEventID).
		// For now only apply to query task to be safe.
		if shouldTruncateHistory(h, maxEventID) && isQueryTask(atDecisionTaskCompletedEventID) {
			first := h.Events[0].GetEventId()
			h.Events = h.Events[:maxEventID-first+1]
			return h, nil, nil
		}

		return h, resp.NextPageToken, nil
	}
}

func shouldTruncateHistory(h *apiv1.History, maxEventID int64) bool {
	size := len(h.Events)
	return size > 0 && maxEventID > 0 && h.Events[size-1].GetEventId() > maxEventID
}

func isQueryTask(atDecisionTaskCompletedEventID int64) bool {
	return atDecisionTaskCompletedEventID == 0
}

func newActivityTaskPoller(taskHandler ActivityTaskHandler, service api.Interface,
	domain string, params workerExecutionParameters) *activityTaskPoller {

	activityTaskPoller := &activityTaskPoller{
		basePoller:          basePoller{shutdownC: params.WorkerStopChannel},
		taskHandler:         taskHandler,
		service:             service,
		domain:              domain,
		taskListName:        params.TaskList,
		identity:            params.Identity,
		logger:              params.Logger,
		metricsScope:        metrics.NewTaggedScope(params.MetricsScope),
		activitiesPerSecond: params.TaskListActivitiesPerSecond,
		featureFlags:        params.FeatureFlags,
	}
	return activityTaskPoller
}

// Poll for a single activity task from the service
func (atp *activityTaskPoller) poll(ctx context.Context) (*apiv1.PollForActivityTaskResponse, time.Time, error) {

	atp.metricsScope.Counter(metrics.ActivityPollCounter).Inc(1)
	startTime := time.Now()

	traceLog(func() {
		atp.logger.Debug("activityTaskPoller::Poll")
	})
	request := &apiv1.PollForActivityTaskRequest{
		Domain:           atp.domain,
		TaskList:         &apiv1.TaskList{Name: atp.taskListName},
		Identity:         atp.identity,
		TaskListMetadata: &apiv1.TaskListMetadata{MaxTasksPerSecond: &types.DoubleValue{Value: atp.activitiesPerSecond}},
	}
	response, err := atp.service.PollForActivityTask(ctx, request, getYarpcCallOptions(atp.featureFlags)...)
	err = api.ConvertError(err)

	if err != nil {
		if isServiceTransientError(err) {
			atp.metricsScope.Counter(metrics.ActivityPollTransientFailedCounter).Inc(1)
		} else {
			atp.metricsScope.Counter(metrics.ActivityPollFailedCounter).Inc(1)
		}
		return nil, startTime, err
	}
	if response == nil || len(response.TaskToken) == 0 {
		atp.metricsScope.Counter(metrics.ActivityPollNoTaskCounter).Inc(1)
		return nil, startTime, nil
	}

	return response, startTime, err
}

type pollFunc func(ctx context.Context) (*apiv1.PollForActivityTaskResponse, time.Time, error)

func (atp *activityTaskPoller) pollWithMetricsFunc(
	pollFunc pollFunc) func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (interface{}, error) { return atp.pollWithMetrics(ctx, pollFunc) }
}

func (atp *activityTaskPoller) pollWithMetrics(ctx context.Context,
	pollFunc func(ctx context.Context) (*apiv1.PollForActivityTaskResponse, time.Time, error)) (interface{}, error) {

	response, startTime, err := pollFunc(ctx)
	if err != nil {
		return nil, err
	}
	if response == nil || len(response.TaskToken) == 0 {
		return &activityTask{}, nil
	}

	workflowType := response.WorkflowType.GetName()
	activityType := response.ActivityType.GetName()
	metricsScope := getMetricsScopeForActivity(atp.metricsScope, workflowType, activityType)
	metricsScope.Counter(metrics.ActivityPollSucceedCounter).Inc(1)
	metricsScope.Timer(metrics.ActivityPollLatency).Record(time.Now().Sub(startTime))

	startedTime := api.TimeFromProto(response.StartedTime)
	scheduledTimeOfThisAttempt := api.TimeFromProto(response.ScheduledTimeOfThisAttempt)
	scheduledToStartLatency := startedTime.Sub(scheduledTimeOfThisAttempt)
	metricsScope.Timer(metrics.ActivityScheduledToStartLatency).Record(scheduledToStartLatency)

	return &activityTask{task: response, pollStartTime: startTime}, nil
}

// PollTask polls a new task
func (atp *activityTaskPoller) PollTask() (interface{}, error) {
	// Get the task.
	activityTask, err := atp.doPoll(atp.featureFlags, atp.pollWithMetricsFunc(atp.poll))
	if err != nil {
		return nil, err
	}
	return activityTask, nil
}

// ProcessTask processes a new task
func (atp *activityTaskPoller) ProcessTask(task interface{}) error {
	if atp.shuttingDown() {
		return errShutdown
	}

	activityTask := task.(*activityTask)
	if activityTask.task == nil {
		// We didn't have task, poll might have timeout.
		traceLog(func() {
			atp.logger.Debug("Activity task unavailable")
		})
		return nil
	}

	workflowType := activityTask.task.WorkflowType.GetName()
	activityType := activityTask.task.ActivityType.GetName()
	metricsScope := getMetricsScopeForActivity(atp.metricsScope, workflowType, activityType)

	executionStartTime := time.Now()
	// Process the activity task.
	request, err := atp.taskHandler.Execute(atp.taskListName, activityTask.task)
	if err != nil {
		metricsScope.Counter(metrics.ActivityExecutionFailedCounter).Inc(1)
		return err
	}
	metricsScope.Timer(metrics.ActivityExecutionLatency).Record(time.Now().Sub(executionStartTime))

	if request == ErrActivityResultPending {
		return nil
	}

	// if worker is shutting down, don't bother reporting activity completion
	if atp.shuttingDown() {
		return errShutdown
	}

	responseStartTime := time.Now()
	reportErr := reportActivityComplete(context.Background(), atp.service, request, metricsScope, atp.featureFlags)
	if reportErr != nil {
		metricsScope.Counter(metrics.ActivityResponseFailedCounter).Inc(1)
		traceLog(func() {
			atp.logger.Debug("reportActivityComplete failed", zap.Error(reportErr))
		})
		return reportErr
	}

	metricsScope.Timer(metrics.ActivityResponseLatency).Record(time.Now().Sub(responseStartTime))
	metricsScope.Timer(metrics.ActivityEndToEndLatency).Record(time.Now().Sub(activityTask.pollStartTime))
	return nil
}

func newLocallyDispatchedActivityTaskPoller(taskHandler ActivityTaskHandler, service api.Interface,
	domain string, params workerExecutionParameters) *locallyDispatchedActivityTaskPoller {
	locallyDispatchedActivityTaskPoller := &locallyDispatchedActivityTaskPoller{
		activityTaskPoller: *newActivityTaskPoller(taskHandler, service, domain, params),
		ldaTunnel:          newLocallyDispatchedActivityTunnel(params.WorkerStopChannel),
	}
	return locallyDispatchedActivityTaskPoller
}

// PollTask polls a new task
func (atp *locallyDispatchedActivityTaskPoller) PollTask() (interface{}, error) {
	// Get the task.
	activityTask, err := atp.doPoll(atp.featureFlags, atp.pollWithMetricsFunc(atp.pollLocallyDispatchedActivity))
	if err != nil {
		return nil, err
	}

	return activityTask, nil
}

func (atp *locallyDispatchedActivityTaskPoller) pollLocallyDispatchedActivity(ctx context.Context) (*apiv1.PollForActivityTaskResponse, time.Time, error) {
	task := atp.ldaTunnel.getTask()
	atp.metricsScope.Counter(metrics.LocallyDispatchedActivityPollCounter).Inc(1)
	// consider to remove the poll latency metric for local dispatch as unnecessary
	startTime := time.Now()
	if task == nil {
		atp.metricsScope.Counter(metrics.LocallyDispatchedActivityPollNoTaskCounter).Inc(1)
		return nil, startTime, nil
	}
	// to be backwards compatible, update total poll counter if optimization succeeded only
	atp.metricsScope.Counter(metrics.ActivityPollCounter).Inc(1)
	atp.metricsScope.Counter(metrics.LocallyDispatchedActivityPollSucceedCounter).Inc(1)
	response := &apiv1.PollForActivityTaskResponse{}
	response.ActivityId = task.ActivityId
	response.ActivityType = task.ActivityType
	response.Header = task.Header
	response.Input = task.Input
	response.WorkflowExecution = task.WorkflowExecution
	response.ScheduledTimeOfThisAttempt = task.ScheduledTimeOfThisAttempt
	response.ScheduledTime = task.ScheduledTime
	response.ScheduleToCloseTimeout = task.ScheduleToCloseTimeout
	response.StartedTime = task.StartedTime
	response.StartToCloseTimeout = task.StartToCloseTimeout
	response.HeartbeatTimeout = task.HeartbeatTimeout
	response.TaskToken = task.TaskToken
	response.WorkflowType = task.WorkflowType
	response.WorkflowDomain = task.WorkflowDomain
	response.Attempt = 0
	return response, startTime, nil
}

func reportActivityComplete(
	ctx context.Context,
	service api.Interface,
	request interface{},
	metricsScope tally.Scope,
	featureFlags FeatureFlags,
) error {
	if request == nil {
		// nothing to report
		return nil
	}

	var reportErr error
	switch request := request.(type) {
	case *apiv1.RespondActivityTaskCanceledRequest:
		reportErr = backoff.Retry(ctx,
			func() error {
				tchCtx, cancel, opt := newChannelContext(ctx, featureFlags)
				defer cancel()

				_, err := service.RespondActivityTaskCanceled(tchCtx, request, opt...)
				return api.ConvertError(err)
			}, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)
	case *apiv1.RespondActivityTaskFailedRequest:
		reportErr = backoff.Retry(ctx,
			func() error {
				tchCtx, cancel, opt := newChannelContext(ctx, featureFlags)
				defer cancel()

				_, err := service.RespondActivityTaskFailed(tchCtx, request, opt...)
				return api.ConvertError(err)
			}, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)
	case *apiv1.RespondActivityTaskCompletedRequest:
		reportErr = backoff.Retry(ctx,
			func() error {
				tchCtx, cancel, opt := newChannelContext(ctx, featureFlags)
				defer cancel()

				_, err := service.RespondActivityTaskCompleted(tchCtx, request, opt...)
				return api.ConvertError(err)
			}, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)
	}
	if reportErr == nil {
		switch request.(type) {
		case *apiv1.RespondActivityTaskCanceledRequest:
			metricsScope.Counter(metrics.ActivityTaskCanceledCounter).Inc(1)
		case *apiv1.RespondActivityTaskFailedRequest:
			metricsScope.Counter(metrics.ActivityTaskFailedCounter).Inc(1)
		case *apiv1.RespondActivityTaskCompletedRequest:
			metricsScope.Counter(metrics.ActivityTaskCompletedCounter).Inc(1)
		}
	}

	return reportErr
}

func reportActivityCompleteByID(
	ctx context.Context,
	service api.Interface,
	request interface{},
	metricsScope tally.Scope,
	featureFlags FeatureFlags,
) error {
	if request == nil {
		// nothing to report
		return nil
	}

	var reportErr error
	switch request := request.(type) {
	case *apiv1.RespondActivityTaskCanceledByIDRequest:
		reportErr = backoff.Retry(ctx,
			func() error {
				tchCtx, cancel, opt := newChannelContext(ctx, featureFlags)
				defer cancel()

				_, err := service.RespondActivityTaskCanceledByID(tchCtx, request, opt...)
				return api.ConvertError(err)
			}, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)
	case *apiv1.RespondActivityTaskFailedByIDRequest:
		reportErr = backoff.Retry(ctx,
			func() error {
				tchCtx, cancel, opt := newChannelContext(ctx, featureFlags)
				defer cancel()

				_, err := service.RespondActivityTaskFailedByID(tchCtx, request, opt...)
				return api.ConvertError(err)
			}, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)
	case *apiv1.RespondActivityTaskCompletedByIDRequest:
		reportErr = backoff.Retry(ctx,
			func() error {
				tchCtx, cancel, opt := newChannelContext(ctx, featureFlags)
				defer cancel()

				_, err := service.RespondActivityTaskCompletedByID(tchCtx, request, opt...)
				return api.ConvertError(err)
			}, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)
	}
	if reportErr == nil {
		switch request.(type) {
		case *apiv1.RespondActivityTaskCanceledByIDRequest:
			metricsScope.Counter(metrics.ActivityTaskCanceledByIDCounter).Inc(1)
		case *apiv1.RespondActivityTaskFailedByIDRequest:
			metricsScope.Counter(metrics.ActivityTaskFailedByIDCounter).Inc(1)
		case *apiv1.RespondActivityTaskCompletedByIDRequest:
			metricsScope.Counter(metrics.ActivityTaskCompletedByIDCounter).Inc(1)
		}
	}

	return reportErr
}

func convertActivityResultToRespondRequest(identity string, taskToken, result []byte, err error,
	dataConverter DataConverter) interface{} {
	if err == ErrActivityResultPending {
		// activity result is pending and will be completed asynchronously.
		// nothing to report at this point
		return ErrActivityResultPending
	}

	if err == nil {
		return &apiv1.RespondActivityTaskCompletedRequest{
			TaskToken: taskToken,
			Result:    &apiv1.Payload{Data: result},
			Identity:  identity}
	}

	reason, details := getErrorDetails(err, dataConverter)
	if _, ok := err.(*CanceledError); ok || err == context.Canceled {
		return &apiv1.RespondActivityTaskCanceledRequest{
			TaskToken: taskToken,
			Details:   &apiv1.Payload{Data: details},
			Identity:  identity}
	}

	return &apiv1.RespondActivityTaskFailedRequest{
		TaskToken: taskToken,
		Failure: &apiv1.Failure{
			Reason: reason,
			Details: details,
		},
		Identity:  identity}
}

func convertActivityResultToRespondRequestByID(identity, domain, workflowID, runID, activityID string,
	result []byte, err error, dataConverter DataConverter) interface{} {
	if err == ErrActivityResultPending {
		// activity result is pending and will be completed asynchronously.
		// nothing to report at this point
		return nil
	}

	if err == nil {
		return &apiv1.RespondActivityTaskCompletedByIDRequest{
			Domain:     domain,
			WorkflowExecution: &apiv1.WorkflowExecution{
				WorkflowId: workflowID,
				RunId: runID,
			},
			ActivityId: activityID,
			Result:     &apiv1.Payload{Data: result},
			Identity:   identity}
	}

	reason, details := getErrorDetails(err, dataConverter)
	if _, ok := err.(*CanceledError); ok || err == context.Canceled {
		return &apiv1.RespondActivityTaskCanceledByIDRequest{
			Domain:     domain,
			WorkflowExecution: &apiv1.WorkflowExecution{
				WorkflowId: workflowID,
				RunId: runID,
			},
			ActivityId: activityID,
			Details:    &apiv1.Payload{Data: details},
			Identity:   identity}
	}

	return &apiv1.RespondActivityTaskFailedByIDRequest{
		Domain:     domain,
		WorkflowExecution: &apiv1.WorkflowExecution{
			WorkflowId: workflowID,
			RunId: runID,
		},
		ActivityId: activityID,
		Failure: &apiv1.Failure{
			Reason:     reason,
			Details:    details,
		},
		Identity: identity}
}
