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

package metrics

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
	"go.uber.org/cadence/v2/internal/api"
	"go.uber.org/yarpc"
)

var (
	safeCharacters = []rune{'_'}

	sanitizeOptions = tally.SanitizeOptions{
		NameCharacters: tally.ValidCharacters{
			Ranges:     tally.AlphanumericRange,
			Characters: safeCharacters,
		},
		KeyCharacters: tally.ValidCharacters{
			Ranges:     tally.AlphanumericRange,
			Characters: safeCharacters,
		},
		ValueCharacters: tally.ValidCharacters{
			Ranges:     tally.AlphanumericRange,
			Characters: safeCharacters,
		},
		ReplacementCharacter: tally.DefaultReplacementCharacter,
	}
)

type testCase struct {
	serviceMethod    string
	callArgs         []interface{}
	mockReturns      []interface{}
	expectedCounters []string
}

func Test_Wrapper(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Minute)
	tests := []testCase{
		// one case for each service call
		{"DeprecateDomain", []interface{}{ctx, &apiv1.DeprecateDomainRequest{}}, []interface{}{&apiv1.DeprecateDomainResponse{}, nil}, []string{CadenceRequest}},
		{"DescribeDomain", []interface{}{ctx, &apiv1.DescribeDomainRequest{}}, []interface{}{&apiv1.DescribeDomainResponse{}, nil}, []string{CadenceRequest}},
		{"GetWorkflowExecutionHistory", []interface{}{ctx, &apiv1.GetWorkflowExecutionHistoryRequest{}}, []interface{}{&apiv1.GetWorkflowExecutionHistoryResponse{}, nil}, []string{CadenceRequest}},
		{"ListClosedWorkflowExecutions", []interface{}{ctx, &apiv1.ListClosedWorkflowExecutionsRequest{}}, []interface{}{&apiv1.ListClosedWorkflowExecutionsResponse{}, nil}, []string{CadenceRequest}},
		{"ListOpenWorkflowExecutions", []interface{}{ctx, &apiv1.ListOpenWorkflowExecutionsRequest{}}, []interface{}{&apiv1.ListOpenWorkflowExecutionsResponse{}, nil}, []string{CadenceRequest}},
		{"PollForActivityTask", []interface{}{ctx, &apiv1.PollForActivityTaskRequest{}}, []interface{}{&apiv1.PollForActivityTaskResponse{}, nil}, []string{CadenceRequest}},
		{"PollForDecisionTask", []interface{}{ctx, &apiv1.PollForDecisionTaskRequest{}}, []interface{}{&apiv1.PollForDecisionTaskResponse{}, nil}, []string{CadenceRequest}},
		{"RecordActivityTaskHeartbeat", []interface{}{ctx, &apiv1.RecordActivityTaskHeartbeatRequest{}}, []interface{}{&apiv1.RecordActivityTaskHeartbeatResponse{}, nil}, []string{CadenceRequest}},
		{"RegisterDomain", []interface{}{ctx, &apiv1.RegisterDomainRequest{}}, []interface{}{&apiv1.RegisterDomainResponse{}, nil}, []string{CadenceRequest}},
		{"RequestCancelWorkflowExecution", []interface{}{ctx, &apiv1.RequestCancelWorkflowExecutionRequest{}}, []interface{}{&apiv1.RequestCancelWorkflowExecutionResponse{}, nil}, []string{CadenceRequest}},
		{"RespondActivityTaskCanceled", []interface{}{ctx, &apiv1.RespondActivityTaskCanceledRequest{}}, []interface{}{&apiv1.RespondActivityTaskCanceledResponse{}, nil}, []string{CadenceRequest}},
		{"RespondActivityTaskCompleted", []interface{}{ctx, &apiv1.RespondActivityTaskCompletedRequest{}}, []interface{}{&apiv1.RespondActivityTaskCompletedResponse{}, nil}, []string{CadenceRequest}},
		{"RespondActivityTaskFailed", []interface{}{ctx, &apiv1.RespondActivityTaskFailedRequest{}}, []interface{}{&apiv1.RespondActivityTaskFailedResponse{}, nil}, []string{CadenceRequest}},
		{"RespondActivityTaskCanceledByID", []interface{}{ctx, &apiv1.RespondActivityTaskCanceledByIDRequest{}}, []interface{}{&apiv1.RespondActivityTaskCanceledByIDResponse{}, nil}, []string{CadenceRequest}},
		{"RespondActivityTaskCompletedByID", []interface{}{ctx, &apiv1.RespondActivityTaskCompletedByIDRequest{}}, []interface{}{&apiv1.RespondActivityTaskCompletedByIDResponse{}, nil}, []string{CadenceRequest}},
		{"RespondActivityTaskFailedByID", []interface{}{ctx, &apiv1.RespondActivityTaskFailedByIDRequest{}}, []interface{}{&apiv1.RespondActivityTaskFailedByIDResponse{}, nil}, []string{CadenceRequest}},
		{"RespondDecisionTaskCompleted", []interface{}{ctx, &apiv1.RespondDecisionTaskCompletedRequest{}}, []interface{}{nil, nil}, []string{CadenceRequest}},
		{"SignalWorkflowExecution", []interface{}{ctx, &apiv1.SignalWorkflowExecutionRequest{}}, []interface{}{&apiv1.SignalWorkflowExecutionResponse{}, nil}, []string{CadenceRequest}},
		{"StartWorkflowExecution", []interface{}{ctx, &apiv1.StartWorkflowExecutionRequest{}}, []interface{}{&apiv1.StartWorkflowExecutionResponse{}, nil}, []string{CadenceRequest}},
		{"TerminateWorkflowExecution", []interface{}{ctx, &apiv1.TerminateWorkflowExecutionRequest{}}, []interface{}{&apiv1.TerminateWorkflowExecutionResponse{}, nil}, []string{CadenceRequest}},
		{"ResetWorkflowExecution", []interface{}{ctx, &apiv1.ResetWorkflowExecutionRequest{}}, []interface{}{&apiv1.ResetWorkflowExecutionResponse{}, nil}, []string{CadenceRequest}},
		{"UpdateDomain", []interface{}{ctx, &apiv1.UpdateDomainRequest{}}, []interface{}{&apiv1.UpdateDomainResponse{}, nil}, []string{CadenceRequest}},
		// one case of invalid request
		{"PollForActivityTask", []interface{}{ctx, &apiv1.PollForActivityTaskRequest{}}, []interface{}{nil, &api.EntityNotExistsError{}}, []string{CadenceRequest, CadenceInvalidRequest}},
		// one case of server error
		{"PollForActivityTask", []interface{}{ctx, &apiv1.PollForActivityTaskRequest{}}, []interface{}{nil, &api.InternalServiceError{}}, []string{CadenceRequest, CadenceError}},
		{"QueryWorkflow", []interface{}{ctx, &apiv1.QueryWorkflowRequest{}}, []interface{}{nil, &api.InternalServiceError{}}, []string{CadenceRequest, CadenceError}},
		{"RespondQueryTaskCompleted", []interface{}{ctx, &apiv1.RespondQueryTaskCompletedRequest{}}, []interface{}{&apiv1.RespondQueryTaskCompletedResponse{}, &api.InternalServiceError{}}, []string{CadenceRequest, CadenceError}},
	}

	// run each test twice - once with the regular scope, once with a sanitized metrics scope
	for _, test := range tests {
		runWrapperTest(t, test, newService, assertMetrics, fmt.Sprintf("%v_normal", test.serviceMethod))
		runWrapperTest(t, test, newPromService, assertPromMetrics, fmt.Sprintf("%v_prom_sanitized", test.serviceMethod))
	}
}

func runWrapperTest(
	t *testing.T,
	test testCase,
	serviceFunc func(*testing.T) (*api.MockInterface, api.Interface, io.Closer, *CapturingStatsReporter),
	validationFunc func(*testing.T, *CapturingStatsReporter, string, []string),
	name string,
) {
	t.Run(name, func(t *testing.T) {
		t.Parallel()
		// gomock mutates the returns slice, which leads to different test values between the two runs.
		// copy the slice until gomock fixes it: https://github.com/golang/mock/issues/353
		returns := append(make([]interface{}, 0, len(test.mockReturns)), test.mockReturns...)

		mockService, wrapperService, closer, reporter := serviceFunc(t)
		switch test.serviceMethod {
		case "DeprecateDomain":
			mockService.EXPECT().DeprecateDomain(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "DescribeDomain":
			mockService.EXPECT().DescribeDomain(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "GetWorkflowExecutionHistory":
			mockService.EXPECT().GetWorkflowExecutionHistory(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "ListClosedWorkflowExecutions":
			mockService.EXPECT().ListClosedWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "ListOpenWorkflowExecutions":
			mockService.EXPECT().ListOpenWorkflowExecutions(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "PollForActivityTask":
			mockService.EXPECT().PollForActivityTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "PollForDecisionTask":
			mockService.EXPECT().PollForDecisionTask(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "RecordActivityTaskHeartbeat":
			mockService.EXPECT().RecordActivityTaskHeartbeat(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "RecordActivityTaskHeartbeatByID":
			mockService.EXPECT().RecordActivityTaskHeartbeatByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "RegisterDomain":
			mockService.EXPECT().RegisterDomain(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "RequestCancelWorkflowExecution":
			mockService.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "RespondActivityTaskCanceled":
			mockService.EXPECT().RespondActivityTaskCanceled(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "RespondActivityTaskCompleted":
			mockService.EXPECT().RespondActivityTaskCompleted(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "RespondActivityTaskFailed":
			mockService.EXPECT().RespondActivityTaskFailed(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "RespondActivityTaskCanceledByID":
			mockService.EXPECT().RespondActivityTaskCanceledByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "RespondActivityTaskCompletedByID":
			mockService.EXPECT().RespondActivityTaskCompletedByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "RespondActivityTaskFailedByID":
			mockService.EXPECT().RespondActivityTaskFailedByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "RespondDecisionTaskCompleted":
			mockService.EXPECT().RespondDecisionTaskCompleted(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "SignalWorkflowExecution":
			mockService.EXPECT().SignalWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "SignaWithStartlWorkflowExecution":
			mockService.EXPECT().SignalWithStartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "StartWorkflowExecution":
			mockService.EXPECT().StartWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "TerminateWorkflowExecution":
			mockService.EXPECT().TerminateWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "ResetWorkflowExecution":
			mockService.EXPECT().ResetWorkflowExecution(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "UpdateDomain":
			mockService.EXPECT().UpdateDomain(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "QueryWorkflow":
			mockService.EXPECT().QueryWorkflow(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		case "RespondQueryTaskCompleted":
			mockService.EXPECT().RespondQueryTaskCompleted(gomock.Any(), gomock.Any(), gomock.Any()).Return(returns...)
		}

		callOption := yarpc.CallOption{}
		inputs := make([]reflect.Value, len(test.callArgs))
		for i, arg := range test.callArgs {
			inputs[i] = reflect.ValueOf(arg)
		}
		inputs = append(inputs, reflect.ValueOf(callOption))
		method := reflect.ValueOf(wrapperService).MethodByName(test.serviceMethod)
		method.Call(inputs)
		require.NoError(t, closer.Close())
		validationFunc(t, reporter, test.serviceMethod, test.expectedCounters)
	})
}

func assertMetrics(t *testing.T, reporter *CapturingStatsReporter, methodName string, counterNames []string) {
	require.Equal(t, len(counterNames), len(reporter.counts))
	for _, name := range counterNames {
		counterName := CadenceMetricsPrefix + methodName + "." + name
		find := false
		// counters are not in order
		for _, counter := range reporter.counts {
			if counterName == counter.name {
				find = true
				break
			}
		}
		require.True(t, find)
	}
	require.Equal(t, 1, len(reporter.timers))
	require.Equal(t, CadenceMetricsPrefix+methodName+"."+CadenceLatency, reporter.timers[0].name)
}

func assertPromMetrics(t *testing.T, reporter *CapturingStatsReporter, methodName string, counterNames []string) {
	require.Equal(t, len(counterNames), len(reporter.counts))
	for _, name := range counterNames {
		counterName := makePromCompatible(CadenceMetricsPrefix + methodName + "." + name)
		find := false
		// counters are not in order
		for _, counter := range reporter.counts {
			if counterName == counter.name {
				find = true
				break
			}
		}
		require.True(t, find)
	}
	require.Equal(t, 1, len(reporter.timers))
	expected := makePromCompatible(CadenceMetricsPrefix + methodName + "." + CadenceLatency)
	require.Equal(t, expected, reporter.timers[0].name)
}

func makePromCompatible(name string) string {
	name = strings.Replace(name, "-", "_", -1)
	name = strings.Replace(name, ".", "_", -1)
	return name
}

func newService(t *testing.T) (
	mockService *api.MockInterface,
	wrapperService api.Interface,
	closer io.Closer,
	reporter *CapturingStatsReporter,
) {
	mockCtrl := gomock.NewController(t)
	mockService = api.NewMockInterface(mockCtrl)
	isReplay := false
	scope, closer, reporter := NewMetricsScope(&isReplay)
	wrapperService = NewWorkflowServiceWrapper(mockService, scope)
	return
}

func newPromService(t *testing.T) (
	mockService *api.MockInterface,
	wrapperService api.Interface,
	closer io.Closer,
	reporter *CapturingStatsReporter,
) {
	mockCtrl := gomock.NewController(t)
	mockService = api.NewMockInterface(mockCtrl)
	isReplay := false
	scope, closer, reporter := newPromScope(&isReplay)
	wrapperService = NewWorkflowServiceWrapper(mockService, scope)
	return
}

func newPromScope(isReplay *bool) (tally.Scope, io.Closer, *CapturingStatsReporter) {
	reporter := &CapturingStatsReporter{}
	opts := tally.ScopeOptions{
		Reporter:        reporter,
		SanitizeOptions: &sanitizeOptions,
	}
	scope, closer := tally.NewRootScope(opts, time.Second)
	return WrapScope(isReplay, scope, &realClock{}), closer, reporter
}
