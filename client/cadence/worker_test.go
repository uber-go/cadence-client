package cadence

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uber-common/bark"
	s "github.com/uber-go/cadence-client/.gen/go/shared"
	"github.com/uber-go/cadence-client/common"
	"github.com/uber-go/cadence-client/mocks"
	"reflect"
)

func getLogger() bark.Logger {
	formatter := &log.TextFormatter{}
	formatter.FullTimestamp = true
	log1 := log.New()
	log1.Level = log.DebugLevel
	log1.Formatter = formatter
	return bark.NewLoggerFromLogrus(log1)
}

type testReplayWorkflow struct {
	t *testing.T
}

func (w testReplayWorkflow) Execute(ctx Context, input []byte) (result []byte, err error) {
	ctx = WithTaskList(ctx, "testTaskList")
	ctx = WithScheduleToStartTimeout(ctx, time.Second)
	ctx = WithScheduleToCloseTimeout(ctx, time.Second)
	ctx = WithStartToCloseTimeout(ctx, time.Second)
	r, err := ExecuteActivity(ctx, "testActivity", []byte("test"))
	require.NoError(w.t, err, err.Error())
	return r, err
}

type testActivity struct {
	t *testing.T
}

func (t testActivity) ActivityType() ActivityType {
	return ActivityType{Name: "testActivity"}
}

func (t testActivity) Execute(ctx context.Context, input []byte) ([]byte, error) {
	return nil, nil
}

func TestWorkflowReplayer(t *testing.T) {
	logger := getLogger()
	taskList := "taskList1"
	testEvents := []*s.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &s.WorkflowExecutionStartedEventAttributes{
			TaskList: &s.TaskList{Name: common.StringPtr(taskList)},
		}),
		createTestEventDecisionTaskScheduled(2, &s.DecisionTaskScheduledEventAttributes{}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &s.DecisionTaskCompletedEventAttributes{}),
		createTestEventActivityTaskScheduled(2, &s.ActivityTaskScheduledEventAttributes{
			ActivityId: common.StringPtr("0"),
		}),
		createTestEventActivityTaskStarted(3, &s.ActivityTaskStartedEventAttributes{}),
	}

	options := WorkflowReplayerOptions{
		Execution: WorkflowExecution{ID: "testID", RunID: "testRunID"},
		Type:      WorkflowType{Name: "testWorkflow"},
		Factory:   func(workflowType WorkflowType) (Workflow, error) { return testReplayWorkflow{t: t}, nil },
		History:   &s.History{Events: testEvents},
	}

	r := NewWorkflowReplayer(options, logger)
	err := r.Process(true)
	require.NoError(t, err)
	require.NotEmpty(t, r.StackTrace(), r.StackTrace())
	require.Contains(t, r.StackTrace(), "cadence.ExecuteActivity")
}

// testSampleWorkflow
func sampleWorkflowExecute(ctx Context, input []byte) (result []byte, err error) {
	ExecuteActivity(ctx, testActivityByteArgs, input)
	ExecuteActivity(ctx, testActivityMultipleArgs, 2, "test", true)
	return []byte("Done"), nil
}

// test activity1
func testActivityByteArgs(ctx context.Context, input []byte) ([]byte, error) {
	fmt.Println("Executing Activity1")
	return nil, nil
}

// test testActivityMultipleArgs
func testActivityMultipleArgs(ctx context.Context, arg1 int, arg2 string, arg3 bool) ([]byte, error) {
	fmt.Println("Executing Activity2")
	return nil, nil
}

func TestCreateWorkersForSingleWorkflowAndActivity(t *testing.T) {
	// Create service endpoint
	service := new(mocks.TChanWorkflowService)
	logger := getLogger()

	// mocks
	service.On("PollForActivityTask", mock.Anything, mock.Anything).Return(&s.PollForActivityTaskResponse{}, nil)
	service.On("RespondActivityTaskCompleted", mock.Anything, mock.Anything).Return(nil)
	service.On("PollForDecisionTask", mock.Anything, mock.Anything).Return(&s.PollForDecisionTaskResponse{}, nil)
	service.On("RespondDecisionTaskCompleted", mock.Anything, mock.Anything).Return(nil)

	// Simulate initialization
	RegisterWorkflow(sampleWorkflowExecute)
	RegisterActivity(testActivityByteArgs)

	// Configure worker options.
	workerOptions := NewWorkerOptions().SetLogger(logger)

	// Start Worker.
	worker := NewWorker(
		service,
		"testGroup",
		workerOptions)
	err := worker.Start()
	require.NoError(t, err)
	worker.Stop()
}

func TestCreateWorkersForManagingMultipleActivities(t *testing.T) {
	// Create service endpoint
	service := new(mocks.TChanWorkflowService)
	logger := getLogger()

	// mocks
	service.On("PollForActivityTask", mock.Anything, mock.Anything).Return(&s.PollForActivityTaskResponse{}, nil)
	service.On("RespondActivityTaskCompleted", mock.Anything, mock.Anything).Return(nil)
	service.On("PollForDecisionTask", mock.Anything, mock.Anything).Return(&s.PollForDecisionTaskResponse{}, nil)
	service.On("RespondDecisionTaskCompleted", mock.Anything, mock.Anything).Return(nil)

	// Simulate initialization
	RegisterWorkflow(sampleWorkflowExecute)
	RegisterActivity(testActivityByteArgs)
	RegisterActivity(testActivityMultipleArgs)

	// Configure worker options.
	workerOptions := NewWorkerOptions().SetLogger(logger).SetActivityExecutionRate(20)

	// Start Worker.
	worker := NewWorker(
		service,
		"testGroupName2",
		workerOptions)
	err := worker.Start()
	require.NoError(t, err)
	worker.Stop()
}

func TestCreateWorkerForWorkflow(t *testing.T) {
	// Create service endpoint
	service := new(mocks.TChanWorkflowService)
	logger := getLogger()

	setHostEnvironment(nil)

	// mocks
	service.On("PollForDecisionTask", mock.Anything, mock.Anything).Return(&s.PollForDecisionTaskResponse{}, nil)
	service.On("RespondDecisionTaskCompleted", mock.Anything, mock.Anything).Return(nil)

	// Simulate initialization
	RegisterWorkflow(sampleWorkflowExecute)

	// Configure worker
	workerOptions := NewWorkerOptions().SetLogger(logger)

	// Start workflow Worker.
	worker := NewWorker(
		service,
		"testGroup",
		workerOptions)
	err := worker.Start()
	require.NoError(t, err)
	worker.Stop()
}

func TestCompleteActivity(t *testing.T) {
	mockService := new(mocks.TChanWorkflowService)
	wfClient := NewWorkflowClient(mockService, nil, "")
	var completedRequest, canceledRequest, failedRequest interface{}
	mockService.On("RespondActivityTaskCompleted", mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			completedRequest = args.Get(1).(*s.RespondActivityTaskCompletedRequest)
		})
	mockService.On("RespondActivityTaskCanceled", mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			canceledRequest = args.Get(1).(*s.RespondActivityTaskCanceledRequest)
		})
	mockService.On("RespondActivityTaskFailed", mock.Anything, mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			failedRequest = args.Get(1).(*s.RespondActivityTaskFailedRequest)
		})

	wfClient.CompleteActivity(nil, nil, nil)
	require.NotNil(t, completedRequest)

	wfClient.CompleteActivity(nil, nil, NewCanceledError())
	require.NotNil(t, canceledRequest)

	wfClient.CompleteActivity(nil, nil, errors.New(""))
	require.NotNil(t, failedRequest)
}

func TestRecordActivityHeartbeat(t *testing.T) {
	mockService := new(mocks.TChanWorkflowService)
	wfClient := NewWorkflowClient(mockService, nil, "")
	var heartbeatRequest *s.RecordActivityTaskHeartbeatRequest
	cancelRequested := false
	heartbeatResponse := s.RecordActivityTaskHeartbeatResponse{CancelRequested: &cancelRequested}
	mockService.On("RecordActivityTaskHeartbeat", mock.Anything, mock.Anything).Return(&heartbeatResponse, nil).
		Run(func(args mock.Arguments) {
			heartbeatRequest = args.Get(1).(*s.RecordActivityTaskHeartbeatRequest)
		})

	wfClient.RecordActivityHeartbeat(nil, nil)
	require.NotNil(t, heartbeatRequest)
}

func testEncodeFunction(t *testing.T, f interface{}, args ...interface{}) string {
	s := fnSignature{FnName: getFunctionName(f), Args: args}
	input, err := getHostEnvironment().Encoder().Marshal(s)
	require.NoError(t, err, err)
	require.NotNil(t, input)

	var s2 fnSignature
	err = getHostEnvironment().Encoder().Unmarshal(input, &s2)
	require.NoError(t, err, err)

	targetArgs := []reflect.Value{}
	for _, arg := range s2.Args {
		targetArgs = append(targetArgs, reflect.ValueOf(arg))
	}
	fnValue := reflect.ValueOf(f)
	retValues := fnValue.Call(targetArgs)
	return retValues[0].Interface().(string)
}

func testEncodeWithName(t *testing.T, args ...interface{}) {
	s := fnSignature{FnName: "test", Args: args}
	input, err := getHostEnvironment().Encoder().Marshal(s)
	require.NoError(t, err, err)
	require.NotNil(t, input)

	var s2 fnSignature
	err = getHostEnvironment().Encoder().Unmarshal(input, &s2)
	require.NoError(t, err, err)

	require.Equal(t, s.FnName, s2.FnName)
	require.Equal(t, len(s.Args), len(s2.Args))
	require.Equal(t, s.Args, s2.Args)
}

func TestEncoder(t *testing.T) {
	testEncodeWithName(t, 2, 3)
	testEncodeWithName(t, nil)
	testEncodeWithName(t)

	getHostEnvironment().Encoder().Register(new(emptyCtx))
	// Two param functor.
	f1 := func(ctx Context, r []byte) string {
		return "result"
	}
	r1 := testEncodeFunction(t, f1, new(emptyCtx), []byte("test"))
	require.Equal(t, r1, "result")
	// No parameters.
	f2 := func() string {
		return "empty-result"
	}
	r2 := testEncodeFunction(t, f2)
	require.Equal(t, r2, "empty-result")
	// Nil parameter.
	f3 := func(r []byte) string {
		return "nil-result"
	}
	r3 := testEncodeFunction(t, f3, []byte(""))
	require.Equal(t, r3, "nil-result")
}

type activitiesCallingOptionsWorkflow struct{
	t *testing.T
}

func (w activitiesCallingOptionsWorkflow) Execute(ctx Context, input []byte) (result []byte, err error) {
	ctx = WithActivityOptions(ctx, NewActivityOptions().
		WithTaskList("exampleTaskList").
		WithScheduleToStartTimeout(10 * time.Second).
		WithStartToCloseTimeout(5 * time.Second).
		WithScheduleToCloseTimeout(10*time.Second))

	// By functions.
	_, err = ExecuteActivity(ctx, testActivityByteArgs, input)
	fmt.Printf("activitiesCallingOptionsWorkflow: %v \n", err)
	require.NoError(w.t, err, err)

	_, err = ExecuteActivity(ctx, testActivityMultipleArgs, 2, "test", true)
	require.NoError(w.t, err, err)

	_, err = ExecuteActivity(ctx, testActivityNoResult, 2, "test")
	require.NoError(w.t, err, err)

	_, err = ExecuteActivity(ctx, testActivityNoContextArg, 2, "test")
	require.NoError(w.t, err, err)

	_, err = ExecuteActivity(ctx, testActivityNoError, 2, "test")
	require.NoError(w.t, err, err)

	_, err = ExecuteActivity(ctx, testActivityNoArgsAndNoResult)
	require.NoError(w.t, err, err)

	// By names.
	_, err = ExecuteActivity(ctx, "testActivityByteArgs", input)
	require.NoError(w.t, err, err)

	_, err = ExecuteActivity(ctx, "testActivityMultipleArgs", 2, "test", true)
	require.NoError(w.t, err, err)

	_, err = ExecuteActivity(ctx, "testActivityNoResult", 2, "test")
	require.NoError(w.t, err, err)

	_, err = ExecuteActivity(ctx, "testActivityNoContextArg", 2, "test")
	require.NoError(w.t, err, err)

	_, err = ExecuteActivity(ctx, "testActivityNoError", 2, "test")
	require.NoError(w.t, err, err)

	_, err = ExecuteActivity(ctx, "testActivityNoArgsAndNoResult")
	require.NoError(w.t, err, err)

	return []byte("Done"), nil
}

// test testActivityNoResult
func testActivityNoResult(ctx context.Context, arg1 int, arg2 string) (error) {
	return nil
}

// test testActivityNoContextArg
func testActivityNoContextArg(arg1 int, arg2 string) (error) {
	return nil
}

// test testActivityNoError
func testActivityNoError(arg1 int, arg2 string) {
	return
}

// test testActivityNoError
func testActivityNoArgsAndNoResult() {
	return
}

func TestVariousActivitySchedulingOption(t *testing.T) {
	w := NewWorkflowDefinition(&activitiesCallingOptionsWorkflow{t:t})
	ctx := &MockWorkflowEnvironment{}
	workflowComplete := make(chan struct{}, 1)

	cbProcessor := newAsyncTestCallbackProcessor()

	ctx.On("ExecuteActivity", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		callback := args.Get(1).(resultHandler)
		cbProcessor.Add(callback, []byte("test"), nil)
	}).Times(12)

	ctx.On("Complete", mock.Anything, mock.Anything).Return().Run(func(args mock.Arguments) {
		if args.Get(1) != nil {
			err := args.Get(1).(Error)
			fmt.Printf("Error: %v, Stack: %v \n", err.Reason(), string(err.Details()))
		}
		workflowComplete <- struct{}{}
	}).Once()

	w.Execute(ctx, []byte(""))

	c := cbProcessor.ProcessOrWait(workflowComplete)
	require.True(t, c, "Workflow failed to complete")
	ctx.AssertExpectations(t)
}
