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
)

func getLogger() bark.Logger {
	formatter := &log.TextFormatter{}
	formatter.FullTimestamp = true
	log.SetFormatter(formatter)
	log.SetLevel(log.DebugLevel)
	return bark.NewLoggerFromLogrus(log.New())

}

type testReplayWorkflow struct {
	t *testing.T
}

func (w testReplayWorkflow) Execute(ctx Context, input []byte) (result []byte, err error) {
	ctx = WithTaskList(ctx, "testTaskList")
	ctx = WithScheduleToStartTimeout(ctx, time.Second)
	ctx = WithScheduleToCloseTimeout(ctx, time.Second)
	ctx = WithStartToCloseTimeout(ctx, time.Second)
	r, err := ExecuteActivity(ctx, ActivityType{Name: "testActivity"}, nil)
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
		Factory:   func(workflowType WorkflowType) (Workflow, error) { return testReplayWorkflow{}, nil },
		History:   &s.History{Events: testEvents},
	}

	r := NewWorkflowReplayer(options, logger)
	err := r.Process(true)
	require.NoError(t, err)
	require.NotEmpty(t, r.StackTrace())
	require.Contains(t, r.StackTrace(), "cadence.ExecuteActivity")
}

// testSampleWorkflow
func sampleWorkflowExecute(ctx Context, input []byte) (result []byte, err error) {
	ExecuteActivityFn(ctx, testActivity1Execute, input)
	ExecuteActivityFn(ctx, testActivity2Execute, input)
	return nil, nil
}

// test activity1
func testActivity1Execute(ctx context.Context, input []byte) ([]byte, error) {
	fmt.Println("Executing Activity1")
	return nil, nil
}

// test activity1
func testActivity2Execute(ctx context.Context, input []byte) ([]byte, error) {
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
	RegisterActivity(testActivity1Execute)

	// Configure task lists and worker
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
	RegisterActivity(testActivity1Execute)
	RegisterActivity(testActivity2Execute)

	// Configure task lists and worker
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
