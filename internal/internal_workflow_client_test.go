// Copyright (c) 2018 Uber Technologies, Inc.
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
	"log"
	"os"
	"testing"
	"time"

	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
	"go.uber.org/cadence/v2/internal/api"
	"go.uber.org/cadence/v2/internal/common/metrics"
	"go.uber.org/cadence/v2/internal/common/serializer"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"
)

const (
	domain                = "some random domain"
	workflowID            = "some random workflow ID"
	workflowType          = "some random workflow type"
	runID                 = "some random run ID"
	tasklist              = "some random tasklist"
	identity              = "some random identity"
	timeoutInSeconds      = 20
	workflowIDReusePolicy = WorkflowIDReusePolicyAllowDuplicateFailedOnly
	testHeader            = "test-header"
)

// historyEventIteratorSuite

type (
	historyEventIteratorSuite struct {
		suite.Suite
		mockCtrl              *gomock.Controller
		workflowServiceClient *api.MockInterface
		wfClient              *workflowClient
	}
)

// stringMapPropagator propagates the list of keys across a workflow,
// interpreting the payloads as strings.
type stringMapPropagator struct {
	keys map[string]struct{}
}

// NewStringMapPropagator returns a context propagator that propagates a set of
// string key-value pairs across a workflow
func NewStringMapPropagator(keys []string) ContextPropagator {
	keyMap := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		keyMap[key] = struct{}{}
	}
	return &stringMapPropagator{keyMap}
}

// Inject injects values from context into headers for propagation
func (s *stringMapPropagator) Inject(ctx context.Context, writer HeaderWriter) error {
	for key := range s.keys {
		value, ok := ctx.Value(contextKey(key)).(string)
		if !ok {
			return fmt.Errorf("unable to extract key from context %v", key)
		}
		writer.Set(key, []byte(value))
	}
	return nil
}

// InjectFromWorkflow injects values from context into headers for propagation
func (s *stringMapPropagator) InjectFromWorkflow(ctx Context, writer HeaderWriter) error {
	for key := range s.keys {
		value, ok := ctx.Value(contextKey(key)).(string)
		if !ok {
			return fmt.Errorf("unable to extract key from context %v", key)
		}
		writer.Set(key, []byte(value))
	}
	return nil
}

// Extract extracts values from headers and puts them into context
func (s *stringMapPropagator) Extract(ctx context.Context, reader HeaderReader) (context.Context, error) {
	if err := reader.ForEachKey(func(key string, value []byte) error {
		if _, ok := s.keys[key]; ok {
			ctx = context.WithValue(ctx, contextKey(key), string(value))
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return ctx, nil
}

// ExtractToWorkflow extracts values from headers and puts them into context
func (s *stringMapPropagator) ExtractToWorkflow(ctx Context, reader HeaderReader) (Context, error) {
	if err := reader.ForEachKey(func(key string, value []byte) error {
		if _, ok := s.keys[key]; ok {
			ctx = WithValue(ctx, contextKey(key), string(value))
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return ctx, nil
}

func TestHistoryEventIteratorSuite(t *testing.T) {
	s := new(historyEventIteratorSuite)
	suite.Run(t, s)
}

func (s *historyEventIteratorSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}
}

func (s *historyEventIteratorSuite) SetupTest() {
	// Create service endpoint
	s.mockCtrl = gomock.NewController(s.T())
	s.workflowServiceClient = api.NewMockInterface(s.mockCtrl)

	s.wfClient = &workflowClient{
		workflowService: s.workflowServiceClient,
		domain:          domain,
	}
}

func (s *historyEventIteratorSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

func (s *historyEventIteratorSuite) TestIterator_NoError() {
	filterType := apiv1.EventFilterType_EVENT_FILTER_TYPE_ALL_EVENT
	request1 := getGetWorkflowExecutionHistoryRequest(filterType)
	response1 := &apiv1.GetWorkflowExecutionHistoryResponse{
		History: &apiv1.History{
			Events: []*apiv1.HistoryEvent{
				// dummy history event
				&apiv1.HistoryEvent{},
			},
		},
		NextPageToken: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	request2 := getGetWorkflowExecutionHistoryRequest(filterType)
	request2.NextPageToken = response1.NextPageToken
	response2 := &apiv1.GetWorkflowExecutionHistoryResponse{
		History: &apiv1.History{
			Events: []*apiv1.HistoryEvent{
				// dummy history event
				&apiv1.HistoryEvent{},
			},
		},
		NextPageToken: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	dummyEvent := []*apiv1.HistoryEvent{
		// dummy history event
		&apiv1.HistoryEvent{},
	}

	blobData := serializeEvents(dummyEvent)
	request3 := getGetWorkflowExecutionHistoryRequest(filterType)
	request3.NextPageToken = response2.NextPageToken
	response3 := &apiv1.GetWorkflowExecutionHistoryResponse{
		RawHistory: []*apiv1.DataBlob{
			// dummy history event
			blobData,
		},
		NextPageToken: nil,
	}

	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), request1, gomock.Any()).Return(response1, nil).Times(1)
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), request2, gomock.Any()).Return(response2, nil).Times(1)
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), request3, gomock.Any()).Return(response3, nil).Times(1)

	events := []*apiv1.HistoryEvent{}
	iter := s.wfClient.GetWorkflowHistory(context.Background(), workflowID, runID, true, apiv1.EventFilterType_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		event, err := iter.Next()
		s.Nil(err)
		events = append(events, event)
	}
	s.Equal(3, len(events))
}

func (s *historyEventIteratorSuite) TestIterator_NoError_EmptyPage() {
	filterType := apiv1.EventFilterType_EVENT_FILTER_TYPE_ALL_EVENT
	request1 := getGetWorkflowExecutionHistoryRequest(filterType)
	response1 := &apiv1.GetWorkflowExecutionHistoryResponse{
		History: &apiv1.History{
			Events: []*apiv1.HistoryEvent{},
		},
		NextPageToken: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	request2 := getGetWorkflowExecutionHistoryRequest(filterType)
	request2.NextPageToken = response1.NextPageToken
	response2 := &apiv1.GetWorkflowExecutionHistoryResponse{
		History: &apiv1.History{
			Events: []*apiv1.HistoryEvent{
				// dummy history event
				&apiv1.HistoryEvent{},
			},
		},
		NextPageToken: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	dummyEvent := []*apiv1.HistoryEvent{
		// dummy history event
		&apiv1.HistoryEvent{},
	}

	blobData := serializeEvents(dummyEvent)
	request3 := getGetWorkflowExecutionHistoryRequest(filterType)
	request3.NextPageToken = response2.NextPageToken
	response3 := &apiv1.GetWorkflowExecutionHistoryResponse{
		RawHistory: []*apiv1.DataBlob{
			// dummy history event
			blobData,
		},
		NextPageToken: nil,
	}

	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), request1, gomock.Any()).Return(response1, nil).Times(1)
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), request2, gomock.Any()).Return(response2, nil).Times(1)
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), request3, gomock.Any()).Return(response3, nil).Times(1)

	events := []*apiv1.HistoryEvent{}
	iter := s.wfClient.GetWorkflowHistory(context.Background(), workflowID, runID, true, apiv1.EventFilterType_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		event, err := iter.Next()
		s.Nil(err)
		events = append(events, event)
	}
	s.Equal(2, len(events))
}

func (s *historyEventIteratorSuite) TestIterator_Error() {
	filterType := apiv1.EventFilterType_EVENT_FILTER_TYPE_ALL_EVENT
	request1 := getGetWorkflowExecutionHistoryRequest(filterType)
	response1 := &apiv1.GetWorkflowExecutionHistoryResponse{
		History: &apiv1.History{
			Events: []*apiv1.HistoryEvent{
				// dummy history event
				&apiv1.HistoryEvent{},
			},
		},
		NextPageToken: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}
	request2 := getGetWorkflowExecutionHistoryRequest(filterType)
	request2.NextPageToken = response1.NextPageToken

	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), request1, gomock.Any()).Return(response1, nil).Times(1)

	iter := s.wfClient.GetWorkflowHistory(context.Background(), workflowID, runID, true, apiv1.EventFilterType_EVENT_FILTER_TYPE_ALL_EVENT)

	s.True(iter.HasNext())
	event, err := iter.Next()
	s.NotNil(event)
	s.Nil(err)

	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), request2, gomock.Any()).Return(nil, &api.EntityNotExistsError{}).Times(1)

	s.True(iter.HasNext())
	event, err = iter.Next()
	s.Nil(event)
	s.NotNil(err)
}

// workflowRunSuite

type (
	workflowRunSuite struct {
		suite.Suite
		mockCtrl              *gomock.Controller
		workflowServiceClient *api.MockInterface
		workflowClient        Client
	}
)

func TestWorkflowRunSuite(t *testing.T) {
	s := new(workflowRunSuite)
	suite.Run(t, s)
}

func (s *workflowRunSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}
}

func (s *workflowRunSuite) TearDownSuite() {

}

func (s *workflowRunSuite) SetupTest() {
	// Create service endpoint
	s.mockCtrl = gomock.NewController(s.T())
	s.workflowServiceClient = api.NewMockInterface(s.mockCtrl)

	metricsScope := metrics.NewTaggedScope(nil)
	options := &ClientOptions{
		MetricsScope: metricsScope,
		Identity:     identity,
	}
	s.workflowClient = NewClient(s.workflowServiceClient, domain, options)
}

func (s *workflowRunSuite) TearDownTest() {
	s.mockCtrl.Finish()
}

func (s *workflowRunSuite) TestExecuteWorkflow_NoDup_Success() {
	createResponse := &apiv1.StartWorkflowExecutionResponse{
		RunId: runID,
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions...).Return(createResponse, nil).Times(1)

	filterType := apiv1.EventFilterType_EVENT_FILTER_TYPE_CLOSE_EVENT
	workflowResult := time.Hour * 59
	encodedResult, _ := encodeArg(getDefaultDataConverter(), workflowResult)
	getRequest := getGetWorkflowExecutionHistoryRequest(filterType)
	getResponse := &apiv1.GetWorkflowExecutionHistoryResponse{
		History: &apiv1.History{
			Events: []*apiv1.HistoryEvent{
				{
					Attributes: &apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes{
						WorkflowExecutionCompletedEventAttributes: &apiv1.WorkflowExecutionCompletedEventAttributes{
							Result: &apiv1.Payload{Data: encodedResult},
						},
					},
				},
			},
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest, callOptions...).Return(getResponse, nil).Times(1)

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.Nil(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.Nil(err)
	s.Equal(workflowResult, decodedResult)
}

func (s *workflowRunSuite) TestExecuteWorkflow_NoDup_RawHistory_Success() {
	createResponse := &apiv1.StartWorkflowExecutionResponse{
		RunId: runID,
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions...).Return(createResponse, nil).Times(1)

	filterType := apiv1.EventFilterType_EVENT_FILTER_TYPE_CLOSE_EVENT
	workflowResult := time.Hour * 59
	encodedResult, _ := encodeArg(getDefaultDataConverter(), workflowResult)
	events := []*apiv1.HistoryEvent{
		{
			Attributes: &apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes{
				WorkflowExecutionCompletedEventAttributes: &apiv1.WorkflowExecutionCompletedEventAttributes{
					Result: &apiv1.Payload{Data: encodedResult},
				},
			},
		},
	}

	blobData := serializeEvents(events)
	getRequest := getGetWorkflowExecutionHistoryRequest(filterType)
	getResponse := &apiv1.GetWorkflowExecutionHistoryResponse{
		RawHistory: []*apiv1.DataBlob{
			blobData,
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest, callOptions...).Return(getResponse, nil).Times(1)

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.Nil(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.Nil(err)
	s.Equal(workflowResult, decodedResult)
}

func (s *workflowRunSuite) TestExecuteWorkflowWorkflowExecutionAlreadyStartedError() {
	alreadyStartedErr := &api.WorkflowExecutionAlreadyStartedError{
		RunID:          runID,
		Message:        "Already Started",
		StartRequestID: uuid.NewRandom().String(),
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions...).
		Return(nil, alreadyStartedErr).Times(1)

	workflowResult := time.Hour * 59
	encodedResult, _ := encodeArg(nil, workflowResult)
	getResponse := &apiv1.GetWorkflowExecutionHistoryResponse{
		History: &apiv1.History{
			Events: []*apiv1.HistoryEvent{
				{
					Attributes: &apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes{
						WorkflowExecutionCompletedEventAttributes: &apiv1.WorkflowExecutionCompletedEventAttributes{
							Result: &apiv1.Payload{Data: encodedResult},
						},
					},
				},
			},
		},
		NextPageToken: nil,
	}
	getHistory := s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).
		Return(getResponse, nil).Times(1)
	getHistory.Do(func(ctx interface{}, getRequest *apiv1.GetWorkflowExecutionHistoryRequest, opt1 interface{}, opt2 interface{}, opt3 interface{}, opt4 interface{}) {
		workflowID := getRequest.WorkflowExecution.WorkflowId
		s.NotEmpty(workflowID)
	})

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.Nil(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.Nil(err)
	s.Equal(workflowResult, decodedResult)
}

func (s *workflowRunSuite) TestExecuteWorkflowWorkflowExecutionAlreadyStartedError_RawHistory() {
	alreadyStartedErr := &api.WorkflowExecutionAlreadyStartedError{
		RunID:          runID,
		Message:        "Already Started",
		StartRequestID: uuid.NewRandom().String(),
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions...).
		Return(nil, alreadyStartedErr).Times(1)

	workflowResult := time.Hour * 59
	encodedResult, _ := encodeArg(nil, workflowResult)
	events := []*apiv1.HistoryEvent{
		{
			Attributes: &apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes{
				WorkflowExecutionCompletedEventAttributes: &apiv1.WorkflowExecutionCompletedEventAttributes{
					Result: &apiv1.Payload{Data: encodedResult},
				},
			},
		},
	}

	blobData := serializeEvents(events)

	getResponse := &apiv1.GetWorkflowExecutionHistoryResponse{
		RawHistory: []*apiv1.DataBlob{
			blobData,
		},
		NextPageToken: nil,
	}
	getHistory := s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).
		Return(getResponse, nil).Times(1)
	getHistory.Do(func(ctx interface{}, getRequest *apiv1.GetWorkflowExecutionHistoryRequest, opt1 interface{}, opt2 interface{}, opt3 interface{}, opt4 interface{}) {
		workflowID := getRequest.WorkflowExecution.WorkflowId
		s.NotEmpty(workflowID)
	})

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.Nil(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.Nil(err)
	s.Equal(workflowResult, decodedResult)
}

// Test for the bug in ExecuteWorkflow.
// When Options.ID was empty then GetWorkflowExecutionHistory was called with an empty WorkflowID.
func (s *workflowRunSuite) TestExecuteWorkflow_NoIdInOptions() {
	createResponse := &apiv1.StartWorkflowExecutionResponse{
		RunId: runID,
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions...).Return(createResponse, nil).Times(1)

	workflowResult := time.Hour * 59
	encodedResult, _ := encodeArg(nil, workflowResult)
	getResponse := &apiv1.GetWorkflowExecutionHistoryResponse{
		History: &apiv1.History{
			Events: []*apiv1.HistoryEvent{
				{
					Attributes: &apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes{
						WorkflowExecutionCompletedEventAttributes: &apiv1.WorkflowExecutionCompletedEventAttributes{
							Result: &apiv1.Payload{Data: encodedResult},
						},
					},
				},
			},
		},
		NextPageToken: nil,
	}
	var wid string
	getHistory := s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).Return(getResponse, nil).Times(1)
	getHistory.Do(func(ctx interface{}, getRequest *apiv1.GetWorkflowExecutionHistoryRequest, opt1 interface{}, opt2 interface{}, opt3 interface{}, opt4 interface{}) {
		wid = getRequest.WorkflowExecution.WorkflowId
		s.NotEmpty(wid)
	})

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.Nil(err)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.Nil(err)
	s.Equal(workflowResult, decodedResult)
	s.Equal(workflowRun.GetID(), wid)
}

// Test for the bug in ExecuteWorkflow in the case of raw history returned from API.
// When Options.ID was empty then GetWorkflowExecutionHistory was called with an empty WorkflowID.
func (s *workflowRunSuite) TestExecuteWorkflow_NoIdInOptions_RawHistory() {
	createResponse := &apiv1.StartWorkflowExecutionResponse{
		RunId: runID,
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions...).Return(createResponse, nil).Times(1)

	workflowResult := time.Hour * 59
	encodedResult, _ := encodeArg(nil, workflowResult)
	events := []*apiv1.HistoryEvent{
		{
			Attributes: &apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes{
				WorkflowExecutionCompletedEventAttributes: &apiv1.WorkflowExecutionCompletedEventAttributes{
					Result: &apiv1.Payload{Data: encodedResult},
				},
			},
		},
	}

	blobData := serializeEvents(events)
	getResponse := &apiv1.GetWorkflowExecutionHistoryResponse{
		RawHistory: []*apiv1.DataBlob{
			blobData,
		},
		NextPageToken: nil,
	}
	var wid string
	getHistory := s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).Return(getResponse, nil).Times(1)
	getHistory.Do(func(ctx interface{}, getRequest *apiv1.GetWorkflowExecutionHistoryRequest, opt1 interface{}, opt2 interface{}, opt3 interface{}, opt4 interface{}) {
		wid = getRequest.WorkflowExecution.WorkflowId
		s.NotEmpty(wid)
	})

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.Nil(err)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.Nil(err)
	s.Equal(workflowResult, decodedResult)
	s.Equal(workflowRun.GetID(), wid)
}

func (s *workflowRunSuite) TestExecuteWorkflow_NoDup_Cancelled() {
	createResponse := &apiv1.StartWorkflowExecutionResponse{
		RunId: runID,
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions...).Return(createResponse, nil).Times(1)

	filterType := apiv1.EventFilterType_EVENT_FILTER_TYPE_CLOSE_EVENT
	details := "some details"
	encodedDetails, _ := encodeArg(getDefaultDataConverter(), details)
	getRequest := getGetWorkflowExecutionHistoryRequest(filterType)
	getResponse := &apiv1.GetWorkflowExecutionHistoryResponse{
		History: &apiv1.History{
			Events: []*apiv1.HistoryEvent{
				{
					Attributes: &apiv1.HistoryEvent_WorkflowExecutionCanceledEventAttributes{
						WorkflowExecutionCanceledEventAttributes: &apiv1.WorkflowExecutionCanceledEventAttributes{
							Details: &apiv1.Payload{Data: encodedDetails},
						},
					},
				},
			},
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest, callOptions...).Return(getResponse, nil).Times(1)

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.Nil(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.NotNil(err)
	_, ok := err.(*CanceledError)
	s.True(ok)
	s.Equal(time.Minute, decodedResult)
}

func (s *workflowRunSuite) TestExecuteWorkflow_NoDup_Failed() {
	createResponse := &apiv1.StartWorkflowExecutionResponse{
		RunId: runID,
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions...).Return(createResponse, nil).Times(1)

	filterType := apiv1.EventFilterType_EVENT_FILTER_TYPE_CLOSE_EVENT
	reason := "some reason"
	details := "some details"
	dataConverter := getDefaultDataConverter()
	encodedDetails, _ := encodeArg(dataConverter, details)
	getRequest := getGetWorkflowExecutionHistoryRequest(filterType)
	getResponse := &apiv1.GetWorkflowExecutionHistoryResponse{
		History: &apiv1.History{
			Events: []*apiv1.HistoryEvent{
				{
					Attributes: &apiv1.HistoryEvent_WorkflowExecutionFailedEventAttributes{
						WorkflowExecutionFailedEventAttributes: &apiv1.WorkflowExecutionFailedEventAttributes{
							Failure: &apiv1.Failure{
								Reason: reason,
								Details: encodedDetails,
							},
						},
					},
				},
			},
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest, callOptions...).Return(getResponse, nil).Times(1)

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.Nil(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.Equal(constructError(reason, encodedDetails, dataConverter), err)
	s.Equal(time.Minute, decodedResult)
}

func (s *workflowRunSuite) TestExecuteWorkflow_NoDup_Terminated() {
	createResponse := &apiv1.StartWorkflowExecutionResponse{
		RunId: runID,
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions...).Return(createResponse, nil).Times(1)

	filterType := apiv1.EventFilterType_EVENT_FILTER_TYPE_CLOSE_EVENT
	getRequest := getGetWorkflowExecutionHistoryRequest(filterType)
	getResponse := &apiv1.GetWorkflowExecutionHistoryResponse{
		History: &apiv1.History{
			Events: []*apiv1.HistoryEvent{
				{
					Attributes: &apiv1.HistoryEvent_WorkflowExecutionTerminatedEventAttributes{
						WorkflowExecutionTerminatedEventAttributes: &apiv1.WorkflowExecutionTerminatedEventAttributes{},
					},
				},
			},
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest, callOptions...).Return(getResponse, nil).Times(1)

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.Nil(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.Equal(newTerminatedError(), err)
	s.Equal(time.Minute, decodedResult)
}

func (s *workflowRunSuite) TestExecuteWorkflow_NoDup_TimedOut() {
	createResponse := &apiv1.StartWorkflowExecutionResponse{
		RunId: runID,
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions...).Return(createResponse, nil).Times(1)

	filterType := apiv1.EventFilterType_EVENT_FILTER_TYPE_CLOSE_EVENT
	timeType := apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_START
	getRequest := getGetWorkflowExecutionHistoryRequest(filterType)
	getResponse := &apiv1.GetWorkflowExecutionHistoryResponse{
		History: &apiv1.History{
			Events: []*apiv1.HistoryEvent{
				{
					Attributes: &apiv1.HistoryEvent_WorkflowExecutionTimedOutEventAttributes{
						WorkflowExecutionTimedOutEventAttributes: &apiv1.WorkflowExecutionTimedOutEventAttributes{
							TimeoutType: timeType,
						},
					},
				},
			},
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest, callOptions...).Return(getResponse, nil).Times(1)

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.Nil(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.NotNil(err)
	_, ok := err.(*TimeoutError)
	s.True(ok)
	s.Equal(timeType, err.(*TimeoutError).TimeoutType())
	s.Equal(time.Minute, decodedResult)
}

func (s *workflowRunSuite) TestExecuteWorkflow_NoDup_ContinueAsNew() {
	createResponse := &apiv1.StartWorkflowExecutionResponse{
		RunId: runID,
	}
	s.workflowServiceClient.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), callOptions...).Return(createResponse, nil).Times(1)

	newRunID := "some other random run ID"
	filterType := apiv1.EventFilterType_EVENT_FILTER_TYPE_CLOSE_EVENT
	getRequest1 := getGetWorkflowExecutionHistoryRequest(filterType)
	getResponse1 := &apiv1.GetWorkflowExecutionHistoryResponse{
		History: &apiv1.History{
			Events: []*apiv1.HistoryEvent{
				{
					Attributes: &apiv1.HistoryEvent_WorkflowExecutionContinuedAsNewEventAttributes{
						WorkflowExecutionContinuedAsNewEventAttributes: &apiv1.WorkflowExecutionContinuedAsNewEventAttributes{
							NewExecutionRunId: newRunID,
						},
					},
				},
			},
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest1, callOptions...).Return(getResponse1, nil).Times(1)

	workflowResult := time.Hour * 59
	encodedResult, _ := encodeArg(getDefaultDataConverter(), workflowResult)
	getRequest2 := getGetWorkflowExecutionHistoryRequest(filterType)
	getRequest2.WorkflowExecution.RunId = newRunID
	getResponse2 := &apiv1.GetWorkflowExecutionHistoryResponse{
		History: &apiv1.History{
			Events: []*apiv1.HistoryEvent{
				{
					Attributes: &apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes{
						WorkflowExecutionCompletedEventAttributes: &apiv1.WorkflowExecutionCompletedEventAttributes{
							Result: &apiv1.Payload{Data: encodedResult},
						},
					},
				},
			},
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest2, callOptions...).Return(getResponse2, nil).Times(1)

	workflowRun, err := s.workflowClient.ExecuteWorkflow(
		context.Background(),
		StartWorkflowOptions{
			ID:                              workflowID,
			TaskList:                        tasklist,
			ExecutionStartToCloseTimeout:    timeoutInSeconds * time.Second,
			DecisionTaskStartToCloseTimeout: timeoutInSeconds * time.Second,
			WorkflowIDReusePolicy:           workflowIDReusePolicy,
		}, workflowType,
	)
	s.Nil(err)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err = workflowRun.Get(context.Background(), &decodedResult)
	s.Nil(err)
	s.Equal(workflowResult, decodedResult)
}

func (s *workflowRunSuite) TestGetWorkflow() {
	filterType := apiv1.EventFilterType_EVENT_FILTER_TYPE_CLOSE_EVENT
	workflowResult := time.Hour * 59
	encodedResult, _ := encodeArg(getDefaultDataConverter(), workflowResult)
	getRequest := getGetWorkflowExecutionHistoryRequest(filterType)
	getResponse := &apiv1.GetWorkflowExecutionHistoryResponse{
		History: &apiv1.History{
			Events: []*apiv1.HistoryEvent{
				{
					Attributes: &apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes{
						WorkflowExecutionCompletedEventAttributes: &apiv1.WorkflowExecutionCompletedEventAttributes{
							Result: &apiv1.Payload{Data: encodedResult},
						},
					},
				},
			},
		},
		NextPageToken: nil,
	}
	s.workflowServiceClient.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), getRequest, callOptions...).Return(getResponse, nil).Times(1)

	workflowID := workflowID
	runID := runID

	workflowRun := s.workflowClient.GetWorkflow(
		context.Background(),
		workflowID,
		runID,
	)
	s.Equal(workflowRun.GetID(), workflowID)
	s.Equal(workflowRun.GetRunID(), runID)
	decodedResult := time.Minute
	err := workflowRun.Get(context.Background(), &decodedResult)
	s.Nil(err)
	s.Equal(workflowResult, decodedResult)
}

func getGetWorkflowExecutionHistoryRequest(filterType apiv1.EventFilterType) *apiv1.GetWorkflowExecutionHistoryRequest {
	isLongPoll := true

	request := &apiv1.GetWorkflowExecutionHistoryRequest{
		Domain: domain,
		WorkflowExecution: &apiv1.WorkflowExecution{
			WorkflowId: workflowID,
			RunId:      runID,
		},
		WaitForNewEvent:        isLongPoll,
		HistoryEventFilterType: filterType,
		SkipArchival:           true,
	}

	return request
}

// workflow client test suite
type (
	workflowClientTestSuite struct {
		suite.Suite
		mockCtrl *gomock.Controller
		service  *api.MockInterface
		client   Client
	}
)

func TestWorkflowClientSuite(t *testing.T) {
	suite.Run(t, new(workflowClientTestSuite))
}

func (s *workflowClientTestSuite) SetupSuite() {
	if testing.Verbose() {
		log.SetOutput(os.Stdout)
	}
}

func (s *workflowClientTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.service = api.NewMockInterface(s.mockCtrl)
	s.client = NewClient(s.service, domain, nil)
}

func (s *workflowClientTestSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

func (s *workflowClientTestSuite) TestSignalWithStartWorkflow() {
	signalName := "my signal"
	signalInput := []byte("my signal input")
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}

	createResponse := &apiv1.SignalWithStartWorkflowExecutionResponse{
		RunId: runID,
	}
	s.service.EXPECT().SignalWithStartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(createResponse, nil).Times(2)

	resp, err := s.client.SignalWithStartWorkflow(context.Background(), workflowID, signalName, signalInput,
		options, workflowType)
	s.Nil(err)
	s.Equal(createResponse.GetRunId(), resp.RunID)

	resp, err = s.client.SignalWithStartWorkflow(context.Background(), "", signalName, signalInput,
		options, workflowType)
	s.Nil(err)
	s.Equal(createResponse.GetRunId(), resp.RunID)
}

func (s *workflowClientTestSuite) TestSignalWithStartWorkflow_Error() {
	signalName := "my signal"
	signalInput := []byte("my signal input")
	options := StartWorkflowOptions{}

	resp, err := s.client.SignalWithStartWorkflow(context.Background(), workflowID, signalName, signalInput,
		options, workflowType)
	s.Equal(errors.New("missing TaskList"), err)
	s.Nil(resp)

	options.TaskList = tasklist
	resp, err = s.client.SignalWithStartWorkflow(context.Background(), workflowID, signalName, signalInput,
		options, workflowType)
	s.NotNil(err)
	s.Nil(resp)

	options.ExecutionStartToCloseTimeout = timeoutInSeconds
	createResponse := &apiv1.SignalWithStartWorkflowExecutionResponse{
		RunId: runID,
	}
	s.service.EXPECT().SignalWithStartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(createResponse, nil)
	resp, err = s.client.SignalWithStartWorkflow(context.Background(), workflowID, signalName, signalInput,
		options, workflowType)
	s.Nil(err)
	s.Equal(createResponse.GetRunId(), resp.RunID)
}

func (s *workflowClientTestSuite) TestStartWorkflow() {
	client, ok := s.client.(*workflowClient)
	s.True(ok)
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	f1 := func(ctx Context, r []byte) string {
		return "result"
	}

	createResponse := &apiv1.StartWorkflowExecutionResponse{
		RunId: runID,
	}
	s.service.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(createResponse, nil)

	resp, err := client.StartWorkflow(context.Background(), options, f1, []byte("test"))
	s.Equal(getDefaultDataConverter(), client.dataConverter)
	s.Nil(err)
	s.Equal(createResponse.GetRunId(), resp.RunID)
}

func (s *workflowClientTestSuite) TestStartWorkflow_WithContext() {
	s.client = NewClient(s.service, domain, &ClientOptions{ContextPropagators: []ContextPropagator{NewStringMapPropagator([]string{testHeader})}})
	client, ok := s.client.(*workflowClient)
	s.True(ok)
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	f1 := func(ctx Context, r []byte) error {
		value := ctx.Value(contextKey(testHeader))
		if val, ok := value.([]byte); ok {
			s.Equal("test-data", string(val))
			return nil
		}
		return fmt.Errorf("context did not propagate to workflow")
	}

	createResponse := &apiv1.StartWorkflowExecutionResponse{
		RunId: runID,
	}
	s.service.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(createResponse, nil)

	resp, err := client.StartWorkflow(context.Background(), options, f1, []byte("test"))
	s.Nil(err)
	s.Equal(createResponse.GetRunId(), resp.RunID)
}

func (s *workflowClientTestSuite) TestStartWorkflow_WithDataConverter() {
	dc := newTestDataConverter()
	s.client = NewClient(s.service, domain, &ClientOptions{DataConverter: dc})
	client, ok := s.client.(*workflowClient)
	s.True(ok)
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
	}
	f1 := func(ctx Context, r []byte) string {
		return "result"
	}
	input := []byte("test")

	createResponse := &apiv1.StartWorkflowExecutionResponse{
		RunId: runID,
	}
	s.service.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(createResponse, nil).
		Do(func(_ interface{}, req *apiv1.StartWorkflowExecutionRequest, _ ...interface{}) {
			dc := client.dataConverter
			encodedArg, _ := dc.ToData(input)
			s.Equal(req.Input.Data, encodedArg)
			var decodedArg []byte
			dc.FromData(req.Input.Data, &decodedArg)
			s.Equal(input, decodedArg)
		})

	resp, err := client.StartWorkflow(context.Background(), options, f1, input)
	s.Equal(newTestDataConverter(), client.dataConverter)
	s.Nil(err)
	s.Equal(createResponse.GetRunId(), resp.RunID)
}

func (s *workflowClientTestSuite) TestStartWorkflow_WithMemoAndSearchAttr() {
	memo := map[string]interface{}{
		"testMemo": "memo value",
	}
	searchAttributes := map[string]interface{}{
		"testAttr": "attr value",
	}
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
		Memo:                            memo,
		SearchAttributes:                searchAttributes,
	}
	wf := func(ctx Context) string {
		return "result"
	}
	startResp := &apiv1.StartWorkflowExecutionResponse{}

	s.service.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(startResp, nil).
		Do(func(_ interface{}, req *apiv1.StartWorkflowExecutionRequest, _ ...interface{}) {
			var resultMemo, resultAttr string
			err := json.Unmarshal(req.Memo.Fields["testMemo"].Data, &resultMemo)
			s.NoError(err)
			s.Equal("memo value", resultMemo)

			err = json.Unmarshal(req.SearchAttributes.IndexedFields["testAttr"].Data, &resultAttr)
			s.NoError(err)
			s.Equal("attr value", resultAttr)
		})
	s.client.StartWorkflow(context.Background(), options, wf)
}

func (s *workflowClientTestSuite) TestSignalWithStartWorkflow_WithMemoAndSearchAttr() {
	memo := map[string]interface{}{
		"testMemo": "memo value",
	}
	searchAttributes := map[string]interface{}{
		"testAttr": "attr value",
	}
	options := StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        tasklist,
		ExecutionStartToCloseTimeout:    timeoutInSeconds,
		DecisionTaskStartToCloseTimeout: timeoutInSeconds,
		Memo:                            memo,
		SearchAttributes:                searchAttributes,
	}
	wf := func(ctx Context) string {
		return "result"
	}
	startResp := &apiv1.SignalWithStartWorkflowExecutionResponse{}

	s.service.EXPECT().SignalWithStartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(startResp, nil).
		Do(func(_ interface{}, req *apiv1.SignalWithStartWorkflowExecutionRequest, _ ...interface{}) {
			var resultMemo, resultAttr string
			err := json.Unmarshal(req.StartRequest.Memo.Fields["testMemo"].Data, &resultMemo)
			s.NoError(err)
			s.Equal("memo value", resultMemo)

			err = json.Unmarshal(req.StartRequest.SearchAttributes.IndexedFields["testAttr"].Data, &resultAttr)
			s.NoError(err)
			s.Equal("attr value", resultAttr)
		})
	s.client.SignalWithStartWorkflow(context.Background(), "wid", "signal", "value", options, wf)
}

func (s *workflowClientTestSuite) TestGetWorkflowMemo() {
	var input1 map[string]interface{}
	result1, err := getWorkflowMemo(input1, nil)
	s.NoError(err)
	s.Nil(result1)

	input1 = make(map[string]interface{})
	result2, err := getWorkflowMemo(input1, nil)
	s.NoError(err)
	s.NotNil(result2)
	s.Equal(0, len(result2.Fields))

	input1["t1"] = "v1"
	result3, err := getWorkflowMemo(input1, nil)
	s.NoError(err)
	s.NotNil(result3)
	s.Equal(1, len(result3.Fields))
	var resultString string
	decodeArg(nil, result3.Fields["t1"].GetData(), &resultString)
	s.Equal("v1", resultString)

	input1["non-serializable"] = make(chan int)
	_, err = getWorkflowMemo(input1, nil)
	s.Error(err)
}

func (s *workflowClientTestSuite) TestSerializeSearchAttributes() {
	var input1 map[string]interface{}
	result1, err := serializeSearchAttributes(input1)
	s.NoError(err)
	s.Nil(result1)

	input1 = make(map[string]interface{})
	result2, err := serializeSearchAttributes(input1)
	s.NoError(err)
	s.NotNil(result2)
	s.Equal(0, len(result2.IndexedFields))

	input1["t1"] = "v1"
	result3, err := serializeSearchAttributes(input1)
	s.NoError(err)
	s.NotNil(result3)
	s.Equal(1, len(result3.IndexedFields))
	var resultString string
	decodeArg(nil, result3.IndexedFields["t1"].GetData(), &resultString)
	s.Equal("v1", resultString)

	input1["non-serializable"] = make(chan int)
	_, err = serializeSearchAttributes(input1)
	s.Error(err)
}

func (s *workflowClientTestSuite) TestListWorkflow() {
	request := &apiv1.ListWorkflowExecutionsRequest{}
	response := &apiv1.ListWorkflowExecutionsResponse{}
	s.service.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(response, nil).
		Do(func(_ interface{}, req *apiv1.ListWorkflowExecutionsRequest, _ ...interface{}) {
			s.Equal(domain, request.GetDomain())
		})
	resp, err := s.client.ListWorkflow(context.Background(), request)
	s.Nil(err)
	s.Equal(response, resp)

	responseErr := &api.BadRequestError{}
	request.Domain = "another"
	s.service.EXPECT().ListWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, responseErr).
		Do(func(_ interface{}, req *apiv1.ListWorkflowExecutionsRequest, _ ...interface{}) {
			s.Equal("another", request.GetDomain())
		})
	resp, err = s.client.ListWorkflow(context.Background(), request)
	s.Equal(responseErr, err)
}

func (s *workflowClientTestSuite) TestListArchivedWorkflow() {
	request := &apiv1.ListArchivedWorkflowExecutionsRequest{}
	response := &apiv1.ListArchivedWorkflowExecutionsResponse{}
	s.service.EXPECT().ListArchivedWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(response, nil).
		Do(func(_ interface{}, req *apiv1.ListArchivedWorkflowExecutionsRequest, _ ...interface{}) {
			s.Equal(domain, request.GetDomain())
		})
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	resp, err := s.client.ListArchivedWorkflow(ctxWithTimeout, request)
	s.Nil(err)
	s.Equal(response, resp)

	responseErr := &api.BadRequestError{}
	request.Domain = "another"
	s.service.EXPECT().ListArchivedWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, responseErr).
		Do(func(_ interface{}, req *apiv1.ListArchivedWorkflowExecutionsRequest, _ ...interface{}) {
			s.Equal("another", request.GetDomain())
		})
	resp, err = s.client.ListArchivedWorkflow(ctxWithTimeout, request)
	s.Equal(responseErr, err)
}

func (s *workflowClientTestSuite) TestScanWorkflow() {
	request := &apiv1.ScanWorkflowExecutionsRequest{}
	response := &apiv1.ScanWorkflowExecutionsResponse{}
	s.service.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(response, nil).
		Do(func(_ interface{}, req *apiv1.ScanWorkflowExecutionsRequest, _ ...interface{}) {
			s.Equal(domain, request.GetDomain())
		})
	resp, err := s.client.ScanWorkflow(context.Background(), request)
	s.Nil(err)
	s.Equal(response, resp)

	responseErr := &api.BadRequestError{}
	request.Domain = "another"
	s.service.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, responseErr).
		Do(func(_ interface{}, req *apiv1.ScanWorkflowExecutionsRequest, _ ...interface{}) {
			s.Equal("another", request.GetDomain())
		})
	resp, err = s.client.ScanWorkflow(context.Background(), request)
	s.Equal(responseErr, err)
}

func (s *workflowClientTestSuite) TestCountWorkflow() {
	request := &apiv1.CountWorkflowExecutionsRequest{}
	response := &apiv1.CountWorkflowExecutionsResponse{}
	s.service.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(response, nil).
		Do(func(_ interface{}, req *apiv1.CountWorkflowExecutionsRequest, _ ...interface{}) {
			s.Equal(domain, request.GetDomain())
		})
	resp, err := s.client.CountWorkflow(context.Background(), request)
	s.Nil(err)
	s.Equal(response, resp)

	responseErr := &api.BadRequestError{}
	request.Domain = "another"
	s.service.EXPECT().CountWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, responseErr).
		Do(func(_ interface{}, req *apiv1.CountWorkflowExecutionsRequest, _ ...interface{}) {
			s.Equal("another", request.GetDomain())
		})
	resp, err = s.client.CountWorkflow(context.Background(), request)
	s.Equal(responseErr, err)
}

func (s *workflowClientTestSuite) TestGetSearchAttributes() {
	response := &apiv1.GetSearchAttributesResponse{}
	s.service.EXPECT().GetSearchAttributes(gomock.Any(), gomock.Any(), gomock.Any()).Return(response, nil)
	resp, err := s.client.GetSearchAttributes(context.Background())
	s.Nil(err)
	s.Equal(response, resp)

	responseErr := &api.BadRequestError{}
	s.service.EXPECT().GetSearchAttributes(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, responseErr)
	resp, err = s.client.GetSearchAttributes(context.Background())
	s.Equal(responseErr, err)
}

func serializeEvents(events []*apiv1.HistoryEvent) *apiv1.DataBlob {

	blob, _ := serializer.SerializeBatchEvents(events, apiv1.EncodingType_ENCODING_TYPE_PROTO3)

	return &apiv1.DataBlob{
		EncodingType: apiv1.EncodingType_ENCODING_TYPE_PROTO3,
		Data:         blob.Data,
	}
}
