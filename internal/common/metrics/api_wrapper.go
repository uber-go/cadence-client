// Copyright (c) 2021 Uber Technologies, Inc.
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
	"sync"
	"time"

	"github.com/uber-go/tally"
	apiv1 "go.uber.org/cadence/v1/.gen/proto/api/v1"
	"go.uber.org/cadence/v1/internal/api"
	"go.uber.org/yarpc"
)

type (
	apiMetricsWrapper struct {
		service     api.Interface
		scope       tally.Scope
		childScopes map[string]tally.Scope
		mutex       sync.Mutex
	}

	apiOperationScope struct {
		scope     tally.Scope
		startTime time.Time
	}
)

// NewApiWrapper creates a new wrapper for cadence API that will emit metrics for each service call.
func NewApiWrapper(service api.Interface, scope tally.Scope) api.Interface {
	return &apiMetricsWrapper{service: service, scope: scope, childScopes: make(map[string]tally.Scope)}
}

func (w *apiMetricsWrapper) getScope(scopeName string) tally.Scope {
	w.mutex.Lock()
	scope, ok := w.childScopes[scopeName]
	if ok {
		w.mutex.Unlock()
		return scope
	}
	scope = w.scope.SubScope(scopeName)
	w.childScopes[scopeName] = scope
	w.mutex.Unlock()
	return scope
}

func (w *apiMetricsWrapper) getOperationScope(scopeName string) *apiOperationScope {
	scope := w.getScope(scopeName)
	scope.Counter(CadenceRequest).Inc(1)

	return &apiOperationScope{scope: scope, startTime: time.Now()}
}

func (s *apiOperationScope) handleError(err error) {
	s.scope.Timer(CadenceLatency).Record(time.Now().Sub(s.startTime))
	if err != nil {
		switch api.ConvertError(err).(type) {
		case *api.EntityNotExistsError,
			*api.BadRequestError,
			*api.DomainAlreadyExistsError,
			*api.WorkflowExecutionAlreadyStartedError,
			*api.WorkflowExecutionAlreadyCompletedError,
			*api.QueryFailedError:
			s.scope.Counter(CadenceInvalidRequest).Inc(1)
		default:
			s.scope.Counter(CadenceError).Inc(1)
		}
	}
}

func (w *apiMetricsWrapper) DeprecateDomain(ctx context.Context, request *apiv1.DeprecateDomainRequest, opts ...yarpc.CallOption) (*apiv1.DeprecateDomainResponse, error) {
	scope := w.getOperationScope(scopeNameDeprecateDomain)
	result, err := w.service.DeprecateDomain(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) ListDomains(ctx context.Context, request *apiv1.ListDomainsRequest, opts ...yarpc.CallOption) (*apiv1.ListDomainsResponse, error) {
	scope := w.getOperationScope(scopeNameListDomains)
	result, err := w.service.ListDomains(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) DescribeDomain(ctx context.Context, request *apiv1.DescribeDomainRequest, opts ...yarpc.CallOption) (*apiv1.DescribeDomainResponse, error) {
	scope := w.getOperationScope(scopeNameDescribeDomain)
	result, err := w.service.DescribeDomain(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) DescribeWorkflowExecution(ctx context.Context, request *apiv1.DescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*apiv1.DescribeWorkflowExecutionResponse, error) {
	scope := w.getOperationScope(scopeNameDescribeWorkflowExecution)
	result, err := w.service.DescribeWorkflowExecution(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) GetWorkflowExecutionHistory(ctx context.Context, request *apiv1.GetWorkflowExecutionHistoryRequest, opts ...yarpc.CallOption) (*apiv1.GetWorkflowExecutionHistoryResponse, error) {
	scope := w.getOperationScope(scopeNameGetWorkflowExecutionHistory)
	result, err := w.service.GetWorkflowExecutionHistory(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) ListClosedWorkflowExecutions(ctx context.Context, request *apiv1.ListClosedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*apiv1.ListClosedWorkflowExecutionsResponse, error) {
	scope := w.getOperationScope(scopeNameListClosedWorkflowExecutions)
	result, err := w.service.ListClosedWorkflowExecutions(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) ListOpenWorkflowExecutions(ctx context.Context, request *apiv1.ListOpenWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*apiv1.ListOpenWorkflowExecutionsResponse, error) {
	scope := w.getOperationScope(scopeNameListOpenWorkflowExecutions)
	result, err := w.service.ListOpenWorkflowExecutions(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) ListWorkflowExecutions(ctx context.Context, request *apiv1.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*apiv1.ListWorkflowExecutionsResponse, error) {
	scope := w.getOperationScope(scopeNameListWorkflowExecutions)
	result, err := w.service.ListWorkflowExecutions(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) ListArchivedWorkflowExecutions(ctx context.Context, request *apiv1.ListArchivedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*apiv1.ListArchivedWorkflowExecutionsResponse, error) {
	scope := w.getOperationScope(scopeNameListArchivedWorkflowExecutions)
	result, err := w.service.ListArchivedWorkflowExecutions(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) ScanWorkflowExecutions(ctx context.Context, request *apiv1.ScanWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*apiv1.ScanWorkflowExecutionsResponse, error) {
	scope := w.getOperationScope(scopeNameScanWorkflowExecutions)
	result, err := w.service.ScanWorkflowExecutions(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) CountWorkflowExecutions(ctx context.Context, request *apiv1.CountWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*apiv1.CountWorkflowExecutionsResponse, error) {
	scope := w.getOperationScope(scopeNameCountWorkflowExecutions)
	result, err := w.service.CountWorkflowExecutions(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) PollForActivityTask(ctx context.Context, request *apiv1.PollForActivityTaskRequest, opts ...yarpc.CallOption) (*apiv1.PollForActivityTaskResponse, error) {
	scope := w.getOperationScope(scopeNamePollForActivityTask)
	result, err := w.service.PollForActivityTask(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) PollForDecisionTask(ctx context.Context, request *apiv1.PollForDecisionTaskRequest, opts ...yarpc.CallOption) (*apiv1.PollForDecisionTaskResponse, error) {
	scope := w.getOperationScope(scopeNamePollForDecisionTask)
	result, err := w.service.PollForDecisionTask(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) RecordActivityTaskHeartbeat(ctx context.Context, request *apiv1.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) (*apiv1.RecordActivityTaskHeartbeatResponse, error) {
	scope := w.getOperationScope(scopeNameRecordActivityTaskHeartbeat)
	result, err := w.service.RecordActivityTaskHeartbeat(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) RecordActivityTaskHeartbeatByID(ctx context.Context, request *apiv1.RecordActivityTaskHeartbeatByIDRequest, opts ...yarpc.CallOption) (*apiv1.RecordActivityTaskHeartbeatByIDResponse, error) {
	scope := w.getOperationScope(scopeNameRecordActivityTaskHeartbeatByID)
	result, err := w.service.RecordActivityTaskHeartbeatByID(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) RegisterDomain(ctx context.Context, request *apiv1.RegisterDomainRequest, opts ...yarpc.CallOption) (*apiv1.RegisterDomainResponse, error) {
	scope := w.getOperationScope(scopeNameRegisterDomain)
	result, err := w.service.RegisterDomain(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) RequestCancelWorkflowExecution(ctx context.Context, request *apiv1.RequestCancelWorkflowExecutionRequest, opts ...yarpc.CallOption) (*apiv1.RequestCancelWorkflowExecutionResponse, error) {
	scope := w.getOperationScope(scopeNameRequestCancelWorkflowExecution)
	result, err := w.service.RequestCancelWorkflowExecution(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) RespondActivityTaskCanceled(ctx context.Context, request *apiv1.RespondActivityTaskCanceledRequest, opts ...yarpc.CallOption) (*apiv1.RespondActivityTaskCanceledResponse, error) {
	scope := w.getOperationScope(scopeNameRespondActivityTaskCanceled)
	result, err := w.service.RespondActivityTaskCanceled(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) RespondActivityTaskCompleted(ctx context.Context, request *apiv1.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption) (*apiv1.RespondActivityTaskCompletedResponse, error) {
	scope := w.getOperationScope(scopeNameRespondActivityTaskCompleted)
	result, err := w.service.RespondActivityTaskCompleted(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) RespondActivityTaskFailed(ctx context.Context, request *apiv1.RespondActivityTaskFailedRequest, opts ...yarpc.CallOption) (*apiv1.RespondActivityTaskFailedResponse, error) {
	scope := w.getOperationScope(scopeNameRespondActivityTaskFailed)
	result, err := w.service.RespondActivityTaskFailed(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) RespondActivityTaskCanceledByID(ctx context.Context, request *apiv1.RespondActivityTaskCanceledByIDRequest, opts ...yarpc.CallOption) (*apiv1.RespondActivityTaskCanceledByIDResponse, error) {
	scope := w.getOperationScope(scopeNameRespondActivityTaskCanceledByID)
	result, err := w.service.RespondActivityTaskCanceledByID(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) RespondActivityTaskCompletedByID(ctx context.Context, request *apiv1.RespondActivityTaskCompletedByIDRequest, opts ...yarpc.CallOption) (*apiv1.RespondActivityTaskCompletedByIDResponse, error) {
	scope := w.getOperationScope(scopeNameRespondActivityTaskCompletedByID)
	result, err := w.service.RespondActivityTaskCompletedByID(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) RespondActivityTaskFailedByID(ctx context.Context, request *apiv1.RespondActivityTaskFailedByIDRequest, opts ...yarpc.CallOption) (*apiv1.RespondActivityTaskFailedByIDResponse, error) {
	scope := w.getOperationScope(scopeNameRespondActivityTaskFailedByID)
	result, err := w.service.RespondActivityTaskFailedByID(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) RespondDecisionTaskCompleted(ctx context.Context, request *apiv1.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption) (*apiv1.RespondDecisionTaskCompletedResponse, error) {
	scope := w.getOperationScope(scopeNameRespondDecisionTaskCompleted)
	response, err := w.service.RespondDecisionTaskCompleted(ctx, request, opts...)
	scope.handleError(err)
	return response, err
}

func (w *apiMetricsWrapper) RespondDecisionTaskFailed(ctx context.Context, request *apiv1.RespondDecisionTaskFailedRequest, opts ...yarpc.CallOption) (*apiv1.RespondDecisionTaskFailedResponse, error) {
	scope := w.getOperationScope(scopeNameRespondDecisionTaskFailed)
	result, err := w.service.RespondDecisionTaskFailed(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) SignalWorkflowExecution(ctx context.Context, request *apiv1.SignalWorkflowExecutionRequest, opts ...yarpc.CallOption) (*apiv1.SignalWorkflowExecutionResponse, error) {
	scope := w.getOperationScope(scopeNameSignalWorkflowExecution)
	result, err := w.service.SignalWorkflowExecution(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) SignalWithStartWorkflowExecution(ctx context.Context, request *apiv1.SignalWithStartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*apiv1.SignalWithStartWorkflowExecutionResponse, error) {
	scope := w.getOperationScope(scopeNameSignalWithStartWorkflowExecution)
	result, err := w.service.SignalWithStartWorkflowExecution(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) StartWorkflowExecution(ctx context.Context, request *apiv1.StartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*apiv1.StartWorkflowExecutionResponse, error) {
	scope := w.getOperationScope(scopeNameStartWorkflowExecution)
	result, err := w.service.StartWorkflowExecution(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) TerminateWorkflowExecution(ctx context.Context, request *apiv1.TerminateWorkflowExecutionRequest, opts ...yarpc.CallOption) (*apiv1.TerminateWorkflowExecutionResponse, error) {
	scope := w.getOperationScope(scopeNameTerminateWorkflowExecution)
	result, err := w.service.TerminateWorkflowExecution(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) ResetWorkflowExecution(ctx context.Context, request *apiv1.ResetWorkflowExecutionRequest, opts ...yarpc.CallOption) (*apiv1.ResetWorkflowExecutionResponse, error) {
	scope := w.getOperationScope(scopeNameResetWorkflowExecution)
	result, err := w.service.ResetWorkflowExecution(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) UpdateDomain(ctx context.Context, request *apiv1.UpdateDomainRequest, opts ...yarpc.CallOption) (*apiv1.UpdateDomainResponse, error) {
	scope := w.getOperationScope(scopeNameUpdateDomain)
	result, err := w.service.UpdateDomain(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) QueryWorkflow(ctx context.Context, request *apiv1.QueryWorkflowRequest, opts ...yarpc.CallOption) (*apiv1.QueryWorkflowResponse, error) {
	scope := w.getOperationScope(scopeNameQueryWorkflow)
	result, err := w.service.QueryWorkflow(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) ResetStickyTaskList(ctx context.Context, request *apiv1.ResetStickyTaskListRequest, opts ...yarpc.CallOption) (*apiv1.ResetStickyTaskListResponse, error) {
	scope := w.getOperationScope(scopeNameResetStickyTaskList)
	result, err := w.service.ResetStickyTaskList(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) DescribeTaskList(ctx context.Context, request *apiv1.DescribeTaskListRequest, opts ...yarpc.CallOption) (*apiv1.DescribeTaskListResponse, error) {
	scope := w.getOperationScope(scopeNameDescribeTaskList)
	result, err := w.service.DescribeTaskList(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) RespondQueryTaskCompleted(ctx context.Context, request *apiv1.RespondQueryTaskCompletedRequest, opts ...yarpc.CallOption) (*apiv1.RespondQueryTaskCompletedResponse, error) {
	scope := w.getOperationScope(scopeNameRespondQueryTaskCompleted)
	result, err := w.service.RespondQueryTaskCompleted(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) GetSearchAttributes(ctx context.Context, request *apiv1.GetSearchAttributesRequest, opts ...yarpc.CallOption) (*apiv1.GetSearchAttributesResponse, error) {
	scope := w.getOperationScope(scopeNameGetSearchAttributes)
	result, err := w.service.GetSearchAttributes(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) ListTaskListPartitions(ctx context.Context, request *apiv1.ListTaskListPartitionsRequest, opts ...yarpc.CallOption) (*apiv1.ListTaskListPartitionsResponse, error) {
	scope := w.getOperationScope(scopeNameListTaskListPartitions)
	result, err := w.service.ListTaskListPartitions(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}

func (w *apiMetricsWrapper) GetClusterInfo(ctx context.Context, request *apiv1.GetClusterInfoRequest,  opts ...yarpc.CallOption) (*apiv1.GetClusterInfoResponse, error) {
	scope := w.getOperationScope(scopeNameGetClusterInfo)
	result, err := w.service.GetClusterInfo(ctx, request, opts...)
	scope.handleError(err)
	return result, err
}
