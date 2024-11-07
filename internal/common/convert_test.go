package common

import (
	s "go.uber.org/cadence/.gen/go/shared"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPtrOf(t *testing.T) {
	assert.Equal(t, "a", *PtrOf("a"))
	assert.Equal(t, 1, *PtrOf(1))
	assert.Equal(t, int32(1), *PtrOf(int32(1)))
	assert.Equal(t, int64(1), *PtrOf(int64(1)))
	assert.Equal(t, float64(1.1), *PtrOf(float64(1.1)))
	assert.Equal(t, true, *PtrOf(true))
}

func TestPtrHelpers(t *testing.T) {
	assert.Equal(t, int32(1), *Int32Ptr(1))
	assert.Equal(t, int64(1), *Int64Ptr(1))
	assert.Equal(t, float64(1.1), *Float64Ptr(1.1))
	assert.Equal(t, true, *BoolPtr(true))
	assert.Equal(t, "a", *StringPtr("a"))
	assert.Equal(t, s.TaskList{Name: PtrOf("a")}, *TaskListPtr(s.TaskList{Name: PtrOf("a")}))
	assert.Equal(t, s.DecisionTypeScheduleActivityTask, *DecisionTypePtr(s.DecisionTypeScheduleActivityTask))
	assert.Equal(t, s.EventTypeWorkflowExecutionStarted, *EventTypePtr(s.EventTypeWorkflowExecutionStarted))
	assert.Equal(t, s.QueryTaskCompletedTypeCompleted, *QueryTaskCompletedTypePtr(s.QueryTaskCompletedTypeCompleted))
	assert.Equal(t, s.TaskListKindNormal, *TaskListKindPtr(s.TaskListKindNormal))
	assert.Equal(t, s.QueryResultTypeFailed, *QueryResultTypePtr(s.QueryResultTypeFailed))
}

func TestCeilHelpers(t *testing.T) {
	assert.Equal(t, int32(2), Int32Ceil(1.1))
	assert.Equal(t, int64(2), Int64Ceil(1.1))
}
