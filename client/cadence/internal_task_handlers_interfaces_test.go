package cadence

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/uber/tchannel-go/thrift"

	m "github.com/uber-go/cadence-client/.gen/go/shared"
	"github.com/uber-go/cadence-client/mocks"
)

type (
	PollLayerInterfacesTestSuite struct {
		suite.Suite
	}
)

// Workflow Context
type testWorkflowContext struct {
}

func (wc testWorkflowContext) WorkflowInfo() WorkflowInfo {
	return WorkflowInfo{}
}
func (wc testWorkflowContext) Complete(result []byte) {
}
func (wc testWorkflowContext) Fail(err error) {
}
func (wc testWorkflowContext) ScheduleActivityTask(parameters executeActivityParameters, callback resultHandler) {

}

// Activity Execution context
type testActivityExecutionContext struct {
}

func (ac testActivityExecutionContext) TaskToken() []byte {
	return []byte("")
}
func (ac testActivityExecutionContext) RecordActivityHeartbeat(details []byte) error {
	return nil
}

// Sample Workflow task handler
type sampleWorkflowTaskHandler struct {
	factory workflowDefinitionFactory
}

func (wth sampleWorkflowTaskHandler) ProcessWorkflowTask(task *m.PollForDecisionTaskResponse, emitStack bool) (*m.RespondDecisionTaskCompletedRequest, string, error) {
	return &m.RespondDecisionTaskCompletedRequest{
		TaskToken: task.TaskToken,
	}, "", nil
}

func (wth sampleWorkflowTaskHandler) LoadWorkflowThroughReplay(task *workflowTask) (workflowDefinition, error) {
	return &helloWorldWorkflow{}, nil
}

func newSampleWorkflowTaskHandler(factory workflowDefinitionFactory) *sampleWorkflowTaskHandler {
	return &sampleWorkflowTaskHandler{factory: factory}
}

// Sample ActivityTaskHandler
type sampleActivityTaskHandler struct {
	activityRegistry map[m.ActivityType]*Activity
}

func newSampleActivityTaskHandler(activityRegistry map[m.ActivityType]*Activity) *sampleActivityTaskHandler {
	return &sampleActivityTaskHandler{activityRegistry: activityRegistry}
}

func (ath sampleActivityTaskHandler) Execute(task *m.PollForActivityTaskResponse) ([]byte, error) {
	//activityImplementation := *ath.activityRegistry[*activityTask.task.ActivityType]
	activityImplementation := &greeterActivity{}
	return activityImplementation.Execute(context.Background(), task.Input)
}

// Test suite.
func (s *PollLayerInterfacesTestSuite) SetupTest() {
}

func TestPollLayerInterfacesTestSuite(t *testing.T) {
	suite.Run(t, new(PollLayerInterfacesTestSuite))
}

func (s *PollLayerInterfacesTestSuite) TestProcessWorkflowTaskInterface() {
	// Create service endpoint and get a workflow task.
	service := new(mocks.TChanWorkflowService)
	ctx, _ := thrift.NewContext(10)

	// mocks
	service.On("PollForDecisionTask", mock.Anything, mock.Anything).Return(&m.PollForDecisionTaskResponse{}, nil)
	service.On("RespondDecisionTaskCompleted", mock.Anything, mock.Anything).Return(nil)

	response, err := service.PollForDecisionTask(ctx, &m.PollForDecisionTaskRequest{})
	s.NoError(err)

	// Process task and respond to the service.
	taskHandler := newSampleWorkflowTaskHandler(testWorkflowDefinitionFactory)
	completionRequest, _, err := taskHandler.ProcessWorkflowTask(response, false)
	s.NoError(err)

	err = service.RespondDecisionTaskCompleted(ctx, completionRequest)
	s.NoError(err)
}

func (s *PollLayerInterfacesTestSuite) TestProcessActivityTaskInterface() {
	// Create service endpoint and get a activity task.
	service := new(mocks.TChanWorkflowService)
	ctx, _ := thrift.NewContext(10)

	// mocks
	service.On("PollForActivityTask", mock.Anything, mock.Anything).Return(&m.PollForActivityTaskResponse{}, nil)
	service.On("RespondActivityTaskCompleted", mock.Anything, mock.Anything).Return(nil)

	response, err := service.PollForActivityTask(ctx, &m.PollForActivityTaskRequest{})
	s.NoError(err)

	// Execute activity task and respond to the service.
	activationRegistry := make(map[m.ActivityType]*Activity)
	taskHandler := newSampleActivityTaskHandler(activationRegistry)
	result, err := taskHandler.Execute(response)
	request := convertActivityResultToRespondRequest("", nil, result, err)

	switch request := request.(type) {
	case *m.RespondActivityTaskCompletedRequest:
		s.NoError(err)
		err = service.RespondActivityTaskCompleted(ctx, request)
		s.NoError(err)
	case *m.RespondActivityTaskFailedRequest:
		s.Error(err)
		err = service.RespondActivityTaskFailed(ctx, request)
		s.NoError(err)
	}
}
