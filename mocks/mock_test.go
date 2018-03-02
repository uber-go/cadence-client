package mocks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/cadence/workflow"
	"go.uber.org/cadence/client"
)

func Test_MockClient(t *testing.T) {
	mockClient := &Client{}

	mockClient.On("StartWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&workflow.Execution{ID: "test-workflowid", RunID: "test-runid"}, nil).Once()

	we, err := mockClient.StartWorkflow(context.Background(), client.StartWorkflowOptions{}, "workflow", "test-input")

	mockClient.AssertExpectations(t)
	require.NoError(t, err)
	require.Equal(t, "test-workflowid", we.ID)
	require.Equal(t, "test-runid", we.RunID)
}
