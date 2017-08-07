package metrics

import (
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uber/tchannel-go/thrift"
	m "go.uber.org/cadence/.gen/go/cadence"
	s "go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/mocks"
)

type testCase struct {
	serviceMethod    string
	callArgs         []interface{}
	mockReturns      []interface{}
	expectedCounters []string
}

func Test_Wrapper(t *testing.T) {
	ctx, _ := thrift.NewContext(time.Minute)
	tests := []testCase{
		// one case for each service call
		{"DeprecateDomain", []interface{}{ctx, &s.DeprecateDomainRequest{}}, []interface{}{nil}, []string{CadenceRequest}},
		{"DescribeDomain", []interface{}{ctx, &s.DescribeDomainRequest{}}, []interface{}{&s.DescribeDomainResponse{}, nil}, []string{CadenceRequest}},
		{"GetWorkflowExecutionHistory", []interface{}{ctx, &s.GetWorkflowExecutionHistoryRequest{}}, []interface{}{&s.GetWorkflowExecutionHistoryResponse{}, nil}, []string{CadenceRequest}},
		{"ListClosedWorkflowExecutions", []interface{}{ctx, &s.ListClosedWorkflowExecutionsRequest{}}, []interface{}{&s.ListClosedWorkflowExecutionsResponse{}, nil}, []string{CadenceRequest}},
		{"ListOpenWorkflowExecutions", []interface{}{ctx, &s.ListOpenWorkflowExecutionsRequest{}}, []interface{}{&s.ListOpenWorkflowExecutionsResponse{}, nil}, []string{CadenceRequest}},
		{"PollForActivityTask", []interface{}{ctx, &s.PollForActivityTaskRequest{}}, []interface{}{&s.PollForActivityTaskResponse{}, nil}, []string{CadenceRequest}},
		{"PollForDecisionTask", []interface{}{ctx, &s.PollForDecisionTaskRequest{}}, []interface{}{&s.PollForDecisionTaskResponse{}, nil}, []string{CadenceRequest}},
		{"RecordActivityTaskHeartbeat", []interface{}{ctx, &s.RecordActivityTaskHeartbeatRequest{}}, []interface{}{&s.RecordActivityTaskHeartbeatResponse{}, nil}, []string{CadenceRequest}},
		{"RegisterDomain", []interface{}{ctx, &s.RegisterDomainRequest{}}, []interface{}{nil}, []string{CadenceRequest}},
		{"RequestCancelWorkflowExecution", []interface{}{ctx, &s.RequestCancelWorkflowExecutionRequest{}}, []interface{}{nil}, []string{CadenceRequest}},
		{"RespondActivityTaskCanceled", []interface{}{ctx, &s.RespondActivityTaskCanceledRequest{}}, []interface{}{nil}, []string{CadenceRequest}},
		{"RespondActivityTaskCompleted", []interface{}{ctx, &s.RespondActivityTaskCompletedRequest{}}, []interface{}{nil}, []string{CadenceRequest}},
		{"RespondActivityTaskFailed", []interface{}{ctx, &s.RespondActivityTaskFailedRequest{}}, []interface{}{nil}, []string{CadenceRequest}},
		{"RespondDecisionTaskCompleted", []interface{}{ctx, &s.RespondDecisionTaskCompletedRequest{}}, []interface{}{nil}, []string{CadenceRequest}},
		{"SignalWorkflowExecution", []interface{}{ctx, &s.SignalWorkflowExecutionRequest{}}, []interface{}{nil}, []string{CadenceRequest}},
		{"StartWorkflowExecution", []interface{}{ctx, &s.StartWorkflowExecutionRequest{}}, []interface{}{&s.StartWorkflowExecutionResponse{}, nil}, []string{CadenceRequest}},
		{"TerminateWorkflowExecution", []interface{}{ctx, &s.TerminateWorkflowExecutionRequest{}}, []interface{}{nil}, []string{CadenceRequest}},
		{"UpdateDomain", []interface{}{ctx, &s.UpdateDomainRequest{}}, []interface{}{&s.UpdateDomainResponse{}, nil}, []string{CadenceRequest}},
		// one case of invalid request
		{"PollForActivityTask", []interface{}{ctx, &s.PollForActivityTaskRequest{}}, []interface{}{nil, &s.EntityNotExistsError{}}, []string{CadenceRequest, CadenceInvalidRequest}},
		// one case of server error
		{"PollForActivityTask", []interface{}{ctx, &s.PollForActivityTaskRequest{}}, []interface{}{nil, &s.InternalServiceError{}}, []string{CadenceRequest, CadenceError}},
	}

	for _, test := range tests {
		mockService, wrapperService, closer, reporter := newService()
		mockService.On(test.serviceMethod, mock.Anything, mock.Anything).Return(test.mockReturns...)
		inputs := make([]reflect.Value, len(test.callArgs))
		for i, arg := range test.callArgs {
			inputs[i] = reflect.ValueOf(arg)
		}
		method := reflect.ValueOf(wrapperService).MethodByName(test.serviceMethod)
		method.Call(inputs)
		closer.Close()
		assertMetrics(t, reporter, test.serviceMethod, test.expectedCounters)
	}
}

func assertMetrics(t *testing.T, reporter *capturingStatsReporter, scopeName string, counterNames []string) {
	require.Equal(t, len(counterNames), len(reporter.counts))
	for i, counterName := range counterNames {
		require.Equal(t, scopeName+"."+counterName, reporter.counts[i].name)
	}
	require.Equal(t, 1, len(reporter.timers))
	require.Equal(t, scopeName+"."+CadenceLatency, reporter.timers[0].name)
}

func newService() (mockService *mocks.TChanWorkflowService,
	wrapperService m.TChanWorkflowService,
	closer io.Closer,
	reporter *capturingStatsReporter,
) {
	mockService = &mocks.TChanWorkflowService{}
	isReplay := false
	scope, closer, reporter := newMetricsScope(&isReplay)
	wrapperService = NewWorkflowServiceWrapper(mockService, scope)
	return
}
