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

package common

import (
	"testing"

	s "go.uber.org/cadence/.gen/go/shared"

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
	assert.Equal(t, 1.1, *Float64Ptr(1.1))
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
