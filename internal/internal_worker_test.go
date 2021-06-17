// Copyright (c) 2017-2020 Uber Technologies Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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
	"errors"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
	"go.uber.org/cadence/v2/internal/api"
	"go.uber.org/yarpc"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func testInternalWorkerRegister(r *registry) {
	r.RegisterWorkflowWithOptions(
		sampleWorkflowExecute,
		RegisterWorkflowOptions{Name: "sampleWorkflowExecute"},
	)
	r.RegisterActivity(testActivityByteArgs)
	r.RegisterActivityWithOptions(
		testActivityMultipleArgs,
		RegisterActivityOptions{Name: "testActivityMultipleArgs"},
	)
	r.RegisterActivity(testActivityMultipleArgsWithStruct)
	r.RegisterActivity(testActivityReturnString)
	r.RegisterActivity(testActivityReturnEmptyString)
	r.RegisterActivity(testActivityReturnEmptyStruct)

	r.RegisterActivity(testActivityNoResult)
	r.RegisterActivity(testActivityNoContextArg)
	r.RegisterActivity(testActivityReturnByteArray)
	r.RegisterActivity(testActivityReturnInt)
	r.RegisterActivity(testActivityReturnNilStructPtr)
	r.RegisterActivity(testActivityReturnStructPtr)
	r.RegisterActivity(testActivityReturnNilStructPtrPtr)
	r.RegisterActivity(testActivityReturnStructPtrPtr)
}

func testInternalWorkerRegisterWithTestEnv(env *TestWorkflowEnvironment) {
	env.RegisterWorkflowWithOptions(
		sampleWorkflowExecute,
		RegisterWorkflowOptions{Name: "sampleWorkflowExecute"},
	)
	env.RegisterActivity(testActivityByteArgs)
	env.RegisterActivityWithOptions(
		testActivityMultipleArgs,
		RegisterActivityOptions{Name: "testActivityMultipleArgs"},
	)
	env.RegisterActivity(testActivityMultipleArgsWithStruct)
	env.RegisterActivity(testActivityReturnString)
	env.RegisterActivity(testActivityReturnEmptyString)
	env.RegisterActivity(testActivityReturnEmptyStruct)

	env.RegisterActivity(testActivityNoResult)
	env.RegisterActivity(testActivityNoContextArg)
	env.RegisterActivity(testActivityReturnByteArray)
	env.RegisterActivity(testActivityReturnInt)
	env.RegisterActivity(testActivityReturnNilStructPtr)
	env.RegisterActivity(testActivityReturnStructPtr)
	env.RegisterActivity(testActivityReturnNilStructPtrPtr)
	env.RegisterActivity(testActivityReturnStructPtrPtr)
}

type internalWorkerTestSuite struct {
	suite.Suite
	mockCtrl *gomock.Controller
	service  *api.MockInterface
	registry *registry
}

func TestInternalWorkerTestSuite(t *testing.T) {
	s := &internalWorkerTestSuite{
		registry: newRegistry(),
	}
	testInternalWorkerRegister(s.registry)
	suite.Run(t, s)
}

func (s *internalWorkerTestSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())
	s.service = api.NewMockInterface(s.mockCtrl)
}

func (s *internalWorkerTestSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

func getTestLogger(t *testing.T) *zap.Logger {
	return zaptest.NewLogger(t)
}

func (s *internalWorkerTestSuite) testDecisionTaskHandlerHelper(params workerExecutionParameters) {
	taskList := "taskList1"
	testEvents := []*apiv1.HistoryEvent{
		createTestEventWorkflowExecutionStarted(1, &apiv1.WorkflowExecutionStartedEventAttributes{
			TaskList: &apiv1.TaskList{Name: taskList},
			Input:    &apiv1.Payload{Data: testEncodeFunctionArgs(params.DataConverter)},
		}),
		createTestEventDecisionTaskScheduled(2, &apiv1.DecisionTaskScheduledEventAttributes{}),
		createTestEventDecisionTaskStarted(3),
	}

	workflowType := "sampleWorkflowExecute"
	workflowID := "testID"
	runID := "testRunID"

	task := &apiv1.PollForDecisionTaskResponse{
		WorkflowExecution:      &apiv1.WorkflowExecution{WorkflowId: workflowID, RunId: runID},
		WorkflowType:           &apiv1.WorkflowType{Name: workflowType},
		History:                &apiv1.History{Events: testEvents},
		PreviousStartedEventId: &types.Int64Value{Value: 0},
	}

	r := newWorkflowTaskHandler(testDomain, params, nil, s.registry)
	_, err := r.ProcessWorkflowTask(&workflowTask{task: task}, nil)
	s.NoError(err)
}

func (s *internalWorkerTestSuite) TestDecisionTaskHandler() {
	params := workerExecutionParameters{
		Identity: "identity",
		Logger:   getTestLogger(s.T()),
	}
	s.testDecisionTaskHandlerHelper(params)
}

func (s *internalWorkerTestSuite) TestDecisionTaskHandler_WithDataConverter() {
	params := workerExecutionParameters{
		Identity:      "identity",
		Logger:        getTestLogger(s.T()),
		DataConverter: newTestDataConverter(),
	}
	s.testDecisionTaskHandlerHelper(params)
}

// testSampleWorkflow
func sampleWorkflowExecute(ctx Context, input []byte) (result []byte, err error) {
	ExecuteActivity(ctx, testActivityByteArgs, input)
	ExecuteActivity(ctx, testActivityMultipleArgs, 2, []string{"test"}, true)
	ExecuteActivity(ctx, testActivityMultipleArgsWithStruct, -8, newTestActivityArg())
	return []byte("Done"), nil
}

// test activity1
func testActivityByteArgs(ctx context.Context, input []byte) ([]byte, error) {
	fmt.Println("Executing Activity1")
	return nil, nil
}

// test testActivityMultipleArgs
func testActivityMultipleArgs(context.Context, int, []string, bool) ([]byte, error) {
	fmt.Println("Executing Activity2")
	return nil, nil
}

// test testActivityMultipleArgsWithStruct
func testActivityMultipleArgsWithStruct(_ context.Context, i int, s testActivityArg) ([]byte, error) {
	fmt.Printf("Executing testActivityMultipleArgsWithStruct: %d, %v\n", i, s)
	return nil, nil
}

func (s *internalWorkerTestSuite) TestCreateWorker() {
	worker := createWorkerWithThrottle(s.service, float64(500.0), nil, nil)
	err := worker.Start()
	require.NoError(s.T(), err)
	time.Sleep(time.Millisecond * 200)
	worker.Stop()
}

func (s *internalWorkerTestSuite) TestCreateWorker_WithDataConverter() {
	worker := createWorkerWithDataConverter(s.service)
	err := worker.Start()
	require.NoError(s.T(), err)
	time.Sleep(time.Millisecond * 200)
	worker.Stop()
}

func (s *internalWorkerTestSuite) TestCreateShadowWorker() {
	worker := createShadowWorker(s.service, &ShadowOptions{})
	s.Nil(worker.workflowWorker)
	s.Nil(worker.activityWorker)
	s.Nil(worker.locallyDispatchedActivityWorker)
	s.Nil(worker.sessionWorker)
}

func (s *internalWorkerTestSuite) TestCreateWorkerRun() {
	// Create service endpoint
	mockCtrl := gomock.NewController(s.T())
	service := api.NewMockInterface(mockCtrl)

	worker := createWorker(service)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		worker.Run()
	}()
	time.Sleep(time.Millisecond * 200)
	p, err := os.FindProcess(os.Getpid())
	assert.NoError(s.T(), err)
	assert.NoError(s.T(), p.Signal(os.Interrupt))
	wg.Wait()
}

func (s *internalWorkerTestSuite) TestNoActivitiesOrWorkflows() {
	t := s.T()
	w := createWorker(s.service)
	w.registry = newRegistry()
	assert.Empty(t, w.registry.getRegisteredActivities())
	assert.Empty(t, w.registry.getRegisteredWorkflowTypes())
	assert.NoError(t, w.Start())
}

func (s *internalWorkerTestSuite) TestWorkerStartFailsWithInvalidDomain() {
	t := s.T()
	testCases := []struct {
		domainErr  error
		isErrFatal bool
	}{
		{&api.EntityNotExistsError{}, true},
		{&api.BadRequestError{}, true},
		{&api.InternalServiceError{}, false},
		{errors.New("unknown"), false},
	}

	mockCtrl := gomock.NewController(t)

	for _, tc := range testCases {
		service := api.NewMockInterface(mockCtrl)
		service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions...).Return(nil, tc.domainErr).Do(
			func(ctx context.Context, request *apiv1.DescribeDomainRequest, opts ...yarpc.CallOption) {
				// log
			}).Times(2)

		worker := createWorker(service)
		if tc.isErrFatal {
			err := worker.Start()
			assert.Error(t, err, "worker.start() MUST fail when domain is invalid")
			errC := make(chan error)
			go func() { errC <- worker.Run() }()
			select {
			case e := <-errC:
				assert.Error(t, e, "worker.Run() MUST fail when domain is invalid")
			case <-time.After(time.Second):
				assert.Fail(t, "worker.Run() MUST fail when domain is invalid")
			}
			continue
		}
		err := worker.Start()
		assert.NoError(t, err, "worker.Start() failed unexpectedly")
		worker.Stop()
	}
}

func (s *internalWorkerTestSuite) TestStartShadowWorkerFailWithInvalidOptions() {
	invalidOptions := []*ShadowOptions{
		{
			Mode: ShadowModeContinuous,
		},
		{
			WorkflowQuery: "workflow query",
			WorkflowTypes: []string{"workflowTypeName"},
		},
	}

	for _, opt := range invalidOptions {
		worker := createShadowWorker(s.service, opt)
		err := worker.Start()
		assert.Error(s.T(), err, "worker.Start() should fail given invalid shadow options")
	}
}

func ofPollForActivityTaskRequest(tps float64) gomock.Matcher {
	return &mockPollForActivityTaskRequest{tps: tps}
}

type mockPollForActivityTaskRequest struct {
	tps float64
}

func (m *mockPollForActivityTaskRequest) Matches(x interface{}) bool {
	v, ok := x.(*apiv1.PollForActivityTaskRequest)
	if !ok {
		return false
	}
	return v.TaskListMetadata.MaxTasksPerSecond.Value == m.tps
}

func (m *mockPollForActivityTaskRequest) String() string {
	return "PollForActivityTaskRequest"
}

func createWorker(
	service *api.MockInterface,
) *aggregatedWorker {
	return createWorkerWithThrottle(service, float64(0.0), nil, nil)
}

func createShadowWorker(
	service *api.MockInterface,
	shadowOptions *ShadowOptions,
) *aggregatedWorker {
	return createWorkerWithThrottle(service, float64(0.0), nil, shadowOptions)
}

func createWorkerWithThrottle(
	service *api.MockInterface,
	activitiesPerSecond float64,
	dc DataConverter,
	shadowOptions *ShadowOptions,
) *aggregatedWorker {
	domain := "testDomain"
	domainStatus := apiv1.DomainStatus_DOMAIN_STATUS_REGISTERED
	domainDesc := &apiv1.DescribeDomainResponse{
		Domain: &apiv1.Domain{
			Name:   domain,
			Status: domainStatus,
		},
	}
	// mocks
	service.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), callOptions...).Return(domainDesc, nil).Do(
		func(ctx context.Context, request *apiv1.DescribeDomainRequest, opts ...yarpc.CallOption) {
			// log
		}).AnyTimes()

	activityTask := &apiv1.PollForActivityTaskResponse{}
	expectedActivitiesPerSecond := activitiesPerSecond
	if expectedActivitiesPerSecond == 0.0 {
		expectedActivitiesPerSecond = defaultTaskListActivitiesPerSecond
	}
	service.EXPECT().PollForActivityTask(
		gomock.Any(), ofPollForActivityTaskRequest(expectedActivitiesPerSecond), callOptions...,
	).Return(activityTask, nil).AnyTimes()
	service.EXPECT().RespondActivityTaskCompleted(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).AnyTimes()

	decisionTask := &apiv1.PollForDecisionTaskResponse{}
	service.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), callOptions...).Return(decisionTask, nil).AnyTimes()
	service.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).AnyTimes()

	// Configure worker options.
	workerOptions := WorkerOptions{}
	workerOptions.WorkerActivitiesPerSecond = 20
	workerOptions.TaskListActivitiesPerSecond = activitiesPerSecond
	if dc != nil {
		workerOptions.DataConverter = dc
	}
	workerOptions.EnableSessionWorker = true

	if shadowOptions != nil {
		workerOptions.EnableShadowWorker = true
		workerOptions.ShadowOptions = *shadowOptions
	}

	// Start Worker.
	worker := NewWorker(
		service,
		domain,
		"testGroupName2",
		workerOptions)
	return worker
}

func createWorkerWithDataConverter(
	service *api.MockInterface,
) *aggregatedWorker {
	return createWorkerWithThrottle(service, float64(0.0), newTestDataConverter(), nil)
}

func (s *internalWorkerTestSuite) testCompleteActivityHelper(opt *ClientOptions) {
	t := s.T()
	domain := "testDomain"
	wfClient := NewClient(s.service, domain, opt)
	var completedRequest, canceledRequest, failedRequest interface{}
	s.service.EXPECT().RespondActivityTaskCompleted(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).Do(
		func(ctx context.Context, request *apiv1.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption) {
			completedRequest = request
		})
	s.service.EXPECT().RespondActivityTaskCanceled(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).Do(
		func(ctx context.Context, request *apiv1.RespondActivityTaskCanceledRequest, opts ...yarpc.CallOption) {
			canceledRequest = request
		})
	s.service.EXPECT().RespondActivityTaskFailed(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).Do(
		func(ctx context.Context, request *apiv1.RespondActivityTaskFailedRequest, opts ...yarpc.CallOption) {
			failedRequest = request
		})

	wfClient.CompleteActivity(context.Background(), []byte("task-token"), nil, nil)
	require.NotNil(t, completedRequest)

	wfClient.CompleteActivity(context.Background(), []byte("task-token"), nil, NewCanceledError())
	require.NotNil(t, canceledRequest)

	wfClient.CompleteActivity(context.Background(), []byte("task-token"), nil, errors.New(""))
	require.NotNil(t, failedRequest)
}

func (s *internalWorkerTestSuite) TestCompleteActivity() {
	s.testCompleteActivityHelper(nil)
}

func (s *internalWorkerTestSuite) TestCompleteActivity_WithDataConverter() {
	opt := &ClientOptions{DataConverter: newTestDataConverter()}
	s.testCompleteActivityHelper(opt)
}

func (s *internalWorkerTestSuite) TestCompleteActivityById() {
	t := s.T()
	domain := "testDomain"
	wfClient := NewClient(s.service, domain, nil)
	var completedRequest, canceledRequest, failedRequest interface{}
	s.service.EXPECT().RespondActivityTaskCompletedByID(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).Do(
		func(ctx context.Context, request *apiv1.RespondActivityTaskCompletedByIDRequest, opts ...yarpc.CallOption) {
			completedRequest = request
		})
	s.service.EXPECT().RespondActivityTaskCanceledByID(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).Do(
		func(ctx context.Context, request *apiv1.RespondActivityTaskCanceledByIDRequest, opts ...yarpc.CallOption) {
			canceledRequest = request
		})
	s.service.EXPECT().RespondActivityTaskFailedByID(gomock.Any(), gomock.Any(), callOptions...).Return(nil, nil).Do(
		func(ctx context.Context, request *apiv1.RespondActivityTaskFailedByIDRequest, opts ...yarpc.CallOption) {
			failedRequest = request
		})

	workflowID := "wid"
	runID := ""
	activityID := "aid"

	wfClient.CompleteActivityByID(context.Background(), domain, workflowID, runID, activityID, nil, nil)
	require.NotNil(t, completedRequest)

	wfClient.CompleteActivityByID(context.Background(), domain, workflowID, runID, activityID, nil, NewCanceledError())
	require.NotNil(t, canceledRequest)

	wfClient.CompleteActivityByID(context.Background(), domain, workflowID, runID, activityID, nil, errors.New(""))
	require.NotNil(t, failedRequest)
}

func (s *internalWorkerTestSuite) TestRecordActivityHeartbeat() {
	domain := "testDomain"
	wfClient := NewClient(s.service, domain, nil)
	var heartbeatRequest *apiv1.RecordActivityTaskHeartbeatRequest
	heartbeatResponse := apiv1.RecordActivityTaskHeartbeatResponse{CancelRequested: false}
	s.service.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).Return(&heartbeatResponse, nil).
		Do(func(ctx context.Context, request *apiv1.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) {
			heartbeatRequest = request
		}).Times(2)

	wfClient.RecordActivityHeartbeat(context.Background(), nil)
	wfClient.RecordActivityHeartbeat(context.Background(), nil, "testStack", "customerObjects", 4)
	require.NotNil(s.T(), heartbeatRequest)
}

func (s *internalWorkerTestSuite) TestRecordActivityHeartbeat_WithDataConverter() {
	t := s.T()
	domain := "testDomain"
	dc := newTestDataConverter()
	opt := &ClientOptions{DataConverter: dc}
	wfClient := NewClient(s.service, domain, opt)
	var heartbeatRequest *apiv1.RecordActivityTaskHeartbeatRequest
	heartbeatResponse := apiv1.RecordActivityTaskHeartbeatResponse{CancelRequested: false}
	detail1 := "testStack"
	detail2 := testStruct{"abc", 123}
	detail3 := 4
	encodedDetail, err := dc.ToData(detail1, detail2, detail3)
	require.Nil(t, err)
	s.service.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), callOptions...).Return(&heartbeatResponse, nil).
		Do(func(ctx context.Context, request *apiv1.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) {
			heartbeatRequest = request
			require.Equal(t, encodedDetail, request.Details.GetData())
		}).Times(1)

	wfClient.RecordActivityHeartbeat(context.Background(), nil, detail1, detail2, detail3)
	require.NotNil(t, heartbeatRequest)
}

func (s *internalWorkerTestSuite) TestRecordActivityHeartbeatByID() {
	domain := "testDomain"
	wfClient := NewClient(s.service, domain, nil)
	var heartbeatRequest *apiv1.RecordActivityTaskHeartbeatByIDRequest
	heartbeatResponse := apiv1.RecordActivityTaskHeartbeatByIDResponse{CancelRequested: false}
	s.service.EXPECT().RecordActivityTaskHeartbeatByID(gomock.Any(), gomock.Any(), callOptions...).Return(&heartbeatResponse, nil).
		Do(func(ctx context.Context, request *apiv1.RecordActivityTaskHeartbeatByIDRequest, opts ...yarpc.CallOption) {
			heartbeatRequest = request
		}).Times(2)

	wfClient.RecordActivityHeartbeatByID(context.Background(), domain, "wid", "rid", "aid")
	wfClient.RecordActivityHeartbeatByID(context.Background(), domain, "wid", "rid", "aid",
		"testStack", "customerObjects", 4)
	require.NotNil(s.T(), heartbeatRequest)
}

type activitiesCallingOptionsWorkflow struct {
	t *testing.T
}

func (w activitiesCallingOptionsWorkflow) Execute(ctx Context, input []byte) (result []byte, err error) {
	ao := ActivityOptions{
		ScheduleToStartTimeout: 10 * time.Second,
		StartToCloseTimeout:    5 * time.Second,
	}
	ctx = WithActivityOptions(ctx, ao)

	// By functions.
	err = ExecuteActivity(ctx, testActivityByteArgs, input).Get(ctx, nil)
	require.NoError(w.t, err, err)

	err = ExecuteActivity(ctx, testActivityMultipleArgs, 2, []string{"test"}, true).Get(ctx, nil)
	require.NoError(w.t, err, err)

	err = ExecuteActivity(ctx, testActivityMultipleArgsWithStruct, -8, newTestActivityArg()).Get(ctx, nil)
	require.NoError(w.t, err, err)

	err = ExecuteActivity(ctx, testActivityNoResult, 2, "test").Get(ctx, nil)
	require.NoError(w.t, err, err)

	err = ExecuteActivity(ctx, testActivityNoContextArg, 2, "test").Get(ctx, nil)
	require.NoError(w.t, err, err)

	f := ExecuteActivity(ctx, testActivityReturnByteArray)
	var r []byte
	err = f.Get(ctx, &r)
	require.NoError(w.t, err, err)
	require.Equal(w.t, []byte("testActivity"), r)

	f = ExecuteActivity(ctx, testActivityReturnInt)
	var rInt int
	err = f.Get(ctx, &rInt)
	require.NoError(w.t, err, err)
	require.Equal(w.t, 5, rInt)

	f = ExecuteActivity(ctx, testActivityReturnString)
	var rString string
	err = f.Get(ctx, &rString)

	require.NoError(w.t, err, err)
	require.Equal(w.t, "testActivity", rString)

	f = ExecuteActivity(ctx, testActivityReturnEmptyString)
	var r2String string
	err = f.Get(ctx, &r2String)
	require.NoError(w.t, err, err)
	require.Equal(w.t, "", r2String)

	f = ExecuteActivity(ctx, testActivityReturnEmptyStruct)
	var r2Struct testActivityResult
	err = f.Get(ctx, &r2Struct)
	require.NoError(w.t, err, err)
	require.Equal(w.t, testActivityResult{}, r2Struct)

	f = ExecuteActivity(ctx, testActivityReturnNilStructPtr)
	var rStructPtr *testActivityResult
	err = f.Get(ctx, &rStructPtr)
	require.NoError(w.t, err, err)
	require.True(w.t, rStructPtr == nil)

	f = ExecuteActivity(ctx, testActivityReturnStructPtr)
	err = f.Get(ctx, &rStructPtr)
	require.NoError(w.t, err, err)
	require.Equal(w.t, *rStructPtr, testActivityResult{Index: 10})

	f = ExecuteActivity(ctx, testActivityReturnNilStructPtrPtr)
	var rStruct2Ptr **testActivityResult
	err = f.Get(ctx, &rStruct2Ptr)
	require.NoError(w.t, err, err)
	require.True(w.t, rStruct2Ptr == nil)

	f = ExecuteActivity(ctx, testActivityReturnStructPtrPtr)
	err = f.Get(ctx, &rStruct2Ptr)
	require.NoError(w.t, err, err)
	require.True(w.t, **rStruct2Ptr == testActivityResult{Index: 10})

	// By names.
	err = ExecuteActivity(ctx, "go.uber.org/cadence/v2/internal.testActivityByteArgs", input).Get(ctx, nil)
	require.NoError(w.t, err, err)

	err = ExecuteActivity(ctx, "testActivityMultipleArgs", 2, []string{"test"}, true).Get(ctx, nil)
	require.NoError(w.t, err, err)

	err = ExecuteActivity(ctx, "go.uber.org/cadence/v2/internal.testActivityNoResult", 2, "test").Get(ctx, nil)
	require.NoError(w.t, err, err)

	err = ExecuteActivity(ctx, "go.uber.org/cadence/v2/internal.testActivityNoContextArg", 2, "test").Get(ctx, nil)
	require.NoError(w.t, err, err)

	f = ExecuteActivity(ctx, "go.uber.org/cadence/v2/internal.testActivityReturnString")
	err = f.Get(ctx, &rString)
	require.NoError(w.t, err, err)
	require.Equal(w.t, "testActivity", rString, rString)

	f = ExecuteActivity(ctx, "go.uber.org/cadence/v2/internal.testActivityReturnEmptyString")
	var r2sString string
	err = f.Get(ctx, &r2String)
	require.NoError(w.t, err, err)
	require.Equal(w.t, "", r2sString)

	f = ExecuteActivity(ctx, "go.uber.org/cadence/v2/internal.testActivityReturnEmptyStruct")
	err = f.Get(ctx, &r2Struct)
	require.NoError(w.t, err, err)
	require.Equal(w.t, testActivityResult{}, r2Struct)

	return []byte("Done"), nil
}

// test testActivityNoResult
func testActivityNoResult(ctx context.Context, arg1 int, arg2 string) error {
	return nil
}

// test testActivityNoContextArg
func testActivityNoContextArg(arg1 int, arg2 string) error {
	return nil
}

// test testActivityReturnByteArray
func testActivityReturnByteArray() ([]byte, error) {
	return []byte("testActivity"), nil
}

// testActivityReturnInt
func testActivityReturnInt() (int, error) {
	return 5, nil
}

// testActivityReturnString
func testActivityReturnString() (string, error) {
	return "testActivity", nil
}

// testActivityReturnEmptyString
func testActivityReturnEmptyString() (string, error) {
	// Return is mocked to retrun nil from server.
	// expect to convert it to appropriate default value.
	return "", nil
}

type testActivityArg struct {
	Index    int
	Name     string
	Data     []byte
	IndexPtr *int
	NamePtr  *string
	DataPtr  *[]byte
}

type testActivityResult struct {
	Index int
}

func newTestActivityArg() *testActivityArg {
	name := "JohnSmith"
	index := 22
	data := []byte{22, 8, 78}

	return &testActivityArg{
		Name:     name,
		Index:    index,
		Data:     data,
		NamePtr:  &name,
		IndexPtr: &index,
		DataPtr:  &data,
	}
}

// testActivityReturnEmptyStruct
func testActivityReturnEmptyStruct() (testActivityResult, error) {
	// Return is mocked to retrun nil from server.
	// expect to convert it to appropriate default value.
	return testActivityResult{}, nil
}
func testActivityReturnNilStructPtr() (*testActivityResult, error) {
	return nil, nil
}
func testActivityReturnStructPtr() (*testActivityResult, error) {
	return &testActivityResult{Index: 10}, nil
}
func testActivityReturnNilStructPtrPtr() (**testActivityResult, error) {
	return nil, nil
}
func testActivityReturnStructPtrPtr() (**testActivityResult, error) {
	r := &testActivityResult{Index: 10}
	return &r, nil
}

func TestVariousActivitySchedulingOption(t *testing.T) {
	w := &activitiesCallingOptionsWorkflow{t: t}

	testVariousActivitySchedulingOption(t, w.Execute)
	testVariousActivitySchedulingOptionWithDataConverter(t, w.Execute)
}

func testVariousActivitySchedulingOption(t *testing.T, wf interface{}) {
	ts := &WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(wf)
	testInternalWorkerRegisterWithTestEnv(env)
	env.ExecuteWorkflow(wf, []byte{1, 2})
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}

func testVariousActivitySchedulingOptionWithDataConverter(t *testing.T, wf interface{}) {
	ts := &WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(WorkerOptions{DataConverter: newTestDataConverter()})
	env.RegisterWorkflow(wf)
	testInternalWorkerRegisterWithTestEnv(env)
	env.ExecuteWorkflow(wf, []byte{1, 2})
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}

func testWorkflowSample(ctx Context, input []byte) (result []byte, err error) {
	return nil, nil
}

func testWorkflowMultipleArgs(ctx Context, arg1 int, arg2 string, arg3 bool) (result []byte, err error) {
	return nil, nil
}

func testWorkflowNoArgs(ctx Context) (result []byte, err error) {
	return nil, nil
}

func testWorkflowReturnInt(ctx Context) (result int, err error) {
	return 5, nil
}

func testWorkflowReturnString(ctx Context, arg1 int) (result string, err error) {
	return "Done", nil
}

type testWorkflowResult struct {
	V int
}

func testWorkflowReturnStruct(ctx Context, arg1 int) (result testWorkflowResult, err error) {
	return testWorkflowResult{}, nil
}

func testWorkflowReturnStructPtr(ctx Context, arg1 int) (result *testWorkflowResult, err error) {
	return &testWorkflowResult{}, nil
}

func testWorkflowReturnStructPtrPtr(ctx Context, arg1 int) (result **testWorkflowResult, err error) {
	return nil, nil
}

func TestRegisterVariousWorkflowTypes(t *testing.T) {
	r := newRegistry()
	r.RegisterWorkflow(testWorkflowSample)
	r.RegisterWorkflow(testWorkflowMultipleArgs)
	r.RegisterWorkflow(testWorkflowNoArgs)
	r.RegisterWorkflow(testWorkflowReturnInt)
	r.RegisterWorkflow(testWorkflowReturnString)
	r.RegisterWorkflow(testWorkflowReturnStruct)
	r.RegisterWorkflow(testWorkflowReturnStructPtr)
	r.RegisterWorkflow(testWorkflowReturnStructPtrPtr)
}

type testErrorDetails struct {
	T string
}

func testActivityErrorWithDetailsHelper(ctx context.Context, t *testing.T, dataConverter DataConverter) {
	a1 := activityExecutor{
		name: "test",
		fn: func(arg1 int) (err error) {
			return NewCustomError("testReason", "testStringDetails")
		}}
	_, e := a1.Execute(ctx, testEncodeFunctionArgs(dataConverter, 1))
	require.Error(t, e)
	errWD := e.(*CustomError)
	require.Equal(t, "testReason", errWD.Reason())
	var strDetails string
	errWD.Details(&strDetails)
	require.Equal(t, "testStringDetails", strDetails)

	a2 := activityExecutor{
		name: "test",
		fn: func(arg1 int) (err error) {
			return NewCustomError("testReason", testErrorDetails{T: "testErrorStack"})
		}}
	_, e = a2.Execute(ctx, testEncodeFunctionArgs(dataConverter, 1))
	require.Error(t, e)
	errWD = e.(*CustomError)
	require.Equal(t, "testReason", errWD.Reason())
	var td testErrorDetails
	errWD.Details(&td)
	require.Equal(t, testErrorDetails{T: "testErrorStack"}, td)

	a3 := activityExecutor{
		name: "test",
		fn: func(arg1 int) (result string, err error) {
			return "testResult", NewCustomError("testReason", testErrorDetails{T: "testErrorStack3"})
		}}
	encResult, e := a3.Execute(ctx, testEncodeFunctionArgs(dataConverter, 1))
	var result string
	err := dataConverter.FromData(encResult, &result)
	require.NoError(t, err)
	require.Equal(t, "testResult", result)
	require.Error(t, e)
	errWD = e.(*CustomError)
	require.Equal(t, "testReason", errWD.Reason())
	errWD.Details(&td)
	require.Equal(t, testErrorDetails{T: "testErrorStack3"}, td)

	a4 := activityExecutor{
		name: "test",
		fn: func(arg1 int) (result string, err error) {
			return "testResult4", NewCustomError("testReason", "testMultipleString", testErrorDetails{T: "testErrorStack4"})
		}}
	encResult, e = a4.Execute(ctx, testEncodeFunctionArgs(dataConverter, 1))
	err = dataConverter.FromData(encResult, &result)
	require.NoError(t, err)
	require.Equal(t, "testResult4", result)
	require.Error(t, e)
	errWD = e.(*CustomError)
	require.Equal(t, "testReason", errWD.Reason())
	var ed string
	errWD.Details(&ed, &td)
	require.Equal(t, "testMultipleString", ed)
	require.Equal(t, testErrorDetails{T: "testErrorStack4"}, td)
}

func TestActivityErrorWithDetailsWithDataConverter(t *testing.T) {
	dc := newTestDataConverter()
	ctx := context.WithValue(context.Background(), activityEnvContextKey, &activityEnvironment{dataConverter: dc})
	testActivityErrorWithDetailsHelper(ctx, t, dc)
}

func testActivityCancelledErrorHelper(ctx context.Context, t *testing.T, dataConverter DataConverter) {
	a1 := activityExecutor{
		name: "test",
		fn: func(arg1 int) (err error) {
			return NewCanceledError("testCancelStringDetails")
		}}
	_, e := a1.Execute(ctx, testEncodeFunctionArgs(dataConverter, 1))
	require.Error(t, e)
	errWD := e.(*CanceledError)
	var strDetails string
	errWD.Details(&strDetails)
	require.Equal(t, "testCancelStringDetails", strDetails)

	a2 := activityExecutor{
		name: "test",
		fn: func(arg1 int) (err error) {
			return NewCanceledError(testErrorDetails{T: "testCancelErrorStack"})
		}}
	_, e = a2.Execute(ctx, testEncodeFunctionArgs(dataConverter, 1))
	require.Error(t, e)
	errWD = e.(*CanceledError)
	var td testErrorDetails
	errWD.Details(&td)
	require.Equal(t, testErrorDetails{T: "testCancelErrorStack"}, td)

	a3 := activityExecutor{
		name: "test",
		fn: func(arg1 int) (result string, err error) {
			return "testResult", NewCanceledError(testErrorDetails{T: "testErrorStack3"})
		}}
	encResult, e := a3.Execute(ctx, testEncodeFunctionArgs(dataConverter, 1))
	var r string
	err := dataConverter.FromData(encResult, &r)
	require.NoError(t, err)
	require.Equal(t, "testResult", r)
	require.Error(t, e)
	errWD = e.(*CanceledError)
	errWD.Details(&td)
	require.Equal(t, testErrorDetails{T: "testErrorStack3"}, td)

	a4 := activityExecutor{
		name: "test",
		fn: func(arg1 int) (result string, err error) {
			return "testResult4", NewCanceledError("testMultipleString", testErrorDetails{T: "testErrorStack4"})
		}}
	encResult, e = a4.Execute(ctx, testEncodeFunctionArgs(dataConverter, 1))
	err = dataConverter.FromData(encResult, &r)
	require.NoError(t, err)
	require.Equal(t, "testResult4", r)
	require.Error(t, e)
	errWD = e.(*CanceledError)
	var ed string
	errWD.Details(&ed, &td)
	require.Equal(t, "testMultipleString", ed)
	require.Equal(t, testErrorDetails{T: "testErrorStack4"}, td)
}

func TestActivityCancelledErrorWithDataConverter(t *testing.T) {
	dc := newTestDataConverter()
	ctx := context.WithValue(context.Background(), activityEnvContextKey, &activityEnvironment{dataConverter: dc})
	testActivityCancelledErrorHelper(ctx, t, dc)
}

func testActivityExecutionVariousTypesHelper(ctx context.Context, t *testing.T, dataConverter DataConverter) {
	a1 := activityExecutor{
		fn: func(ctx context.Context, arg1 string) (*testWorkflowResult, error) {
			return &testWorkflowResult{V: 1}, nil
		}}
	encResult, e := a1.Execute(ctx, testEncodeFunctionArgs(dataConverter, "test"))
	require.NoError(t, e)
	var r *testWorkflowResult
	err := dataConverter.FromData(encResult, &r)
	require.NoError(t, err)
	require.Equal(t, 1, r.V)

	a2 := activityExecutor{
		fn: func(ctx context.Context, arg1 *testWorkflowResult) (*testWorkflowResult, error) {
			return &testWorkflowResult{V: 2}, nil
		}}
	encResult, e = a2.Execute(ctx, testEncodeFunctionArgs(dataConverter, r))
	require.NoError(t, e)
	err = dataConverter.FromData(encResult, &r)
	require.NoError(t, err)
	require.Equal(t, 2, r.V)
}

func TestActivityExecutionVariousTypesWithDataConverter(t *testing.T) {
	dc := newTestDataConverter()
	ctx := context.WithValue(context.Background(), activityEnvContextKey, &activityEnvironment{
		dataConverter: dc,
	})
	testActivityExecutionVariousTypesHelper(ctx, t, dc)
}

func TestActivityNilArgs(t *testing.T) {
	nilErr := errors.New("nils")
	activityFn := func(name string, idx int, strptr *string) error {
		if name == "" && idx == 0 && strptr == nil {
			return nilErr
		}
		return nil
	}

	args := []interface{}{nil, nil, nil}
	_, err := getValidatedActivityFunction(activityFn, args, newRegistry())
	require.NoError(t, err)

	dataConverter := getDefaultDataConverter()
	data, _ := encodeArgs(dataConverter, args)
	reflectArgs, err := decodeArgs(dataConverter, reflect.TypeOf(activityFn), data)
	require.NoError(t, err)

	reflectResults := reflect.ValueOf(activityFn).Call(reflectArgs)
	require.Equal(t, nilErr, reflectResults[0].Interface())
}

func TestWorkerOptionDefaults(t *testing.T) {
	domain := "worker-options-test"
	taskList := "worker-options-tl"
	aggWorker := newAggregatedWorker(nil, domain, taskList, WorkerOptions{})
	decisionWorker := aggWorker.workflowWorker
	require.True(t, decisionWorker.executionParameters.Identity != "")
	require.NotNil(t, decisionWorker.executionParameters.Logger)
	require.NotNil(t, decisionWorker.executionParameters.MetricsScope)
	require.Nil(t, decisionWorker.executionParameters.ContextPropagators)

	expected := workerExecutionParameters{
		TaskList:                             taskList,
		MaxConcurrentActivityPollers:         defaultConcurrentPollRoutineSize,
		MaxConcurrentDecisionPollers:         defaultConcurrentPollRoutineSize,
		ConcurrentLocalActivityExecutionSize: defaultMaxConcurrentLocalActivityExecutionSize,
		ConcurrentActivityExecutionSize:      defaultMaxConcurrentActivityExecutionSize,
		ConcurrentDecisionTaskExecutionSize:  defaultMaxConcurrentTaskExecutionSize,
		WorkerActivitiesPerSecond:            defaultTaskListActivitiesPerSecond,
		WorkerDecisionTasksPerSecond:         defaultWorkerTaskExecutionRate,
		TaskListActivitiesPerSecond:          defaultTaskListActivitiesPerSecond,
		WorkerLocalActivitiesPerSecond:       defaultWorkerLocalActivitiesPerSecond,
		StickyScheduleToStartTimeout:         stickyDecisionScheduleToStartTimeoutSeconds * time.Second,
		DataConverter:                        getDefaultDataConverter(),
		Tracer:                               opentracing.NoopTracer{},
		Logger:                               decisionWorker.executionParameters.Logger,
		MetricsScope:                         decisionWorker.executionParameters.MetricsScope,
		Identity:                             decisionWorker.executionParameters.Identity,
		UserContext:                          decisionWorker.executionParameters.UserContext,
	}

	assertWorkerExecutionParamsEqual(t, expected, decisionWorker.executionParameters)

	activityWorker := aggWorker.activityWorker
	require.True(t, activityWorker.executionParameters.Identity != "")
	require.NotNil(t, activityWorker.executionParameters.Logger)
	require.NotNil(t, activityWorker.executionParameters.MetricsScope)
	require.Nil(t, activityWorker.executionParameters.ContextPropagators)
	assertWorkerExecutionParamsEqual(t, expected, activityWorker.executionParameters)
}

func TestWorkerOptionNonDefaults(t *testing.T) {
	domain := "worker-options-test"
	taskList := "worker-options-tl"

	options := WorkerOptions{
		Identity:                                "143@worker-options-test-1",
		TaskListActivitiesPerSecond:             8888,
		MaxConcurrentSessionExecutionSize:       3333,
		MaxConcurrentDecisionTaskExecutionSize:  2222,
		MaxConcurrentActivityExecutionSize:      1111,
		MaxConcurrentLocalActivityExecutionSize: 101,
		MaxConcurrentDecisionTaskPollers:        11,
		MaxConcurrentActivityTaskPollers:        12,
		WorkerLocalActivitiesPerSecond:          222,
		WorkerDecisionTasksPerSecond:            111,
		WorkerActivitiesPerSecond:               99,
		StickyScheduleToStartTimeout:            555 * time.Minute,
		DataConverter:                           &defaultDataConverter{},
		BackgroundActivityContext:               context.Background(),
		Logger:                                  zap.NewNop(),
		MetricsScope:                            tally.NoopScope,
		Tracer:                                  opentracing.NoopTracer{},
	}

	aggWorker := newAggregatedWorker(nil, domain, taskList, options)
	decisionWorker := aggWorker.workflowWorker
	require.True(t, len(decisionWorker.executionParameters.ContextPropagators) > 0)

	expected := workerExecutionParameters{
		TaskList:                             taskList,
		MaxConcurrentActivityPollers:         options.MaxConcurrentActivityTaskPollers,
		MaxConcurrentDecisionPollers:         options.MaxConcurrentDecisionTaskPollers,
		ConcurrentLocalActivityExecutionSize: options.MaxConcurrentLocalActivityExecutionSize,
		ConcurrentActivityExecutionSize:      options.MaxConcurrentActivityExecutionSize,
		ConcurrentDecisionTaskExecutionSize:  options.MaxConcurrentDecisionTaskExecutionSize,
		WorkerActivitiesPerSecond:            options.WorkerActivitiesPerSecond,
		WorkerDecisionTasksPerSecond:         options.WorkerDecisionTasksPerSecond,
		TaskListActivitiesPerSecond:          options.TaskListActivitiesPerSecond,
		WorkerLocalActivitiesPerSecond:       options.WorkerLocalActivitiesPerSecond,
		StickyScheduleToStartTimeout:         options.StickyScheduleToStartTimeout,
		DataConverter:                        options.DataConverter,
		Tracer:                               options.Tracer,
		Logger:                               options.Logger,
		MetricsScope:                         options.MetricsScope,
		Identity:                             options.Identity,
	}

	assertWorkerExecutionParamsEqual(t, expected, decisionWorker.executionParameters)

	activityWorker := aggWorker.activityWorker
	require.True(t, len(activityWorker.executionParameters.ContextPropagators) > 0)
	assertWorkerExecutionParamsEqual(t, expected, activityWorker.executionParameters)
}

func assertWorkerExecutionParamsEqual(t *testing.T, paramsA workerExecutionParameters, paramsB workerExecutionParameters) {
	require.Equal(t, paramsA.TaskList, paramsA.TaskList)
	require.Equal(t, paramsA.Identity, paramsB.Identity)
	require.Equal(t, paramsA.DataConverter, paramsB.DataConverter)
	require.Equal(t, paramsA.Tracer, paramsB.Tracer)
	require.Equal(t, paramsA.ConcurrentLocalActivityExecutionSize, paramsB.ConcurrentLocalActivityExecutionSize)
	require.Equal(t, paramsA.ConcurrentActivityExecutionSize, paramsB.ConcurrentActivityExecutionSize)
	require.Equal(t, paramsA.ConcurrentDecisionTaskExecutionSize, paramsB.ConcurrentDecisionTaskExecutionSize)
	require.Equal(t, paramsA.WorkerActivitiesPerSecond, paramsB.WorkerActivitiesPerSecond)
	require.Equal(t, paramsA.WorkerDecisionTasksPerSecond, paramsB.WorkerDecisionTasksPerSecond)
	require.Equal(t, paramsA.TaskListActivitiesPerSecond, paramsB.TaskListActivitiesPerSecond)
	require.Equal(t, paramsA.StickyScheduleToStartTimeout, paramsB.StickyScheduleToStartTimeout)
	require.Equal(t, paramsA.MaxConcurrentDecisionPollers, paramsB.MaxConcurrentDecisionPollers)
	require.Equal(t, paramsA.MaxConcurrentActivityPollers, paramsB.MaxConcurrentActivityPollers)
	require.Equal(t, paramsA.NonDeterministicWorkflowPolicy, paramsB.NonDeterministicWorkflowPolicy)
	require.Equal(t, paramsA.EnableLoggingInReplay, paramsB.EnableLoggingInReplay)
	require.Equal(t, paramsA.DisableStickyExecution, paramsB.DisableStickyExecution)
}

/*
var testWorkflowID1 = s.WorkflowExecution{WorkflowId: common.StringPtr("testWID"), RunId: common.StringPtr("runID")}
var testWorkflowID2 = s.WorkflowExecution{WorkflowId: common.StringPtr("testWID2"), RunId: common.StringPtr("runID2")}
var thriftEncodingTests = []encodingTest{
	{&thriftEncoding{}, []interface{}{&testWorkflowID1}},
	{&thriftEncoding{}, []interface{}{&testWorkflowID1, &testWorkflowID2}},
	{&thriftEncoding{}, []interface{}{&testWorkflowID1, &testWorkflowID2, &testWorkflowID1}},
}

// TODO: Disable until thriftrw encoding support is added to cadence client.(follow up change)
func _TestThriftEncoding(t *testing.T) {
	// Success tests.
	for _, et := range thriftEncodingTests {
		data, err := et.encoding.Marshal(et.input)
		require.NoError(t, err)

		var result []interface{}
		for _, v := range et.input {
			arg := reflect.New(reflect.ValueOf(v).Type()).Interface()
			result = append(result, arg)
		}
		err = et.encoding.Unmarshal(data, result)
		require.NoError(t, err)

		for i := 0; i < len(et.input); i++ {
			vat := reflect.ValueOf(result[i]).Elem().Interface()
			require.Equal(t, et.input[i], vat)
		}
	}

	// Failure tests.
	enc := &thriftEncoding{}
	_, err := enc.Marshal([]interface{}{testWorkflowID1})
	require.Contains(t, err.Error(), "pointer to thrift.TStruct type is required")

	err = enc.Unmarshal([]byte("dummy"), []interface{}{testWorkflowID1})
	require.Contains(t, err.Error(), "pointer to pointer thrift.TStruct type is required")

	err = enc.Unmarshal([]byte("dummy"), []interface{}{&testWorkflowID1})
	require.Contains(t, err.Error(), "pointer to pointer thrift.TStruct type is required")

	_, err = enc.Marshal([]interface{}{testWorkflowID1, &testWorkflowID2})
	require.Contains(t, err.Error(), "pointer to thrift.TStruct type is required")

	err = enc.Unmarshal([]byte("dummy"), []interface{}{testWorkflowID1, &testWorkflowID2})
	require.Contains(t, err.Error(), "pointer to pointer thrift.TStruct type is required")
}
*/

// Encode function args
func testEncodeFunctionArgs(dataConverter DataConverter, args ...interface{}) []byte {
	input, err := encodeArgs(dataConverter, args)
	if err != nil {
		fmt.Println(err)
		panic("Failed to encode arguments")
	}
	return input
}

func TestIsNonRetriableError(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{
			err:      nil,
			expected: false,
		},
		{
			err:      &api.ServiceBusyError{},
			expected: false,
		},
		{
			err:      &api.BadRequestError{},
			expected: true,
		},
		{
			err:      &api.ClientVersionNotSupportedError{},
			expected: true,
		},
	}

	for _, test := range tests {
		require.Equal(t, test.expected, isNonRetriableError(test.err))
	}
}
