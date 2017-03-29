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

	// Simulate initialization
	workflowFactory := func(wt WorkflowType) (Workflow, error) {
		return testSampleWorkflow{}, nil
	}
	RegisterWorkflow("taskListWorkflow", workflowFactory)
	RegisterActivity("taskListActivity", []Activity{&testActivity{}})

	// Configure task lists and worker
	workflowTask := NewWorkflowTask("taskListWorkflow")
	activityTask := NewActivityTask("taskListActivity").SetActivityExecutionRate(20)
	workerOptions := NewWorkerOptions().SetLogger(logger)

	// Start Worker.
	worker := NewWorker(
		service,
		[]WorkflowTask{workflowTask},
		[]ActivityTask{activityTask},
		workerOptions)
	err := worker.Start()
	require.NoError(t, err)
}

func TestCreateWorkersForManagingTwoTaskLists(t *testing.T) {
	// Create service endpoint
	service := new(mocks.TChanWorkflowService)
	logger := getLogger()

	// Simulate initialization
	workflowFactory := func(wt WorkflowType) (Workflow, error) {
		return testSampleWorkflow{}, nil
	}
	RegisterWorkflow("taskList1", workflowFactory)
	RegisterWorkflow("taskList2", workflowFactory)
	RegisterActivity("taskList-act-1", []Activity{&testActivity{}})
	RegisterActivity("taskList-act-2", []Activity{&testActivity2{}})

	// Configure task lists and worker
	wt1 := NewWorkflowTask("taskList1")
	wt2 := NewWorkflowTask("taskList2")
	at1 := NewActivityTask("taskList-act-1").SetActivityExecutionRate(20)
	at2 := NewActivityTask("taskList-act-2").SetActivityExecutionRate(20)
	workerOptions := NewWorkerOptions().SetLogger(logger)

	// Start Worker.
	worker := NewWorker(
		service,
		[]WorkflowTask{wt1, wt2},
		[]ActivityTask{at1, at2},
		workerOptions)
	err := worker.Start()
	require.NoError(t, err)
}

func TestCreateWorkerSeparatelyForWorkflowAndActivityWorker(t *testing.T) {
	// Create service endpoint
	service := new(mocks.TChanWorkflowService)
	logger := getLogger()

	// Simulate initialization
	workflowFactory := func(wt WorkflowType) (Workflow, error) {
		return testSampleWorkflow{}, nil
	}
	RegisterWorkflow("taskList1", workflowFactory)
	RegisterWorkflow("taskList2", workflowFactory)
	RegisterActivity("taskList-act-1", []Activity{&testActivity{}})
	RegisterActivity("taskList-act-2", []Activity{&testActivity2{}})

	// Configure task lists and worker
	wt1 := NewWorkflowTask("taskList1")
	wt2 := NewWorkflowTask("taskList2")
	workerOptions := NewWorkerOptions().SetLogger(logger)

	// Start workflow Worker.
	worker := NewWorker(
		service,
		[]WorkflowTask{wt1, wt2},
		nil,
		workerOptions)
	err := worker.Start()
	require.NoError(t, err)

	at1 := NewActivityTask("taskList-act-1").SetActivityExecutionRate(20)
	at2 := NewActivityTask("taskList-act-2").SetActivityExecutionRate(20).SetAutoHeartBeat(true)

	// Start activity Worker.
	aWorker := NewWorker(
		service,
		nil,
		[]ActivityTask{at1, at2},
		workerOptions)
	err = aWorker.Start()
	require.NoError(t, err)
}
