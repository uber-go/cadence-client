package internal

import (
	"testing"
	"time"

	"github.com/facebookgo/clock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
)

type workflowShadowerSuite struct {
	*require.Assertions
	suite.Suite

	controller  *gomock.Controller
	mockService *workflowservicetest.MockClient

	testShadower        *WorkflowShadower
	testWorkflowHistory *shared.History
	testTimestamp       time.Time
}

func TestWorkflowShadowerSuite(t *testing.T) {
	s := new(workflowShadowerSuite)
	suite.Run(t, s)
}

func (s *workflowShadowerSuite) SetupTest() {
	s.Assertions = require.New(s.T())

	s.controller = gomock.NewController(s.T())
	s.mockService = workflowservicetest.NewMockClient(s.controller)

	var err error
	s.testShadower, err = NewWorkflowShadower(s.mockService, &WorkflowShadowerOptions{
		Domain: "testDomain",
	})
	s.NoError(err)

	// overwrite shadower clock to be a mock clock
	s.testShadower.clock = clock.NewMock()
	// register test workflow
	s.testShadower.RegisterWorkflow(testReplayWorkflow)

	// TODO: refactor test for workflow replayer and delete the duplicated test workflow history defined below
	taskList := "taskList1"
	s.testWorkflowHistory = &shared.History{Events: []*shared.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &shared.WorkflowExecutionStartedEventAttributes{
			WorkflowType: &shared.WorkflowType{Name: common.StringPtr("go.uber.org/cadence/internal.testReplayWorkflow")},
			TaskList:     &shared.TaskList{Name: common.StringPtr(taskList)},
			Input:        testEncodeFunctionArgs(getDefaultDataConverter()),
		}),
		createTestEventDecisionTaskScheduled(2, &shared.DecisionTaskScheduledEventAttributes{}),
		createTestEventDecisionTaskStarted(3),
		createTestEventDecisionTaskCompleted(4, &shared.DecisionTaskCompletedEventAttributes{}),
		createTestEventActivityTaskScheduled(5, &shared.ActivityTaskScheduledEventAttributes{
			ActivityId:   common.StringPtr("0"),
			ActivityType: &shared.ActivityType{Name: common.StringPtr("testActivity")},
			TaskList:     &shared.TaskList{Name: &taskList},
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
	}}

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
			msg: "both forms are used",
			timeFilter: TimeFilter{
				TimeRange:    "24h",
				MinTimestamp: time.Now().Add(-24 * time.Hour),
				MaxTimestamp: time.Now(),
			},
			expectErr: true,
		},
		{
			msg:        "neither form is used",
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
		{
			msg: "time range",
			timeFilter: TimeFilter{
				TimeRange: "24h",
			},
			expectErr: false,
			validationFn: func(f TimeFilter) {
				s.Equal(24*time.Hour, f.MaxTimestamp.Sub(f.MinTimestamp))
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

func (s *workflowShadowerSuite) TestShadowerOptionValidation() {
	testCases := []struct {
		msg          string
		options      WorkflowShadowerOptions
		expectErr    bool
		validationFn func(*WorkflowShadowerOptions)
	}{
		{
			msg:       "domain not specified",
			options:   WorkflowShadowerOptions{},
			expectErr: true,
		},
		{
			msg: "populate sampling rate and logger",
			options: WorkflowShadowerOptions{
				Domain: "testDomain",
			},
			expectErr: false,
			validationFn: func(options *WorkflowShadowerOptions) {
				s.Empty(options.WorkflowQuery)
				s.Equal(1.0, options.SamplingRate)
				s.NotNil(options.Logger)
			},
		},
		{
			msg: "construct query",
			options: WorkflowShadowerOptions{
				Domain:         "testDomain",
				WorkflowTypes:  []string{"testWorkflowType"},
				WorkflowStatus: []string{"open"},
				WorkflowStartTimeFilter: &TimeFilter{
					MinTimestamp: s.testTimestamp.Add(-time.Hour),
					MaxTimestamp: s.testTimestamp,
				},
			},
			expectErr: false,
			validationFn: func(options *WorkflowShadowerOptions) {
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

	s.testShadower.options.ExitCondition = &WorkflowShadowerExitCondition{
		ExpirationTime: expirationTime,
	}

	s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(&shared.ListWorkflowExecutionsResponse{
		Executions:    s.newTestWorkflowExecutions(totalWorkflows),
		NextPageToken: nil,
	}, nil).Times(1)
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(...interface{}) (*shared.GetWorkflowExecutionHistoryResponse, error) {
		s.testShadower.clock.(*clock.Mock).Add(timePerWorkflow)
		return &shared.GetWorkflowExecutionHistoryResponse{
			History: s.testWorkflowHistory,
		}, nil
	}).Times(int(expirationTime/timePerWorkflow) + 1)

	s.NoError(s.testShadower.shadowWorker())
}

func (s *workflowShadowerSuite) TestShadowWorkerExitCondition_MaxShadowingCount() {
	maxShadowingCount := 50

	s.testShadower.options.ExitCondition = &WorkflowShadowerExitCondition{
		MaxShadowingCount: maxShadowingCount,
	}

	s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(&shared.ListWorkflowExecutionsResponse{
		Executions:    s.newTestWorkflowExecutions(maxShadowingCount * 2),
		NextPageToken: []byte{1, 2, 3},
	}, nil).Times(1)
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), gomock.Any()).Return(&shared.GetWorkflowExecutionHistoryResponse{
		History: s.testWorkflowHistory,
	}, nil).Times(maxShadowingCount)

	s.NoError(s.testShadower.shadowWorker())
}

func (s *workflowShadowerSuite) TestShadowWorkerExitCondition_FullScan() {
	workflowExecutions := s.newTestWorkflowExecutions(10)
	numScan := 3
	totalWorkflows := len(workflowExecutions) * numScan

	for i := 0; i != numScan; i++ {
		scanResp := &shared.ListWorkflowExecutionsResponse{
			Executions:    workflowExecutions,
			NextPageToken: []byte{1, 2, 3},
		}
		if i == numScan-1 {
			scanResp.NextPageToken = nil
		}
		s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(scanResp, nil).Times(1)
	}

	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), gomock.Any()).Return(&shared.GetWorkflowExecutionHistoryResponse{
		History: s.testWorkflowHistory,
	}, nil).Times(totalWorkflows)

	s.NoError(s.testShadower.shadowWorker())
}

func (s *workflowShadowerSuite) TestShadowWorkerExitCondition_ReplayFailed() {
	successfullyReplayed := 5
	s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(&shared.ListWorkflowExecutionsResponse{
		Executions:    s.newTestWorkflowExecutions(successfullyReplayed * 2),
		NextPageToken: []byte{1, 2, 3},
	}, nil).Times(1)
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), gomock.Any()).Return(&shared.GetWorkflowExecutionHistoryResponse{
		History: s.testWorkflowHistory,
	}, nil).Times(successfullyReplayed)
	s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), gomock.Any()).Return(&shared.GetWorkflowExecutionHistoryResponse{
		History: &shared.History{Events: []*shared.HistoryEvent{}},
	}, nil).Times(1)

	s.Error(s.testShadower.shadowWorker())
}

func (s *workflowShadowerSuite) TestShadowWorkerExitCondition_ExpectedReplayError() {
	testCases := []struct {
		msg                string
		getHistoryErr      error
		getHistoryResponse *shared.GetWorkflowExecutionHistoryResponse
	}{
		{
			msg:           "only workflow started event", // for example cron workflow
			getHistoryErr: nil,
			getHistoryResponse: &shared.GetWorkflowExecutionHistoryResponse{
				History: &shared.History{Events: []*shared.HistoryEvent{
					createTestEventWorkflowExecutionStarted(1, &shared.WorkflowExecutionStartedEventAttributes{
						WorkflowType: &shared.WorkflowType{Name: common.StringPtr("testWorkflow")},
						TaskList:     &shared.TaskList{Name: common.StringPtr("taskList")},
						Input:        testEncodeFunctionArgs(getDefaultDataConverter()),
					}),
				},
				},
			},
		},
		{
			msg:                "workflow not exist",
			getHistoryErr:      &shared.EntityNotExistsError{Message: "Workflow passed retention date"},
			getHistoryResponse: nil,
		},
		{
			msg:                "corrupted workflow history", // for example cron workflow
			getHistoryErr:      &shared.InternalServiceError{Message: "History events not continuous"},
			getHistoryResponse: nil,
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			s.mockService.EXPECT().ScanWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(&shared.ListWorkflowExecutionsResponse{
				Executions:    s.newTestWorkflowExecutions(1),
				NextPageToken: nil,
			}, nil).Times(1)
			s.mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), gomock.Any()).Return(test.getHistoryResponse, test.getHistoryErr).Times(1)

			s.NoError(s.testShadower.shadowWorker())
		})
	}
}

func (s *workflowShadowerSuite) newTestWorkflowExecutions(size int) []*shared.WorkflowExecutionInfo {
	executions := make([]*shared.WorkflowExecutionInfo, size)
	for i := 0; i != size; i++ {
		executions[i] = &shared.WorkflowExecutionInfo{
			Execution: &shared.WorkflowExecution{
				WorkflowId: common.StringPtr("workflowID"),
				RunId:      common.StringPtr("runID"),
			},
		}
	}
	return executions
}
