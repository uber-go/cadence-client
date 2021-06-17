// Copyright (c) 2017 Uber Technologies, Inc.
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

package util

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
)

func Test_byteSliceToString(t *testing.T) {
	data := []byte("blob-data")
	v := reflect.ValueOf(data)
	strVal := valueToString(v)

	require.Equal(t, "[blob-data]", strVal)

	intBlob := []int32{1, 2, 3}
	v2 := reflect.ValueOf(intBlob)
	strVal2 := valueToString(v2)

	require.Equal(t, "[len=3]", strVal2)
}

func Test_GetHistoryEventType(t *testing.T) {
	event := apiv1.HistoryEvent{Attributes: &apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes{}}
	eventType := GetHistoryEventType(&event)
	require.Equal(t, "WorkflowExecutionCompleted", eventType)
}

func Test_GetDecisionType(t *testing.T) {
	decision := apiv1.Decision{Attributes: &apiv1.Decision_StartChildWorkflowExecutionDecisionAttributes{}}
	eventType := GetDecisionType(&decision)
	require.Equal(t, "StartChildWorkflowExecution", eventType)
}

func Test_HistoryEventToString(t *testing.T) {
	event := apiv1.HistoryEvent{Attributes: &apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes{
		WorkflowExecutionCompletedEventAttributes: &apiv1.WorkflowExecutionCompletedEventAttributes {
			Result: &apiv1.Payload{Data: []byte{'a','b','c'}},
			DecisionTaskCompletedEventId: 123,
		},
	}}
	str := HistoryEventToString(&event)
	require.Equal(t, "WorkflowExecutionCompleted: (WorkflowExecutionCompletedEventAttributes:(Result:(Data:[abc]), DecisionTaskCompletedEventId:123))", str)
}