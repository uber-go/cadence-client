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
type testSampleWorkflow struct {
	t *testing.T
}

func (w testSampleWorkflow) Execute(ctx Context, input []byte) (result []byte, err error) {
	return nil, nil
}

// testActivity2
type testActivity2 struct {
	t *testing.T
}

func (t testActivity2) ActivityType() ActivityType {
	return ActivityType{Name: "testActivity2"}
}

func (t testActivity2) Execute(ctx context.Context, input []byte) ([]byte, error) {
	return nil, nil
}

func TestCreateWorkersForSingleTaskList(t *testing.T) {
	// Create service endpoint
	service := new(mocks.TChanWorkflowService)
	logger := getLogger()

	workflowFactory := func(wt WorkflowType) (Workflow, error) {
		return testSampleWorkflow{}, nil
	}

	// Workflow execution parameters.
	workerOptions := GetWorkerOptions().
		AddWorkflow("taskList", workflowFactory).
		AddActivity("taskList", []Activity{&testActivity{}}).
		WithConcurrentPollSize(10).WithLogger(logger)

	// Launch worker.
	worker := NewWorker(service, workerOptions)
	worker.Start()
}

func TestCreateWorkersForManagingTwoTaskLists(t *testing.T) {
	// Create service endpoint
	service := new(mocks.TChanWorkflowService)
	logger := getLogger()

	workflowFactory := func(wt WorkflowType) (Workflow, error) {
		return testSampleWorkflow{}, nil
	}

	// Workflow execution parameters.
	workerOptions := GetWorkerOptions().
		AddWorkflow("taskList1", workflowFactory).
		AddWorkflow("taskList2", workflowFactory).
		AddActivity("taskList1", []Activity{&testActivity{}}).
		AddActivity("taskList2", []Activity{&testActivity2{}}).
		WithConcurrentPollSize(10).WithLogger(logger)

	// Launch worker.
	worker := NewWorker(service, workerOptions)
	worker.Start()
}

func TestCreateWorkerSeparatelyForWorkflowAndActivityWorker(t *testing.T) {
	// Create service endpoint
	service := new(mocks.TChanWorkflowService)
	logger := getLogger()

	workflowFactory := func(wt WorkflowType) (Workflow, error) {
		return testSampleWorkflow{}, nil
	}

	// Workflow execution parameters.
	workerOptions := GetWorkerOptions().
		AddWorkflow("taskList1", workflowFactory).
		AddWorkflow("taskList2", workflowFactory).
		WithConcurrentPollSize(10).WithLogger(logger)

	// Launch Workflow worker.
	workflowWorker := NewWorker(service, workerOptions)
	workflowWorker.Start()

	// Activity execution parameters.
	activityWorkerOptions := GetWorkerOptions().
		AddActivity("taskList1", []Activity{&testActivity{}}).
		AddActivity("taskList2", []Activity{&testActivity2{}}).
		WithConcurrentPollSize(10).
		WithConcurrentActivityExecutionSize(20).WithLogger(logger)

	// Launch Activity worker.
	activityWorker := NewWorker(service, activityWorkerOptions)
	activityWorker.Start()
}
