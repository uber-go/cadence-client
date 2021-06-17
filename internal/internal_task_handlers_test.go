// Copyright (c) 2017-2020 Uber Technologies Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/opentracing/opentracing-go"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
	"go.uber.org/cadence/v2/internal/api"
	"go.uber.org/zap"
)

const (
	testDomain = "test-domain"
)

type (
	TaskHandlersTestSuite struct {
		suite.Suite
		logger   *zap.Logger
		service  *api.MockInterface
		registry *registry
	}
)

func registerWorkflows(r *registry) {
	r.RegisterWorkflowWithOptions(
		helloWorldWorkflowFunc,
		RegisterWorkflowOptions{Name: "HelloWorld_Workflow"},
	)
	r.RegisterWorkflowWithOptions(
		helloWorldWorkflowCancelFunc,
		RegisterWorkflowOptions{Name: "HelloWorld_WorkflowCancel"},
	)
	r.RegisterWorkflowWithOptions(
		returnPanicWorkflowFunc,
		RegisterWorkflowOptions{Name: "ReturnPanicWorkflow"},
	)
	r.RegisterWorkflowWithOptions(
		panicWorkflowFunc,
		RegisterWorkflowOptions{Name: "PanicWorkflow"},
	)
	r.RegisterWorkflowWithOptions(
		getWorkflowInfoWorkflowFunc,
		RegisterWorkflowOptions{Name: "GetWorkflowInfoWorkflow"},
	)
	r.RegisterWorkflowWithOptions(
		querySignalWorkflowFunc,
		RegisterWorkflowOptions{Name: "QuerySignalWorkflow"},
	)
	r.RegisterActivityWithOptions(
		greeterActivityFunc,
		RegisterActivityOptions{Name: "Greeter_Activity"},
	)
	r.RegisterWorkflowWithOptions(
		binaryChecksumWorkflowFunc,
		RegisterWorkflowOptions{Name: "BinaryChecksumWorkflow"},
	)
}

func returnPanicWorkflowFunc(ctx Context, input []byte) error {
	return newPanicError("panicError", "stackTrace")
}

func panicWorkflowFunc(ctx Context, input []byte) error {
	panic("panicError")
}

func getWorkflowInfoWorkflowFunc(ctx Context, expectedLastCompletionResult string) (info *WorkflowInfo, err error) {
	result := GetWorkflowInfo(ctx)
	var lastCompletionResult string
	err = getDefaultDataConverter().FromData(result.lastCompletionResult, &lastCompletionResult)
	if err != nil {
		return nil, err
	}
	if lastCompletionResult != expectedLastCompletionResult {
		return nil, errors.New("lastCompletionResult is not " + expectedLastCompletionResult)
	}
	return result, nil
}

// Test suite.
func (t *TaskHandlersTestSuite) SetupTest() {
}

func (t *TaskHandlersTestSuite) SetupSuite() {
	logger, _ := zap.NewDevelopment()
	t.logger = logger
	registerWorkflows(t.registry)
}

func TestTaskHandlersTestSuite(t *testing.T) {
	suite.Run(t, &TaskHandlersTestSuite{
		registry: newRegistry(),
	})
}

func createTestEventWorkflowExecutionCompleted(eventID int64, attr *apiv1.WorkflowExecutionCompletedEventAttributes) *apiv1.HistoryEvent {
	return &apiv1.HistoryEvent{
		EventId: eventID,
		Attributes: &apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes{WorkflowExecutionCompletedEventAttributes: attr},
	}
}

func createTestEventWorkflowExecutionContinuedAsNew(eventID int64, attr *apiv1.WorkflowExecutionContinuedAsNewEventAttributes) *apiv1.HistoryEvent {
	return &apiv1.HistoryEvent{
		EventId: eventID,
		Attributes: &apiv1.HistoryEvent_WorkflowExecutionContinuedAsNewEventAttributes{WorkflowExecutionContinuedAsNewEventAttributes: attr},
	}
}

func createTestEventWorkflowExecutionStarted(eventID int64, attr *apiv1.WorkflowExecutionStartedEventAttributes) *apiv1.HistoryEvent {
	return &apiv1.HistoryEvent{
		EventId: eventID,
		Attributes: &apiv1.HistoryEvent_WorkflowExecutionStartedEventAttributes{WorkflowExecutionStartedEventAttributes: attr},
	}
}

func createTestEventLocalActivity(eventID int64, attr *apiv1.MarkerRecordedEventAttributes) *apiv1.HistoryEvent {
	return &apiv1.HistoryEvent{
		EventId:    eventID,
		Attributes: &apiv1.HistoryEvent_MarkerRecordedEventAttributes{MarkerRecordedEventAttributes: attr},
	}
}

func createTestEventActivityTaskScheduled(eventID int64, attr *apiv1.ActivityTaskScheduledEventAttributes) *apiv1.HistoryEvent {
	return &apiv1.HistoryEvent{
		EventId:    eventID,
		Attributes: &apiv1.HistoryEvent_ActivityTaskScheduledEventAttributes{ActivityTaskScheduledEventAttributes: attr},
	}
}

func createTestEventActivityTaskStarted(eventID int64, attr *apiv1.ActivityTaskStartedEventAttributes) *apiv1.HistoryEvent {
	return &apiv1.HistoryEvent{
		EventId:    eventID,
		Attributes: &apiv1.HistoryEvent_ActivityTaskStartedEventAttributes{ActivityTaskStartedEventAttributes: attr},
	}
}

func createTestEventActivityTaskCompleted(eventID int64, attr *apiv1.ActivityTaskCompletedEventAttributes) *apiv1.HistoryEvent {
	return &apiv1.HistoryEvent{
		EventId:                              eventID,
		Attributes: &apiv1.HistoryEvent_ActivityTaskCompletedEventAttributes{ActivityTaskCompletedEventAttributes: attr},
	}
}

func createTestEventActivityTaskTimedOut(eventID int64, attr *apiv1.ActivityTaskTimedOutEventAttributes) *apiv1.HistoryEvent {
	return &apiv1.HistoryEvent{
		EventId:                             eventID,
		Attributes: &apiv1.HistoryEvent_ActivityTaskTimedOutEventAttributes{ActivityTaskTimedOutEventAttributes: attr},
	}
}

func createTestEventDecisionTaskScheduled(eventID int64, attr *apiv1.DecisionTaskScheduledEventAttributes) *apiv1.HistoryEvent {
	return &apiv1.HistoryEvent{
		EventId:                              eventID,
		Attributes: &apiv1.HistoryEvent_DecisionTaskScheduledEventAttributes{DecisionTaskScheduledEventAttributes: attr},
	}
}

func createTestEventDecisionTaskStarted(eventID int64) *apiv1.HistoryEvent {
	return &apiv1.HistoryEvent{
		EventId:   eventID,
		Attributes: &apiv1.HistoryEvent_DecisionTaskStartedEventAttributes{
			DecisionTaskStartedEventAttributes: &apiv1.DecisionTaskStartedEventAttributes{},
		},
	}
}

func createTestEventWorkflowExecutionSignaled(eventID int64, signalName string) *apiv1.HistoryEvent {
	return createTestEventWorkflowExecutionSignaledWithPayload(eventID, signalName, nil)
}

func createTestEventWorkflowExecutionSignaledWithPayload(eventID int64, signalName string, payload []byte) *apiv1.HistoryEvent {
	return &apiv1.HistoryEvent{
		EventId:   eventID,
		Attributes: &apiv1.HistoryEvent_WorkflowExecutionSignaledEventAttributes{
			WorkflowExecutionSignaledEventAttributes: &apiv1.WorkflowExecutionSignaledEventAttributes{
				SignalName: signalName,
				Input:      &apiv1.Payload{Data: payload},
				Identity:   "test-identity",
			},
		},
	}
}

func createTestEventDecisionTaskCompleted(eventID int64, attr *apiv1.DecisionTaskCompletedEventAttributes) *apiv1.HistoryEvent {
	return &apiv1.HistoryEvent{
		EventId:    eventID,
		Attributes: &apiv1.HistoryEvent_DecisionTaskCompletedEventAttributes{DecisionTaskCompletedEventAttributes: attr},
	}
}

func createTestEventDecisionTaskFailed(eventID int64, attr *apiv1.DecisionTaskFailedEventAttributes) *apiv1.HistoryEvent {
	return &apiv1.HistoryEvent{
		EventId:    eventID,
		Attributes: &apiv1.HistoryEvent_DecisionTaskFailedEventAttributes{DecisionTaskFailedEventAttributes: attr},
	}
}

func createTestEventSignalExternalWorkflowExecutionFailed(eventID int64, attr *apiv1.SignalExternalWorkflowExecutionFailedEventAttributes) *apiv1.HistoryEvent {
	return &apiv1.HistoryEvent{
		EventId:   eventID,
		Attributes: &apiv1.HistoryEvent_SignalExternalWorkflowExecutionFailedEventAttributes{SignalExternalWorkflowExecutionFailedEventAttributes: attr},
	}
}

func createWorkflowTask(
	events []*apiv1.HistoryEvent,
	previousStartEventID int64,
	workflowName string,
) *apiv1.PollForDecisionTaskResponse {
	return createWorkflowTaskWithQueries(events, previousStartEventID, workflowName, nil)
}

func createWorkflowTaskWithQueries(
	events []*apiv1.HistoryEvent,
	previousStartEventID int64,
	workflowName string,
	queries map[string]*apiv1.WorkflowQuery,
) *apiv1.PollForDecisionTaskResponse {
	eventsCopy := make([]*apiv1.HistoryEvent, len(events))
	copy(eventsCopy, events)
	return &apiv1.PollForDecisionTaskResponse{
		PreviousStartedEventId: &types.Int64Value{Value: previousStartEventID},
		WorkflowType:           &apiv1.WorkflowType{Name: workflowName},
		History:                &apiv1.History{Events: eventsCopy},
		WorkflowExecution: &apiv1.WorkflowExecution{
			WorkflowId: "fake-workflow-id",
			RunId:      uuid.New(),
		},
		Queries: queries,
	}
}

func createQueryTask(
	events []*apiv1.HistoryEvent,
	previousStartEventID int64,
	workflowName string,
	queryType string,
) *apiv1.PollForDecisionTaskResponse {
	task := createWorkflowTask(events, previousStartEventID, workflowName)
	task.Query = &apiv1.WorkflowQuery{
		QueryType: queryType,
	}
	return task
}

func createTestEventTimerStarted(eventID int64, id int) *apiv1.HistoryEvent {
	timerID := fmt.Sprintf("%v", id)
	attr := &apiv1.TimerStartedEventAttributes{
		TimerId:                      timerID,
		StartToFireTimeout:           nil,
		DecisionTaskCompletedEventId: 0,
	}
	return &apiv1.HistoryEvent{
		EventId: eventID,
		Attributes: &apiv1.HistoryEvent_TimerStartedEventAttributes{TimerStartedEventAttributes: attr},
	}
}

func createTestEventTimerFired(eventID int64, id int) *apiv1.HistoryEvent {
	timerID := fmt.Sprintf("%v", id)
	attr := &apiv1.TimerFiredEventAttributes{
		TimerId: timerID,
	}

	return &apiv1.HistoryEvent{
		EventId:    eventID,
		Attributes: &apiv1.HistoryEvent_TimerFiredEventAttributes{TimerFiredEventAttributes: attr},
	}
}

var testWorkflowTaskTasklist = "tl1"

func (t *TaskHandlersTestSuite) testWorkflowTaskWorkflowExecutionStartedHelper(params workerExecutionParameters) {
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{TaskList: &apiv1.TaskList{Name: testWorkflowTaskTasklist}}),
	}
	task := createWorkflowTask(testEvents, 0, "HelloWorld_Workflow")
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	response := request.(*apiv1.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
	t.Equal(1, len(response.Decisions))
	t.NotNil(response.Decisions[0].GetScheduleActivityTaskDecisionAttributes())
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_WorkflowExecutionStarted() {
	params := workerExecutionParameters{
		TaskList: testWorkflowTaskTasklist,
		Identity: "test-id-1",
		Logger:   t.logger,
	}
	t.testWorkflowTaskWorkflowExecutionStartedHelper(params)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_WorkflowExecutionStarted_WithDataConverter() {
	params := workerExecutionParameters{
		TaskList:      testWorkflowTaskTasklist,
		Identity:      "test-id-1",
		Logger:        t.logger,
		DataConverter: newTestDataConverter(),
	}
	t.testWorkflowTaskWorkflowExecutionStartedHelper(params)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_BinaryChecksum() {
	taskList := "tl1"
	checksum1 := "chck1"
	checksum2 := "chck2"
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{
			ScheduledEventId: 2, BinaryChecksum: checksum1}),
		createTestEventTimerStarted(5, 0),
		createTestEventTimerFired(6, 0),
		createTestEventDecisionTaskScheduled(7, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(8),
		createTestEventDecisionTaskCompleted(9, &apiv1.DecisionTaskCompletedEventAttributes{
			ScheduledEventId: 7, BinaryChecksum: checksum2}),
		createTestEventTimerStarted(10, 1),
		createTestEventTimerFired(11, 1),
		createTestEventDecisionTaskScheduled(12, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(13),
	}
	task := createWorkflowTask(testEvents, 8, "BinaryChecksumWorkflow")
	params := workerExecutionParameters{
		TaskList: taskList,
		Identity: "test-id-1",
		Logger:   t.logger,
	}
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	response := request.(*apiv1.RespondDecisionTaskCompletedRequest)

	t.NoError(err)
	t.NotNil(response)
	t.Equal(1, len(response.Decisions))
	t.NotNil(response.Decisions[0].GetCompleteWorkflowExecutionDecisionAttributes())
	checksumsBytes := response.Decisions[0].GetCompleteWorkflowExecutionDecisionAttributes().Result.GetData()
	var checksums []string
	json.Unmarshal(checksumsBytes, &checksums)
	t.Equal(3, len(checksums))
	t.Equal("chck1", checksums[0])
	t.Equal("chck2", checksums[1])
	t.Equal(getBinaryChecksum(), checksums[2])
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_ActivityTaskScheduled() {
	// Schedule an activity and see if we complete workflow.
	taskList := "tl1"
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{ScheduledEventId: 2}),
		createTestEventActivityTaskScheduled(5, &apiv1.ActivityTaskScheduledEventAttributes{
			ActivityId:   "0",
			ActivityType: &apiv1.ActivityType{Name: "Greeter_Activity"},
			TaskList:     &apiv1.TaskList{Name: taskList},
		}),
		createTestEventActivityTaskStarted(6, &apiv1.ActivityTaskStartedEventAttributes{}),
		createTestEventActivityTaskCompleted(7, &apiv1.ActivityTaskCompletedEventAttributes{ScheduledEventId: 5}),
		createTestEventDecisionTaskStarted(8),
	}
	task := createWorkflowTask(testEvents[0:3], 0, "HelloWorld_Workflow")
	params := workerExecutionParameters{
		TaskList: taskList,
		Identity: "test-id-1",
		Logger:   t.logger,
	}
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	response := request.(*apiv1.RespondDecisionTaskCompletedRequest)

	t.NoError(err)
	t.NotNil(response)
	t.Equal(1, len(response.Decisions))
	t.NotNil(response.Decisions[0].GetScheduleActivityTaskDecisionAttributes())

	// Schedule an activity and see if we complete workflow, Having only one last decision.
	task = createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	request, err = taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	response = request.(*apiv1.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
	t.Equal(1, len(response.Decisions))
	t.NotNil(response.Decisions[0].GetCompleteWorkflowExecutionDecisionAttributes())
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_QueryWorkflow_Sticky() {
	// Schedule an activity and see if we complete workflow.
	taskList := "sticky-tl"
	execution := &apiv1.WorkflowExecution{
		WorkflowId: "fake-workflow-id",
		RunId:      uuid.New(),
	}
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{ScheduledEventId: 2}),
		createTestEventActivityTaskScheduled(5, &apiv1.ActivityTaskScheduledEventAttributes{
			ActivityId:   "0",
			ActivityType: &apiv1.ActivityType{Name: "Greeter_Activity"},
			TaskList:     &apiv1.TaskList{Name: taskList},
		}),
		createTestEventActivityTaskStarted(6, &apiv1.ActivityTaskStartedEventAttributes{}),
		createTestEventActivityTaskCompleted(7, &apiv1.ActivityTaskCompletedEventAttributes{ScheduledEventId: 5}),
	}
	params := workerExecutionParameters{
		TaskList: taskList,
		Identity: "test-id-1",
		Logger:   t.logger,
	}
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)

	// first make progress on the workflow
	task := createWorkflowTask(testEvents[0:1], 0, "HelloWorld_Workflow")
	task.StartedEventId = 1
	task.WorkflowExecution = execution
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	response := request.(*apiv1.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
	t.Equal(1, len(response.Decisions))
	t.NotNil(response.Decisions[0].GetScheduleActivityTaskDecisionAttributes())

	// then check the current state using query task
	task = createQueryTask([]*apiv1.HistoryEvent{}, 6, "HelloWorld_Workflow", queryType)
	task.WorkflowExecution = execution
	queryResp, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.NoError(err)
	t.verifyQueryResult(queryResp, "waiting-activity-result")
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_QueryWorkflow_NonSticky() {
	// Schedule an activity and see if we complete workflow.
	taskList := "tl1"
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{ScheduledEventId: 2}),
		createTestEventActivityTaskScheduled(5, &apiv1.ActivityTaskScheduledEventAttributes{
			ActivityId:   "0",
			ActivityType: &apiv1.ActivityType{Name: "Greeter_Activity"},
			TaskList:     &apiv1.TaskList{Name: taskList},
		}),
		createTestEventActivityTaskStarted(6, &apiv1.ActivityTaskStartedEventAttributes{}),
		createTestEventActivityTaskCompleted(7, &apiv1.ActivityTaskCompletedEventAttributes{ScheduledEventId: 5}),
		createTestEventDecisionTaskStarted(8),
		createTestEventWorkflowExecutionSignaled(9, "test-signal"),
	}
	params := workerExecutionParameters{
		TaskList: taskList,
		Identity: "test-id-1",
		Logger:   t.logger,
	}

	// TODO: the following comment is not true, previousStartEventID is either 0 or points to a valid decisionTaskStartedID
	// we need to fix the test
	//
	// query after first decision task (notice the previousStartEventID is always the last eventID for query task)
	task := createQueryTask(testEvents[0:3], 3, "HelloWorld_Workflow", queryType)
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	response, _ := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.verifyQueryResult(response, "waiting-activity-result")

	// query after activity task complete but before second decision task started
	task = createQueryTask(testEvents[0:7], 7, "HelloWorld_Workflow", queryType)
	taskHandler = newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	response, _ = taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.verifyQueryResult(response, "waiting-activity-result")

	// query after second decision task
	task = createQueryTask(testEvents[0:8], 8, "HelloWorld_Workflow", queryType)
	taskHandler = newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	response, _ = taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.verifyQueryResult(response, "done")

	// query after second decision task with extra events
	task = createQueryTask(testEvents[0:9], 9, "HelloWorld_Workflow", queryType)
	taskHandler = newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	response, _ = taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.verifyQueryResult(response, "done")

	task = createQueryTask(testEvents[0:9], 9, "HelloWorld_Workflow", "invalid-query-type")
	taskHandler = newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	response, _ = taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.NotNil(response)
	queryResp, ok := response.(*apiv1.RespondQueryTaskCompletedRequest)
	t.True(ok)
	t.NotNil(queryResp.Result.ErrorMessage)
	t.Contains(queryResp.Result.ErrorMessage, "unknown queryType")
}

func (t *TaskHandlersTestSuite) verifyQueryResult(response interface{}, expectedResult string) {
	t.NotNil(response)
	queryResp, ok := response.(*apiv1.RespondQueryTaskCompletedRequest)
	t.True(ok)
	t.Empty(queryResp.Result.ErrorMessage)
	t.NotNil(queryResp.Result.Answer)
	encodedValue := newEncodedValue(queryResp.Result.Answer.GetData(), nil)
	var queryResult string
	err := encodedValue.Get(&queryResult)
	t.NoError(err)
	t.Equal(expectedResult, queryResult)
}

func (t *TaskHandlersTestSuite) TestCacheEvictionWhenErrorOccurs() {
	taskList := "taskList"
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{ScheduledEventId: 2}),
		createTestEventActivityTaskScheduled(5, &apiv1.ActivityTaskScheduledEventAttributes{
			ActivityId:   "0",
			ActivityType: &apiv1.ActivityType{Name: "pkg.Greeter_Activity"},
			TaskList:     &apiv1.TaskList{Name: taskList},
		}),
	}
	params := workerExecutionParameters{
		TaskList:                       taskList,
		Identity:                       "test-id-1",
		Logger:                         zap.NewNop(),
		NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	// now change the history event so it does not match to decision produced via replay
	testEvents[4].GetActivityTaskScheduledEventAttributes().ActivityType.Name = "some-other-activity"
	task := createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	// newWorkflowTaskWorkerInternal will set the laTunnel in taskHandler, without it, ProcessWorkflowTask()
	// will fail as it can't find laTunnel in getWorkflowCache().
	newWorkflowTaskWorkerInternal(taskHandler, t.service, testDomain, params, make(chan struct{}), nil)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)

	t.Error(err)
	t.Nil(request)
	t.Contains(err.Error(), "nondeterministic")

	// There should be nothing in the cache.
	t.EqualValues(getWorkflowCache().Size(), 0)
}

func (t *TaskHandlersTestSuite) TestWithMissingHistoryEvents() {
	taskList := "taskList"
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{ScheduledEventId: 2}),
		createTestEventDecisionTaskScheduled(6, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(7),
	}
	params := workerExecutionParameters{
		TaskList:                       taskList,
		Identity:                       "test-id-1",
		Logger:                         zap.NewNop(),
		NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
	}

	for _, startEventID := range []int64{0, 3} {
		taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
		task := createWorkflowTask(testEvents, startEventID, "HelloWorld_Workflow")
		// newWorkflowTaskWorkerInternal will set the laTunnel in taskHandler, without it, ProcessWorkflowTask()
		// will fail as it can't find laTunnel in getWorkflowCache().
		newWorkflowTaskWorkerInternal(taskHandler, t.service, testDomain, params, make(chan struct{}), nil)
		request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)

		t.Error(err)
		t.Nil(request)
		t.Contains(err.Error(), "missing history events")

		// There should be nothing in the cache.
		t.EqualValues(getWorkflowCache().Size(), 0)
	}
}

func (t *TaskHandlersTestSuite) TestWithTruncatedHistory() {
	taskList := "taskList"
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskFailed(4, &apiv1.DecisionTaskFailedEventAttributes{ScheduledEventId: 2}),
		createTestEventDecisionTaskScheduled(5, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(6),
		createTestEventDecisionTaskCompleted(7, &apiv1.DecisionTaskCompletedEventAttributes{ScheduledEventId: 5}),
		createTestEventActivityTaskScheduled(8, &apiv1.ActivityTaskScheduledEventAttributes{
			ActivityId:   "0",
			ActivityType: &apiv1.ActivityType{Name: "pkg.Greeter_Activity"},
			TaskList:     &apiv1.TaskList{Name: taskList},
		}),
	}
	params := workerExecutionParameters{
		TaskList:                       taskList,
		Identity:                       "test-id-1",
		Logger:                         zap.NewNop(),
		NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
	}

	testCases := []struct {
		startedEventID         int64
		previousStartedEventID int64
		isResultErr            bool
	}{
		{0, 6, false},
		{10, 0, true},
		{10, 6, true},
	}

	for i, tc := range testCases {
		taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
		task := createWorkflowTask(testEvents, tc.previousStartedEventID, "HelloWorld_Workflow")
		task.StartedEventId = tc.startedEventID
		// newWorkflowTaskWorkerInternal will set the laTunnel in taskHandler, without it, ProcessWorkflowTask()
		// will fail as it can't find laTunnel in getWorkflowCache().
		newWorkflowTaskWorkerInternal(taskHandler, t.service, testDomain, params, make(chan struct{}), nil)
		request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)

		if tc.isResultErr {
			t.Error(err, "testcase %v failed", i)
			t.Nil(request)
			t.Contains(err.Error(), "premature end of stream")
			t.EqualValues(getWorkflowCache().Size(), 1)
			continue
		}

		t.NoError(err, "testcase %v failed", i)
	}
}

func (t *TaskHandlersTestSuite) TestSideEffectDefer_Sticky() {
	t.testSideEffectDeferHelper(false)
}

func (t *TaskHandlersTestSuite) TestSideEffectDefer_NonSticky() {
	t.testSideEffectDeferHelper(true)
}

func (t *TaskHandlersTestSuite) testSideEffectDeferHelper(disableSticky bool) {
	value := "should not be modified"
	expectedValue := value
	doneCh := make(chan struct{})

	workflowFunc := func(ctx Context) error {
		defer func() {
			if !IsReplaying(ctx) {
				// This is an side effect op
				value = ""
			}
			close(doneCh)
		}()
		Sleep(ctx, 1*time.Second)
		return nil
	}
	workflowName := fmt.Sprintf("SideEffectDeferWorkflow-Sticky=%v", disableSticky)
	t.registry.RegisterWorkflowWithOptions(
		workflowFunc,
		RegisterWorkflowOptions{Name: workflowName},
	)

	taskList := "taskList"
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
	}

	params := workerExecutionParameters{
		TaskList:               taskList,
		Identity:               "test-id-1",
		Logger:                 zap.NewNop(),
		DisableStickyExecution: disableSticky,
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	task := createWorkflowTask(testEvents, 0, workflowName)
	_, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.Nil(err)

	if !params.DisableStickyExecution {
		// 1. We can't set cache size in the test to 1, otherwise other tests will break.
		// 2. We need to make sure cache is empty when the test is completed,
		// So manually trigger a delete.
		getWorkflowCache().Delete(task.WorkflowExecution.GetRunId())
	}
	// Make sure the workflow coroutine has exited.
	<-doneCh
	// The side effect op should not be executed.
	t.Equal(expectedValue, value)

	// There should be nothing in the cache.
	t.EqualValues(0, getWorkflowCache().Size())
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_NondeterministicDetection() {
	taskList := "taskList"
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{ScheduledEventId: 2}),
		createTestEventActivityTaskScheduled(5, &apiv1.ActivityTaskScheduledEventAttributes{
			ActivityId:   "0",
			ActivityType: &apiv1.ActivityType{Name: "pkg.Greeter_Activity"},
			TaskList:     &apiv1.TaskList{Name: taskList},
		}),
	}
	task := createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	stopC := make(chan struct{})
	params := workerExecutionParameters{
		TaskList:                       taskList,
		Identity:                       "test-id-1",
		Logger:                         zap.NewNop(),
		NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
		WorkerStopChannel:              stopC,
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	response := request.(*apiv1.RespondDecisionTaskCompletedRequest)
	// there should be no error as the history events matched the decisions.
	t.NoError(err)
	t.NotNil(response)

	// now change the history event so it does not match to decision produced via replay
	testEvents[4].GetActivityTaskScheduledEventAttributes().ActivityType.Name = "some-other-activity"
	task = createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	// newWorkflowTaskWorkerInternal will set the laTunnel in taskHandler, without it, ProcessWorkflowTask()
	// will fail as it can't find laTunnel in getWorkflowCache().
	newWorkflowTaskWorkerInternal(taskHandler, t.service, testDomain, params, stopC, nil)
	request, err = taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.Error(err)
	t.Nil(request)
	t.Contains(err.Error(), "nondeterministic")

	// now, create a new task handler with fail nondeterministic workflow policy
	// and verify that it handles the mismatching history correctly.
	params.NonDeterministicWorkflowPolicy = NonDeterministicWorkflowPolicyFailWorkflow
	failOnNondeterminismTaskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	task = createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	request, err = failOnNondeterminismTaskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	// When FailWorkflow policy is set, task handler does not return an error,
	// because it will indicate non determinism in the request.
	t.NoError(err)
	// Verify that request is a RespondDecisionTaskCompleteRequest
	response, ok := request.(*apiv1.RespondDecisionTaskCompletedRequest)
	t.True(ok)
	// Verify there's at least 1 decision
	// and the last last decision is to fail workflow
	// and contains proper justification.(i.e. nondeterminism).
	t.True(len(response.Decisions) > 0)
	closeDecision := response.Decisions[len(response.Decisions)-1]
	t.NotNil(closeDecision.GetFailWorkflowExecutionDecisionAttributes())
	t.Contains(closeDecision.GetFailWorkflowExecutionDecisionAttributes().Failure.GetReason(), "NonDeterministicWorkflowPolicyFailWorkflow")

	// now with different package name to activity type
	testEvents[4].GetActivityTaskScheduledEventAttributes().ActivityType.Name = "new-package.Greeter_Activity"
	task = createWorkflowTask(testEvents, 3, "HelloWorld_Workflow")
	request, err = taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.NoError(err)
	t.NotNil(request)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_WorkflowReturnsPanicError() {
	taskList := "taskList"
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
	}
	task := createWorkflowTask(testEvents, 3, "ReturnPanicWorkflow")
	params := workerExecutionParameters{
		TaskList:                       taskList,
		Identity:                       "test-id-1",
		Logger:                         zap.NewNop(),
		NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.NoError(err)
	t.NotNil(request)
	r, ok := request.(*apiv1.RespondDecisionTaskCompletedRequest)
	t.True(ok)
	t.NotNil(r.Decisions[0].GetFailWorkflowExecutionDecisionAttributes())
	attr := r.Decisions[0].GetFailWorkflowExecutionDecisionAttributes()
	t.EqualValues("cadenceInternal:Panic", attr.Failure.GetReason())
	details := string(attr.Failure.GetDetails())
	t.True(strings.HasPrefix(details, "\"panicError"), details)
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_WorkflowPanics() {
	taskList := "taskList"
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
	}
	task := createWorkflowTask(testEvents, 3, "PanicWorkflow")
	params := workerExecutionParameters{
		TaskList:                       taskList,
		Identity:                       "test-id-1",
		Logger:                         zap.NewNop(),
		NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.NoError(err)
	t.NotNil(request)
	r, ok := request.(*apiv1.RespondDecisionTaskFailedRequest)
	t.True(ok)
	t.EqualValues(apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_WORKFLOW_WORKER_UNHANDLED_FAILURE, r.Cause)
	t.EqualValues("panicError", string(r.Details.GetData()))
}

func (t *TaskHandlersTestSuite) TestGetWorkflowInfo() {
	taskList := "taskList"
	parentID := "parentID"
	parentRunID := "parentRun"
	cronSchedule := "5 4 * * *"
	continuedRunID := uuid.New()
	parentExecution := &apiv1.WorkflowExecution{
		WorkflowId: parentID,
		RunId:      parentRunID,
	}
	parentDomain := "parentDomain"
	var attempt int32 = 123
	var executionTimeout int32 = 213456
	var taskTimeout int32 = 21
	workflowType := "GetWorkflowInfoWorkflow"
	lastCompletionResult, err := getDefaultDataConverter().ToData("lastCompletionData")
	t.NoError(err)
	retryPolicy := &apiv1.RetryPolicy{
		InitialInterval:    api.DurationToProto(time.Second * 1),
		BackoffCoefficient: 1.0,
		MaximumInterval:    api.DurationToProto(time.Second * 1),
		MaximumAttempts:    3,
	}
	startedEventAttributes := &apiv1.WorkflowExecutionStartedEventAttributes{
		Input:    &apiv1.Payload{Data: lastCompletionResult},
		TaskList: &apiv1.TaskList{Name: taskList},
		ParentExecutionInfo: &apiv1.ParentExecutionInfo{
			DomainName:        parentDomain,
			WorkflowExecution: parentExecution,
		},
		CronSchedule:                 cronSchedule,
		ContinuedExecutionRunId:      continuedRunID,
		Attempt:                      attempt,
		ExecutionStartToCloseTimeout: api.SecondsToProto(executionTimeout),
		TaskStartToCloseTimeout:      api.SecondsToProto(taskTimeout),
		LastCompletionResult:         &apiv1.Payload{Data: lastCompletionResult},
		RetryPolicy:                  retryPolicy,
	}
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, startedEventAttributes),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
	}
	task := createWorkflowTask(testEvents, 3, workflowType)
	params := workerExecutionParameters{
		TaskList:                       taskList,
		Identity:                       "test-id-1",
		Logger:                         zap.NewNop(),
		NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	t.NoError(err)
	t.NotNil(request)
	r, ok := request.(*apiv1.RespondDecisionTaskCompletedRequest)
	t.True(ok)
	t.NotNil(r.Decisions[0].GetCompleteWorkflowExecutionDecisionAttributes())
	attr := r.Decisions[0].GetCompleteWorkflowExecutionDecisionAttributes()
	var result WorkflowInfo
	t.NoError(getDefaultDataConverter().FromData(attr.Result.GetData(), &result))
	t.EqualValues(taskList, result.TaskListName)
	t.EqualValues(parentID, result.ParentWorkflowExecution.ID)
	t.EqualValues(parentRunID, result.ParentWorkflowExecution.RunID)
	t.EqualValues(cronSchedule, *result.CronSchedule)
	t.EqualValues(continuedRunID, *result.ContinuedExecutionRunID)
	t.EqualValues(parentDomain, *result.ParentWorkflowDomain)
	t.EqualValues(attempt, result.Attempt)
	t.EqualValues(executionTimeout, result.ExecutionStartToCloseTimeoutSeconds)
	t.EqualValues(taskTimeout, result.TaskStartToCloseTimeoutSeconds)
	t.EqualValues(workflowType, result.WorkflowType.Name)
	t.EqualValues(testDomain, result.Domain)
	t.EqualValues(retryPolicy, result.RetryPolicy)
}

func (t *TaskHandlersTestSuite) TestConsistentQuery_InvalidQueryTask() {
	taskList := "taskList"
	params := workerExecutionParameters{
		TaskList:                       taskList,
		Identity:                       "test-id-1",
		Logger:                         zap.NewNop(),
		NonDeterministicWorkflowPolicy: NonDeterministicWorkflowPolicyBlockWorkflow,
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	task := createWorkflowTask(nil, 3, "HelloWorld_Workflow")
	task.Query = &apiv1.WorkflowQuery{}
	task.Queries = map[string]*apiv1.WorkflowQuery{"query_id": {}}
	newWorkflowTaskWorkerInternal(taskHandler, t.service, testDomain, params, make(chan struct{}), nil)
	// query and queries are both specified so this is an invalid task
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)

	t.Error(err)
	t.Nil(request)
	t.Contains(err.Error(), "invalid query decision task")

	// There should be nothing in the cache.
	t.EqualValues(getWorkflowCache().Size(), 0)
}

func (t *TaskHandlersTestSuite) TestConsistentQuery_Success() {
	taskList := "tl1"
	checksum1 := "chck1"
	numberOfSignalsToComplete, err := getDefaultDataConverter().ToData(2)
	t.NoError(err)
	signal, err := getDefaultDataConverter().ToData("signal data")
	t.NoError(err)
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{
			TaskList: &apiv1.TaskList{Name: taskList},
			Input:    &apiv1.Payload{Data: numberOfSignalsToComplete},
		}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{
			ScheduledEventId: 2, BinaryChecksum: checksum1}),
		createTestEventWorkflowExecutionSignaledWithPayload(5, signalCh, signal),
		createTestEventDecisionTaskScheduled(6, &apiv1.DecisionTaskScheduledEventAttributes{}),
		createTestEventDecisionTaskStarted(7),
	}

	queries := map[string]*apiv1.WorkflowQuery{
		"id1": {QueryType: queryType},
		"id2": {QueryType: errQueryType},
	}
	task := createWorkflowTaskWithQueries(testEvents[0:3], 0, "QuerySignalWorkflow", queries)

	params := workerExecutionParameters{
		TaskList: taskList,
		Identity: "test-id-1",
		Logger:   t.logger,
	}

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	response := request.(*apiv1.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
	t.Len(response.Decisions, 0)
	expectedQueryResults := map[string]*apiv1.WorkflowQueryResult{
		"id1": {
			ResultType: apiv1.QueryResultType_QUERY_RESULT_TYPE_ANSWERED,
			Answer:     &apiv1.Payload{Data: []byte(fmt.Sprintf("\"%v\"\n", startingQueryValue))},
		},
		"id2": {
			ResultType:   apiv1.QueryResultType_QUERY_RESULT_TYPE_FAILED,
			ErrorMessage: queryErr,
		},
	}
	t.assertQueryResultsEqual(expectedQueryResults, response.QueryResults)

	secondTask := createWorkflowTaskWithQueries(testEvents, 3, "QuerySignalWorkflow", queries)
	secondTask.WorkflowExecution.RunId = task.WorkflowExecution.RunId
	request, err = taskHandler.ProcessWorkflowTask(&workflowTask{task: secondTask}, nil)
	response = request.(*apiv1.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
	t.Len(response.Decisions, 1)
	expectedQueryResults = map[string]*apiv1.WorkflowQueryResult{
		"id1": {
			ResultType: apiv1.QueryResultType_QUERY_RESULT_TYPE_ANSWERED,
			Answer:     &apiv1.Payload{Data: []byte(fmt.Sprintf("\"%v\"\n", "signal data"))},
		},
		"id2": {
			ResultType:   apiv1.QueryResultType_QUERY_RESULT_TYPE_FAILED,
			ErrorMessage: queryErr,
		},
	}
	t.assertQueryResultsEqual(expectedQueryResults, response.QueryResults)

	// clean up workflow left in cache
	getWorkflowCache().Delete(task.WorkflowExecution.RunId)
}

func (t *TaskHandlersTestSuite) assertQueryResultsEqual(expected map[string]*apiv1.WorkflowQueryResult, actual map[string]*apiv1.WorkflowQueryResult) {
	t.Equal(len(expected), len(actual))
	for expectedID, expectedResult := range expected {
		t.Contains(actual, expectedID)
		t.Equal(expectedResult, actual[expectedID])
	}
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_CancelActivityBeforeSent() {
	// Schedule an activity and see if we complete workflow.
	taskList := "tl1"
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{}),
		createTestEventDecisionTaskStarted(3),
	}
	task := createWorkflowTask(testEvents, 0, "HelloWorld_WorkflowCancel")

	params := workerExecutionParameters{
		TaskList: taskList,
		Identity: "test-id-1",
		Logger:   t.logger,
	}
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	response := request.(*apiv1.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
	t.Equal(1, len(response.Decisions))
	t.NotNil(response.Decisions[0].GetCompleteWorkflowExecutionDecisionAttributes())
}

func (t *TaskHandlersTestSuite) TestWorkflowTask_PageToken() {
	// Schedule a decision activity and see if we complete workflow.
	taskList := "tl1"
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{}),
	}
	task := createWorkflowTask(testEvents, 0, "HelloWorld_Workflow")
	task.NextPageToken = []byte("token")

	params := workerExecutionParameters{
		TaskList: taskList,
		Identity: "test-id-1",
		Logger:   t.logger,
	}

	nextEvents := []*apiv1.HistoryEvent{
		createTestEventDecisionTaskStarted(3),
	}

	historyIterator := &historyIteratorImpl{
		iteratorFunc: func(nextToken []byte) (*apiv1.History, []byte, error) {
			return &apiv1.History{Events: nextEvents}, nil, nil
		},
	}
	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task, historyIterator: historyIterator}, nil)
	response := request.(*apiv1.RespondDecisionTaskCompletedRequest)
	t.NoError(err)
	t.NotNil(response)
}

func (t *TaskHandlersTestSuite) TestLocalActivityRetry_DecisionHeartbeatFail() {
	backoffIntervalInSeconds := int32(1)
	backoffDuration := time.Second * time.Duration(backoffIntervalInSeconds)
	workflowComplete := false

	retryLocalActivityWorkflowFunc := func(ctx Context, intput []byte) error {
		ao := LocalActivityOptions{
			ScheduleToCloseTimeout: time.Minute,
			RetryPolicy: &RetryPolicy{
				InitialInterval:    backoffDuration,
				BackoffCoefficient: 1.1,
				MaximumInterval:    time.Minute,
				ExpirationInterval: time.Minute,
			},
		}
		ctx = WithLocalActivityOptions(ctx, ao)

		err := ExecuteLocalActivity(ctx, func() error {
			return errors.New("some random error")
		}).Get(ctx, nil)
		workflowComplete = true
		return err
	}
	t.registry.RegisterWorkflowWithOptions(
		retryLocalActivityWorkflowFunc,
		RegisterWorkflowOptions{Name: "RetryLocalActivityWorkflow"},
	)

	decisionTaskStartedEvent := createTestEventDecisionTaskStarted(3)
	decisionTaskStartedEvent.EventTime = api.TimeToProto(time.Now())
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{
			// make sure the timeout is same as the backoff interval
			TaskStartToCloseTimeout: api.SecondsToProto(backoffIntervalInSeconds),
			TaskList:                &apiv1.TaskList{Name: testWorkflowTaskTasklist}},
		),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{}),
		decisionTaskStartedEvent,
	}

	task := createWorkflowTask(testEvents, 0, "RetryLocalActivityWorkflow")
	stopCh := make(chan struct{})
	params := workerExecutionParameters{
		TaskList:          testWorkflowTaskTasklist,
		Identity:          "test-id-1",
		Logger:            t.logger,
		Tracer:            opentracing.NoopTracer{},
		WorkerStopChannel: stopCh,
	}
	defer close(stopCh)

	taskHandler := newWorkflowTaskHandler(testDomain, params, nil, t.registry)
	laTunnel := newLocalActivityTunnel(params.WorkerStopChannel)
	taskHandlerImpl, ok := taskHandler.(*workflowTaskHandlerImpl)
	t.True(ok)
	taskHandlerImpl.laTunnel = laTunnel

	laTaskPoller := newLocalActivityPoller(params, laTunnel)
	doneCh := make(chan struct{})
	go func() {
		// laTaskPoller needs to poll the local activity and process it
		task, err := laTaskPoller.PollTask()
		t.NoError(err)
		err = laTaskPoller.ProcessTask(task)
		t.NoError(err)

		// before clearing workflow state, a reset sticky task will be sent to this chan,
		// drain the chan so that workflow state can be cleared
		<-laTunnel.resultCh

		close(doneCh)
	}()

	laResultCh := make(chan *localActivityResult)
	response, err := taskHandler.ProcessWorkflowTask(
		&workflowTask{
			task:       task,
			laResultCh: laResultCh,
		},
		func(response interface{}, startTime time.Time) (*workflowTask, error) {
			return nil, &api.EntityNotExistsError{Message: "Decision task not found."}
		})
	t.Nil(response)
	t.Error(err)

	// wait for the retry timer to fire
	time.Sleep(backoffDuration)
	t.False(workflowComplete)
	<-doneCh
}

func (t *TaskHandlersTestSuite) TestHeartBeat_NoError() {
	mockCtrl := gomock.NewController(t.T())
	mockService := api.NewMockInterface(mockCtrl)

	heartbeatResponse := apiv1.RecordActivityTaskHeartbeatResponse{CancelRequested: false}
	mockService.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).Return(&heartbeatResponse, nil)

	cadenceInvoker := &cadenceInvoker{
		identity:  "Test_Cadence_Invoker",
		service:   mockService,
		taskToken: nil,
	}

	heartbeatErr := cadenceInvoker.BatchHeartbeat(nil)
	t.NoError(heartbeatErr)
}

func newHeartbeatRequestMatcher(details []byte) gomock.Matcher {
	return &recordHeartbeatRequestMatcher{details: details}
}

type recordHeartbeatRequestMatcher struct {
	details []byte
}

func (r *recordHeartbeatRequestMatcher) String() string {
	panic("implement me")
}

func (r *recordHeartbeatRequestMatcher) Matches(x interface{}) bool {
	req, ok := x.(*apiv1.RecordActivityTaskHeartbeatRequest)
	if !ok {
		return false
	}

	return reflect.DeepEqual(r.details, req.Details.GetData())
}

func (t *TaskHandlersTestSuite) TestHeartBeat_Interleaved() {
	mockCtrl := gomock.NewController(t.T())
	mockService := api.NewMockInterface(mockCtrl)

	heartbeatResponse := apiv1.RecordActivityTaskHeartbeatResponse{CancelRequested: false}
	mockService.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), newHeartbeatRequestMatcher([]byte("1")), callOptions...).Return(&heartbeatResponse, nil).Times(3)
	mockService.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), newHeartbeatRequestMatcher([]byte("2")), callOptions...).Return(&heartbeatResponse, nil).Times(3)

	cadenceInvoker := &cadenceInvoker{
		identity:              "Test_Cadence_Invoker",
		service:               mockService,
		taskToken:             nil,
		heartBeatTimeoutInSec: 3,
	}

	heartbeatErr := cadenceInvoker.BatchHeartbeat([]byte("1"))
	t.NoError(heartbeatErr)
	time.Sleep(1 * time.Second)

	for i := 0; i < 4; i++ {
		heartbeatErr = cadenceInvoker.BackgroundHeartbeat()
		t.NoError(heartbeatErr)
		time.Sleep(time.Second)
	}

	heartbeatErr = cadenceInvoker.BatchHeartbeat([]byte("2"))
	t.NoError(heartbeatErr)
	time.Sleep(1 * time.Second)

	for i := 0; i < 4; i++ {
		heartbeatErr = cadenceInvoker.BackgroundHeartbeat()
		t.NoError(heartbeatErr)
		time.Sleep(time.Second)
	}
}

func (t *TaskHandlersTestSuite) TestHeartBeat_NilResponseWithError() {
	mockCtrl := gomock.NewController(t.T())
	mockService := api.NewMockInterface(mockCtrl)

	entityNotExistsError := &api.EntityNotExistsError{}
	mockService.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).Return(nil, entityNotExistsError)

	cadenceInvoker := newServiceInvoker(
		nil,
		"Test_Cadence_Invoker",
		mockService,
		func() {},
		0,
		make(chan struct{}),
		featureFlags,
	)

	heartbeatErr := cadenceInvoker.BatchHeartbeat(nil)
	t.NotNil(heartbeatErr)
	_, ok := (heartbeatErr).(*api.EntityNotExistsError)
	t.True(ok, "heartbeatErr must be EntityNotExistsError.")
}

func (t *TaskHandlersTestSuite) TestHeartBeat_NilResponseWithDomainNotActiveError() {
	mockCtrl := gomock.NewController(t.T())
	mockService := api.NewMockInterface(mockCtrl)

	domainNotActiveError := &api.DomainNotActiveError{}
	mockService.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).Return(nil, domainNotActiveError)

	called := false
	cancelHandler := func() { called = true }

	cadenceInvoker := newServiceInvoker(
		nil,
		"Test_Cadence_Invoker",
		mockService,
		cancelHandler,
		0,
		make(chan struct{}),
		featureFlags,
	)

	heartbeatErr := cadenceInvoker.BatchHeartbeat(nil)
	t.NotNil(heartbeatErr)
	_, ok := (heartbeatErr).(*api.DomainNotActiveError)
	t.True(ok, "heartbeatErr must be DomainNotActiveError.")
	t.True(called)
}

type testActivityDeadline struct {
	logger *zap.Logger
	d      time.Duration
}

func (t *testActivityDeadline) Execute(ctx context.Context, input []byte) ([]byte, error) {
	if d, _ := ctx.Deadline(); d.IsZero() {
		panic("invalid deadline provided")
	}
	if t.d != 0 {
		// Wait till deadline expires.
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return nil, nil
}

func (t *testActivityDeadline) ActivityType() ActivityType {
	return ActivityType{Name: "test"}
}

func (t *testActivityDeadline) GetFunction() interface{} {
	return t.Execute
}

func (t *testActivityDeadline) GetOptions() RegisterActivityOptions {
	return RegisterActivityOptions{}
}

type deadlineTest struct {
	actWaitDuration  time.Duration
	ScheduleTS       time.Time
	ScheduleDuration int32
	StartTS          time.Time
	StartDuration    int32
	err              error
}

func (t *TaskHandlersTestSuite) TestActivityExecutionDeadline() {
	deadlineTests := []deadlineTest{
		{time.Duration(0), time.Now(), 3, time.Now(), 3, nil},
		{time.Duration(0), time.Now(), 4, time.Now(), 3, nil},
		{time.Duration(0), time.Now(), 3, time.Now(), 4, nil},
		{time.Duration(0), time.Now().Add(-1 * time.Second), 1, time.Now(), 1, context.DeadlineExceeded},
		{time.Duration(0), time.Now(), 1, time.Now().Add(-1 * time.Second), 1, context.DeadlineExceeded},
		{time.Duration(0), time.Now().Add(-1 * time.Second), 1, time.Now().Add(-1 * time.Second), 1, context.DeadlineExceeded},
		{time.Duration(1 * time.Second), time.Now(), 1, time.Now(), 1, context.DeadlineExceeded},
		{time.Duration(1 * time.Second), time.Now(), 2, time.Now(), 1, context.DeadlineExceeded},
		{time.Duration(1 * time.Second), time.Now(), 1, time.Now(), 2, context.DeadlineExceeded},
	}
	a := &testActivityDeadline{logger: t.logger}
	registry := t.registry
	registry.addActivityWithLock(a.ActivityType().Name, a)

	mockCtrl := gomock.NewController(t.T())
	mockService := api.NewMockInterface(mockCtrl)

	for i, d := range deadlineTests {
		a.d = d.actWaitDuration
		wep := workerExecutionParameters{
			Logger:        t.logger,
			DataConverter: getDefaultDataConverter(),
			Tracer:        opentracing.NoopTracer{},
		}
		activityHandler := newActivityTaskHandler(mockService, wep, registry)
		pats := &apiv1.PollForActivityTaskResponse{
			TaskToken: []byte("token"),
			WorkflowExecution: &apiv1.WorkflowExecution{
				WorkflowId: "wID",
				RunId:      "rID"},
			ActivityType:               &apiv1.ActivityType{Name: "test"},
			ActivityId:                 uuid.New(),
			ScheduledTime:              api.TimeToProto(d.ScheduleTS),
			ScheduledTimeOfThisAttempt: api.TimeToProto(d.ScheduleTS),
			ScheduleToCloseTimeout:     api.SecondsToProto(d.ScheduleDuration),
			StartedTime:                api.TimeToProto(d.StartTS),
			StartToCloseTimeout:        api.SecondsToProto(d.StartDuration),
			WorkflowType: &apiv1.WorkflowType{
				Name: "wType",
			},
			WorkflowDomain: "domain",
		}
		td := fmt.Sprintf("testIndex: %v, testDetails: %v", i, d)
		r, err := activityHandler.Execute(tasklist, pats)
		t.logger.Info(fmt.Sprintf("test: %v, result: %v err: %v", td, r, err))
		t.Equal(d.err, err, td)
		if err != nil {
			t.Nil(r, td)
		}
	}
}

func activityWithWorkerStop(ctx context.Context) error {
	fmt.Println("Executing Activity with worker stop")
	workerStopCh := GetWorkerStopChannel(ctx)

	select {
	case <-workerStopCh:
		return nil
	case <-time.NewTimer(time.Second * 5).C:
		return fmt.Errorf("Activity failed to handle worker stop event")
	}
}

func (t *TaskHandlersTestSuite) TestActivityExecutionWorkerStop() {
	a := &testActivityDeadline{logger: t.logger}
	registry := t.registry
	registry.RegisterActivityWithOptions(
		activityWithWorkerStop,
		RegisterActivityOptions{Name: a.ActivityType().Name, DisableAlreadyRegisteredCheck: true},
	)

	mockCtrl := gomock.NewController(t.T())
	mockService := api.NewMockInterface(mockCtrl)
	workerStopCh := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	wep := workerExecutionParameters{
		Logger:            t.logger,
		DataConverter:     getDefaultDataConverter(),
		UserContext:       ctx,
		UserContextCancel: cancel,
		WorkerStopChannel: workerStopCh,
	}
	activityHandler := newActivityTaskHandler(mockService, wep, registry)
	pats := &apiv1.PollForActivityTaskResponse{
		TaskToken: []byte("token"),
		WorkflowExecution: &apiv1.WorkflowExecution{
			WorkflowId: "wID",
			RunId:      "rID"},
		ActivityType:               &apiv1.ActivityType{Name: "test"},
		ActivityId:                 uuid.New(),
		ScheduledTime:              api.TimeToProto(time.Now()),
		ScheduledTimeOfThisAttempt: api.TimeToProto(time.Now()),
		ScheduleToCloseTimeout:     api.SecondsToProto(1),
		StartedTime:                api.TimeToProto(time.Now()),
		StartToCloseTimeout:        api.SecondsToProto(1),
		WorkflowType: &apiv1.WorkflowType{
			Name: "wType",
		},
		WorkflowDomain: "domain",
	}
	close(workerStopCh)
	r, err := activityHandler.Execute(tasklist, pats)
	t.NoError(err)
	t.NotNil(r)
}

func Test_NonDeterministicCheck(t *testing.T) {
	decisionTypes := (&apiv1.Decision{}).XXX_OneofWrappers()
	require.Equal(t, 13, len(decisionTypes), "If you see this error, you are adding new decision type. "+
		"Before updating the number to make this test pass, please make sure you update isDecisionMatchEvent() method "+
		"to check the new decision type. Otherwise the replay will fail on the new decision event.")

	eventTypes := (&apiv1.HistoryEvent{}).XXX_OneofWrappers()
	decisionEventTypeCount := 0
	for _, et := range eventTypes {
		if isDecisionEventAttributes(et) {
			decisionEventTypeCount++
		}
	}
	// CancelTimer has 2 corresponding events.
	require.Equal(t, len(decisionTypes)+1, decisionEventTypeCount, "Every decision type must have one matching event type. "+
		"If you add new decision type, you need to update isDecisionEvent() method to include that new event type as well.")
}

func Test_IsDecisionMatchEvent_UpsertWorkflowSearchAttributes(t *testing.T) {
	strictMode := false

	testCases := []struct {
		name     string
		decision *apiv1.Decision
		event    *apiv1.HistoryEvent
		expected bool
	}{
		{
			name: "event type not match",
			decision: &apiv1.Decision{
				Attributes: &apiv1.Decision_UpsertWorkflowSearchAttributesDecisionAttributes{
					UpsertWorkflowSearchAttributesDecisionAttributes: &apiv1.UpsertWorkflowSearchAttributesDecisionAttributes{
						SearchAttributes: &apiv1.SearchAttributes{},
					},
				},
			},
			event:    &apiv1.HistoryEvent{},
			expected: false,
		},
		{
			name: "attributes not match",
			decision: &apiv1.Decision{
				Attributes: &apiv1.Decision_UpsertWorkflowSearchAttributesDecisionAttributes{
					UpsertWorkflowSearchAttributesDecisionAttributes: &apiv1.UpsertWorkflowSearchAttributesDecisionAttributes{
						SearchAttributes: &apiv1.SearchAttributes{},
					},
				},
			},
			event: &apiv1.HistoryEvent{
				Attributes: &apiv1.HistoryEvent_UpsertWorkflowSearchAttributesEventAttributes{
					UpsertWorkflowSearchAttributesEventAttributes: &apiv1.UpsertWorkflowSearchAttributesEventAttributes{},
				},
			},
			expected: true,
		},
		{
			name: "attributes match",
			decision: &apiv1.Decision{
				Attributes: &apiv1.Decision_UpsertWorkflowSearchAttributesDecisionAttributes{
					UpsertWorkflowSearchAttributesDecisionAttributes: &apiv1.UpsertWorkflowSearchAttributesDecisionAttributes{
						SearchAttributes: &apiv1.SearchAttributes{},
					},
				},
			},
			event: &apiv1.HistoryEvent{
				Attributes: &apiv1.HistoryEvent_UpsertWorkflowSearchAttributesEventAttributes{
					UpsertWorkflowSearchAttributesEventAttributes: &apiv1.UpsertWorkflowSearchAttributesEventAttributes{
						SearchAttributes: &apiv1.SearchAttributes{},
					},
				},
			},
			expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.expected, isDecisionMatchEvent(testCase.decision, testCase.event, strictMode))
		})
	}

	strictMode = true

	testCases = []struct {
		name     string
		decision *apiv1.Decision
		event    *apiv1.HistoryEvent
		expected bool
	}{
		{
			name: "attributes not match",
			decision: &apiv1.Decision{
				Attributes: &apiv1.Decision_UpsertWorkflowSearchAttributesDecisionAttributes{
					UpsertWorkflowSearchAttributesDecisionAttributes: &apiv1.UpsertWorkflowSearchAttributesDecisionAttributes{
						SearchAttributes: &apiv1.SearchAttributes{},
					},
				},
			},
			event: &apiv1.HistoryEvent{
				Attributes: &apiv1.HistoryEvent_UpsertWorkflowSearchAttributesEventAttributes{
					UpsertWorkflowSearchAttributesEventAttributes: &apiv1.UpsertWorkflowSearchAttributesEventAttributes{},
				},
			},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.expected, isDecisionMatchEvent(testCase.decision, testCase.event, strictMode))
		})
	}
}

func Test_IsSearchAttributesMatched(t *testing.T) {
	testCases := []struct {
		name     string
		lhs      *apiv1.SearchAttributes
		rhs      *apiv1.SearchAttributes
		expected bool
	}{
		{
			name:     "both nil",
			lhs:      nil,
			rhs:      nil,
			expected: true,
		},
		{
			name:     "left nil",
			lhs:      nil,
			rhs:      &apiv1.SearchAttributes{},
			expected: false,
		},
		{
			name:     "right nil",
			lhs:      &apiv1.SearchAttributes{},
			rhs:      nil,
			expected: false,
		},
		{
			name: "not match",
			lhs: &apiv1.SearchAttributes{
				IndexedFields: map[string]*apiv1.Payload{
					"key1": {Data: []byte("1")},
					"key2": {Data: []byte("abc")},
				},
			},
			rhs:      &apiv1.SearchAttributes{},
			expected: false,
		},
		{
			name: "match",
			lhs: &apiv1.SearchAttributes{
				IndexedFields: map[string]*apiv1.Payload{
					"key1": {Data: []byte("1")},
					"key2": {Data: []byte("abc")},
				},
			},
			rhs: &apiv1.SearchAttributes{
				IndexedFields: map[string]*apiv1.Payload{
					"key2": {Data: []byte("abc")},
					"key1": {Data: []byte("1")},
				},
			},
			expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.expected, isSearchAttributesMatched(testCase.lhs, testCase.rhs))
		})
	}
}
