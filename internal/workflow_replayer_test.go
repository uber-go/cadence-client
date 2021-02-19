package internal

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/zap"
)

type workflowReplayerSuite struct {
	*require.Assertions
	suite.Suite

	replayer *WorkflowReplayer
	logger   *zap.Logger
}

var (
	testTaskList = "taskList"
)

func TestWorkflowReplayerSuite(t *testing.T) {
	s := new(workflowReplayerSuite)
	suite.Run(t, s)
}

func (s *workflowReplayerSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.logger = getTestLogger(s.T())
	s.replayer = NewWorkflowReplayer()
	s.replayer.RegisterWorkflow(testReplayWorkflow)
	s.replayer.RegisterWorkflow(testReplayWorkflowLocalActivity)
	s.replayer.RegisterWorkflow(testReplayWorkflowFromFile)
	s.replayer.RegisterWorkflow(testReplayWorkflowFromFileParent)
	s.replayer.RegisterWorkflow(localActivitiesCallingOptionsWorkflow{s.T()}.Execute)
}

func (s *workflowReplayerSuite) TestReplayWorkflowHistory_Full() {
	err := s.replayer.ReplayWorkflowHistory(s.logger, getTestReplayWorkflowFullHistory())
	s.NoError(err)
}

func (s *workflowReplayerSuite) TestReplayWorkflowHistory_Partial() {
	err := s.replayer.ReplayWorkflowHistory(s.logger, getTestReplayWorkflowPartialHistory())
	s.NoError(err)
}

func (s *workflowReplayerSuite) TestReplayWorkflowHistory_LocalActivity() {
	err := s.replayer.ReplayWorkflowHistory(s.logger, getTestReplayWorkflowLocalActivityHistory())
	s.NoError(err)
}

func (s *workflowReplayerSuite) TestReplayWorkflowHistory_LocalActivity_Result_Mismatch() {
	err := s.replayer.ReplayWorkflowHistory(s.logger, getTestReplayWorkflowLocalActivityResultMismatchHistory())
	s.Error(err)
}

func (s *workflowReplayerSuite) TestReplayWorkflowHistory_LocalActivity_Activity_Type_Mismatch() {
	err := s.replayer.ReplayWorkflowHistory(s.logger, getTestReplayWorkflowLocalActivityTypeMismatchHistory())
	s.Error(err)
}

func (s *workflowReplayerSuite) TestReplayWorkflowHistoryFromFileLocalActivities() {
	err := s.replayer.ReplayWorkflowHistoryFromJSONFile(s.logger, "testdata/localActivities.json")
	s.NoError(err)
}

func (s *workflowReplayerSuite) TestReplayWorkflowHistoryFromFileParent() {
	err := s.replayer.ReplayWorkflowHistoryFromJSONFile(s.logger, "testdata/parentWF.json")
	s.NoError(err)
}

func (s *workflowReplayerSuite) TestReplayWorkflowHistoryFromFile() {
	err := s.replayer.ReplayWorkflowHistoryFromJSONFile(s.logger, "testdata/sampleHistory.json")
	s.NoError(err)
}

func testReplayWorkflow(ctx Context) error {
	ao := ActivityOptions{
		ScheduleToStartTimeout: time.Second,
		StartToCloseTimeout:    time.Second,
	}
	ctx = WithActivityOptions(ctx, ao)
	err := ExecuteActivity(ctx, "testActivity").Get(ctx, nil)
	if err != nil {
		GetLogger(ctx).Error("activity failed with error.", zap.Error(err))
		panic("Failed workflow")
	}
	return err
}

func testReplayWorkflowLocalActivity(ctx Context) error {
	ao := LocalActivityOptions{
		ScheduleToCloseTimeout: time.Second,
	}
	ctx = WithLocalActivityOptions(ctx, ao)
	err := ExecuteLocalActivity(ctx, testActivity).Get(ctx, nil)
	if err != nil {
		GetLogger(ctx).Error("activity failed with error.", zap.Error(err))
		panic("Failed workflow")
	}
	return err
}

func testReplayWorkflowFromFile(ctx Context) error {
	ao := ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       20 * time.Second,
		WaitForCancellation:    true,
	}
	ctx = WithActivityOptions(ctx, ao)
	err := ExecuteActivity(ctx, "testActivityMultipleArgs", 2, "test", true).Get(ctx, nil)
	if err != nil {
		GetLogger(ctx).Error("activity failed with error.", zap.Error(err))
		panic("Failed workflow")
	}
	return err
}

func testReplayWorkflowFromFileParent(ctx Context) error {
	execution := GetWorkflowInfo(ctx).WorkflowExecution
	childID := fmt.Sprintf("child_workflow:%v", execution.RunID)
	cwo := ChildWorkflowOptions{
		WorkflowID:                   childID,
		ExecutionStartToCloseTimeout: time.Minute,
	}
	ctx = WithChildWorkflowOptions(ctx, cwo)
	var result string
	cwf := ExecuteChildWorkflow(ctx, testReplayWorkflowFromFile)
	f1 := cwf.SignalChildWorkflow(ctx, "test-signal", "test-data")
	err := f1.Get(ctx, nil)
	if err != nil {
		return err
	}
	return cwf.Get(ctx, &result)
}

func testActivity(ctx context.Context) error {
	return nil
}

type localActivitiesCallingOptionsWorkflow struct {
	t *testing.T
}

func (w localActivitiesCallingOptionsWorkflow) Execute(ctx Context, input []byte) (result []byte, err error) {
	ao := LocalActivityOptions{
		ScheduleToCloseTimeout: time.Second,
	}
	ctx = WithLocalActivityOptions(ctx, ao)

	// By functions.
	err = ExecuteLocalActivity(ctx, testActivityByteArgs, input).Get(ctx, nil)
	require.NoError(w.t, err, err)

	err = ExecuteLocalActivity(ctx, testActivityMultipleArgs, 2, []string{"test"}, true).Get(ctx, nil)
	require.NoError(w.t, err, err)

	err = ExecuteLocalActivity(ctx, testActivityNoResult, 2, "test").Get(ctx, nil)
	require.NoError(w.t, err, err)

	err = ExecuteLocalActivity(ctx, testActivityNoContextArg, 2, "test").Get(ctx, nil)
	require.NoError(w.t, err, err)

	f := ExecuteLocalActivity(ctx, testActivityReturnByteArray)
	var r []byte
	err = f.Get(ctx, &r)
	require.NoError(w.t, err, err)
	require.Equal(w.t, []byte("testActivity"), r)

	f = ExecuteLocalActivity(ctx, testActivityReturnInt)
	var rInt int
	err = f.Get(ctx, &rInt)
	require.NoError(w.t, err, err)
	require.Equal(w.t, 5, rInt)

	f = ExecuteLocalActivity(ctx, testActivityReturnString)
	var rString string
	err = f.Get(ctx, &rString)

	require.NoError(w.t, err, err)
	require.Equal(w.t, "testActivity", rString)

	f = ExecuteLocalActivity(ctx, testActivityReturnEmptyString)
	var r2String string
	err = f.Get(ctx, &r2String)
	require.NoError(w.t, err, err)
	require.Equal(w.t, "", r2String)

	f = ExecuteLocalActivity(ctx, testActivityReturnEmptyStruct)
	var r2Struct testActivityResult
	err = f.Get(ctx, &r2Struct)
	require.NoError(w.t, err, err)
	require.Equal(w.t, testActivityResult{}, r2Struct)

	f = ExecuteLocalActivity(ctx, testActivityReturnNilStructPtr)
	var rStructPtr *testActivityResult
	err = f.Get(ctx, &rStructPtr)
	require.NoError(w.t, err, err)
	require.True(w.t, rStructPtr == nil)

	f = ExecuteLocalActivity(ctx, testActivityReturnStructPtr)
	err = f.Get(ctx, &rStructPtr)
	require.NoError(w.t, err, err)
	require.Equal(w.t, *rStructPtr, testActivityResult{Index: 10})

	f = ExecuteLocalActivity(ctx, testActivityReturnNilStructPtrPtr)
	var rStruct2Ptr **testActivityResult
	err = f.Get(ctx, &rStruct2Ptr)
	require.NoError(w.t, err, err)
	require.True(w.t, rStruct2Ptr == nil)

	f = ExecuteLocalActivity(ctx, testActivityReturnStructPtrPtr)
	err = f.Get(ctx, &rStruct2Ptr)
	require.NoError(w.t, err, err)
	require.True(w.t, **rStruct2Ptr == testActivityResult{Index: 10})

	return []byte("Done"), nil
}

func getTestReplayWorkflowFullHistory() *shared.History {
	return &shared.History{
		Events: []*shared.HistoryEvent{
			createTestEventWorkflowExecutionStarted(1, &shared.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &shared.WorkflowType{Name: common.StringPtr("go.uber.org/cadence/internal.testReplayWorkflow")},
				TaskList:     &shared.TaskList{Name: common.StringPtr(testTaskList)},
				Input:        testEncodeFunctionArgs(getDefaultDataConverter()),
			}),
			createTestEventDecisionTaskScheduled(2, &shared.DecisionTaskScheduledEventAttributes{}),
			createTestEventDecisionTaskStarted(3),
			createTestEventDecisionTaskCompleted(4, &shared.DecisionTaskCompletedEventAttributes{}),
			createTestEventActivityTaskScheduled(5, &shared.ActivityTaskScheduledEventAttributes{
				ActivityId:   common.StringPtr("0"),
				ActivityType: &shared.ActivityType{Name: common.StringPtr("testActivity")},
				TaskList:     &shared.TaskList{Name: &testTaskList},
			}),
			createTestEventActivityTaskStarted(6, &shared.ActivityTaskStartedEventAttributes{
				ScheduledEventId: common.Int64Ptr(5),
			}),
			createTestEventActivityTaskCompleted(7, &shared.ActivityTaskCompletedEventAttributes{
				ScheduledEventId: common.Int64Ptr(5),
				StartedEventId:   common.Int64Ptr(6),
			}),
			createTestEventDecisionTaskScheduled(8, &shared.DecisionTaskScheduledEventAttributes{}),
			createTestEventDecisionTaskStarted(9),
			createTestEventDecisionTaskCompleted(10, &shared.DecisionTaskCompletedEventAttributes{
				ScheduledEventId: common.Int64Ptr(8),
				StartedEventId:   common.Int64Ptr(9),
			}),
			createTestEventWorkflowExecutionCompleted(11, &shared.WorkflowExecutionCompletedEventAttributes{
				DecisionTaskCompletedEventId: common.Int64Ptr(10),
			}),
		},
	}
}

func getTestReplayWorkflowPartialHistory() *shared.History {
	return &shared.History{
		Events: []*shared.HistoryEvent{
			createTestEventWorkflowExecutionStarted(1, &shared.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &shared.WorkflowType{Name: common.StringPtr("go.uber.org/cadence/internal.testReplayWorkflow")},
				TaskList:     &shared.TaskList{Name: common.StringPtr(testTaskList)},
				Input:        testEncodeFunctionArgs(getDefaultDataConverter()),
			}),
			createTestEventDecisionTaskScheduled(2, &shared.DecisionTaskScheduledEventAttributes{}),
			createTestEventDecisionTaskStarted(3),
			createTestEventDecisionTaskCompleted(4, &shared.DecisionTaskCompletedEventAttributes{}),
			createTestEventActivityTaskScheduled(5, &shared.ActivityTaskScheduledEventAttributes{
				ActivityId:   common.StringPtr("0"),
				ActivityType: &shared.ActivityType{Name: common.StringPtr("testActivity-fm")},
				TaskList:     &shared.TaskList{Name: &testTaskList},
			}),
		},
	}
}

func getTestReplayWorkflowLocalActivityHistory() *shared.History {
	return &shared.History{
		Events: []*shared.HistoryEvent{
			createTestEventWorkflowExecutionStarted(1, &shared.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &shared.WorkflowType{Name: common.StringPtr("go.uber.org/cadence/internal.testReplayWorkflowLocalActivity")},
				TaskList:     &shared.TaskList{Name: common.StringPtr(testTaskList)},
				Input:        testEncodeFunctionArgs(getDefaultDataConverter()),
			}),
			createTestEventDecisionTaskScheduled(2, &shared.DecisionTaskScheduledEventAttributes{}),
			createTestEventDecisionTaskStarted(3),
			createTestEventDecisionTaskCompleted(4, &shared.DecisionTaskCompletedEventAttributes{}),

			createTestEventLocalActivity(5, &shared.MarkerRecordedEventAttributes{
				MarkerName:                   common.StringPtr(localActivityMarkerName),
				Details:                      createLocalActivityMarkerDataForTest("0", "go.uber.org/cadence/internal.testActivity"),
				DecisionTaskCompletedEventId: common.Int64Ptr(4),
			}),

			createTestEventWorkflowExecutionCompleted(6, &shared.WorkflowExecutionCompletedEventAttributes{
				DecisionTaskCompletedEventId: common.Int64Ptr(4),
			}),
		},
	}
}

func getTestReplayWorkflowLocalActivityResultMismatchHistory() *shared.History {
	return &shared.History{
		Events: []*shared.HistoryEvent{
			createTestEventWorkflowExecutionStarted(1, &shared.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &shared.WorkflowType{Name: common.StringPtr("go.uber.org/cadence/internal.testReplayWorkflowLocalActivity")},
				TaskList:     &shared.TaskList{Name: common.StringPtr(testTaskList)},
				Input:        testEncodeFunctionArgs(getDefaultDataConverter()),
			}),
			createTestEventDecisionTaskScheduled(2, &shared.DecisionTaskScheduledEventAttributes{}),
			createTestEventDecisionTaskStarted(3),
			createTestEventDecisionTaskCompleted(4, &shared.DecisionTaskCompletedEventAttributes{}),

			createTestEventLocalActivity(5, &shared.MarkerRecordedEventAttributes{
				MarkerName:                   common.StringPtr(localActivityMarkerName),
				Details:                      createLocalActivityMarkerDataForTest("0", ""),
				DecisionTaskCompletedEventId: common.Int64Ptr(4),
			}),

			createTestEventWorkflowExecutionCompleted(6, &shared.WorkflowExecutionCompletedEventAttributes{
				Result:                       []byte("some-incorrect-result"),
				DecisionTaskCompletedEventId: common.Int64Ptr(4),
			}),
		},
	}
}

func getTestReplayWorkflowLocalActivityTypeMismatchHistory() *shared.History {
	return &shared.History{
		Events: []*shared.HistoryEvent{
			createTestEventWorkflowExecutionStarted(1, &shared.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &shared.WorkflowType{Name: common.StringPtr("go.uber.org/cadence/internal.testReplayWorkflowLocalActivity")},
				TaskList:     &shared.TaskList{Name: common.StringPtr(testTaskList)},
				Input:        testEncodeFunctionArgs(getDefaultDataConverter()),
			}),
			createTestEventDecisionTaskScheduled(2, &shared.DecisionTaskScheduledEventAttributes{}),
			createTestEventDecisionTaskStarted(3),
			createTestEventDecisionTaskCompleted(4, &shared.DecisionTaskCompletedEventAttributes{}),

			createTestEventLocalActivity(5, &shared.MarkerRecordedEventAttributes{
				MarkerName:                   common.StringPtr(localActivityMarkerName),
				Details:                      createLocalActivityMarkerDataForTest("0", "different-activity-type"),
				DecisionTaskCompletedEventId: common.Int64Ptr(4),
			}),

			createTestEventWorkflowExecutionCompleted(6, &shared.WorkflowExecutionCompletedEventAttributes{
				DecisionTaskCompletedEventId: common.Int64Ptr(4),
			}),
		},
	}
}

func createLocalActivityMarkerDataForTest(activityID, activityType string) []byte {
	lamd := localActivityMarkerData{
		ActivityID:   activityID,
		ActivityType: activityType,
		ReplayTime:   time.Now(),
	}

	// encode marker data
	markerData, err := encodeArg(nil, lamd)
	if err != nil {
		panic(fmt.Sprintf("error encoding local activity marker data: %v", err))
	}
	return markerData
}
