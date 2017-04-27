package cadence

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	s "github.com/uber-go/cadence-client/.gen/go/shared"
	"github.com/uber-go/cadence-client/mocks"
	"github.com/uber-go/cadence-client/common/backoff"
	"github.com/uber-go/cadence-client/common"
)

func TestActivityHeartbeat(t *testing.T) {
	service := new(mocks.TChanWorkflowService)
	ctx, cancel := context.WithCancel(context.Background())
	invoker := newServiceInvoker([]byte("task-token"), "identity", service, cancel)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{serviceInvoker: invoker})

	service.On("RecordActivityTaskHeartbeat", mock.Anything, mock.Anything).
		Return(&s.RecordActivityTaskHeartbeatResponse{}, nil).Once()

	err := RecordActivityHeartbeat(ctx, "testDetails")
	require.NoError(t, err)
}

func TestActivityHeartbeat_InternalError(t *testing.T) {
	p := backoff.NewExponentialRetryPolicy(time.Millisecond)
	p.SetMaximumInterval(100 * time.Millisecond)
	p.SetExpirationInterval(100 * time.Millisecond)

	service := new(mocks.TChanWorkflowService)
	ctx, cancel := context.WithCancel(context.Background())
	invoker := newServiceInvoker([]byte("task-token"), "identity", service, cancel)
	invoker.(*cadenceInvoker).retryPolicy = p
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{serviceInvoker: invoker})

	service.On("RecordActivityTaskHeartbeat", mock.Anything, mock.Anything).
		Return(nil, s.NewInternalServiceError())

	err := RecordActivityHeartbeat(ctx, "testDetails")
	require.NoError(t, err)
}

func TestActivityHeartbeat_CancelRequested(t *testing.T) {
	service := new(mocks.TChanWorkflowService)
	ctx, cancel := context.WithCancel(context.Background())
	invoker := newServiceInvoker([]byte("task-token"), "identity", service, cancel)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{serviceInvoker: invoker})

	service.On("RecordActivityTaskHeartbeat", mock.Anything, mock.Anything).
		Return(&s.RecordActivityTaskHeartbeatResponse{CancelRequested: common.BoolPtr(true)}, nil).Once()

	err := RecordActivityHeartbeat(ctx, "testDetails")
	require.NoError(t, err)
	<-ctx.Done()
	require.Equal(t, ctx.Err(), context.Canceled)
}

