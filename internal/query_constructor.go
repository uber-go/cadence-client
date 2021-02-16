package internal

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/cadence/.gen/go/shared"
)

type (
	// WorkflowStatus represents the status of a workflow
	WorkflowStatus string
)

const (
	// WorkflowStatusOpen is the WorkflowStatus for open workflows
	WorkflowStatusOpen WorkflowStatus = "OPEN"
	// WorkflowStatusClosed is the WorkflowStatus for closed workflows
	WorkflowStatusClosed WorkflowStatus = "CLOSED"
)

var (
	// WorkflowStatusCompleted is the WorkflowStatus for completed workflow
	WorkflowStatusCompleted = WorkflowStatus(shared.WorkflowExecutionCloseStatusCompleted.String())
	// WorkflowStatusFailed is the WorkflowStatus for failed workflows
	WorkflowStatusFailed = WorkflowStatus(shared.WorkflowExecutionCloseStatusFailed.String())
	// WorkflowStatusCanceled is the WorkflowStatus for canceled workflows
	WorkflowStatusCanceled = WorkflowStatus(shared.WorkflowExecutionCloseStatusCanceled.String())
	// WorkflowStatusTerminated is the WorkflowStatus for terminated workflows
	WorkflowStatusTerminated = WorkflowStatus(shared.WorkflowExecutionCloseStatusTerminated.String())
	// WorkflowStatusContinuedAsNew is the WorkflowStatus for continuedAsNew workflows
	WorkflowStatusContinuedAsNew = WorkflowStatus(shared.WorkflowExecutionCloseStatusContinuedAsNew.String())
	// WorkflowStatusTimedOut is the WorkflowStatus for timedout workflows
	WorkflowStatusTimedOut = WorkflowStatus(shared.WorkflowExecutionCloseStatusTimedOut.String())
)

const (
	keyWorkflowType = "WorkflowType"
	keyCloseStatus  = "CloseStatus"
	keyStartTime    = "StartTime"
	keyCloseTime    = "CloseTime"
)

type (
	// QueryConstructor constructs visibility query
	QueryConstructor interface {
		WorkflowTypes([]string) QueryConstructor
		WorkflowStatus([]WorkflowStatus) QueryConstructor
		StartTime(time.Time, time.Time) QueryConstructor
		Query() string
	}

	queryConstructorImpl struct {
		builder strings.Builder
	}
)

// NewQueryConstructor creates a new visibility QueryConstructor
func NewQueryConstructor() QueryConstructor {
	return &queryConstructorImpl{}
}

func (q *queryConstructorImpl) WorkflowTypes(types []string) QueryConstructor {
	workflowTypeQueries := make([]string, 0, len(types))
	for _, workflowType := range types {
		workflowTypeQueries = append(workflowTypeQueries, fmt.Sprintf(keyWorkflowType+` = "%v"`, workflowType))
	}
	q.appendPartialQuery(strings.Join(workflowTypeQueries, " or "))
	return q
}

func (q *queryConstructorImpl) WorkflowStatus(statuses []WorkflowStatus) QueryConstructor {
	workflowStatusQueries := make([]string, 0, len(statuses))
	for _, status := range statuses {
		var statusQuery string
		switch status {
		case WorkflowStatusOpen:
			statusQuery = keyCloseTime + " = missing"
		case WorkflowStatusClosed:
			statusQuery = keyCloseTime + " != missing"
		default:
			statusQuery = keyCloseStatus + ` = "` + string(status) + `"`
		}
		workflowStatusQueries = append(workflowStatusQueries, statusQuery)
	}
	q.appendPartialQuery(strings.Join(workflowStatusQueries, " or "))
	return q
}

func (q *queryConstructorImpl) StartTime(minStartTime, maxStartTime time.Time) QueryConstructor {
	startTimeQueries := make([]string, 0, 2)
	if !minStartTime.IsZero() {
		startTimeQueries = append(startTimeQueries, fmt.Sprintf(keyStartTime+` >= %v`, minStartTime.UnixNano()))
	}
	if !maxStartTime.Equal(maxTimestamp) {
		startTimeQueries = append(startTimeQueries, fmt.Sprintf(keyStartTime+` <= %v`, maxStartTime.UnixNano()))
	}

	q.appendPartialQuery(strings.Join(startTimeQueries, " and "))
	return q
}

func (q *queryConstructorImpl) Query() string {
	return q.builder.String()
}

func (q *queryConstructorImpl) appendPartialQuery(query string) {
	if len(query) == 0 {
		return
	}

	if q.builder.Len() != 0 {
		q.builder.WriteString(" and ")
	}

	q.builder.WriteRune('(')
	q.builder.WriteString(query)
	q.builder.WriteRune(')')
}

// ToWorkflowStatus converts workflow status from string type to WorkflowStatus type
func ToWorkflowStatus(statusString string) (WorkflowStatus, error) {
	status := WorkflowStatus(strings.ToUpper(statusString))
	switch status {
	case WorkflowStatusOpen, WorkflowStatusClosed, WorkflowStatusCompleted,
		WorkflowStatusFailed, WorkflowStatusCanceled, WorkflowStatusTerminated,
		WorkflowStatusContinuedAsNew, WorkflowStatusTimedOut:
		return status, nil
	default:
		return "", fmt.Errorf("unknown workflow status: %v", statusString)
	}
}
