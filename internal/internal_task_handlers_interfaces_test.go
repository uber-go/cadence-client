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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/uber/tchannel-go/thrift"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
	"go.uber.org/cadence/v2/internal/api"
	"golang.org/x/net/context"
)

type (
	PollLayerInterfacesTestSuite struct {
		suite.Suite
		mockCtrl *gomock.Controller
		service *api.MockInterface
	}
)

// Sample Workflow task handler
type sampleWorkflowTaskHandler struct {
}

func (wth sampleWorkflowTaskHandler) ProcessWorkflowTask(
	workflowTask *workflowTask,
	d decisionHeartbeatFunc,
) (interface{}, error) {
	return &apiv1.RespondDecisionTaskCompletedRequest{
		TaskToken: workflowTask.task.TaskToken,
	}, nil
}

func newSampleWorkflowTaskHandler() *sampleWorkflowTaskHandler {
	return &sampleWorkflowTaskHandler{}
}

// Sample ActivityTaskHandler
type sampleActivityTaskHandler struct {
}

func newSampleActivityTaskHandler() *sampleActivityTaskHandler {
	return &sampleActivityTaskHandler{}
}

func (ath sampleActivityTaskHandler) Execute(taskList string, task *apiv1.PollForActivityTaskResponse) (interface{}, error) {
	activityImplementation := &greeterActivity{}
	result, err := activityImplementation.Execute(context.Background(), task.Input.GetData())
	if err != nil {
		reason := err.Error()
		return &apiv1.RespondActivityTaskFailedRequest{
			TaskToken: task.TaskToken,
			Failure: &apiv1.Failure{
				Reason: reason,
			},
		}, nil
	}
	return &apiv1.RespondActivityTaskCompletedRequest{
		TaskToken: task.TaskToken,
		Result:    &apiv1.Payload{Data: result},
	}, nil
}

// Test suite.
func TestPollLayerInterfacesTestSuite(t *testing.T) {
	suite.Run(t, new(PollLayerInterfacesTestSuite))
}

func (s *PollLayerInterfacesTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.service = api.NewMockInterface(s.mockCtrl)
}

func (s *PollLayerInterfacesTestSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

func (s *PollLayerInterfacesTestSuite) TestProcessWorkflowTaskInterface() {
	ctx, cancel := thrift.NewContext(10)
	defer cancel()

	// mocks
	s.service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any()).Return(&apiv1.PollForDecisionTaskResponse{}, nil)
	s.service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any()).Return(nil, nil)

	response, err := s.service.PollForDecisionTask(ctx, &apiv1.PollForDecisionTaskRequest{})
	s.NoError(err)

	// Process task and respond to the service.
	taskHandler := newSampleWorkflowTaskHandler()
	request, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: response}, nil)
	completionRequest := request.(*apiv1.RespondDecisionTaskCompletedRequest)
	s.NoError(err)

	_, err = s.service.RespondDecisionTaskCompleted(ctx, completionRequest)
	s.NoError(err)
}

func (s *PollLayerInterfacesTestSuite) TestProcessActivityTaskInterface() {
	ctx, cancel := thrift.NewContext(10)
	defer cancel()

	// mocks
	s.service.EXPECT().PollForActivityTask(gomock.Any(), gomock.Any()).Return(&apiv1.PollForActivityTaskResponse{}, nil)
	s.service.EXPECT().RespondActivityTaskCompleted(gomock.Any(), gomock.Any()).Return(nil, nil)

	response, err := s.service.PollForActivityTask(ctx, &apiv1.PollForActivityTaskRequest{})
	s.NoError(err)

	// Execute activity task and respond to the service.
	taskHandler := newSampleActivityTaskHandler()
	request, err := taskHandler.Execute(tasklist, response)
	s.NoError(err)
	switch request.(type) {
	case *apiv1.RespondActivityTaskCompletedRequest:
		_, err = s.service.RespondActivityTaskCompleted(ctx, request.(*apiv1.RespondActivityTaskCompletedRequest))
		s.NoError(err)
	case *apiv1.RespondActivityTaskFailedRequest: // shouldn't happen
		_, err = s.service.RespondActivityTaskFailed(ctx, request.(*apiv1.RespondActivityTaskFailedRequest))
		s.NoError(err)
	}
}

func (s *PollLayerInterfacesTestSuite) TestGetNextDecisions() {
	// Schedule an activity and see if we complete workflow.
	taskList := "tl1"
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(3),
		{
			EventId:   4,
			Attributes: &apiv1.HistoryEvent_DecisionTaskFailedEventAttributes{
				DecisionTaskFailedEventAttributes: &apiv1.DecisionTaskFailedEventAttributes{},
			},
		},
		{
			EventId:   5,
			Attributes: &apiv1.HistoryEvent_WorkflowExecutionSignaledEventAttributes{
				WorkflowExecutionSignaledEventAttributes: &apiv1.WorkflowExecutionSignaledEventAttributes{},
			},
		},
		createTestEventDecisionTaskScheduled(6, &apiv1.DecisionTaskScheduledEventAttributes{TaskList: &apiv1.TaskList{Name: taskList}}),
		createTestEventDecisionTaskStarted(7),
	}
	task := createWorkflowTask(testEvents[0:3], 0, "HelloWorld_Workflow")

	historyIterator := &historyIteratorImpl{
		iteratorFunc: func(nextToken []byte) (*apiv1.History, []byte, error) {
			return &apiv1.History{
				Events: testEvents[3:],
			}, nil, nil
		},
		nextPageToken: []byte("test"),
	}

	workflowTask := &workflowTask{task: task, historyIterator: historyIterator}

	eh := newHistory(workflowTask, nil)

	events, _, _, err := eh.NextDecisionEvents()

	s.NoError(err)
	s.Equal(3, len(events))
	s.NotNil(events[1].GetWorkflowExecutionSignaledEventAttributes())
	s.NotNil(events[2].GetDecisionTaskStartedEventAttributes())
	s.Equal(int64(7), events[2].GetEventId())
}
