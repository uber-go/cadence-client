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

package internal

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
	"go.uber.org/cadence/v2/internal/api"
	"go.uber.org/yarpc"
)

type activityTestSuite struct {
	suite.Suite
	mockCtrl *gomock.Controller
	service *api.MockInterface
}

func TestActivityTestSuite(t *testing.T) {
	s := new(activityTestSuite)
	suite.Run(t, s)
}

func (s *activityTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.service = api.NewMockInterface(s.mockCtrl)
}

func (s *activityTestSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

// this is the mock for yarpcCallOptions, make sure length are the same
var callOptions = []interface{}{gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()}

var featureFlags = FeatureFlags{}

func (s *activityTestSuite) TestActivityHeartbeat() {
	ctx, cancel := context.WithCancel(context.Background())
	invoker := newServiceInvoker([]byte("task-token"), "identity", s.service, cancel, 1, make(chan struct{}), featureFlags)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{serviceInvoker: invoker})

	s.service.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).
		Return(&apiv1.RecordActivityTaskHeartbeatResponse{}, nil).Times(1)

	RecordActivityHeartbeat(ctx, "testDetails")
}

func (s *activityTestSuite) TestActivityHeartbeat_InternalError() {
	ctx, cancel := context.WithCancel(context.Background())
	invoker := newServiceInvoker([]byte("task-token"), "identity", s.service, cancel, 1, make(chan struct{}), featureFlags)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{
		serviceInvoker: invoker,
		logger:         getTestLogger(s.T())})

	s.service.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).
		Return(nil, &api.InternalServiceError{}).
		Do(func(ctx context.Context, request *apiv1.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) {
			fmt.Println("MOCK RecordActivityTaskHeartbeat executed")
		}).AnyTimes()

	RecordActivityHeartbeat(ctx, "testDetails")
}

func (s *activityTestSuite) TestActivityHeartbeat_CancelRequested() {
	ctx, cancel := context.WithCancel(context.Background())
	invoker := newServiceInvoker([]byte("task-token"), "identity", s.service, cancel, 1, make(chan struct{}), featureFlags)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{
		serviceInvoker: invoker,
		logger:         getTestLogger(s.T())})

	s.service.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).
		Return(&apiv1.RecordActivityTaskHeartbeatResponse{CancelRequested: true}, nil).Times(1)

	RecordActivityHeartbeat(ctx, "testDetails")
	<-ctx.Done()
	require.Equal(s.T(), ctx.Err(), context.Canceled)
}

func (s *activityTestSuite) TestActivityHeartbeat_EntityNotExist() {
	ctx, cancel := context.WithCancel(context.Background())
	invoker := newServiceInvoker([]byte("task-token"), "identity", s.service, cancel, 1, make(chan struct{}), featureFlags)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{
		serviceInvoker: invoker,
		logger:         getTestLogger(s.T())})

	s.service.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).
		Return(&apiv1.RecordActivityTaskHeartbeatResponse{}, &api.EntityNotExistsError{}).Times(1)

	RecordActivityHeartbeat(ctx, "testDetails")
	<-ctx.Done()
	require.Equal(s.T(), ctx.Err(), context.Canceled)
}

func (s *activityTestSuite) TestActivityHeartbeat_SuppressContinousInvokes() {
	ctx, cancel := context.WithCancel(context.Background())
	invoker := newServiceInvoker([]byte("task-token"), "identity", s.service, cancel, 2, make(chan struct{}), featureFlags)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{
		serviceInvoker: invoker,
		logger:         getTestLogger(s.T())})

	// Multiple calls but only one call is made.
	s.service.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).
		Return(&apiv1.RecordActivityTaskHeartbeatResponse{}, nil).Times(1)
	RecordActivityHeartbeat(ctx, "testDetails")
	RecordActivityHeartbeat(ctx, "testDetails")
	RecordActivityHeartbeat(ctx, "testDetails")
	invoker.Close(false)

	// No HB timeout configured.
	service2 := api.NewMockInterface(s.mockCtrl)
	invoker2 := newServiceInvoker([]byte("task-token"), "identity", service2, cancel, 0, make(chan struct{}), featureFlags)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{
		serviceInvoker: invoker2,
		logger:         getTestLogger(s.T())})
	service2.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).
		Return(&apiv1.RecordActivityTaskHeartbeatResponse{}, nil).Times(1)
	RecordActivityHeartbeat(ctx, "testDetails")
	RecordActivityHeartbeat(ctx, "testDetails")
	invoker2.Close(false)

	// simulate batch picks before expiry.
	waitCh := make(chan struct{})
	service3 := api.NewMockInterface(s.mockCtrl)
	invoker3 := newServiceInvoker([]byte("task-token"), "identity", service3, cancel, 2, make(chan struct{}), featureFlags)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{
		serviceInvoker: invoker3,
		logger:         getTestLogger(s.T())})
	service3.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).
		Return(&apiv1.RecordActivityTaskHeartbeatResponse{}, nil).Times(1)

	service3.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).
		Return(&apiv1.RecordActivityTaskHeartbeatResponse{}, nil).
		Do(func(ctx context.Context, request *apiv1.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) {
			ev := newEncodedValues(request.Details.GetData(), nil)
			var progress string
			err := ev.Get(&progress)
			if err != nil {
				panic(err)
			}
			require.Equal(s.T(), "testDetails-expected", progress)
			waitCh <- struct{}{}
		}).Times(1)

	RecordActivityHeartbeat(ctx, "testDetails")
	RecordActivityHeartbeat(ctx, "testDetails2")
	RecordActivityHeartbeat(ctx, "testDetails3")
	RecordActivityHeartbeat(ctx, "testDetails-expected")
	<-waitCh
	invoker3.Close(false)

	// simulate batch picks before expiry, with out any progress specified.
	waitCh2 := make(chan struct{})
	service4 := api.NewMockInterface(s.mockCtrl)
	invoker4 := newServiceInvoker([]byte("task-token"), "identity", service4, cancel, 2, make(chan struct{}), featureFlags)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{
		serviceInvoker: invoker4,
		logger:         getTestLogger(s.T())})
	service4.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).
		Return(&apiv1.RecordActivityTaskHeartbeatResponse{}, nil).Times(1)
	service4.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).
		Return(&apiv1.RecordActivityTaskHeartbeatResponse{}, nil).
		Do(func(ctx context.Context, request *apiv1.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) {
			require.Nil(s.T(), request.Details.GetData())
			waitCh2 <- struct{}{}
		}).Times(1)

	RecordActivityHeartbeat(ctx, nil)
	RecordActivityHeartbeat(ctx, nil)
	RecordActivityHeartbeat(ctx, nil)
	RecordActivityHeartbeat(ctx, nil)
	<-waitCh2
	invoker4.Close(false)
}

func (s *activityTestSuite) TestActivityHeartbeat_WorkerStop() {
	ctx, cancel := context.WithCancel(context.Background())
	workerStopChannel := make(chan struct{})
	invoker := newServiceInvoker([]byte("task-token"), "identity", s.service, cancel, 5, workerStopChannel, featureFlags)
	ctx = context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{serviceInvoker: invoker})

	heartBeatDetail := "testDetails"
	waitCh := make(chan struct{}, 1)
	waitCh <- struct{}{}
	waitC2 := make(chan struct{}, 1)
	s.service.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).
		Return(&apiv1.RecordActivityTaskHeartbeatResponse{}, nil).
		Do(func(ctx context.Context, request *apiv1.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) {
			if _, ok := <-waitCh; ok {
				close(waitCh)
				return
			}
			close(waitC2)
		}).Times(2)
	RecordActivityHeartbeat(ctx, heartBeatDetail)
	RecordActivityHeartbeat(ctx, "testDetails")
	close(workerStopChannel)
	<-waitC2
}

func (s *activityTestSuite) TestGetWorkerStopChannel() {
	ch := make(chan struct{}, 1)
	ctx := context.WithValue(context.Background(), activityEnvContextKey, &activityEnvironment{workerStopChannel: ch})
	channel := GetWorkerStopChannel(ctx)
	s.NotNil(channel)
}
