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

// This test must be its own package because workflow execution cache
// is package-level global variable, so any tests against it should belong to
// its own package to avoid inter-test interference because "go test" command
// builds one test binary per go package(even if the tests in the package are split
// among multiple .go source files) and then uses reflection on the per package
// binary to run tests.
// This means any test whose result hinges on having its own exclusive own of globals
// should be put in its own package to avoid conflicts in global variable accesses.
package evictiontest

import (
	"strconv"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"go.uber.org/atomic"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
	"go.uber.org/cadence/v2/internal"
	"go.uber.org/cadence/v2/internal/api"
	"go.uber.org/cadence/v2/worker"
	"go.uber.org/yarpc"
	"golang.org/x/net/context"
)

func testReplayWorkflow(ctx internal.Context) error {
	ao := internal.ActivityOptions{
		ScheduleToStartTimeout: time.Second,
		StartToCloseTimeout:    time.Second,
	}
	ctx = internal.WithActivityOptions(ctx, ao)
	err := internal.ExecuteActivity(ctx, "testActivity").Get(ctx, nil)
	if err != nil {
		panic("Failed workflow")
	}
	return err
}

type (
	CacheEvictionSuite struct {
		suite.Suite
		mockCtrl *gomock.Controller
		service *api.MockInterface
	}
)

// Test suite.
func (s *CacheEvictionSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.service = api.NewMockInterface(s.mockCtrl)
}

func (s *CacheEvictionSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

func TestWorkersTestSuite(t *testing.T) {
	formatter := &log.TextFormatter{}
	formatter.FullTimestamp = true
	log.SetFormatter(formatter)
	log.SetLevel(log.DebugLevel)
	suite.Run(t, new(CacheEvictionSuite))
}

// this is the mock for yarpcCallOptions, make sure length are the same
var callOptions = []interface{}{gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()}

func createTestEventWorkflowExecutionStarted(eventID int64, attr *apiv1.WorkflowExecutionStartedEventAttributes) *apiv1.HistoryEvent {
	return &apiv1.HistoryEvent{
		EventId:    eventID,
		Attributes: &apiv1.HistoryEvent_WorkflowExecutionStartedEventAttributes{WorkflowExecutionStartedEventAttributes: attr},
	}
}

func createTestEventDecisionTaskScheduled(eventID int64, attr *apiv1.DecisionTaskScheduledEventAttributes) *apiv1.HistoryEvent {
	return &apiv1.HistoryEvent{
		EventId:                              eventID,
		Attributes: &apiv1.HistoryEvent_DecisionTaskScheduledEventAttributes{DecisionTaskScheduledEventAttributes: attr},
	}
}

func (s *CacheEvictionSuite) TestResetStickyOnEviction() {
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{
			TaskList: &apiv1.TaskList{Name: "tasklist"},
		}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{}),
	}

	var taskCounter atomic.Int32 // lambda variable to keep count
	// mock that manufactures unique decision tasks
	mockPollForDecisionTask := func(
		ctx context.Context,
		_PollRequest *apiv1.PollForDecisionTaskRequest,
		opts ...yarpc.CallOption,
	) (success *apiv1.PollForDecisionTaskResponse, err error) {
		taskID := taskCounter.Inc()
		workflowID := "testID" + strconv.Itoa(int(taskID))
		runID := "runID" + strconv.Itoa(int(taskID))
		// how we initialize the response here is the result of a series of trial and error
		// the goal is we want to fabricate a response that looks real enough to our worker
		// that it will actually go along with processing it instead of just tossing it out
		// after polling it or giving an error
		ret := &apiv1.PollForDecisionTaskResponse{
			TaskToken:              make([]byte, 5),
			WorkflowExecution:      &apiv1.WorkflowExecution{WorkflowId: workflowID, RunId: runID},
			WorkflowType:           &apiv1.WorkflowType{Name: "go.uber.org/cadence/v2/evictiontest.testReplayWorkflow"},
			History:                &apiv1.History{Events: testEvents},
			PreviousStartedEventId: &types.Int64Value{Value: 5}}
		return ret, nil
	}

	resetStickyAPICalled := make(chan struct{})
	mockResetStickyTaskList := func(
		ctx context.Context,
		_ResetRequest *apiv1.ResetStickyTaskListRequest,
		opts ...yarpc.CallOption,
	) (success *apiv1.ResetStickyTaskListResponse, err error) {
		resetStickyAPICalled <- struct{}{}
		return &apiv1.ResetStickyTaskListResponse{}, nil
	}
	// pick 5 as cache size because it's not too big and not too small.
	cacheSize := 5
	internal.SetStickyWorkflowCacheSize(cacheSize)
	// once for workflow worker because we disable activity worker
	s.service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).Times(1)
	// feed our worker exactly *cacheSize* "legit" decision tasks
	// these are handcrafted decision tasks that are not blatantly obviously mocks
	// the goal is to trick our worker into thinking they are real so it
	// actually goes along with processing these and puts their execution in the cache.
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions...).DoAndReturn(mockPollForDecisionTask).Times(cacheSize)
	// after *cacheSize* "legit" tasks are fed to our worker, start feeding our worker empty responses.
	// these will get tossed away immediately after polled, but we still need them so gomock doesn't compain about unexpected calls.
	// this is because our worker's poller doesn't stop, it keeps polling on the service client as long
	// as Stop() is not called on the worker
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.PollForDecisionTaskResponse{}, nil).AnyTimes()
	// this gets called after polled decision tasks are processed, any number of times doesn't matter
	s.service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.RespondDecisionTaskCompletedResponse{}, nil).AnyTimes()
	// this is the critical point of the test.
	// ResetSticky should be called exactly once because our workflow cache evicts when full
	// so if our worker puts *cacheSize* entries in the cache, it should evict exactly one
	s.service.EXPECT().ResetStickyTaskList(gomock.Any(), gomock.Any(), callOptions...).DoAndReturn(mockResetStickyTaskList).Times(1)

	workflowWorker := internal.NewWorker(s.service, "test-domain", "tasklist", worker.Options{DisableActivityWorker: true})
	// this is an arbitrary workflow we use for this test
	// NOTE: a simple helloworld that doesn't execute an activity
	// won't work because the workflow will simply just complete
	// and won't stay in the cache.
	// for this test, we need a workflow that "blocks" either by
	// running an activity or waiting on a timer so that its execution
	// context sticks around in the cache.
	workflowWorker.RegisterWorkflow(testReplayWorkflow)

	workflowWorker.Start()

	testTimedOut := false
	select {
	case <-time.After(time.Second * 5):
		testTimedOut = true
	case <-resetStickyAPICalled:
		// success
	}

	workflowWorker.Stop()
	s.Equal(testTimedOut, false)
}
