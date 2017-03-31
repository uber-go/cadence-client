package cadence

import (
	"context"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/uber-common/bark"
	s "github.com/uber-go/cadence-client/.gen/go/shared"
	"github.com/uber-go/cadence-client/common"
	"github.com/uber-go/cadence-client/mocks"
	"github.com/stretchr/testify/mock"
	"fmt"
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
	r, err := ExecuteActivity(ctx, ExecuteActivityParameters{
		ActivityType: ActivityType{Name: "testActivity"},
		TaskListName: "testTaskList"})
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
	ExecuteActivity(ctx, )
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
