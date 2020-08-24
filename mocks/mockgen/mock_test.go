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

package mockgen

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"testing"
)

func Test_MockClient(t *testing.T) {
	testWorkflowID := "test-workflowid"
	testRunID := "test-runid"
	testWorkflowName := "workflow"
	testWorkflowInput := "input"

	mockController := gomock.NewController(t)

	mockClient := NewMockClient(mockController)
	mockClient.EXPECT().StartWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&workflow.Execution{ID: testWorkflowID, RunID: testRunID}, nil).Times(1)
	we, err := mockClient.StartWorkflow(context.Background(), client.StartWorkflowOptions{}, testWorkflowName, testWorkflowInput)
	require.NoError(t, err)
	require.Equal(t, testWorkflowID, we.ID)
	require.Equal(t, testRunID, we.RunID)

	mockClient.EXPECT().SignalWithStartWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&workflow.Execution{ID: testWorkflowID, RunID: testRunID}, nil).Times(1)
	we, err = mockClient.SignalWithStartWorkflow(context.Background(), "wid", "signal", "val", client.StartWorkflowOptions{}, testWorkflowName, testWorkflowInput)
	require.NoError(t, err)
	require.Equal(t, testWorkflowID, we.ID)
	require.Equal(t, testRunID, we.RunID)

	mockWfRun := NewMockWorkflowRun(mockController)
	mockClient.EXPECT().ExecuteWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockWfRun, nil).Times(1)
	wfRun, err := mockClient.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{}, testWorkflowName, testWorkflowInput)
	require.NoError(t, err)
	require.Equal(t, testWorkflowID, we.ID)
	require.Equal(t, testRunID, we.RunID)

	mockWfRun.EXPECT().GetID().Return(testWorkflowID).Times(1)
	mockWfRun.EXPECT().GetRunID().Return(testRunID).Times(1)
	mockWfRun.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	require.Equal(t, testWorkflowID, wfRun.GetID())
	require.Equal(t, testRunID, wfRun.GetRunID())
	require.NoError(t, wfRun.Get(context.Background(), &testWorkflowID))

	mockWfRun.EXPECT().GetID().Return(testWorkflowID).Times(1)
	mockWfRun.EXPECT().GetRunID().Return(testRunID).Times(1)
	mockWfRun.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	mockClient.EXPECT().GetWorkflow(gomock.Any(), gomock.Any(), gomock.Any()).Return(mockWfRun).Times(1)
	wfRun = mockClient.GetWorkflow(context.Background(), testWorkflowID, testRunID)
	require.Equal(t, testWorkflowID, wfRun.GetID())
	require.Equal(t, testRunID, wfRun.GetRunID())
	require.NoError(t, wfRun.Get(context.Background(), &testWorkflowID))

	mockHistoryIter := NewMockHistoryEventIterator(mockController)
	mockHistoryIter.EXPECT().HasNext().Return(true).Times(1)
	mockHistoryIter.EXPECT().Next().Return(&shared.HistoryEvent{}, nil).Times(1)
	mockClient.EXPECT().GetWorkflowHistory(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockHistoryIter).Times(1)
	historyIter := mockClient.GetWorkflowHistory(context.Background(), testWorkflowID, testRunID, true, shared.HistoryEventFilterTypeCloseEvent)
	require.NotNil(t, historyIter)
	require.Equal(t, true, historyIter.HasNext())
	next, err := historyIter.Next()
	require.NotNil(t, next)
	require.NoError(t, err)

	mockController.Finish()
}
