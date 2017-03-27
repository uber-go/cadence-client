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

	options := NewTaskOptions().SetLogger(logger)

	// Register with cadence
	dt := NewDecisionTask("taskListWorkflow").RegisterWorkflow(workflowFactory).SetTaskOptions(options)
	at := NewActivityTask("taskListActivity").RegisterActivity([]Activity{&testActivity{}}).SetTaskOptions(options)

	// Start Cadence.
	_, err := Start(service, []DecisionTask{dt}, []ActivityTask{at})
	require.NoError(t, err)
}

func TestCreateWorkersForManagingTwoTaskLists(t *testing.T) {
	// Create service endpoint
	service := new(mocks.TChanWorkflowService)
	logger := getLogger()
	options := NewTaskOptions().SetLogger(logger)

	workflowFactory := func(wt WorkflowType) (Workflow, error) {
		return testSampleWorkflow{}, nil
	}

	// Register with cadence
	dt1 := NewDecisionTask("taskList1").RegisterWorkflow(workflowFactory).SetTaskOptions(options)
	dt2 := NewDecisionTask("taskList2").RegisterWorkflow(workflowFactory).SetTaskOptions(options)
	at1 := NewActivityTask("taskList-act-1").RegisterActivity([]Activity{&testActivity{}}).SetTaskOptions(options)
	at2 := NewActivityTask("taskList-act-2").RegisterActivity([]Activity{&testActivity2{}}).SetTaskOptions(options)

	// Start Cadence.
	_, err := Start(service, []DecisionTask{dt1, dt2}, []ActivityTask{at1, at2})
	require.NoError(t, err)
}

func TestCreateWorkerSeparatelyForWorkflowAndActivityWorker(t *testing.T) {
	// Create service endpoint
	service := new(mocks.TChanWorkflowService)
	logger := getLogger()
	options := NewTaskOptions().SetLogger(logger)

	workflowFactory := func(wt WorkflowType) (Workflow, error) {
		return testSampleWorkflow{}, nil
	}

	// Register workflow with cadence
	dt1 := NewDecisionTask("taskList1").RegisterWorkflow(workflowFactory).SetTaskOptions(options)
	dt2 := NewDecisionTask("taskList2").RegisterWorkflow(workflowFactory).SetTaskOptions(options)

	// Start Cadence hosting workflows.
	_, err := Start(service, []DecisionTask{dt1, dt2}, nil)
	require.NoError(t, err)

	// Register activities with cadence
	at1 := NewActivityTask("taskList-act-1").RegisterActivity([]Activity{&testActivity{}}).SetTaskOptions(options)
	at2 := NewActivityTask("taskList-act-2").RegisterActivity([]Activity{&testActivity2{}}).
		SetTaskOptions(options).
		SetAutoHeartBeat(true).
		SetMaxConcurrentActivityExecutionSize(5).
		SetActivityExecutionRate(100)

	// Start Cadence hosting activites.
	_, err = Start(service, nil, []ActivityTask{at1, at2})
	require.NoError(t, err)
}
