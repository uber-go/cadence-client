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
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
	"go.uber.org/cadence/v2/internal/api"
	"go.uber.org/yarpc"
	"go.uber.org/zap"
)

// ActivityTaskHandler never returns response
type noResponseActivityTaskHandler struct {
	isExecuteCalled chan struct{}
}

func newNoResponseActivityTaskHandler() *noResponseActivityTaskHandler {
	return &noResponseActivityTaskHandler{isExecuteCalled: make(chan struct{})}
}

func (ath noResponseActivityTaskHandler) Execute(taskList string, task *apiv1.PollForActivityTaskResponse) (interface{}, error) {
	close(ath.isExecuteCalled)
	c := make(chan struct{})
	<-c
	return nil, nil
}

func (ath noResponseActivityTaskHandler) BlockedOnExecuteCalled() error {
	<-ath.isExecuteCalled
	return nil
}

type (
	WorkersTestSuite struct {
		suite.Suite
		mockCtrl *gomock.Controller
		service  *api.MockInterface
	}
)

// Test suite.
func (s *WorkersTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.service = api.NewMockInterface(s.mockCtrl)
}

func (s *WorkersTestSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

func TestWorkersTestSuite(t *testing.T) {
	formatter := &log.TextFormatter{}
	formatter.FullTimestamp = true
	log.SetFormatter(formatter)
	log.SetLevel(log.DebugLevel)

	suite.Run(t, new(WorkersTestSuite))
}

func (s *WorkersTestSuite) TestWorkflowWorker() {
	domain := "testDomain"
	logger, _ := zap.NewDevelopment()

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil)
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.PollForDecisionTaskResponse{}, nil).AnyTimes()
	s.service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).AnyTimes()

	ctx, cancel := context.WithCancel(context.Background())
	executionParameters := workerExecutionParameters{
		TaskList:                     "testTaskList",
		MaxConcurrentDecisionPollers: 5,
		Logger:                       logger,
		UserContext:                  ctx,
		UserContextCancel:            cancel,
	}
	overrides := &workerOverrides{workflowTaskHandler: newSampleWorkflowTaskHandler()}
	workflowWorker := newWorkflowWorkerInternal(
		s.service, domain, executionParameters, nil, overrides, newRegistry(), nil,
	)
	workflowWorker.Start()
	workflowWorker.Stop()

	s.Nil(ctx.Err())
}

func (s *WorkersTestSuite) TestActivityWorker() {
	s.testActivityWorker(false)
}

func (s *WorkersTestSuite) TestActivityWorkerWithLocalActivityDispatch() {
	s.testActivityWorker(true)
}

func (s *WorkersTestSuite) testActivityWorker(useLocallyDispatched bool) {
	domain := "testDomain"
	logger, _ := zap.NewDevelopment()

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil)
	s.service.EXPECT().PollForActivityTask(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.PollForActivityTaskResponse{}, nil).AnyTimes()
	s.service.EXPECT().RespondActivityTaskCompleted(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).AnyTimes()

	executionParameters := workerExecutionParameters{
		TaskList:                     "testTaskList",
		MaxConcurrentActivityPollers: 5,
		Logger:                       logger,
	}
	overrides := &workerOverrides{activityTaskHandler: newSampleActivityTaskHandler(), useLocallyDispatchedActivityPoller: useLocallyDispatched}
	a := &greeterActivity{}
	registry := newRegistry()
	registry.addActivityWithLock(a.ActivityType().Name, a)
	activityWorker := newActivityWorker(
		s.service, domain, executionParameters, overrides, registry, nil,
	)
	activityWorker.Start()
	activityWorker.Stop()
}

func (s *WorkersTestSuite) TestActivityWorkerStop() {
	domain := "testDomain"
	logger, _ := zap.NewDevelopment()

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

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil)
	s.service.EXPECT().PollForActivityTask(gomock.Any(), gomock.Any(), callOptions...).Return(pats, nil).AnyTimes()
	s.service.EXPECT().RespondActivityTaskCompleted(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).AnyTimes()

	stopC := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	executionParameters := workerExecutionParameters{
		TaskList:                        "testTaskList",
		MaxConcurrentActivityPollers:    5,
		ConcurrentActivityExecutionSize: 2,
		Logger:                          logger,
		UserContext:                     ctx,
		UserContextCancel:               cancel,
		WorkerStopTimeout:               time.Second * 2,
		WorkerStopChannel:               stopC,
	}
	activityTaskHandler := newNoResponseActivityTaskHandler()
	overrides := &workerOverrides{activityTaskHandler: activityTaskHandler}
	a := &greeterActivity{}
	registry := newRegistry()
	registry.addActivityWithLock(a.ActivityType().Name, a)
	worker := newActivityWorker(
		s.service, domain, executionParameters, overrides, registry, nil,
	)
	worker.Start()
	activityTaskHandler.BlockedOnExecuteCalled()
	go worker.Stop()

	<-worker.worker.shutdownCh
	err := ctx.Err()
	s.NoError(err)

	<-ctx.Done()
	err = ctx.Err()
	s.Error(err)
}

func (s *WorkersTestSuite) TestPollForDecisionTask_InternalServiceError() {
	domain := "testDomain"

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil)
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.PollForDecisionTaskResponse{}, &api.InternalServiceError{}).AnyTimes()

	executionParameters := workerExecutionParameters{
		TaskList:                     "testDecisionTaskList",
		MaxConcurrentDecisionPollers: 5,
		Logger:                       zap.NewNop(),
	}
	overrides := &workerOverrides{workflowTaskHandler: newSampleWorkflowTaskHandler()}
	workflowWorker := newWorkflowWorkerInternal(
		s.service, domain, executionParameters, nil, overrides, newRegistry(), nil,
	)
	workflowWorker.Start()
	workflowWorker.Stop()
}

func (s *WorkersTestSuite) TestLongRunningDecisionTask() {
	localActivityCalledCount := 0
	localActivitySleep := func(duration time.Duration) error {
		time.Sleep(duration)
		localActivityCalledCount++
		return nil
	}

	doneCh := make(chan struct{})

	isWorkflowCompleted := false
	longDecisionWorkflowFn := func(ctx Context, input []byte) error {
		lao := LocalActivityOptions{
			ScheduleToCloseTimeout: time.Second * 2,
		}
		ctx = WithLocalActivityOptions(ctx, lao)
		err := ExecuteLocalActivity(ctx, localActivitySleep, time.Second).Get(ctx, nil)

		if err != nil {
			return err
		}

		err = ExecuteLocalActivity(ctx, localActivitySleep, time.Second).Get(ctx, nil)
		isWorkflowCompleted = true
		return err
	}

	domain := "testDomain"
	taskList := "long-running-decision-tl"
	testEvents := []*apiv1.HistoryEvent{
		{
			EventId: 1,
			Attributes: &apiv1.HistoryEvent_WorkflowExecutionStartedEventAttributes{
				WorkflowExecutionStartedEventAttributes: &apiv1.WorkflowExecutionStartedEventAttributes{
					TaskList:                     &apiv1.TaskList{Name: taskList},
					ExecutionStartToCloseTimeout: api.SecondsToProto(10),
					TaskStartToCloseTimeout:      api.SecondsToProto(2),
					WorkflowType:                 &apiv1.WorkflowType{Name: "long-running-decision-workflow-type"},
				},
			},
		},
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{ScheduledEventId: 2}),
		{
			EventId: 5,
			Attributes: &apiv1.HistoryEvent_MarkerRecordedEventAttributes{
				MarkerRecordedEventAttributes: &apiv1.MarkerRecordedEventAttributes{
					MarkerName:                   localActivityMarkerName,
					Details:                      &apiv1.Payload{Data: s.createLocalActivityMarkerDataForTest("0")},
					DecisionTaskCompletedEventId: 4,
				},
			},
		},
		createTestEventDecisionTaskScheduled(6, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(7),
		createTestEventDecisionTaskCompleted(8, &apiv1.DecisionTaskCompletedEventAttributes{ScheduledEventId: 2}),
		{
			EventId: 9,
			Attributes: &apiv1.HistoryEvent_MarkerRecordedEventAttributes{
				MarkerRecordedEventAttributes: &apiv1.MarkerRecordedEventAttributes{
					MarkerName:                   localActivityMarkerName,
					Details:                      &apiv1.Payload{Data: s.createLocalActivityMarkerDataForTest("1")},
					DecisionTaskCompletedEventId: 8,
				},
			},
		},
		createTestEventDecisionTaskScheduled(10, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(11),
	}

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).AnyTimes()
	task := &apiv1.PollForDecisionTaskResponse{
		TaskToken: []byte("test-token"),
		WorkflowExecution: &apiv1.WorkflowExecution{
			WorkflowId: "long-running-decision-workflow-id",
			RunId:      "long-running-decision-workflow-run-id",
		},
		WorkflowType: &apiv1.WorkflowType{
			Name: "long-running-decision-workflow-type",
		},
		PreviousStartedEventId: &types.Int64Value{Value: 0},
		StartedEventId:         3,
		History:                &apiv1.History{Events: testEvents[0:3]},
		NextPageToken:          nil,
		NextEventId:            4,
	}
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions...).Return(task, nil).Times(1)
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.PollForDecisionTaskResponse{}, &api.InternalServiceError{}).AnyTimes()

	respondCounter := 0
	s.service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), callOptions...).DoAndReturn(func(ctx context.Context, request *apiv1.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption,
	) (success *apiv1.RespondDecisionTaskCompletedResponse, err error) {
		respondCounter++
		switch respondCounter {
		case 1:
			s.Equal(1, len(request.Decisions))
			s.NotNil(request.Decisions[0].GetRecordMarkerDecisionAttributes())
			task.PreviousStartedEventId = &types.Int64Value{Value: 3}
			task.StartedEventId = 7
			task.History.Events = testEvents[3:7]
			return &apiv1.RespondDecisionTaskCompletedResponse{DecisionTask: task}, nil
		case 2:
			s.Equal(2, len(request.Decisions))
			s.NotNil(request.Decisions[0].GetRecordMarkerDecisionAttributes())
			s.NotNil(request.Decisions[1].GetCompleteWorkflowExecutionDecisionAttributes())
			task.PreviousStartedEventId = &types.Int64Value{Value: 7}
			task.StartedEventId = 11
			task.History.Events = testEvents[7:11]
			close(doneCh)
			return nil, nil
		default:
			panic("unexpected RespondDecisionTaskCompleted")
		}
	}).Times(2)

	options := WorkerOptions{
		Logger:                zap.NewNop(),
		DisableActivityWorker: true,
		Identity:              "test-worker-identity",
	}
	worker := newAggregatedWorker(s.service, domain, taskList, options)
	worker.RegisterWorkflowWithOptions(
		longDecisionWorkflowFn,
		RegisterWorkflowOptions{Name: "long-running-decision-workflow-type"},
	)
	worker.RegisterActivity(localActivitySleep)

	worker.Start()
	// wait for test to complete
	select {
	case <-doneCh:
		break
	case <-time.After(time.Second * 4):
	}
	worker.Stop()

	s.True(isWorkflowCompleted)
	s.Equal(2, localActivityCalledCount)
}

func (s *WorkersTestSuite) TestQueryTask_WorkflowCacheEvicted() {
	domain := "testDomain"
	taskList := "query-task-cache-evicted-tl"
	workflowType := "query-task-cache-evicted-workflow"
	workflowID := "query-task-cache-evicted-workflow-id"
	runID := "query-task-cache-evicted-workflow-run-id"
	activityType := "query-task-cache-evicted-activity"
	queryType := "state"
	doneCh := make(chan struct{})

	activityFn := func(ctx context.Context) error {
		return nil
	}

	queryWorkflowFn := func(ctx Context) error {
		queryResult := "started"
		// setup query handler for query type "state"
		if err := SetQueryHandler(ctx, queryType, func(input []byte) (string, error) {
			return queryResult, nil
		}); err != nil {
			return err
		}

		queryResult = "waiting on timer"
		NewTimer(ctx, time.Minute*2).Get(ctx, nil)

		queryResult = "waiting on activity"
		ctx = WithActivityOptions(ctx, ActivityOptions{
			ScheduleToStartTimeout: 10 * time.Second,
			StartToCloseTimeout:    10 * time.Second,
		})
		if err := ExecuteActivity(ctx, activityFn).Get(ctx, nil); err != nil {
			return err
		}
		queryResult = "done"
		return nil
	}

	testEvents := []*apiv1.HistoryEvent{
		{
			EventId: 1,
			Attributes: &apiv1.HistoryEvent_WorkflowExecutionStartedEventAttributes{
				WorkflowExecutionStartedEventAttributes: &apiv1.WorkflowExecutionStartedEventAttributes{
					TaskList:                     &apiv1.TaskList{Name: taskList},
					ExecutionStartToCloseTimeout: api.SecondsToProto(180),
					TaskStartToCloseTimeout:      api.SecondsToProto(2),
					WorkflowType:                 &apiv1.WorkflowType{Name: workflowType},
				},
			},
		},
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{ScheduledEventId: 2}),
		{
			EventId: 5,
			Attributes: &apiv1.HistoryEvent_TimerStartedEventAttributes{
				TimerStartedEventAttributes: &apiv1.TimerStartedEventAttributes{
					TimerId:                      "0",
					StartToFireTimeout:           api.SecondsToProto(120),
					DecisionTaskCompletedEventId: 4,
				},
			},
		},
		{
			EventId: 6,
			Attributes: &apiv1.HistoryEvent_TimerFiredEventAttributes{
				TimerFiredEventAttributes: &apiv1.TimerFiredEventAttributes{
					TimerId:        "0",
					StartedEventId: 5,
				},
			},
		},
		createTestEventDecisionTaskScheduled(7, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(8),
		createTestEventDecisionTaskCompleted(9, &apiv1.DecisionTaskCompletedEventAttributes{ScheduledEventId: 2}),
		createTestEventActivityTaskScheduled(10, &apiv1.ActivityTaskScheduledEventAttributes{
			ActivityId: "1",
			ActivityType: &apiv1.ActivityType{
				Name: activityType,
			},
			Domain: domain,
			TaskList: &apiv1.TaskList{
				Name: taskList,
			},
			ScheduleToStartTimeout:       api.SecondsToProto(10),
			StartToCloseTimeout:          api.SecondsToProto(10),
			DecisionTaskCompletedEventId: 9,
		}),
	}

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).AnyTimes()
	task := &apiv1.PollForDecisionTaskResponse{
		TaskToken: []byte("test-token"),
		WorkflowExecution: &apiv1.WorkflowExecution{
			WorkflowId: workflowID,
			RunId:      runID,
		},
		WorkflowType: &apiv1.WorkflowType{
			Name: workflowType,
		},
		PreviousStartedEventId: &types.Int64Value{Value: 0},
		StartedEventId:         3,
		History:                &apiv1.History{Events: testEvents[0:3]},
		NextPageToken:          nil,
		NextEventId:            4,
	}
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions...).Return(task, nil).Times(1)
	s.service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), callOptions...).DoAndReturn(func(ctx context.Context, request *apiv1.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption,
	) (success *apiv1.RespondDecisionTaskCompletedResponse, err error) {
		s.Equal(1, len(request.Decisions))
		s.NotNil(request.Decisions[0].GetStartTimerDecisionAttributes())
		return &apiv1.RespondDecisionTaskCompletedResponse{}, nil
	}).Times(1)
	queryTask := &apiv1.PollForDecisionTaskResponse{
		TaskToken: []byte("test-token"),
		WorkflowExecution: &apiv1.WorkflowExecution{
			WorkflowId: workflowID,
			RunId:      runID,
		},
		WorkflowType: &apiv1.WorkflowType{
			Name: workflowType,
		},
		PreviousStartedEventId: &types.Int64Value{Value: 3},
		History:                &apiv1.History{}, // sticky query, so there's no history
		NextPageToken:          nil,
		NextEventId:            5,
		Query: &apiv1.WorkflowQuery{
			QueryType: queryType,
		},
	}
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions...).DoAndReturn(func(ctx context.Context, request *apiv1.PollForDecisionTaskRequest, opts ...yarpc.CallOption,
	) (success *apiv1.PollForDecisionTaskResponse, err error) {
		getWorkflowCache().Delete(runID) // force remove the workflow state
		return queryTask, nil
	}).Times(1)
	s.service.EXPECT().ResetStickyTaskList(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.ResetStickyTaskListResponse{}, nil).AnyTimes()
	s.service.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.GetWorkflowExecutionHistoryResponse{
		History: &apiv1.History{Events: testEvents}, // workflow has made progress, return all available events
	}, nil).Times(1)
	dc := getDefaultDataConverter()
	expectedResult, err := dc.ToData("waiting on timer")
	s.NoError(err)
	s.service.EXPECT().RespondQueryTaskCompleted(gomock.Any(), gomock.Any(), callOptions...).DoAndReturn(func(ctx context.Context, request *apiv1.RespondQueryTaskCompletedRequest, opts ...yarpc.CallOption) error {
		s.Equal(apiv1.QueryResultType_QUERY_RESULT_TYPE_ANSWERED, request.Result.GetResultType())
		s.Equal(expectedResult, request.Result.Answer.GetData())
		close(doneCh)
		return nil
	}).Times(1)
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.PollForDecisionTaskResponse{}, &api.InternalServiceError{}).AnyTimes()

	options := WorkerOptions{
		Logger:                zap.NewNop(),
		DisableActivityWorker: true,
		Identity:              "test-worker-identity",
		DataConverter:         dc,
	}
	worker := newAggregatedWorker(s.service, domain, taskList, options)
	worker.RegisterWorkflowWithOptions(
		queryWorkflowFn,
		RegisterWorkflowOptions{Name: workflowType},
	)
	worker.RegisterActivityWithOptions(
		activityFn,
		RegisterActivityOptions{Name: activityType},
	)

	worker.Start()
	<-doneCh
	worker.Stop()
}

func (s *WorkersTestSuite) TestMultipleLocalActivities() {
	localActivityCalledCount := 0
	localActivitySleep := func(duration time.Duration) error {
		time.Sleep(duration)
		localActivityCalledCount++
		return nil
	}

	doneCh := make(chan struct{})

	isWorkflowCompleted := false
	longDecisionWorkflowFn := func(ctx Context, input []byte) error {
		lao := LocalActivityOptions{
			ScheduleToCloseTimeout: time.Second * 2,
		}
		ctx = WithLocalActivityOptions(ctx, lao)
		err := ExecuteLocalActivity(ctx, localActivitySleep, time.Second).Get(ctx, nil)

		if err != nil {
			return err
		}

		err = ExecuteLocalActivity(ctx, localActivitySleep, time.Second).Get(ctx, nil)
		isWorkflowCompleted = true
		return err
	}

	domain := "testDomain"
	taskList := "multiple-local-activities-tl"
	testEvents := []*apiv1.HistoryEvent{
		{
			EventId: 1,
			Attributes: &apiv1.HistoryEvent_WorkflowExecutionStartedEventAttributes{
				WorkflowExecutionStartedEventAttributes: &apiv1.WorkflowExecutionStartedEventAttributes{
					TaskList:                     &apiv1.TaskList{Name: taskList},
					ExecutionStartToCloseTimeout: api.SecondsToProto(10),
					TaskStartToCloseTimeout:      api.SecondsToProto(3),
					WorkflowType:                 &apiv1.WorkflowType{Name: "multiple-local-activities-workflow-type"},
				},
			},
		},
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &apiv1.DecisionTaskCompletedEventAttributes{ScheduledEventId: 2}),
		{
			EventId: 5,
			Attributes: &apiv1.HistoryEvent_MarkerRecordedEventAttributes{
				MarkerRecordedEventAttributes: &apiv1.MarkerRecordedEventAttributes{
					MarkerName:                   localActivityMarkerName,
					Details:                      &apiv1.Payload{Data: s.createLocalActivityMarkerDataForTest("0")},
					DecisionTaskCompletedEventId: 4,
				},
			},
		},
		createTestEventDecisionTaskScheduled(6, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(7),
		createTestEventDecisionTaskCompleted(8, &apiv1.DecisionTaskCompletedEventAttributes{ScheduledEventId: 2}),
		{
			EventId: 9,
			Attributes: &apiv1.HistoryEvent_MarkerRecordedEventAttributes{
				MarkerRecordedEventAttributes: &apiv1.MarkerRecordedEventAttributes{
					MarkerName:                   localActivityMarkerName,
					Details:                      &apiv1.Payload{Data: s.createLocalActivityMarkerDataForTest("1")},
					DecisionTaskCompletedEventId: 8,
				},
			},
		},
		createTestEventDecisionTaskScheduled(10, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(11),
	}

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).AnyTimes()
	task := &apiv1.PollForDecisionTaskResponse{
		TaskToken: []byte("test-token"),
		WorkflowExecution: &apiv1.WorkflowExecution{
			WorkflowId: "multiple-local-activities-workflow-id",
			RunId:      "multiple-local-activities-workflow-run-id",
		},
		WorkflowType: &apiv1.WorkflowType{
			Name: "multiple-local-activities-workflow-type",
		},
		PreviousStartedEventId: &types.Int64Value{Value: 0},
		StartedEventId:         3,
		History:                &apiv1.History{Events: testEvents[0:3]},
		NextPageToken:          nil,
		NextEventId:            4,
	}
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions...).Return(task, nil).Times(1)
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.PollForDecisionTaskResponse{}, &api.InternalServiceError{}).AnyTimes()

	respondCounter := 0
	s.service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), callOptions...).DoAndReturn(func(ctx context.Context, request *apiv1.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption,
	) (success *apiv1.RespondDecisionTaskCompletedResponse, err error) {
		respondCounter++
		switch respondCounter {
		case 1:
			s.Equal(3, len(request.Decisions))
			s.NotNil(request.Decisions[0].GetRecordMarkerDecisionAttributes())
			task.PreviousStartedEventId = &types.Int64Value{Value: 3}
			task.StartedEventId = 7
			task.History.Events = testEvents[3:11]
			close(doneCh)
			return nil, nil
		default:
			panic("unexpected RespondDecisionTaskCompleted")
		}
	}).Times(1)

	options := WorkerOptions{
		Logger:                zap.NewNop(),
		DisableActivityWorker: true,
		Identity:              "test-worker-identity",
	}
	worker := newAggregatedWorker(s.service, domain, taskList, options)
	worker.RegisterWorkflowWithOptions(
		longDecisionWorkflowFn,
		RegisterWorkflowOptions{Name: "multiple-local-activities-workflow-type"},
	)
	worker.RegisterActivity(localActivitySleep)

	worker.Start()
	// wait for test to complete
	select {
	case <-doneCh:
		break
	case <-time.After(time.Second * 5):
	}
	worker.Stop()

	s.True(isWorkflowCompleted)
	s.Equal(2, localActivityCalledCount)
}

func (s *WorkersTestSuite) createLocalActivityMarkerDataForTest(activityID string) []byte {
	lamd := localActivityMarkerData{
		ActivityID: activityID,
		ReplayTime: time.Now(),
	}

	// encode marker data
	markerData, err := encodeArg(nil, lamd)
	s.NoError(err)
	return markerData
}

func (s *WorkersTestSuite) TestLocallyDispatchedActivity() {
	activityCalledCount := 0
	activitySleep := func(duration time.Duration) error {
		time.Sleep(duration)
		activityCalledCount++
		return nil
	}

	doneCh := make(chan struct{})

	isActivityResponseCompleted := false
	workflowFn := func(ctx Context, input []byte) error {
		ao := ActivityOptions{
			ScheduleToCloseTimeout: 1 * time.Second,
			ScheduleToStartTimeout: 1 * time.Second,
			StartToCloseTimeout:    1 * time.Second,
		}
		ctx = WithActivityOptions(ctx, ao)
		err := ExecuteActivity(ctx, activitySleep, 500*time.Millisecond).Get(ctx, nil)
		return err
	}

	domain := "testDomain"
	workflowType := "locally-dispatched-activity-workflow-type"
	taskList := "locally-dispatched-activity-tl"
	testEvents := []*apiv1.HistoryEvent{
		{
			EventId: 1,
			Attributes: &apiv1.HistoryEvent_WorkflowExecutionStartedEventAttributes{
				WorkflowExecutionStartedEventAttributes: &apiv1.WorkflowExecutionStartedEventAttributes{
					TaskList:                     &apiv1.TaskList{Name: taskList},
					ExecutionStartToCloseTimeout: api.SecondsToProto(10),
					TaskStartToCloseTimeout:      api.SecondsToProto(2),
					WorkflowType:                 &apiv1.WorkflowType{Name: workflowType},
				},
			},
		},
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
	}

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).AnyTimes()
	task := &apiv1.PollForDecisionTaskResponse{
		TaskToken: []byte("test-token"),
		WorkflowExecution: &apiv1.WorkflowExecution{
			WorkflowId: "locally-dispatched-activity-workflow-id",
			RunId:      "locally-dispatched-activity-workflow-run-id",
		},
		WorkflowType: &apiv1.WorkflowType{
			Name: workflowType,
		},
		PreviousStartedEventId: &types.Int64Value{Value: 0},
		StartedEventId:         3,
		History:                &apiv1.History{Events: testEvents[0:3]},
		NextPageToken:          nil,
		NextEventId:            4,
	}
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions...).Return(task, nil).Times(1)
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.PollForDecisionTaskResponse{}, &api.InternalServiceError{}).AnyTimes()
	s.service.EXPECT().PollForActivityTask(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).AnyTimes()
	s.service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), callOptions...).DoAndReturn(func(ctx context.Context, request *apiv1.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption,
	) (success *apiv1.RespondDecisionTaskCompletedResponse, err error) {
		s.Equal(1, len(request.Decisions))
		activitiesToDispatchLocally := make(map[string]*apiv1.ActivityLocalDispatchInfo)
		d := request.Decisions[0]
		s.NotNil(d.GetScheduleActivityTaskDecisionAttributes())
		activitiesToDispatchLocally[d.GetScheduleActivityTaskDecisionAttributes().ActivityId] =
			&apiv1.ActivityLocalDispatchInfo{
				ActivityId:                 d.GetScheduleActivityTaskDecisionAttributes().ActivityId,
				ScheduledTime:              api.TimeToProto(time.Now()),
				ScheduledTimeOfThisAttempt: api.TimeToProto(time.Now()),
				StartedTime:                api.TimeToProto(time.Now()),
				TaskToken:                  []byte("test-token")}
		return &apiv1.RespondDecisionTaskCompletedResponse{ActivitiesToDispatchLocally: activitiesToDispatchLocally}, nil
	}).Times(1)
	s.service.EXPECT().RespondActivityTaskCompleted(gomock.Any(), gomock.Any(), callOptions...).DoAndReturn(func(ctx context.Context, request *apiv1.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption,
	) (*apiv1.RespondActivityTaskCompletedResponse, error) {
		defer close(doneCh)
		isActivityResponseCompleted = true
		return nil, nil
	}).Times(1)

	options := WorkerOptions{
		Logger:   zap.NewNop(),
		Identity: "test-worker-identity",
	}
	worker := newAggregatedWorker(s.service, domain, taskList, options)
	worker.RegisterWorkflowWithOptions(
		workflowFn,
		RegisterWorkflowOptions{Name: workflowType},
	)
	worker.RegisterActivityWithOptions(activitySleep, RegisterActivityOptions{Name: "activitySleep"})

	worker.Start()
	// wait for test to complete
	select {
	case <-doneCh:
		break
	case <-time.After(1 * time.Second):
	}
	worker.Stop()

	s.True(isActivityResponseCompleted)
	s.Equal(1, activityCalledCount)
}

func (s *WorkersTestSuite) TestMultipleLocallyDispatchedActivity() {
	var activityCalledCount uint32 = 0
	activitySleep := func(duration time.Duration) error {
		time.Sleep(duration)
		atomic.AddUint32(&activityCalledCount, 1)
		return nil
	}

	doneCh := make(chan struct{})

	var activityCount uint32 = 5
	workflowFn := func(ctx Context, input []byte) error {
		ao := ActivityOptions{
			ScheduleToCloseTimeout: 1 * time.Second,
			ScheduleToStartTimeout: 1 * time.Second,
			StartToCloseTimeout:    1 * time.Second,
		}
		ctx = WithActivityOptions(ctx, ao)
		for i := 1; i < int(activityCount); i++ {
			ExecuteActivity(ctx, activitySleep, 500*time.Millisecond)
		}
		ExecuteActivity(ctx, activitySleep, 500*time.Millisecond).Get(ctx, nil)
		return nil
	}

	domain := "testDomain"
	workflowType := "locally-dispatched-multiple-activity-workflow-type"
	taskList := "locally-dispatched-multiple-activity-tl"
	testEvents := []*apiv1.HistoryEvent{
		{
			EventId: 1,
			Attributes: &apiv1.HistoryEvent_WorkflowExecutionStartedEventAttributes{
				WorkflowExecutionStartedEventAttributes: &apiv1.WorkflowExecutionStartedEventAttributes{
					TaskList:                     &apiv1.TaskList{Name: taskList},
					ExecutionStartToCloseTimeout: api.SecondsToProto(10),
					TaskStartToCloseTimeout:      api.SecondsToProto(2),
					WorkflowType:                 &apiv1.WorkflowType{Name: workflowType},
				},
			},
		},
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
	}

	options := WorkerOptions{
		Logger:   zap.NewNop(),
		Identity: "test-worker-identity",
	}

	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).AnyTimes()
	task := &apiv1.PollForDecisionTaskResponse{
		TaskToken: []byte("test-token"),
		WorkflowExecution: &apiv1.WorkflowExecution{
			WorkflowId: "locally-dispatched-multiple-activity-workflow-id",
			RunId:      "locally-dispatched-multiple-activity-workflow-run-id",
		},
		WorkflowType: &apiv1.WorkflowType{
			Name: workflowType,
		},
		PreviousStartedEventId: &types.Int64Value{Value: 0},
		StartedEventId:         3,
		History:                &apiv1.History{Events: testEvents[0:3]},
		NextPageToken:          nil,
		NextEventId:            4,
	}
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions...).Return(task, nil).Times(1)
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.PollForDecisionTaskResponse{}, &api.InternalServiceError{}).AnyTimes()
	s.service.EXPECT().PollForActivityTask(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).AnyTimes()
	s.service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), callOptions...).DoAndReturn(func(ctx context.Context, request *apiv1.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption,
	) (success *apiv1.RespondDecisionTaskCompletedResponse, err error) {
		s.Equal(int(activityCount), len(request.Decisions))
		activitiesToDispatchLocally := make(map[string]*apiv1.ActivityLocalDispatchInfo)
		for _, d := range request.Decisions {
			s.NotNil(d.GetScheduleActivityTaskDecisionAttributes())
			activitiesToDispatchLocally[d.GetScheduleActivityTaskDecisionAttributes().ActivityId] =
				&apiv1.ActivityLocalDispatchInfo{
					ActivityId:                 d.GetScheduleActivityTaskDecisionAttributes().ActivityId,
					ScheduledTime:              api.TimeToProto(time.Now()),
					ScheduledTimeOfThisAttempt: api.TimeToProto(time.Now()),
					StartedTime:                api.TimeToProto(time.Now()),
					TaskToken:                  []byte("test-token")}
		}
		return &apiv1.RespondDecisionTaskCompletedResponse{ActivitiesToDispatchLocally: activitiesToDispatchLocally}, nil
	}).Times(1)
	var activityResponseCompletedCount uint32 = 0
	s.service.EXPECT().RespondActivityTaskCompleted(gomock.Any(), gomock.Any(), callOptions...).DoAndReturn(func(ctx context.Context, request *apiv1.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption,
	) (*apiv1.RespondActivityTaskCompletedResponse, error) {
		defer func() {
			if atomic.LoadUint32(&activityResponseCompletedCount) == activityCount {
				close(doneCh)
			}
		}()
		atomic.AddUint32(&activityResponseCompletedCount, 1)
		return nil, nil
	}).MinTimes(1)

	worker := newAggregatedWorker(s.service, domain, taskList, options)
	worker.RegisterWorkflowWithOptions(
		workflowFn,
		RegisterWorkflowOptions{Name: workflowType},
	)
	worker.RegisterActivityWithOptions(activitySleep, RegisterActivityOptions{Name: "activitySleep"})
	s.NotNil(worker.locallyDispatchedActivityWorker)
	worker.Start()

	// wait for test to complete
	select {
	case <-doneCh:
		break
	case <-time.After(1 * time.Second):
	}
	worker.Stop()

	// for currently unbuffered channel at least one activity should be sent
	s.True(activityResponseCompletedCount > 0)
	s.True(activityCalledCount > 0)
}
