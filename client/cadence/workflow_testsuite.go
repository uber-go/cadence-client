package cadence

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/facebookgo/clock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	m "github.com/uber-go/cadence-client/.gen/go/cadence"
	"github.com/uber-go/cadence-client/.gen/go/shared"
	"github.com/uber-go/cadence-client/common"
	"github.com/uber-go/cadence-client/mocks"
	"github.com/uber/tchannel-go/thrift"
	"go.uber.org/zap"
)

const (
	defaultTestTaskList   = "default-test-tasklist"
	defaultTestWorkflowID = "default-test-workflow-id"
	defaultTestRunID      = "default-test-run-id"
)

type (
	// EncodedValue is type alias used to encapsulate/extract encoded result from workflow/activity.
	EncodedValue []byte

	// EncodedValues is a type alias used to encapsulate/extract encoded arguments from workflow/activity.
	EncodedValues []byte

	timerHandle struct {
		callback resultHandler
		timer    *clock.Timer
		duration time.Duration
		timerId  int
	}

	activityHandle struct {
		callback     resultHandler
		activityType string
	}

	callbackHandle struct {
		callback          func()
		startDecisionTask bool // start a new decision task after callback() is handled.
	}

	activityExecutorWrapper struct {
		activity
		env *testWorkflowEnvironmentImpl
	}

	taskListSpecificActivity struct {
		fn        interface{}
		taskLists map[string]struct{}
	}

	// WorkflowTestSuite is the test suite to run unit tests for workflow/activity.
	WorkflowTestSuite struct {
		suite.Suite
		logger                     *zap.Logger
		hostEnv                    *hostEnvImpl
		taskListSpecificActivities map[string]*taskListSpecificActivity
	}

	// TestWorkflowEnvironment is the environment that you use to test workflow
	TestWorkflowEnvironment struct {
		impl *testWorkflowEnvironmentImpl
	}

	// TestActivityEnviornment is the environment that you use to test activity
	TestActivityEnviornment struct {
		impl *testWorkflowEnvironmentImpl
	}

	// testWorkflowEnvironmentImpl is the environment that runs the workflow/activity unit tests.
	testWorkflowEnvironmentImpl struct {
		testSuite          *WorkflowTestSuite
		overrodeActivities map[string]interface{} // map of registered-fnName -> fakeActivityFn

		service       m.TChanWorkflowService
		workerOptions WorkerOptions
		logger        *zap.Logger
		clock         clock.Clock

		workflowInfo          *WorkflowInfo
		workflowDef           workflowDefinition
		counterID             int
		workflowCancelHandler func()

		locker              *sync.Mutex
		scheduledActivities map[string]*activityHandle
		scheduledTimers     map[string]*timerHandle

		onActivityStartedListener func(ctx context.Context, args EncodedValues)
		onActivityEndedListener   func(result EncodedValue, err error, activityType string)
		onTimerScheduledListener  func(timerID string, duration time.Duration)
		onTimerFiredListener      func(timerID string)
		onTimerCancelledListener  func(timerID string)

		callbackChannel chan callbackHandle
		idleTimeout     time.Duration
		isTestCompleted bool
		testResult      EncodedValue
		testError       error

		enableFastForwardClock bool
	}
)

// Get extract data from encoded data to desired value type. valuePtr is pointer to the actual value type.
func (b EncodedValue) Get(valuePtr interface{}) error {
	return getHostEnvironment().decodeArg(b, valuePtr)
}

// Get extract values from encoded data to desired value type. valuePtrs are pointers to the actual value types.
func (b EncodedValues) Get(valuePtrs ...interface{}) error {
	return getHostEnvironment().decode(b, valuePtrs)
}

// SetT sets the testing.T instance. This method is called by testify to setup the testing.T for test suite.
func (s *WorkflowTestSuite) SetT(t *testing.T) {
	s.Suite.SetT(t)
	if s.hostEnv == nil {
		s.hostEnv = &hostEnvImpl{
			workflowFuncMap: make(map[string]interface{}),
			activityFuncMap: make(map[string]interface{}),
			encoding:        gobEncoding{},
		}
		s.taskListSpecificActivities = make(map[string]*taskListSpecificActivity)
	}
}

// NewWorkflowTestSuite creates a WorkflowTestSuite
func NewWorkflowTestSuite() *WorkflowTestSuite {
	return &WorkflowTestSuite{
		hostEnv: &hostEnvImpl{
			workflowFuncMap: make(map[string]interface{}),
			activityFuncMap: make(map[string]interface{}),
			encoding:        gobEncoding{},
		},
		taskListSpecificActivities: make(map[string]*taskListSpecificActivity),
	}
}

// NewTestWorkflowEnvironment creates a new instance of TestWorkflowEnvironment. You can use the returned TestWorkflowEnvironment
// to run your workflow in the test environment.
func (s *WorkflowTestSuite) NewTestWorkflowEnvironment() *TestWorkflowEnvironment {
	return &TestWorkflowEnvironment{impl: s.newTestWorkflowEnvironmentImpl()}
}

// NewTestActivityEnvironment creates a new instance of TestActivityEnviornment. You can use the returned TestActivityEnviornment
// to run your activity in the test environment.
func (s *WorkflowTestSuite) NewTestActivityEnvironment() *TestActivityEnviornment {
	return &TestActivityEnviornment{impl: s.newTestWorkflowEnvironmentImpl()}
}

func (s *WorkflowTestSuite) newTestWorkflowEnvironmentImpl() *testWorkflowEnvironmentImpl {
	env := &testWorkflowEnvironmentImpl{
		testSuite: s,
		logger:    s.logger,

		overrodeActivities: make(map[string]interface{}),

		workflowInfo: &WorkflowInfo{
			WorkflowExecution: WorkflowExecution{
				ID:    defaultTestWorkflowID,
				RunID: defaultTestRunID,
			},
			WorkflowType: WorkflowType{Name: "workflow-type-not-specified"},
			TaskListName: defaultTestTaskList,
		},

		locker:              &sync.Mutex{},
		scheduledActivities: make(map[string]*activityHandle),
		scheduledTimers:     make(map[string]*timerHandle),
		callbackChannel:     make(chan callbackHandle, 1000),
		idleTimeout:         time.Second * 3,

		enableFastForwardClock: true,
	}

	if env.logger == nil {
		logger, _ := zap.NewDevelopment()
		env.logger = logger
	}

	if env.service == nil {
		mockService := new(mocks.TChanWorkflowService)
		mockHeartbeatFn := func(c thrift.Context, r *shared.RecordActivityTaskHeartbeatRequest) error {
			activityID := string(r.TaskToken)
			env.locker.Lock() // need lock as this is running in activity worker's goroutinue
			_, ok := env.scheduledActivities[activityID]
			env.locker.Unlock()
			if !ok {
				env.logger.Debug("RecordActivityTaskHeartbeat: ActivityID not found, could be already completed or cancelled.",
					zap.String(tagActivityID, activityID))
				return shared.NewEntityNotExistsError()
			}
			env.logger.Debug("RecordActivityTaskHeartbeat", zap.String(tagActivityID, activityID))
			return nil
		}

		mockService.On("RecordActivityTaskHeartbeat", mock.Anything, mock.Anything).Return(
			&shared.RecordActivityTaskHeartbeatResponse{CancelRequested: common.BoolPtr(false)},
			mockHeartbeatFn)
		env.service = mockService
	}
	if env.workerOptions == nil {
		env.workerOptions = NewWorkerOptions().SetLogger(env.logger)
	}

	if env.clock == nil {
		env.clock = clock.NewMock()
	}

	return env
}

// RegisterWorkflow registers a workflow that could be used by tests of this WorkflowTestSuite instance. All workflow registered
// via cadence.RegisterWorkflow() are still valid and will be available to all tests of all instance of WorkflowTestSuite.
// Workflow registration is only required if you are invoking workflow by name.
func (s *WorkflowTestSuite) RegisterWorkflow(workflowFn interface{}) {
	err := s.hostEnv.RegisterWorkflow(workflowFn)
	if err != nil {
		panic(err)
	}
}

// RegisterActivity registers an activity with given task list that could be used by tests of this WorkflowTestSuite instance
// for that given tasklist. Activities registered via cadence.RegisterActivity() are still valid and will be available
// to all tests of all instances of WorkflowTestSuite for any tasklist. However, if an activity is registered via this
// WorkflowTestSuite.RegisterActivity() with a given tasklist, that activity will only be visible to that registered tasklist
// in tests of this WorkflowTestSuite instance, even if that activity was also registered via cadence.RegisterActivity().
// Example: ActivityA, ActivityB are registered via cadence.RegisterActivity(). These 2 activities will be visible to all
// tests of any instance of WorkflowTestSuite. If testSuite1 register ActivityA to task-list-1, then ActivityA will only
// be visible to task-list-1 for tests of testSuite1, and ActivityB will be visible to all tasklist for tests of testSuite1.
func (s *WorkflowTestSuite) RegisterActivity(activityFn interface{}, taskList string) {
	fnName := getFunctionName(activityFn)

	_, ok := s.hostEnv.activityFuncMap[fnName]
	if !ok {
		// activity not registered yet, register now
		err := s.hostEnv.RegisterActivity(activityFn)
		if err != nil {
			panic(err)
		}
	}

	taskListActivity, ok := s.taskListSpecificActivities[fnName]
	if !ok {
		taskListActivity = &taskListSpecificActivity{fn: activityFn, taskLists: make(map[string]struct{})}
		s.taskListSpecificActivities[fnName] = taskListActivity
	}
	taskListActivity.taskLists[taskList] = struct{}{}
}

// SetLogger sets the logger for this WorkflowTestSuite
func (t *WorkflowTestSuite) SetLogger(logger *zap.Logger) {
	t.logger = logger
}

// EnableClockFastForwardWhenBlockedByTimer enables auto clock fast forward when dispatcher is blocked by timer.
// This flag is ignored if the autoStartDecisionTask is set to false.
// Default is true. If you set this flag to false, your timer will be fired by below 2 cases:
//  1) Use real clock, and timer is fired when time goes by.
//  2) Use mock clock, and you need to manually move forward mock clock to fire timer.
func (t *TestWorkflowEnvironment) EnableClockFastForwardWhenBlockedByTimer(enable bool) *TestWorkflowEnvironment {
	t.impl.enableFastForwardClock = enable
	return t
}

// ExecuteActivity executes an activity. The tested activity will be executed synchronously in the calling goroutinue.
// Caller should use EncodedValue.Get() to extract strong typed result value.
func (t *TestActivityEnviornment) ExecuteActivity(activityFn interface{}, args ...interface{}) (EncodedValue, error) {
	env := t.impl
	fnName := getFunctionName(activityFn)

	input, err := getHostEnvironment().encodeArgs(args)
	if err != nil {
		panic(err)
	}

	task := newTestActivityTask(
		defaultTestWorkflowID,
		defaultTestRunID,
		"0",
		fnName,
		input,
	)

	// ensure activityFn is registered to defaultTestTaskList
	env.testSuite.RegisterActivity(activityFn, defaultTestTaskList)
	taskHandler := env.newTestActivityTaskHandler(defaultTestTaskList)
	result, err := taskHandler.Execute(task)
	switch request := result.(type) {
	case *shared.RespondActivityTaskCanceledRequest:
		return nil, NewCanceledError(request.Details)
	case *shared.RespondActivityTaskFailedRequest:
		return nil, NewErrorWithDetails(*request.Reason, request.Details)
	case *shared.RespondActivityTaskCompletedRequest:
		return EncodedValue(request.Result_), nil
	default:
		// will never happen
		return nil, fmt.Errorf("unsupported respond type %T", result)
	}
}

// SetWorkerOption sets the WorkerOptions that will be use by TestActivityEnviornment. TestActivityEnviornment will
// use options set by SetIdentity(), SetMetrics(), and WithActivityContext() on the WorkerOptions. Other options are ignored.
func (t *TestActivityEnviornment) SetWorkerOption(options WorkerOptions) *TestActivityEnviornment {
	t.impl.workerOptions = options
	return t
}

// ExecuteWorkflow executes a workflow, wait until workflow complete. It returns nil if workflow completed. Otherwise,
// it returns error if workflow is blocked and cannot complete.
func (t *TestWorkflowEnvironment) ExecuteWorkflow(workflowFn interface{}, args ...interface{}) error {
	env := t.impl
	env.startWorkflow(workflowFn, args...)
	executeErr := env.Execute()
	if executeErr != nil {
		return executeErr
	}
	return nil
}

// OverrideActivity overrides an actual activity with a fake activity. The fake activity will be invoked in place where the
// actual activity should have been invoked.
func (t *TestWorkflowEnvironment) OverrideActivity(activityFn, fakeActivityFn interface{}) {
	t.impl.OverrideActivity(activityFn, fakeActivityFn)
}

func (t *TestWorkflowEnvironment) Now() time.Time {
	return t.impl.Now()
}

// SetWorkflowService sets the m.TChanWorkflowService for TestWorkflowEnvironment
func (t *TestWorkflowEnvironment) SetWorkflowService(service m.TChanWorkflowService) *TestWorkflowEnvironment {
	t.impl.service = service
	return t
}

// SetClock sets the clock that will be used for testWorkflowEnvironmentImpl
func (t *TestWorkflowEnvironment) SetClock(clock clock.Clock) *TestWorkflowEnvironment {
	t.impl.clock = clock
	return t
}

// SetWorkerOption sets the WorkerOptions for TestWorkflowEnvironment. TestWorkflowEnvironment will use options set by
// SetIdentity(), SetMetrics(), and WithActivityContext() on the WorkerOptions. Other options are ignored.
func (t *TestWorkflowEnvironment) SetWorkerOption(options WorkerOptions) *TestWorkflowEnvironment {
	t.impl.workerOptions = options
	return t
}

// SetIdleTimeout sets the idle timeout for this TestWorkflowEnvironment. Idle means workflow is blocked and cannot make
// progress. This could happen if workflow is waiting for something, like activity result, timer, or signal. This is
// real wall clock time, not the workflow time (a.k.a not cadence.Now() time).
func (t *TestWorkflowEnvironment) SetIdleTimeout(idleTimeout time.Duration) *TestWorkflowEnvironment {
	t.impl.idleTimeout = idleTimeout
	return t
}

// SetOnActivityStartedListener sets a listener that will be called when an activity task started.
func (t *TestWorkflowEnvironment) SetOnActivityStartedListener(
	listener func(ctx context.Context, args EncodedValues)) *TestWorkflowEnvironment {
	t.impl.onActivityStartedListener = listener
	return t
}

// SetOnActivityEndedListener sets a listener that will be called when an activity task ended.
func (t *TestWorkflowEnvironment) SetOnActivityEndedListener(
	listener func(result EncodedValue, err error, activityType string)) *TestWorkflowEnvironment {
	t.impl.onActivityEndedListener = listener
	return t
}

// SetOnTimerScheduledListener sets a listener that will be called when a timer is scheduled.
func (t *TestWorkflowEnvironment) SetOnTimerScheduledListener(
	listener func(timerID string, duration time.Duration)) *TestWorkflowEnvironment {
	t.impl.onTimerScheduledListener = listener
	return t
}

// SetOnTimerFiredListener sets a listener that will be called after a timer is fired.
func (t *TestWorkflowEnvironment) SetOnTimerFiredListener(listener func(timerID string)) *TestWorkflowEnvironment {
	t.impl.onTimerFiredListener = listener
	return t
}

// SetOnTimerCancelledListener sets a listener that will be called when a timer is cancelled
func (t *TestWorkflowEnvironment) SetOnTimerCancelledListener(listener func(timerID string)) *TestWorkflowEnvironment {
	t.impl.onTimerCancelledListener = listener
	return t
}

// IsWorkflowCompleted check if test is completed or not
func (t *TestWorkflowEnvironment) IsWorkflowCompleted() bool {
	return t.impl.isTestCompleted
}

// GetWorkflowResult return the encoded result from test workflow
func (t *TestWorkflowEnvironment) GetWorkflowResult(valuePtr interface{}) {
	s, r := t.impl.testSuite, t.impl.testResult
	s.NotNil(r)
	err := r.Get(valuePtr)
	s.NoError(err)
}

// GetWorkflowError return the error from test workflow
func (t *TestWorkflowEnvironment) GetWorkflowError() error {
	return t.impl.testError
}

// CompleteActivity complete an activity that had returned ErrActivityResultPending error
func (t *TestWorkflowEnvironment) CompleteActivity(taskToken []byte, result interface{}, err error) error {
	return t.impl.CompleteActivity(taskToken, result, err)
}

// CancelWorkflow cancels the currently running test workflow.
func (t *TestWorkflowEnvironment) CancelWorkflow() {
	t.impl.RequestCancelWorkflow("test-domain",
		t.impl.workflowInfo.WorkflowExecution.ID,
		t.impl.workflowInfo.WorkflowExecution.RunID)
}

// RegisterDelayedCallback creates a new timer with specified delayDuration using workflow clock (not wall clock). When
// the timer fires, the callback will be called. By default, this test suite uses mock clock which automatically move
// forward to fire next timer when workflow is blocked. You can use this API to make some event happen at desired time.
func (t *TestWorkflowEnvironment) RegisterDelayedCallback(callback func(), delayDuration time.Duration) {
	t.impl.registerDelayedCallback(callback, delayDuration)
}

func (env *testWorkflowEnvironmentImpl) startWorkflow(workflowFn interface{}, args ...interface{}) {
	s := env.testSuite
	var workflowType string
	fnType := reflect.TypeOf(workflowFn)
	switch fnType.Kind() {
	case reflect.String:
		workflowType = workflowFn.(string)
	case reflect.Func:
		// auto register workflow if it is not already registered
		fnName := getFunctionName(workflowFn)
		if _, ok := s.hostEnv.getWorkflowFn(fnName); !ok {
			s.RegisterWorkflow(workflowFn)
		}
		workflowType = getFunctionName(workflowFn)
	default:
		panic("unsupported workflowFn")
	}

	env.workflowInfo.WorkflowType.Name = workflowType
	factory := getWorkflowDefinitionFactory(s.hostEnv.newRegisteredWorkflowFactory())
	workflowDefinition, err := factory(env.workflowInfo.WorkflowType)
	if err != nil {
		// try to get workflow from global registered workflows
		factory = getWorkflowDefinitionFactory(getHostEnvironment().newRegisteredWorkflowFactory())
		workflowDefinition, err = factory(env.workflowInfo.WorkflowType)
		if err != nil {
			panic(err)
		}
	}
	env.workflowDef = workflowDefinition

	input, err := s.hostEnv.encodeArgs(args)
	if err != nil {
		panic(err)
	}
	env.workflowDef.Execute(env, input)
}

func (env *testWorkflowEnvironmentImpl) OverrideActivity(activityFn, fakeActivityFn interface{}) {
	// verify both functions are valid activity func
	actualFnType := reflect.TypeOf(activityFn)
	if err := validateFnFormat(actualFnType, false); err != nil {
		panic(err)
	}
	fakeFnType := reflect.TypeOf(fakeActivityFn)
	if err := validateFnFormat(fakeFnType, false); err != nil {
		panic(err)
	}

	// verify signature of registeredActivityFn and fakeActivityFn are the same.
	if actualFnType != fakeFnType {
		panic("activityFn and fakeActivityFn have different func signature")
	}

	fnName := getFunctionName(activityFn)
	env.overrodeActivities[fnName] = fakeActivityFn
}

// startDecisionTask will trigger OnDecisionTaskStart() on the workflow which will execute the dispatcher until all
// coroutinues are blocked. This method is only necessary when you disable the auto start decision task to have full
// control of the workflow execution on when to start a decision task.
func (env *testWorkflowEnvironmentImpl) startDecisionTask() {
	// post an empty callback to event loop, and request OnDecisionTaskStarted to be triggered after that empty callback
	env.postCallback(func() {}, true /* to start decision task */)
}

func (env *testWorkflowEnvironmentImpl) Execute() error {
	if env.workflowDef == nil {
		return errors.New("nothing to execute, you need to call startWorkflow() before you execute")
	}
	// kick off the initial decision task
	env.startDecisionTask()

	for {
		// use non-blocking-select to check if there is anything pending in the main thread.
		select {
		case c := <-env.callbackChannel:
			// this will drain the callbackChannel
			env.processCallback(c)
		default:
			// nothing to process, main thread is blocked at this moment, now check if we should auto fire next timer
			if !env.autoFireNextTimer() {
				if env.isTestCompleted {
					return nil
				}

				// no timer to fire, wait for things to do or timeout.
				select {
				case c := <-env.callbackChannel:
					env.processCallback(c)
				case <-time.After(env.idleTimeout):
					st := env.workflowDef.StackTrace()
					env.logger.Debug("Dispatcher idle timeout.",
						zap.Duration("IdleTimeout", env.idleTimeout),
						zap.String("DispatcherStack", st))
					return fmt.Errorf("Workflow cannot complete. Timeout: %v, WorkflowStack: %v", env.idleTimeout, st)
				}
			}
		}
	}
}

func (env *testWorkflowEnvironmentImpl) registerDelayedCallback(f func(), delayDuration time.Duration) {
	env.postCallback(func() {
		env.NewTimer(delayDuration, func(result []byte, err error) {
			f()
		})
	}, true)
}

func (env *testWorkflowEnvironmentImpl) processCallback(c callbackHandle) {
	// locker is needed to prevent race condition between dispatcher loop goroutinue and activity worker goroutinues.
	// The activity workers could call into Heartbeat which by default is mocked in this test suite. The mock needs to
	// access s.scheduledActivities map, that could cause data race warning.
	env.locker.Lock()
	defer env.locker.Unlock()
	c.callback()
	if c.startDecisionTask {
		env.workflowDef.OnDecisionTaskStarted() // this will execute dispatcher
	}
}

func (env *testWorkflowEnvironmentImpl) autoFireNextTimer() bool {
	if !env.enableFastForwardClock || len(env.scheduledTimers) == 0 {
		return false
	}

	// find next timer
	var tofire *timerHandle
	for _, t := range env.scheduledTimers {
		if tofire == nil {
			tofire = t
		} else if t.duration < tofire.duration ||
			(t.duration == tofire.duration && t.timerId < tofire.timerId) {
			tofire = t
		}
	}

	d := tofire.duration
	env.logger.Debug("Auto fire timer", zap.Int(tagTimerID, tofire.timerId), zap.Duration("Duration", tofire.duration))

	// Move mock clock forward, this will fire the timer, and the timer callback will remove timer from scheduledTimers.
	mockClock, ok := env.clock.(*clock.Mock)
	if !ok {
		panic("configured clock does not support fast forward, must use a mock clock for fast forward")
	}
	mockClock.Add(d)

	// reduce all pending timer's duration by d
	for _, t := range env.scheduledTimers {
		t.duration -= d
	}
	return true
}

func (env *testWorkflowEnvironmentImpl) postCallback(cb func(), startDecisionTask bool) {
	env.callbackChannel <- callbackHandle{callback: cb, startDecisionTask: startDecisionTask}
}

func (env *testWorkflowEnvironmentImpl) RequestCancelActivity(activityID string) {
	handle, ok := env.scheduledActivities[activityID]
	if !ok {
		env.logger.Debug("RequestCancelActivity failed, Activity not exists or already completed.", zap.String(tagActivityID, activityID))
		return
	}
	env.logger.Debug("RequestCancelActivity", zap.String(tagActivityID, activityID))
	delete(env.scheduledActivities, activityID)
	env.postCallback(func() {
		cancelledErr := NewCanceledError()
		handle.callback(nil, cancelledErr)
		if env.onActivityEndedListener != nil {
			env.onActivityEndedListener(nil, cancelledErr, handle.activityType)
		}
	}, true)
}

// RequestCancelTimer request to cancel timer on this testWorkflowEnvironmentImpl.
func (env *testWorkflowEnvironmentImpl) RequestCancelTimer(timerID string) {
	env.logger.Debug("RequestCancelTimer", zap.String(tagTimerID, timerID))
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
	}, true)
}

func (env *testWorkflowEnvironmentImpl) Complete(result []byte, err error) {
	if env.isTestCompleted {
		env.logger.Debug("Workflow already completed.")
		return
	}
	env.isTestCompleted = true
	env.testResult = EncodedValue(result)
	env.testError = err

	if err == ErrCanceled && env.workflowCancelHandler != nil {
		env.workflowCancelHandler()
	}
}

func (env *testWorkflowEnvironmentImpl) CompleteActivity(taskToken []byte, result interface{}, err error) error {
	if taskToken == nil {
		return errors.New("nil task token provided")
	}
	var data []byte
	if result != nil {
		var encodeErr error
		data, encodeErr = getHostEnvironment().encodeArg(result)
		if encodeErr != nil {
			return encodeErr
		}
	}

	activityID := string(taskToken)
	env.postCallback(func() {
		activityHandle, ok := env.scheduledActivities[activityID]
		if !ok {
			env.logger.Debug("CompleteActivity: ActivityID not found, could be already completed or cancelled.",
				zap.String(tagActivityID, activityID))
			return
		}
		request := convertActivityResultToRespondRequest("test-identity", taskToken, data, err)
		env.handleActivityResult(activityID, request, activityHandle.activityType)
	}, false /* do not auto schedule decision task, because activity might be still pending */)

	return nil
}

func (env *testWorkflowEnvironmentImpl) GetLogger() *zap.Logger {
	return env.logger
}

func (env *testWorkflowEnvironmentImpl) ExecuteActivity(parameters executeActivityParameters, callback resultHandler) *activityInfo {
	activityInfo := &activityInfo{fmt.Sprintf("%d", env.nextId())}

	task := newTestActivityTask(
		defaultTestWorkflowID,
		defaultTestRunID,
		activityInfo.activityID,
		parameters.ActivityType.Name,
		parameters.Input,
	)

	taskHandler := env.newTestActivityTaskHandler(parameters.TaskListName)
	activityHandle := &activityHandle{callback: callback, activityType: parameters.ActivityType.Name}
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
		}, false /* do not auto schedule decision task, because activity might be still pending */)
	}()

	return activityInfo
}

func (env *testWorkflowEnvironmentImpl) handleActivityResult(activityID string, result interface{}, activityType string) {
	env.logger.Debug(fmt.Sprintf("handleActivityResult: %T.", result),
		zap.String(tagActivityID, activityID), zap.String(tagActivityType, activityType))
	if result == nil {
		// In case activity returns ErrActivityResultPending, the respond will be nil, and we don's need to do anything.
		// Activity will need to complete asynchronously using CompleteActivity().
		if env.onActivityEndedListener != nil {
			env.onActivityEndedListener(nil, ErrActivityResultPending, activityType)
		}
		return
	}

	// this is running in dispatcher
	activityHandle, ok := env.scheduledActivities[activityID]
	if !ok {
		env.logger.Debug("handleActivityResult: ActivityID not exists, could be already completed or cancelled.",
			zap.String(tagActivityID, activityID))
		return
	}

	delete(env.scheduledActivities, activityID)

	var blob []byte
	var err error

	switch request := result.(type) {
	case *shared.RespondActivityTaskCanceledRequest:
		err = NewCanceledError(request.Details)
		activityHandle.callback(nil, err)
	case *shared.RespondActivityTaskFailedRequest:
		err = NewErrorWithDetails(*request.Reason, request.Details)
		activityHandle.callback(nil, err)
	case *shared.RespondActivityTaskCompletedRequest:
		blob = request.Result_
		activityHandle.callback(blob, nil)
	default:
		panic(fmt.Sprintf("unsupported respond type %T", result))
	}

	if env.onActivityEndedListener != nil {
		env.onActivityEndedListener(EncodedValue(blob), err, activityType)
	}

	env.startDecisionTask()
}

// Execute executes the activity code. This is the wrapper where we call ActivityTaskStartedListener hook.
func (a *activityExecutorWrapper) Execute(ctx context.Context, input []byte) ([]byte, error) {
	if a.env.onActivityStartedListener != nil {
		a.env.postCallback(func() {
			a.env.onActivityStartedListener(ctx, EncodedValues(input))
		}, false)
	}
	return a.activity.Execute(ctx, input)
}

func (env *testWorkflowEnvironmentImpl) newTestActivityTaskHandler(taskList string) ActivityTaskHandler {
	wOptions := env.workerOptions.(*workerOptions)
	params := workerExecutionParameters{
		TaskList:     taskList,
		Identity:     wOptions.identity,
		MetricsScope: wOptions.metricsScope,
		Logger:       env.logger,
		UserContext:  wOptions.userContext,
	}
	ensureRequiredParams(&params)

	var activities []activity
	for fnName, tasklistActivity := range env.testSuite.taskListSpecificActivities {
		if _, ok := tasklistActivity.taskLists[taskList]; ok {
			activities = append(activities, env.wrapActivity(&activityExecutor{name: fnName, fn: tasklistActivity.fn}))
		}
	}

	// get global registered activities
	globalActivities := getHostEnvironment().getRegisteredActivities()
	for _, a := range globalActivities {
		fnName := a.ActivityType().Name
		if _, ok := env.testSuite.taskListSpecificActivities[fnName]; ok {
			// activity is registered to a specific taskList, so ignore it from the global registered activities.
			continue
		}
		activities = append(activities, env.wrapActivity(a))
	}

	if len(activities) == 0 {
		panic(fmt.Sprintf("no activity is registered for tasklist '%v'", taskList))
	}

	taskHandler := newActivityTaskHandler(activities, env.service, params)
	return taskHandler
}

func (env *testWorkflowEnvironmentImpl) wrapActivity(a activity) *activityExecutorWrapper {
	fnName := a.ActivityType().Name
	if overrideFn, ok := env.overrodeActivities[fnName]; ok {
		// override activity
		a = &activityExecutor{name: fnName, fn: overrideFn}
	}

	activityWrapper := &activityExecutorWrapper{activity: a, env: env}
	return activityWrapper
}

func newTestActivityTask(workflowID, runID, activityID, activityType string, input []byte) *shared.PollForActivityTaskResponse {
	task := &shared.PollForActivityTaskResponse{
		WorkflowExecution: &shared.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
		ActivityId:   common.StringPtr(activityID),
		TaskToken:    []byte(activityID), // use activityID as TaskToken so we can map TaskToken in heartbeat calls.
		ActivityType: &shared.ActivityType{Name: common.StringPtr(activityType)},
		Input:        input,
	}
	return task
}

func (env *testWorkflowEnvironmentImpl) NewTimer(d time.Duration, callback resultHandler) *timerInfo {
	nextId := env.nextId()
	timerInfo := &timerInfo{fmt.Sprintf("%d", nextId)}
	timer := env.clock.AfterFunc(d, func() {
		env.postCallback(func() {
			delete(env.scheduledTimers, timerInfo.timerID)
			callback(nil, nil)
			if env.onTimerFiredListener != nil {
				env.onTimerFiredListener(timerInfo.timerID)
			}
		}, true)
	})
	env.scheduledTimers[timerInfo.timerID] = &timerHandle{timer: timer, callback: callback, duration: d, timerId: nextId}
	if env.onTimerScheduledListener != nil {
		env.onTimerScheduledListener(timerInfo.timerID, d)
	}
	return timerInfo
}

func (env *testWorkflowEnvironmentImpl) Now() time.Time {
	return env.clock.Now()
}

func (env *testWorkflowEnvironmentImpl) WorkflowInfo() *WorkflowInfo {
	return env.workflowInfo
}

func (env *testWorkflowEnvironmentImpl) RegisterCancel(handler func()) {
	env.workflowCancelHandler = handler
}

func (env *testWorkflowEnvironmentImpl) RequestCancelWorkflow(domainName, workflowID, runID string) error {
	env.workflowCancelHandler()
	return nil
}

func (env *testWorkflowEnvironmentImpl) nextId() int {
	activityID := env.counterID
	env.counterID++
	return activityID
}

// make sure interface is implemented
var _ workflowEnvironment = (*testWorkflowEnvironmentImpl)(nil)
