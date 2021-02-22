package internal

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type queryBuilderSuite struct {
	*require.Assertions
	suite.Suite
}

func TestQueryBuilderSuite(t *testing.T) {
	s := new(queryBuilderSuite)
	suite.Run(t, s)
}

func (s *queryBuilderSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func (s *queryBuilderSuite) TestWorkflowTypeQuery() {
	testCases := []struct {
		msg           string
		workflowTypes []string
		expectedQuery string
	}{
		{
			msg:           "empty workflowTypes",
			workflowTypes: nil,
			expectedQuery: "",
		},
		{
			msg:           "single workflowType",
			workflowTypes: []string{"testWorkflowType"},
			expectedQuery: `(WorkflowType = "testWorkflowType")`,
		},
		{
			msg:           "multiple workflowTypes",
			workflowTypes: []string{"testWorkflowType1", "testWorkflowType2"},
			expectedQuery: `(WorkflowType = "testWorkflowType1" or WorkflowType = "testWorkflowType2")`,
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			builder := NewQueryBuilder()
			builder.WorkflowTypes(test.workflowTypes)
			s.Equal(test.expectedQuery, builder.Build())
		})
	}
}

func (s *queryBuilderSuite) TestWorkflowStatusQuery() {
	testCases := []struct {
		msg              string
		workflowStatuses []WorkflowStatus
		expectedQuery    string
	}{
		{
			msg:              "empty workflow status",
			workflowStatuses: []WorkflowStatus{},
			expectedQuery:    "",
		},
		{
			msg:              "open workflow",
			workflowStatuses: []WorkflowStatus{WorkflowStatusOpen},
			expectedQuery:    "(CloseTime = missing)",
		},
		{
			msg:              "closed workflow",
			workflowStatuses: []WorkflowStatus{WorkflowStatusClosed},
			expectedQuery:    "(CloseTime != missing)",
		},
		{
			msg:              "multiple workflow statuses",
			workflowStatuses: []WorkflowStatus{WorkflowStatusFailed, WorkflowStatusTimedOut},
			expectedQuery:    `(CloseStatus = "FAILED" or CloseStatus = "TIMED_OUT")`,
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			builder := NewQueryBuilder()
			builder.WorkflowStatus(test.workflowStatuses)
			s.Equal(test.expectedQuery, builder.Build())
		})
	}
}

func (s *queryBuilderSuite) TestStartTimeQuery() {
	testTimestamp := time.Now()
	testCases := []struct {
		msg           string
		minStartTime  time.Time
		maxStartTime  time.Time
		expectedQuery string
	}{
		{
			msg:           "empty minTimestamp",
			maxStartTime:  testTimestamp,
			expectedQuery: fmt.Sprintf("(StartTime <= %v)", testTimestamp.UnixNano()),
		},
		{
			msg:           "max maxTimestamp",
			minStartTime:  testTimestamp,
			maxStartTime:  maxTimestamp,
			expectedQuery: fmt.Sprintf("(StartTime >= %v)", testTimestamp.UnixNano()),
		},
		{
			msg:           "both timestamps are used",
			minStartTime:  testTimestamp.Add(-time.Hour),
			maxStartTime:  testTimestamp,
			expectedQuery: fmt.Sprintf("(StartTime >= %v and StartTime <= %v)", testTimestamp.Add(-time.Hour).UnixNano(), testTimestamp.UnixNano()),
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			builder := NewQueryBuilder()
			builder.StartTime(test.minStartTime, test.maxStartTime)
			s.Equal(test.expectedQuery, builder.Build())
		})
	}
}

func (s *queryBuilderSuite) TestMultipleFilters() {
	maxStartTime := time.Now()
	minStartTime := maxStartTime.Add(-time.Hour)

	builder := NewQueryBuilder().
		WorkflowTypes([]string{"testWorkflowType1", "testWorkflowType2"}).
		WorkflowStatus([]WorkflowStatus{WorkflowStatusOpen}).
		StartTime(minStartTime, maxStartTime)

	expectedQuery := fmt.Sprintf(`(WorkflowType = "testWorkflowType1" or WorkflowType = "testWorkflowType2") and (CloseTime = missing) and (StartTime >= %v and StartTime <= %v)`,
		minStartTime.UnixNano(),
		maxStartTime.UnixNano(),
	)
	s.Equal(expectedQuery, builder.Build())
}

func (s *queryBuilderSuite) TestToWorkflowStatus() {
	testCases := []struct {
		msg            string
		statusString   string
		expectErr      bool
		expectedStatus WorkflowStatus
	}{
		{
			msg:          "unknown status",
			statusString: "unknown",
			expectErr:    true,
		},
		{
			msg:            "lower case status string",
			statusString:   "open",
			expectErr:      false,
			expectedStatus: WorkflowStatusOpen,
		},
		{
			msg:            "mixed case status string",
			statusString:   "Timed_Out",
			expectErr:      false,
			expectedStatus: WorkflowStatusTimedOut,
		},
		{

			msg:            "upper case status string",
			statusString:   "TERMINATED",
			expectErr:      false,
			expectedStatus: WorkflowStatusTerminated,
		},
	}

	for _, test := range testCases {
		s.T().Run(test.msg, func(t *testing.T) {
			actualStatus, err := ToWorkflowStatus(test.statusString)
			if test.expectErr {
				s.Error(err)
				return
			}

			s.Equal(test.expectedStatus, actualStatus)
		})
	}
}
