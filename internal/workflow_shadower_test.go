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
	"sync"
	"testing"
	"time"

	"github.com/facebookgo/clock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
	"go.uber.org/cadence/v2/internal/api"
)

type workflowShadowerSuite struct {
	*require.Assertions
	suite.Suite

	controller  *gomock.Controller
	mockService *api.MockInterface

	testShadower        *WorkflowShadower
	testWorkflowHistory *apiv1.History
	testTimestamp       time.Time
}

func TestWorkflowShadowerSuite(t *testing.T) {
	s := new(workflowShadowerSuite)
	suite.Run(t, s)
}

func (s *workflowShadowerSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockService = api.NewMockInterface(s.controller)

	var err error
	s.testShadower, err = NewWorkflowShadower(s.mockService, "testDomain", ShadowOptions{}, ReplayOptions{}, nil)
	s.NoError(err)

	// overwrite shadower clock to be a mock clock
	s.testShadower.clock = clock.NewMock()
	// register test workflow
	s.testShadower.RegisterWorkflow(testReplayWorkflow)

	s.testWorkflowHistory = getTestReplayWorkflowFullHistory()

	s.testTimestamp = time.Now()
}

func (s *workflowShadowerSuite) TearDownTest() {
	s.controller.Finish()
}

func (s *workflowShadowerSuite) TestTimeFilterValidation() {
	testCases := []struct {
		msg          string
		timeFilter   TimeFilter
		expectErr    bool
		validationFn func(TimeFilter)
	}{
		{
			msg: "maxTimestamp before minTimestamp",
			timeFilter: TimeFilter{
				MinTimestamp: s.testTimestamp.Add(time.Hour),
				MaxTimestamp: s.testTimestamp,
			},
			expectErr: true,
		},
		{
			msg:        "neither timestamp is specified",
			timeFilter: TimeFilter{},
			expectErr:  false,
			validationFn: func(f TimeFilter) {
				s.True(f.MinTimestamp.IsZero())
				s.True(f.MaxTimestamp.Equal(maxTimestamp))
			},
		},
		{
			msg: "only min timestamp is specified",
			timeFilter: TimeFilter{
				MinTimestamp: s.testTimestamp,
			},
			expectErr: false,
			validationFn: func(f TimeFilter) {
				s.True(f.MinTimestamp.Equal(s.testTimestamp))
				s.True(f.MaxTimestamp.Equal(maxTimestamp))
			},
		},
		{
			msg: "only max timestamp is specified",
			timeFilter: TimeFilter{
				MaxTimestamp: s.testTimestamp,
			},
			expectErr: false,
			validationFn: func(f TimeFilter) {
				s.True(f.MinTimestamp.IsZero())
				s.True(f.MaxTimestamp.Equal(s.testTimestamp))
			},
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			err := test.timeFilter.validateAndPopulateFields()
			if test.expectErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			test.validationFn(test.timeFilter)
		})
	}
}

func (s *workflowShadowerSuite) TestTimeFilterIsEmpty() {
	testCases := []struct {
		msg     string
		filter  *TimeFilter
		isEmpty bool
	}{
		{
			msg:     "nil pointer",
			filter:  nil,
			isEmpty: true,
		},
		{
			msg:     "neither field is specified",
			filter:  &TimeFilter{},
			isEmpty: true,
		},
		{
			msg: "not empty",
			filter: &TimeFilter{
				MaxTimestamp: time.Now(),
			},
			isEmpty: false,
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			s.Equal(test.isEmpty, test.filter.isEmpty())
		})
	}
}

func (s *workflowShadowerSuite) TestShadowOptionsValidation() {
	testCases := []struct {
		msg          string
		options      ShadowOptions
		expectErr    bool
		validationFn func(*ShadowOptions)
	}{
		{
			msg: "exit condition not specified in continuous mode",
			options: ShadowOptions{
				Mode: ShadowModeContinuous,
			},
			expectErr: true,
		},
		{
			msg: "both query and other filters are specified",
			options: ShadowOptions{
				WorkflowQuery: "some random query",
				WorkflowStartTimeFilter: TimeFilter{
					MinTimestamp: time.Now(),
				},
			},
			expectErr: true,
		},
		{
			msg:       "populate sampling rate, concurrency and status",
			options:   ShadowOptions{},
			expectErr: false,
			validationFn: func(options *ShadowOptions) {
				s.Equal("(CloseTime = missing)", options.WorkflowQuery)
				s.Equal(1.0, options.SamplingRate)
				s.Equal(1, options.Concurrency)
			},
		},
		{
			msg: "construct query",
			options: ShadowOptions{
				WorkflowTypes:  []string{"testWorkflowType"},
				WorkflowStatus: []string{"open"},
				WorkflowStartTimeFilter: TimeFilter{
					MinTimestamp: s.testTimestamp.Add(-time.Hour),
					MaxTimestamp: s.testTimestamp,
				},
			},
			expectErr: false,
			validationFn: func(options *ShadowOptions) {
				expectedQuery := NewQueryBuilder().
					WorkflowTypes([]string{"testWorkflowType"}).
					WorkflowStatus([]WorkflowStatus{WorkflowStatusOpen}).
					StartTime(
						s.testTimestamp.Add(-time.Hour),
						s.testTimestamp,
					).Build()

				s.Equal(expectedQuery, options.WorkflowQuery)
			},
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			err := test.options.validateAndPopulateFields()
			if test.expectErr {
				s.Error(err)
				return
			}

			s.NoError(err)
			test.validationFn(&test.options)
		})
	}
}

func (s *workflowShadowerSuite) TestShadowWorkerExitCondition_ExpirationTime() {
	totalWorkflows := 50
	timePerWorkflow := 7 * time.Second
	expirationTime := time.Minute

	s.testShadower.shadowOptions.ExitCondition = ShadowExitCondition{
		ExpirationInterval: expirationTime,
	}

	s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.ScanWorkflowExecutionsResponse{
		Executions:    newTestWorkflowExecutions(totalWorkflows),
		NextPageToken: nil,
	}, nil).Times(1)
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).DoAndReturn(func(...interface{}) (*apiv1.GetWorkflowExecutionHistoryResponse, error) {
		s.testShadower.clock.(*clock.Mock).Add(timePerWorkflow)
		return &apiv1.GetWorkflowExecutionHistoryResponse{
			History: s.testWorkflowHistory,
		}, nil
	}).Times(int(expirationTime/timePerWorkflow) + 1)

	s.NoError(s.testShadower.shadowWorker())
}

func (s *workflowShadowerSuite) TestShadowWorkerExitCondition_MaxShadowingCount() {
	maxShadowCount := 50

	s.testShadower.shadowOptions.ExitCondition = ShadowExitCondition{
		ShadowCount: maxShadowCount,
	}

	s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.ScanWorkflowExecutionsResponse{
		Executions:    newTestWorkflowExecutions(maxShadowCount * 2),
		NextPageToken: []byte{1, 2, 3},
	}, nil).Times(1)
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.GetWorkflowExecutionHistoryResponse{
		History: s.testWorkflowHistory,
	}, nil).Times(maxShadowCount)

	s.NoError(s.testShadower.shadowWorker())
}

func (s *workflowShadowerSuite) TestShadowWorker_NormalMode() {
	workflowExecutions := newTestWorkflowExecutions(10)
	numScan := 3
	totalWorkflows := len(workflowExecutions) * numScan

	for i := 0; i != numScan; i++ {
		scanResp := &apiv1.ScanWorkflowExecutionsResponse{
			Executions:    workflowExecutions,
			NextPageToken: []byte{1, 2, 3},
		}
		if i == numScan-1 {
			scanResp.NextPageToken = nil
		}
		s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions...).Return(scanResp, nil).Times(1)
	}

	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.GetWorkflowExecutionHistoryResponse{
		History: s.testWorkflowHistory,
	}, nil).Times(totalWorkflows)

	s.NoError(s.testShadower.shadowWorker())
}

func (s *workflowShadowerSuite) TestShadowWorker_ContinuousMode() {
	workflowExecutions := newTestWorkflowExecutions(10)
	numScan := 3
	totalWorkflows := len(workflowExecutions) * numScan

	s.testShadower.shadowOptions.Mode = ShadowModeContinuous
	s.testShadower.shadowOptions.ExitCondition = ShadowExitCondition{
		ShadowCount: totalWorkflows,
	}

	for i := 0; i != numScan; i++ {
		scanResp := &apiv1.ScanWorkflowExecutionsResponse{
			Executions: workflowExecutions,
		}
		s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions...).Return(scanResp, nil).Times(1)
	}

	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.GetWorkflowExecutionHistoryResponse{
		History: s.testWorkflowHistory,
	}, nil).Times(totalWorkflows)

	doneCh := make(chan struct{})
	var advanceTimeWG sync.WaitGroup
	advanceTimeWG.Add(1)
	go func() {
		defer advanceTimeWG.Done()
		for {
			time.Sleep(100 * time.Millisecond)
			select {
			case <-doneCh:
				return
			default:
				s.testShadower.clock.(*clock.Mock).Add(defaultWaitDurationPerIteration)
			}
		}
	}()

	s.NoError(s.testShadower.shadowWorker())
	close(doneCh)
	advanceTimeWG.Wait()
}

func (s *workflowShadowerSuite) TestShadowWorker_ReplayFailed() {
	successfullyReplayed := 5
	s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.ScanWorkflowExecutionsResponse{
		Executions:    newTestWorkflowExecutions(successfullyReplayed * 2),
		NextPageToken: []byte{1, 2, 3},
	}, nil).Times(1)
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.GetWorkflowExecutionHistoryResponse{
		History: s.testWorkflowHistory,
	}, nil).Times(successfullyReplayed)
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.GetWorkflowExecutionHistoryResponse{
		History: getTestReplayWorkflowMismatchHistory(),
	}, nil).Times(1)

	s.Error(s.testShadower.shadowWorker())
}

func (s *workflowShadowerSuite) TestShadowWorker_ExpectedReplayError() {
	testCases := []struct {
		msg                string
		getHistoryErr      error
		getHistoryResponse *apiv1.GetWorkflowExecutionHistoryResponse
	}{
		{
			msg:           "only workflow started event", // for example cron workflow
			getHistoryErr: nil,
			getHistoryResponse: &apiv1.GetWorkflowExecutionHistoryResponse{
				History: &apiv1.History{Events: []*apiv1.HistoryEvent{
					createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{
						WorkflowType: &apiv1.WorkflowType{Name: "testWorkflow"},
						TaskList:     &apiv1.TaskList{Name: "taskList"},
						Input:        &apiv1.Payload{Data: testEncodeFunctionArgs(getDefaultDataConverter())},
						CronSchedule: "* * * * *",
					}),
				},
				},
			},
		},
		{
			msg:                "workflow not exist",
			getHistoryErr:      &api.EntityNotExistsError{Message: "Workflow passed retention date"},
			getHistoryResponse: nil,
		},
		{
			msg:                "corrupted workflow history", // for example cron workflow
			getHistoryErr:      &api.InternalServiceError{Message: "History events not continuous"},
			getHistoryResponse: nil,
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), callOptions...).Return(&apiv1.ScanWorkflowExecutionsResponse{
				Executions:    newTestWorkflowExecutions(1),
				NextPageToken: nil,
			}, nil).Times(1)
			s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), callOptions...).Return(test.getHistoryResponse, test.getHistoryErr).Times(1)

			s.NoError(s.testShadower.shadowWorker())
		})
	}
}

func newTestWorkflowExecutions(size int) []*apiv1.WorkflowExecutionInfo {
	executions := make([]*apiv1.WorkflowExecutionInfo, size)
	for i := 0; i != size; i++ {
		executions[i] = &apiv1.WorkflowExecutionInfo{
			WorkflowExecution: &apiv1.WorkflowExecution{
				WorkflowId: "workflowID",
				RunId:      "runID",
			},
		}
	}
	return executions
}
