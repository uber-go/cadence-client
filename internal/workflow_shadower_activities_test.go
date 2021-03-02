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
	"math/rand"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/zap/zaptest"
)

type workflowShadowerActivitiesSuite struct {
	*require.Assertions
	suite.Suite
	WorkflowTestSuite

	controller  *gomock.Controller
	mockService *workflowservicetest.MockClient

	env                 *TestActivityEnvironment
	testReplayer        *WorkflowReplayer
	testWorkflowHistory *shared.History
}

func TestWorkflowShadowerActivitiesSuite(t *testing.T) {
	s := new(workflowShadowerActivitiesSuite)
	suite.Run(t, s)
}

func (s *workflowShadowerActivitiesSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockService = workflowservicetest.NewMockClient(s.controller)

	s.env = s.NewTestActivityEnvironment()
	s.testReplayer = NewWorkflowReplayer()
	s.testReplayer.RegisterWorkflow(testReplayWorkflow)
	s.testWorkflowHistory = getTestReplayWorkflowFullHistory()

	activityContext := context.Background()
	activityContext = context.WithValue(activityContext, serviceClientContextKey, s.mockService)
	activityContext = context.WithValue(activityContext, workflowReplayerContextKey, s.testReplayer)
	s.env.SetWorkerOptions(WorkerOptions{
		BackgroundActivityContext: activityContext,
		Logger:                    zaptest.NewLogger(s.T()),
	})
	s.env.RegisterActivityWithOptions(scanWorkflowActivity, RegisterActivityOptions{
		Name: scanWorkflowActivityName,
	})
	s.env.RegisterActivityWithOptions(replayWorkflowExecutionActivity, RegisterActivityOptions{
		Name: replayWorkflowExecutionActivityName,
	})
}

func (s *workflowShadowerActivitiesSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *workflowShadowerActivitiesSuite) TestScanWorkflowActivity_Succeed() {
	numExecutions := 1000
	s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions...).Return(&shared.ListWorkflowExecutionsResponse{
		Executions:    newTestWorkflowExecutions(numExecutions),
		NextPageToken: []byte{1, 2, 3},
	}, nil).Times(1)

	params := scanWorkflowActivityParams{
		Domain:        defaultTestDomain,
		WorkflowQuery: "some random workflow visibility query",
		SamplingRate:  0.5,
	}

	resultValue, err := s.env.ExecuteActivity(scanWorkflowActivityName, params)
	s.NoError(err)

	var result scanWorkflowActivityResult
	s.NoError(resultValue.Get(&result))
	s.NotNil(result.NextPageToken)
	s.True(len(result.Executions) > 0)
	s.True(len(result.Executions) < numExecutions)
}

func (s *workflowShadowerActivitiesSuite) TestScanWorkflowActivity_InvalidQuery() {
	s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions...).Return(nil, &shared.BadRequestError{
		Message: "invalid query",
	}).Times(1)

	params := scanWorkflowActivityParams{
		Domain:        defaultTestDomain,
		WorkflowQuery: "invalid workflow visibility query",
	}

	_, err := s.env.ExecuteActivity(scanWorkflowActivityName, params)
	s.Error(err)
	s.Equal(errMsgInvalidQuery, err.Error())
}

func (s *workflowShadowerActivitiesSuite) TestReplayWorkflowExecutionActivity_NoPreviousProgress() {
	numExecutions := 10
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).Return(&shared.GetWorkflowExecutionHistoryResponse{
		History: s.testWorkflowHistory,
	}, nil).Times(numExecutions)

	params := replayWorkflowActivityParams{
		Domain:     defaultTestDomain,
		Executions: make([]WorkflowExecution, numExecutions),
	}

	resultValue, err := s.env.ExecuteActivity(replayWorkflowExecutionActivityName, params)
	s.NoError(err)

	var result replayWorkflowActivityResult
	s.NoError(resultValue.Get(&result))
	s.Equal(numExecutions, result.Succeed)
	s.Equal(0, result.Failed)
}

func (s *workflowShadowerActivitiesSuite) TestReplayWorkflowExecutionActivity_WithPreviousProgress() {
	progress := replayWorkflowActivityProgress{
		Result: replayWorkflowActivityResult{
			Succeed: 3,
			Failed:  0,
		},
		NextExecutionIdx: 5,
	}
	s.env.SetHeartbeatDetails(progress)

	numExecutions := 10
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).Return(&shared.GetWorkflowExecutionHistoryResponse{
		History: s.testWorkflowHistory,
	}, nil).Times(numExecutions - progress.NextExecutionIdx)

	params := replayWorkflowActivityParams{
		Domain:     defaultTestDomain,
		Executions: make([]WorkflowExecution, numExecutions),
	}

	resultValue, err := s.env.ExecuteActivity(replayWorkflowExecutionActivityName, params)
	s.NoError(err)

	var result replayWorkflowActivityResult
	s.NoError(resultValue.Get(&result))
	s.Equal(progress.Result.Succeed+numExecutions-progress.NextExecutionIdx, result.Succeed)
	s.Equal(0, result.Failed)
}

func (s *workflowShadowerActivitiesSuite) TestReplayWorkflowExecutionActivity_RandomReplayResult() {
	numSucceed := 0
	numFailed := 0
	numSkipped := 0
	numExecutions := 100

	mismatchWorkflowHistory := getTestReplayWorkflowMismatchHistory()
	for i := 0; i != numExecutions; i++ {
		switch rand.Intn(3) {
		case 0:
			numSucceed++
			s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).Return(&shared.GetWorkflowExecutionHistoryResponse{
				History: s.testWorkflowHistory,
			}, nil).Times(1)
		case 1:
			numSkipped++
			s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).Return(nil, &shared.EntityNotExistsError{}).Times(1)
		case 2:
			numFailed++
			s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).Return(&shared.GetWorkflowExecutionHistoryResponse{
				History: mismatchWorkflowHistory,
			}, nil).Times(1)
		}
	}

	params := replayWorkflowActivityParams{
		Domain:     defaultTestDomain,
		Executions: make([]WorkflowExecution, numExecutions),
	}

	resultValue, err := s.env.ExecuteActivity(replayWorkflowExecutionActivityName, params)
	s.NoError(err)

	var result replayWorkflowActivityResult
	s.NoError(resultValue.Get(&result))
	s.Equal(numSucceed, result.Succeed)
	s.Equal(numFailed, result.Failed)
	s.Equal(numSkipped, numExecutions-result.Succeed-result.Failed)
}

func (s *workflowShadowerActivitiesSuite) TestReplayWorkflowExecutionActivity_WorkflowNotRegistered() {
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).Return(&shared.GetWorkflowExecutionHistoryResponse{
		History: getTestReplayWorkflowLocalActivityHistory(), // this workflow type is not registered
	}, nil).Times(1)

	params := replayWorkflowActivityParams{
		Domain:     defaultTestDomain,
		Executions: make([]WorkflowExecution, 5),
	}

	_, err := s.env.ExecuteActivity(replayWorkflowExecutionActivityName, params)
	s.Error(err)
	s.Equal(errMsgWorkflowTypeNotRegistered, err.Error())
}
