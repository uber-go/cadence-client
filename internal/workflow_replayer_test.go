// Copyright (c) 2017-2021 Uber Technologies Inc.
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

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
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
	s.replayer = NewWorkflowReplayerWithOptions(ReplayOptions{
		ContextPropagators: []ContextPropagator{
			NewStringMapPropagator([]string{testHeader}),
		},
	})
	s.replayer.RegisterWorkflow(testReplayWorkflow)
	s.replayer.RegisterWorkflow(testReplayWorkflowLocalActivity)
	s.replayer.RegisterWorkflow(testReplayWorkflowContextPropagator)
	s.replayer.RegisterWorkflow(testReplayWorkflowFromFile)
	s.replayer.RegisterWorkflow(testReplayWorkflowFromFileParent)
	s.replayer.RegisterWorkflow(localActivitiesCallingOptionsWorkflow{s.T()}.Execute)
}

func (s *workflowReplayerSuite) TestReplayWorkflowHistory_Full() {
	err := s.replayer.ReplayWorkflowHistory(s.logger, getTestReplayWorkflowFullHistory())
	s.NoError(err)
}

func (s *workflowReplayerSuite) TestReplayWorkflowHistory_Full_ResultMisMatch() {
	fullHistory := getTestReplayWorkflowFullHistory()
	completedEvent := fullHistory.Events[len(fullHistory.Events)-1]
	s.NotNil(completedEvent.GetWorkflowExecutionCompletedEventAttributes())
	completedEvent.GetWorkflowExecutionCompletedEventAttributes().Result = &apiv1.Payload{Data: []byte("some random result")}

	err := s.replayer.ReplayWorkflowHistory(s.logger, fullHistory)
	s.Error(err)
}

func (s *workflowReplayerSuite) TestReplayWorkflowHistory_Full_ContinueAsNew() {
	fullHistory := getTestReplayWorkflowFullHistory()
	completedEventIdx := len(fullHistory.Events) - 1
	s.NotNil(fullHistory.Events[completedEventIdx].GetWorkflowExecutionCompletedEventAttributes())
	fullHistory.Events[completedEventIdx] = createTestEventWorkflowExecutionContinuedAsNew(int64(completedEventIdx+1), nil)

	err := s.replayer.ReplayWorkflowHistory(s.logger, fullHistory)
	s.NoError(err)
}

func (s *workflowReplayerSuite) TestReplayWorkflowHistory_Partial_WithDecisionEvents() {
	err := s.replayer.ReplayWorkflowHistory(s.logger, getTestReplayWorkflowPartialHistoryWithDecisionEvents())
	s.NoError(err)
}

func (s *workflowReplayerSuite) TestReplayWorkflowHistory_Partial_NoDecisionEvents() {
	err := s.replayer.ReplayWorkflowHistory(s.logger, getTestReplayWorkflowPartialHistoryNoDecisionEvents())
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

func (s *workflowReplayerSuite) TestReplayWorkflowHistory_ContextPropagator() {
	err := s.replayer.ReplayWorkflowHistory(s.logger, getTestReplayWorkflowContextPropagatorHistory())
	s.NoError(err)
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

func testReplayWorkflowContextPropagator(ctx Context) error {
	value := ctx.Value(contextKey(testHeader))
	if val, ok := value.(string); ok && val != "" {
		testReplayWorkflow(ctx)
		return nil
	}

	return errors.New("context propagator is not setup correctly for workflow replayer")
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

func getTestReplayWorkflowFullHistory() *apiv1.History {
	return &apiv1.History{
		Events: []*apiv1.HistoryEvent{
			createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &apiv1.WorkflowType{Name: "go.uber.org/cadence/v2/internal.testReplayWorkflow"},
				TaskList:     &apiv1.TaskList{Name: testTaskList},
				Input:        &apiv1.Payload{Data: testEncodeFunctionArgs(getDefaultDataConverter())},
			}),
			createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{}),
			createTestEventDecisionTaskStarted(3),
			createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{}),
			createTestEventActivityTaskScheduled(5, &apiv1.ActivityTaskScheduledEventAttributes{
				ActivityId:   "0",
				ActivityType: &apiv1.ActivityType{Name: "testActivity"},
				TaskList:     &apiv1.TaskList{Name: testTaskList},
			}),
			createTestEventActivityTaskStarted(6, &apiv1.ActivityTaskStartedEventAttributes{
				ScheduledEventId: 5,
			}),
			createTestEventActivityTaskCompleted(7, &apiv1.ActivityTaskCompletedEventAttributes{
				ScheduledEventId: 5,
				StartedEventId:   6,
			}),
			createTestEventDecisionTaskScheduled(8, &apiv1.DecisionTaskScheduledEventAttributes{}),
			createTestEventDecisionTaskStarted(9),
			createTestEventDecisionTaskCompleted(10, &apiv1.DecisionTaskCompletedEventAttributes{
				ScheduledEventId: 8,
				StartedEventId:   9,
			}),
			createTestEventWorkflowExecutionCompleted(11, &apiv1.WorkflowExecutionCompletedEventAttributes{
				DecisionTaskCompletedEventId: 10,
			}),
		},
	}
}

func getTestReplayWorkflowPartialHistoryWithDecisionEvents() *apiv1.History {
	return &apiv1.History{
		Events: []*apiv1.HistoryEvent{
			createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &apiv1.WorkflowType{Name: "go.uber.org/cadence/v2/internal.testReplayWorkflow"},
				TaskList:     &apiv1.TaskList{Name: testTaskList},
				Input:        &apiv1.Payload{Data: testEncodeFunctionArgs(getDefaultDataConverter())},
			}),
			createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{}),
			createTestEventDecisionTaskStarted(3),
			createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{}),
			createTestEventActivityTaskScheduled(5, &apiv1.ActivityTaskScheduledEventAttributes{
				ActivityId:   "0",
				ActivityType: &apiv1.ActivityType{Name: "testActivity-fm"},
				TaskList:     &apiv1.TaskList{Name: testTaskList},
			}),
		},
	}
}

func getTestReplayWorkflowPartialHistoryNoDecisionEvents() *apiv1.History {
	return &apiv1.History{
		Events: []*apiv1.HistoryEvent{
			createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &apiv1.WorkflowType{Name: "go.uber.org/cadence/v2/internal.testReplayWorkflow"},
				TaskList:     &apiv1.TaskList{Name: testTaskList},
				Input:        &apiv1.Payload{Data: testEncodeFunctionArgs(getDefaultDataConverter())},
			}),
			createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{}),
			createTestEventDecisionTaskStarted(3),
			createTestEventDecisionTaskFailed(4, &apiv1.DecisionTaskFailedEventAttributes{ScheduledEventId: 2}),
			createTestEventDecisionTaskScheduled(5, &apiv1.DecisionTaskScheduledEventAttributes{}),
			createTestEventDecisionTaskStarted(6),
		},
	}
}

func getTestReplayWorkflowMismatchHistory() *apiv1.History {
	return &apiv1.History{
		Events: []*apiv1.HistoryEvent{
			createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &apiv1.WorkflowType{Name: "go.uber.org/cadence/v2/internal.testReplayWorkflow"},
				TaskList:     &apiv1.TaskList{Name: "taskList"},
				Input:        &apiv1.Payload{Data: testEncodeFunctionArgs(getDefaultDataConverter())},
			}),
			createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{}),
			createTestEventDecisionTaskStarted(3),
			createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{}),
			createTestEventActivityTaskScheduled(5, &apiv1.ActivityTaskScheduledEventAttributes{
				ActivityId:   "0",
				ActivityType: &apiv1.ActivityType{Name: "unknownActivityType"},
				TaskList:     &apiv1.TaskList{Name: "taskList"},
			}),
		},
	}
}

func getTestReplayWorkflowLocalActivityHistory() *apiv1.History {
	return &apiv1.History{
		Events: []*apiv1.HistoryEvent{
			createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &apiv1.WorkflowType{Name: "go.uber.org/cadence/v2/internal.testReplayWorkflowLocalActivity"},
				TaskList:     &apiv1.TaskList{Name: testTaskList},
				Input:        &apiv1.Payload{Data: testEncodeFunctionArgs(getDefaultDataConverter())},
			}),
			createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{}),
			createTestEventDecisionTaskStarted(3),
			createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{}),

			createTestEventLocalActivity(5, &apiv1.MarkerRecordedEventAttributes{
				MarkerName:                   localActivityMarkerName,
				Details:                      &apiv1.Payload{Data: createLocalActivityMarkerDataForTest("0", "go.uber.org/cadence/v2/internal.testActivity")},
				DecisionTaskCompletedEventId: 4,
			}),

			createTestEventWorkflowExecutionCompleted(6, &apiv1.WorkflowExecutionCompletedEventAttributes{
				DecisionTaskCompletedEventId: 4,
			}),
		},
	}
}

func getTestReplayWorkflowLocalActivityResultMismatchHistory() *apiv1.History {
	return &apiv1.History{
		Events: []*apiv1.HistoryEvent{
			createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &apiv1.WorkflowType{Name: "go.uber.org/cadence/v2/internal.testReplayWorkflowLocalActivity"},
				TaskList:     &apiv1.TaskList{Name: testTaskList},
				Input:        &apiv1.Payload{Data: testEncodeFunctionArgs(getDefaultDataConverter())},
			}),
			createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{}),
			createTestEventDecisionTaskStarted(3),
			createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{}),

			createTestEventLocalActivity(5, &apiv1.MarkerRecordedEventAttributes{
				MarkerName:                   localActivityMarkerName,
				Details:                      &apiv1.Payload{Data: createLocalActivityMarkerDataForTest("0", "")},
				DecisionTaskCompletedEventId: 4,
			}),

			createTestEventWorkflowExecutionCompleted(6, &apiv1.WorkflowExecutionCompletedEventAttributes{
				Result:                       &apiv1.Payload{Data: []byte("some-incorrect-result")},
				DecisionTaskCompletedEventId: 4,
			}),
		},
	}
}

func getTestReplayWorkflowLocalActivityTypeMismatchHistory() *apiv1.History {
	return &apiv1.History{
		Events: []*apiv1.HistoryEvent{
			createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{
				WorkflowType: &apiv1.WorkflowType{Name: "go.uber.org/cadence/v2/internal.testReplayWorkflowLocalActivity"},
				TaskList:     &apiv1.TaskList{Name: testTaskList},
				Input:        &apiv1.Payload{Data: testEncodeFunctionArgs(getDefaultDataConverter())},
			}),
			createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{}),
			createTestEventDecisionTaskStarted(3),
			createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{}),

			createTestEventLocalActivity(5, &apiv1.MarkerRecordedEventAttributes{
				MarkerName:                   localActivityMarkerName,
				Details:                      &apiv1.Payload{Data: createLocalActivityMarkerDataForTest("0", "different-activity-type")},
				DecisionTaskCompletedEventId: 4,
			}),

			createTestEventWorkflowExecutionCompleted(6, &apiv1.WorkflowExecutionCompletedEventAttributes{
				DecisionTaskCompletedEventId: 4,
			}),
		},
	}
}

func getTestReplayWorkflowContextPropagatorHistory() *apiv1.History {
	history := getTestReplayWorkflowFullHistory()
	history.Events[0].GetWorkflowExecutionStartedEventAttributes().WorkflowType.Name = "go.uber.org/cadence/v2/internal.testReplayWorkflowContextPropagator"
	history.Events[0].GetWorkflowExecutionStartedEventAttributes().Header = &apiv1.Header{
		Fields: map[string]*apiv1.Payload{testHeader: {Data:[]byte("testValue")}},
	}
	return history
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
