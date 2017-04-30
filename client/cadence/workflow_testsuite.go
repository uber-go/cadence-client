package cadence

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/facebookgo/clock"
	"github.com/stretchr/testify/mock"
	m "github.com/uber-go/cadence-client/.gen/go/cadence"
	s "github.com/uber-go/cadence-client/.gen/go/shared"
	"github.com/uber-go/cadence-client/common"
	"github.com/uber-go/cadence-client/mocks"
	"github.com/uber/tchannel-go/thrift"
	"go.uber.org/zap"
)

const (
	// DefaultTestTaskList is the default task list used to register activity.
	DefaultTestTaskList = "default-test-tasklist"
	// DefaultWorkflowID is the default workflow id used by the TestWorkflowEnvironment
	DefaultTestWorkflowID = "default-test-workflow-id"
	// DefaultTestRunID is the default run id used by the TestWorkflowEnvironment
	DefaultTestRunID = "default-test-run-id"
)

type (
	timerHandle struct {
		callback resultHandler
		timer    *clock.Timer
	}

	activityHandle struct {
		callback    resultHandler
		isCancelled bool
		isDone      bool
	}

	callbackHandle struct {
		callback          func()
		startDecisionTask bool
	}

	activityExecutorWrapper struct {
		activity
		fn  interface{}
		env *TestWorkflowEnvironment
	}

	// WorkflowTestSuite is the test suite for workflow unit test. Workflow and activities that will be tested must be
	// registered through test suite. After registration, use WorkflowTestSuite.TestWorkflow() to test your workflow.
	WorkflowTestSuite struct {
		hostEnv              *hostEnvImpl
		registeredActivities map[string]map[string]bool // map of taskList -> (map of fnName of registered activities)
		overrodeActivities   map[string]interface{}     // map of registered-fnName -> mockActivityFn

		service       m.TChanWorkflowService
		workerOptions WorkerOptions
		logger        *zap.Logger
		clock         clock.Clock
	}

	// BlobResult is a type used to encapsulate encoded result from workflow/activity. Use BlobResult.GetResult() to
	// extract and decode result into typed value.
	BlobResult []byte

	// BlobArgs is a type used to encapsulate encoded arguments from workflow/activity. Use BlobArgs.GetArgs() to
	// extract and decode args into typed values.
	BlobArgs []byte

	TestWorkflowEnvironment struct {
		testSuite *WorkflowTestSuite

		service       m.TChanWorkflowService
		workerOptions WorkerOptions
		logger        *zap.Logger
		clock         clock.Clock

		workflowInfo          *WorkflowInfo
		workflowDef           workflowDefinition
		counterID             int32
		workflowCancelHandler func()

		locker              *sync.Mutex
		scheduledActivities map[string]*activityHandle
		scheduledTimers     map[string]*timerHandle

		onActivityTaskStartedListener func(ctx context.Context, args BlobArgs, activityType string)
		onActivityTaskEndedListener   func(result BlobResult, err error, activityType string)
		onTimerScheduledListener      func(timerID string, duration time.Duration)
		onTimerFiredListener          func(timerID string)
		onTimerCancelledListener      func(timerID string)

		testResult            BlobResult
		testError             error
		callbackChannel       chan callbackHandle
		isTestCompleted       bool
		autoStartDecisionTask bool
	}
)

// GetResult extract data from encoded blob to desired value type. valuePtr is pointer to the actual result struct.
func (b BlobResult) GetResult(valuePtr interface{}) error {
	return getHostEnvironment().decodeArg(b, valuePtr)
}

// GetArgs extract arguments from encoded blob to desired value type. valuePtrs are pointers to the actual arguments.
func (b BlobArgs) GetArgs(valuePtrs ...interface{}) error {
	return getHostEnvironment().decode(b, valuePtrs)
}

// NewWorkflowTestSuite create a new WorkflowTestSuite
func NewWorkflowTestSuite() *WorkflowTestSuite {
	suite := &WorkflowTestSuite{
		hostEnv: &hostEnvImpl{
			workflowFuncMap: make(map[string]interface{}),
			activityFuncMap: make(map[string]interface{}),
			encoding:        gobEncoding{},
		},
		registeredActivities: make(map[string]map[string]bool),
		overrodeActivities:   make(map[string]interface{}),
	}
	return suite
}

// SetLogger sets the logger that will be used when creating new TestWorkflowEnvironment from this suite
func (t *WorkflowTestSuite) SetLogger(logger *zap.Logger) *WorkflowTestSuite {
	t.logger = logger
	return t
}

// SetService sets the m.TChanWorkflowService that will be used when creating new TestWorkflowEnvironment from this suite
func (t *WorkflowTestSuite) SetService(service m.TChanWorkflowService) *WorkflowTestSuite {
	t.service = service
	return t
}

// SetClock sets the clock that will be used when creating new TestWorkflowEnvironment from this suite
func (t *WorkflowTestSuite) SetClock(clock clock.Clock) *WorkflowTestSuite {
	t.clock = clock
	return t
}

// SetWorkerOption sets the WorkerOptions that will be used when creating new TestWorkflowEnvironment from this suite
func (t *WorkflowTestSuite) SetWorkerOption(options WorkerOptions) *WorkflowTestSuite {
	t.workerOptions = options
	return t
}

// NewTestWorkflowEnvironment create a new TestWorkflowEnvironment that can be use to test activity.
// See more: TestWorkflowEnvironment.TestActivity()
func (t *WorkflowTestSuite) NewTestWorkflowEnvironment() *TestWorkflowEnvironment {
	env := &TestWorkflowEnvironment{
		testSuite:     t,
		service:       t.service,
		workerOptions: t.workerOptions,
		logger:        t.logger,
		clock:         t.clock,

		workflowInfo: &WorkflowInfo{
			WorkflowExecution: WorkflowExecution{
				ID:    DefaultTestWorkflowID,
				RunID: DefaultTestRunID,
			},
			WorkflowType: WorkflowType{Name: "workflow-type-not-specified"},
			TaskListName: DefaultTestTaskList,
		},

		locker:                &sync.Mutex{},
		scheduledActivities:   make(map[string]*activityHandle),
		scheduledTimers:       make(map[string]*timerHandle),
		callbackChannel:       make(chan callbackHandle, 100),
		autoStartDecisionTask: true,
	}

	if env.logger == nil {
		logger, _ := zap.NewDevelopment()
		env.logger = logger
	}

	if env.service == nil {
		mockService := new(mocks.TChanWorkflowService)
		mockHeartbeatFn := func(c thrift.Context, r *s.RecordActivityTaskHeartbeatRequest) *s.RecordActivityTaskHeartbeatResponse {
			activityID := string(r.TaskToken)
			env.locker.Lock()
			defer env.locker.Unlock()
			activityHandle, ok := env.scheduledActivities[activityID]
			if !ok {
				return nil
			}
			env.logger.Debug("RecordActivityTaskHeartbeat", zap.Bool("MockResponse_CancelRequested", activityHandle.isCancelled))
			return &s.RecordActivityTaskHeartbeatResponse{CancelRequested: common.BoolPtr(activityHandle.isCancelled)}
		}

		mockService.On("RecordActivityTaskHeartbeat", mock.Anything, mock.Anything).Return(mockHeartbeatFn, nil)
		env.service = mockService
	}
	if env.workerOptions == nil {
		env.workerOptions = NewWorkerOptions().SetLogger(env.logger)
	}

	if env.clock == nil {
		env.clock = clock.New()
	}

	return env
}

// RegisterWorkflow register a workflow to this suite
func (t *WorkflowTestSuite) RegisterWorkflow(workflowFunc interface{}) {
	err := t.hostEnv.RegisterWorkflow(workflowFunc)
	if err != nil {
		panic(err)
	}
}

// RegisterActivity register an activity to this suite using given task list
func (t *WorkflowTestSuite) RegisterActivity(activityFn interface{}, taskList string) {
	fnName := getFunctionName(activityFn)

	_, ok := t.hostEnv.activityFuncMap[fnName]
	if !ok {
		// activity not registered yet, register now
		err := t.hostEnv.RegisterActivity(activityFn)
		if err != nil {
			panic(err)
		}
	}

	taskListActivities, ok := t.registeredActivities[taskList]
	if !ok {
		// this is the first activity registered to this tasklist
		taskListActivities = make(map[string]bool)
		t.registeredActivities[taskList] = taskListActivities
	}

	taskListActivities[fnName] = true
}

// Override overrides an actual activity with a mock activity. The mock activity will be invoked in place where the
// actual activity should have been invoked.
func (t *WorkflowTestSuite) Override(activityFn, mockActivityFn interface{}) {
	// verify both functions are valid activity func
	actualFnType := reflect.TypeOf(activityFn)
	if err := validateFnFormat(actualFnType, false); err != nil {
		panic(err)
	}
	mockFnType := reflect.TypeOf(mockActivityFn)
	if err := validateFnFormat(mockFnType, false); err != nil {
		panic(err)
	}

	// verify signature of registeredActivityFn and mockActivityFn are the same.
	if actualFnType.NumIn() != mockFnType.NumIn() {
		panic("different input parameters")
	}
	if actualFnType.NumOut() != mockFnType.NumOut() {
		panic("different output parameters")
	}
	for i := 0; i < actualFnType.NumIn(); i++ {
		if actualFnType.In(i) != mockFnType.In(i) {
			panic("different input parameters")
		}
	}
	for i := 0; i < actualFnType.NumOut(); i++ {
		if actualFnType.Out(i) != mockFnType.Out(i) {
			panic("different output parameters")
		}
	}

	fnName := getFunctionName(activityFn)
	t.overrodeActivities[fnName] = mockActivityFn
}

// TestWorkflow creates a new TestWorkflowEnvironment that is prepared and ready to run the given workflow.
func (t *WorkflowTestSuite) TestWorkflow(workflowFn interface{}, args ...interface{}) *TestWorkflowEnvironment {
	var workflowType string
	fnType := reflect.TypeOf(workflowFn)
	switch fnType.Kind() {
	case reflect.String:
		workflowType = workflowFn.(string)
	case reflect.Func:
		workflowType = getFunctionName(workflowFn)
	default:
		panic("unsupported workflowFn")
	}

	env := t.NewTestWorkflowEnvironment()
	env.workflowInfo.WorkflowType.Name = workflowType
	factory := getWorkflowDefinitionFactory(t.hostEnv.newRegisteredWorkflowFactory())
	workflowDefinition, err := factory(env.workflowInfo.WorkflowType)
	if err != nil {
		panic(err)
	}
	env.workflowDef = workflowDefinition

	input, err := t.hostEnv.encodeArgs(args)
	if err != nil {
		panic(err)
	}
	env.workflowDef.Execute(env, input)
	return env
}

// TestWorkflowPart wraps a function and test it just as if it is a workflow. You don't need to register workflowPartFn.
func (t *WorkflowTestSuite) TestWorkflowPart(workflowPartFn interface{}, args ...interface{}) *TestWorkflowEnvironment {
	fnType := reflect.TypeOf(workflowPartFn)
	if fnType.Kind() != reflect.Func ||
		fnType.NumIn() == 0 ||
		fnType.In(0) != reflect.TypeOf((*Context)(nil)).Elem() {
		panic("workflowPartFn has to be a function with cadence.Context as its first input parameter")
	}

	workflowWrapperFn := func(ctx Context, args []interface{}) ([]byte, error) {
		valueArgs := []reflect.Value{reflect.ValueOf(ctx)}
		for _, arg := range args {
			valueArgs = append(valueArgs, reflect.ValueOf(arg))
		}

		fnValue := reflect.ValueOf(workflowPartFn)
		retValues := fnValue.Call(valueArgs)
		return validateFunctionAndGetResults(workflowPartFn, retValues)
	}

	// register wrapper workflow if it has not been registered yet.
	fnName := getFunctionName(workflowWrapperFn)
	if _, ok := t.hostEnv.getWorkflowFn(fnName); !ok {
		t.RegisterWorkflow(workflowWrapperFn)
	}

	return t.TestWorkflow(workflowWrapperFn, args)
}

// TestActivity tests an activity. The tested activity will be executed synchronously in the calling goroutinue. Listener
// registered via SetOnActivityTaskStartedListener will be called before the tested activity is executed in this same
// calling goroutinue. However, listener registered via SetOnActivityTaskEndedListener won't be called in this case because
// ActivityTaskEndedListener is only called by workflow dispatcher goroutinue. There is no workflow involved when test
// activity using this TestActivity() method. You can create a new TestWorkflowEnvironment by
// WorkflowTestSuite.NewTestWorkflowEnvironment() and use it to test your activity.
func (env *TestWorkflowEnvironment) TestActivity(activityFn interface{}, args ...interface{}) (BlobResult, error) {
	fnName := getFunctionName(activityFn)
	activityInfo := &activityInfo{env.nextId()}

	input, err := env.testSuite.hostEnv.encodeArgs(args)
	if err != nil {
		panic(err)
	}

	task := newTestActivityTask(
		DefaultTestWorkflowID,
		DefaultTestRunID,
		activityInfo.activityID,
		fnName,
		input,
	)

	// ensure activityFn is registered to DefaultTestTaskList
	env.testSuite.RegisterActivity(activityFn, DefaultTestTaskList)
	taskHandler := env.newTestActivityTaskHandler(DefaultTestTaskList)
	result, err := taskHandler.Execute(task)
	switch request := result.(type) {
	case *s.RespondActivityTaskCanceledRequest:
		return nil, NewCanceledError(request.Details)
	case *s.RespondActivityTaskFailedRequest:
		return nil, NewErrorWithDetails(*request.Reason, request.Details)
	case *s.RespondActivityTaskCompletedRequest:
		return BlobResult(request.Result_), nil
	default:
		return nil, fmt.Errorf("unsupported respond type %T", result)
	}
}

// StartDecisionTask will trigger OnDecisionTaskStart() on the workflow which will execute the dispatcher until all
// coroutinues are blocked. This method is only necessary when you disable the auto start decision task to have full
// control of the workflow execution on when to start a decision task.
func (env *TestWorkflowEnvironment) StartDecisionTask() {
	// post an empty callback to event loop, and request OnDecisionTaskStarted to be triggered after that empty callback
	// is handled.
	env.postCallback(func() {}, true /* to start decision task */)
}

// EnableAutoStartDecisionTask enables auto start decision task on every events. It is equivalent to immediately schedule
// a new decision task on new history events (like activity completed/failed/cancelled, timer fired/cancelled, etc).
// Default is true. Only set this to false if you need to precisely control when to schedule new decision task. For example,
// if your workflow code starts 2 activities, and you only want a new decision task scheduled after both of them are complete.
// Your test can set a listener by using SetOnActivityTaskEndedListener() to know when your activities are done to determine
// when to schedule a new decision task.
func (env *TestWorkflowEnvironment) EnableAutoStartDecisionTask(enable bool) *TestWorkflowEnvironment {
	env.autoStartDecisionTask = enable
	return env
}

// SetOnActivityTaskStartedListener sets a listener that will be called when an activity task started. The listener will
// be called by activity worker goroutinue outside of workflow dispatcher. The listener could be called by multiple goroutinues
// at the same time.
func (env *TestWorkflowEnvironment) SetOnActivityTaskStartedListener(listener func(ctx context.Context, args BlobArgs, activityType string)) {
	env.onActivityTaskStartedListener = listener
}

// SetOnActivityTaskEndedListener sets a listener that will be called when an activity task ended.
// The listener will be called by workflow dispatcher goroutinue.
func (env *TestWorkflowEnvironment) SetOnActivityTaskEndedListener(listener func(result BlobResult, err error, activityType string)) {
	env.onActivityTaskEndedListener = listener
}

// SetOnTimerScheduledListener sets a listener that will be called when a timer is scheduled.
// The listener will be called by workflow dispatcher goroutinue.
func (env *TestWorkflowEnvironment) SetOnTimerScheduledListener(listener func(timerID string, duration time.Duration)) {
	env.onTimerScheduledListener = listener
}

// SetOnTimerFiredListener sets a listener that will be called when a timer is fired
// The listener will be called by workflow dispatcher goroutinue.
func (env *TestWorkflowEnvironment) SetOnTimerFiredListener(listener func(timerID string)) {
	env.onTimerFiredListener = listener
}

// SetOnTimerCancelledListener sets a listener that will be called when a timer is cancelled
// The listener will be called by workflow dispatcher goroutinue.
func (env *TestWorkflowEnvironment) SetOnTimerCancelledListener(listener func(timerID string)) {
	env.onTimerCancelledListener = listener
}

// StartDispatcherLoop starts the main loop that drives workflow dispatcher. The main loop runs in the calling goroutinue
// and it blocked until tested workflow is completed.
func (env *TestWorkflowEnvironment) StartDispatcherLoop(idleTimeout time.Duration) bool {
	// kick off the first decision task
	env.StartDecisionTask()

	for {
		if env.isTestCompleted {
			env.logger.Debug("Workflow completed, stop callback processing...")
			return true
		}

		select {
		case c := <-env.callbackChannel:
			env.processCallback(c)
		case <-time.After(idleTimeout):
			env.logger.Debug("Idle timeout, exiting.", zap.Duration("Timeout", idleTimeout))
			return false
		}
	}
}

func (env *TestWorkflowEnvironment) processCallback(c callbackHandle) {
	// locker is needed to prevent race condition between dispatcher loop goroutinue and activity worker goroutinues.
	// The activity workers could call into Heartbeat which by default is mocked in this test suite. The mock needs to
	// access env.scheduledActivities map, that could cause data race warning.
	env.locker.Lock()
	defer env.locker.Unlock()
	c.callback()
	if c.startDecisionTask {
		env.workflowDef.OnDecisionTaskStarted() // this will execute dispatcher
	}
}

func (env *TestWorkflowEnvironment) postCallback(cb func(), startDecisionTask bool) {
	env.callbackChannel <- callbackHandle{callback: cb, startDecisionTask: startDecisionTask}
}

// RequestCancelActivity request to cancel an activity by activityID
func (env *TestWorkflowEnvironment) RequestCancelActivity(activityID string) {
	env.logger.Sugar().Debugf("RequestCancelActivity %v", activityID)
	handle, ok := env.scheduledActivities[activityID]
	if !ok {
		env.logger.Debug("RequestCancelActivity failed, ActivityID not exists.", zap.String(tagActivityID, activityID))
		return
	}
	if handle.isCancelled {
		env.logger.Debug("RequestCancelActivity failed, Activity already cancelled.", zap.String(tagActivityID, activityID))
		return
	}
	if handle.isDone {
		env.logger.Debug("RequestCancelActivity failed, Activity already completed.", zap.String(tagActivityID, activityID))
		return
	}

	handle.isCancelled = true
	env.postCallback(func() {
		handle.callback(nil, NewCanceledError())
	}, env.autoStartDecisionTask)
}

// RequestCancelTimer request to cancel a timer by timerID
func (env *TestWorkflowEnvironment) RequestCancelTimer(timerID string) {
	env.logger.Sugar().Debugf("RequestCancelTimer %v", timerID)
	timerHandle, ok := env.scheduledTimers[timerID]
	if !ok {
		env.logger.Debug("RequestCancelTimer failed, TimerID not exists.", zap.String(tagTimerID, timerID))
	}

	delete(env.scheduledTimers, timerID)
	timerHandle.timer.Stop()
	env.postCallback(func() {
		timerHandle.callback(nil, NewCanceledError())
		if env.onTimerCancelledListener != nil {
			env.onTimerCancelledListener(timerID)
		}
	}, env.autoStartDecisionTask)
}

// Complete completes workflow running with this TestWorkflowEnvironment
func (env *TestWorkflowEnvironment) Complete(result []byte, err error) {
	if env.isTestCompleted {
		env.logger.Debug("Workflow already completed.")
		return
	}
	env.isTestCompleted = true
	env.testResult = BlobResult(result)
	env.testError = err

	if err == ErrCanceled && env.workflowCancelHandler != nil {
		env.workflowCancelHandler()
	}
}

// IsTestCompleted check if test is completed or not
func (env *TestWorkflowEnvironment) IsTestCompleted() bool {
	return env.isTestCompleted
}

// GetTestResult return the blob result from test workflow
func (env *TestWorkflowEnvironment) GetTestResult() BlobResult {
	return env.testResult
}

// GetTestError return the error from test workflow
func (env *TestWorkflowEnvironment) GetTestError() error {
	return env.testError
}

// GetLogger returns a logger from TestWorkflowEnvironment
func (env *TestWorkflowEnvironment) GetLogger() *zap.Logger {
	return env.logger
}

// ExecuteActivity executes requested activity in TestWorkflowEnvironment
func (env *TestWorkflowEnvironment) ExecuteActivity(parameters executeActivityParameters, callback resultHandler) *activityInfo {
	activityInfo := &activityInfo{env.nextId()}

	task := newTestActivityTask(
		DefaultTestWorkflowID,
		DefaultTestRunID,
		activityInfo.activityID,
		parameters.ActivityType.Name,
		parameters.Input,
	)

	taskHandler := env.newTestActivityTaskHandler(parameters.TaskListName)
	activityHandle := &activityHandle{callback: callback}
	env.scheduledActivities[activityInfo.activityID] = activityHandle

	// activity runs in separate goroutinue outside of workflow dispatcher
	go func() {
		result, err := taskHandler.Execute(task)
		if err != nil {
			panic(err)
		}
		// post activity result to workflow dispatcher
		env.postCallback(func() {
			env.handleActivityResult(activityInfo.activityID, result, parameters.ActivityType.Name)
		}, env.autoStartDecisionTask)
	}()

	return activityInfo
}

func (env *TestWorkflowEnvironment) handleActivityResult(activityID string, result interface{}, activityType string) {
	// this is running in dispatcher
	activityHandle, ok := env.scheduledActivities[activityID]
	if !ok {
		env.logger.Debug("handleActivityResult failed, ActivityID not exists.", zap.String(tagActivityID, activityID))
		return
	}
	if activityHandle.isCancelled {
		env.logger.Debug("Activit is cancelled, ignore the result.", zap.String(tagActivityID, activityID))
		return
	}
	if activityHandle.isDone {
		env.logger.Debug("Activit is already completed, ignore the result.", zap.String(tagActivityID, activityID))
		return
	}
	activityHandle.isDone = true

	var blob []byte
	var err error

	switch request := result.(type) {
	case *s.RespondActivityTaskCanceledRequest:
		err = NewCanceledError(request.Details)
		activityHandle.callback(nil, err)
	case *s.RespondActivityTaskFailedRequest:
		err = NewErrorWithDetails(*request.Reason, request.Details)
		activityHandle.callback(nil, err)
	case *s.RespondActivityTaskCompletedRequest:
		blob = request.Result_
		activityHandle.callback(blob, nil)
	default:
		panic(fmt.Sprintf("unsupported respond type %T", result))
	}

	if env.onActivityTaskEndedListener != nil {
		env.onActivityTaskEndedListener(BlobResult(blob), err, activityType)
	}
}

func (a *activityExecutorWrapper) Execute(ctx context.Context, input []byte) ([]byte, error) {
	if a.env.onActivityTaskStartedListener != nil {
		a.env.onActivityTaskStartedListener(ctx, BlobArgs(input), a.ActivityType().Name)
	}
	return a.activity.Execute(ctx, input)
}

func (env *TestWorkflowEnvironment) newTestActivityTaskHandler(taskList string) ActivityTaskHandler {
	taskListActivities, ok := env.testSuite.registeredActivities[taskList]
	if !ok {
		panic(fmt.Sprintf("no activity is registered with tasklist '%v'", taskList))
	}

	wOptions := env.workerOptions.(*workerOptions)
	params := workerExecutionParameters{
		TaskList:     taskList,
		Identity:     wOptions.identity,
		MetricsScope: wOptions.metricsScope,
		Logger:       wOptions.logger,
		UserContext:  wOptions.userContext,
	}
	if params.Logger == nil && env.logger != nil {
		params.Logger = env.logger
	}
	ensureRequiredParams(&params)

	registeredActivities := env.testSuite.hostEnv.getRegisteredActivities()
	var activities []activity
	for _, a := range registeredActivities {
		fnName := a.ActivityType().Name
		if _, ok := taskListActivities[fnName]; !ok {
			// activity is not registered with this tasklist, skip it
			continue
		}
		activity := a
		overrideFn, ok := env.testSuite.overrodeActivities[fnName]
		if ok {
			activity = &activityExecutor{name: fnName, fn: overrideFn}
		}

		activityWrapper := &activityExecutorWrapper{activity: activity, fn: a.Execute, env: env}
		activities = append(activities, activityWrapper)
	}

	taskHandler := newActivityTaskHandler(activities, env.service, params)
	return taskHandler
}

func newTestActivityTask(workflowID, runID, activityID, activityType string, input []byte) *s.PollForActivityTaskResponse {
	task := &s.PollForActivityTaskResponse{
		WorkflowExecution: &s.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
		ActivityId:   common.StringPtr(activityID),
		TaskToken:    []byte(activityID), // use activityID as TaskToken so we can map TaskToken in heartbeat calls.
		ActivityType: &s.ActivityType{Name: common.StringPtr(activityType)},
		Input:        input,
	}
	return task
}

// NewTimer creates a new timer
func (env *TestWorkflowEnvironment) NewTimer(d time.Duration, callback resultHandler) *timerInfo {
	timerInfo := &timerInfo{env.nextId()}
	timer := env.clock.AfterFunc(d, func() {
		delete(env.scheduledTimers, timerInfo.timerID)
		env.postCallback(func() {
			callback(nil, nil)
			if env.onTimerFiredListener != nil {
				env.onTimerFiredListener(timerInfo.timerID)
			}
		}, env.autoStartDecisionTask)
	})
	env.scheduledTimers[timerInfo.timerID] = &timerHandle{timer: timer, callback: callback}
	if env.onTimerScheduledListener != nil {
		env.onTimerScheduledListener(timerInfo.timerID, d)
	}
	return timerInfo
}

// Now returns current time for this TestWorkflowEnvironment
func (env *TestWorkflowEnvironment) Now() time.Time {
	return env.clock.Now()
}

// WorkflowInfo returns WorkflowInfo for TestWorkflowEnvironment
func (env *TestWorkflowEnvironment) WorkflowInfo() *WorkflowInfo {
	return env.workflowInfo
}

// RegisterCancel registers a handler that will be invoked when workflow completed with ErrCanceled error.
func (env *TestWorkflowEnvironment) RegisterCancel(handler func()) {
	env.workflowCancelHandler = handler
}

func (env *TestWorkflowEnvironment) RequestCancelWorkflow(domainName, workflowID, runID string) error {
	// NOT IMPLEMENTED
	return nil
}

func (env *TestWorkflowEnvironment) nextId() string {
	activityID := env.counterID
	env.counterID++
	return fmt.Sprintf("%d", activityID)
}

// make sure interface is implemented
var _ workflowEnvironment = (*TestWorkflowEnvironment)(nil)
