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

package compatibility

import (
	"context"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
	apiv1 "go.uber.org/cadence/.gen/proto/api/v1"
	"go.uber.org/yarpc"
)

type thrift2protoAdapter struct {
	domain     apiv1.DomainAPIYARPCClient
	workflow   apiv1.WorkflowAPIYARPCClient
	worker     apiv1.WorkerAPIYARPCClient
	visibility apiv1.VisibilityAPIYARPCClient
}

func NewThrift2ProtoAdapter(
	domain apiv1.DomainAPIYARPCClient,
	workflow apiv1.WorkflowAPIYARPCClient,
	worker apiv1.WorkerAPIYARPCClient,
	visibility apiv1.VisibilityAPIYARPCClient,
) workflowserviceclient.Interface {
	return thrift2protoAdapter{domain, workflow, worker, visibility}
}

func (a thrift2protoAdapter) CountWorkflowExecutions(ctx context.Context, request *shared.CountWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.CountWorkflowExecutionsResponse, error) {
	response, err := a.visibility.CountWorkflowExecutions(ctx, protoCountWorkflowExecutionsRequest(request), opts...)
	return thriftCountWorkflowExecutionsResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) DeprecateDomain(ctx context.Context, request *shared.DeprecateDomainRequest, opts ...yarpc.CallOption) error {
	_, err := a.domain.DeprecateDomain(ctx, protoDeprecateDomainRequest(request), opts...)
	return thriftError(err)
}

func (a thrift2protoAdapter) DescribeDomain(ctx context.Context, request *shared.DescribeDomainRequest, opts ...yarpc.CallOption) (*shared.DescribeDomainResponse, error) {
	response, err := a.domain.DescribeDomain(ctx, protoDescribeDomainRequest(request), opts...)
	return thriftDescribeDomainResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) DescribeTaskList(ctx context.Context, request *shared.DescribeTaskListRequest, opts ...yarpc.CallOption) (*shared.DescribeTaskListResponse, error) {
	response, err := a.workflow.DescribeTaskList(ctx, protoDescribeTaskListRequest(request), opts...)
	return thriftDescribeTaskListResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) DescribeWorkflowExecution(ctx context.Context, request *shared.DescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.DescribeWorkflowExecutionResponse, error) {
	response, err := a.workflow.DescribeWorkflowExecution(ctx, protoDescribeWorkflowExecutionRequest(request), opts...)
	return thriftDescribeWorkflowExecutionResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) GetClusterInfo(ctx context.Context, opts ...yarpc.CallOption) (*shared.ClusterInfo, error) {
	response, err := a.workflow.GetClusterInfo(ctx, &apiv1.GetClusterInfoRequest{}, opts...)
	return thriftGetClusterInfoResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) GetSearchAttributes(ctx context.Context, opts ...yarpc.CallOption) (*shared.GetSearchAttributesResponse, error) {
	response, err := a.visibility.GetSearchAttributes(ctx, &apiv1.GetSearchAttributesRequest{}, opts...)
	return thriftGetSearchAttributesResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) GetWorkflowExecutionHistory(ctx context.Context, request *shared.GetWorkflowExecutionHistoryRequest, opts ...yarpc.CallOption) (*shared.GetWorkflowExecutionHistoryResponse, error) {
	response, err := a.workflow.GetWorkflowExecutionHistory(ctx, protoGetWorkflowExecutionHistoryRequest(request), opts...)
	return thriftGetWorkflowExecutionHistoryResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) ListArchivedWorkflowExecutions(ctx context.Context, request *shared.ListArchivedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListArchivedWorkflowExecutionsResponse, error) {
	response, err := a.visibility.ListArchivedWorkflowExecutions(ctx, protoListArchivedWorkflowExecutionsRequest(request), opts...)
	return thriftListArchivedWorkflowExecutionsResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) ListClosedWorkflowExecutions(ctx context.Context, request *shared.ListClosedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListClosedWorkflowExecutionsResponse, error) {
	response, err := a.visibility.ListClosedWorkflowExecutions(ctx, protoListClosedWorkflowExecutionsRequest(request), opts...)
	return thriftListClosedWorkflowExecutionsResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) ListDomains(ctx context.Context, request *shared.ListDomainsRequest, opts ...yarpc.CallOption) (*shared.ListDomainsResponse, error) {
	response, err := a.domain.ListDomains(ctx, protoListDomainsRequest(request), opts...)
	return thriftListDomainsResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) ListOpenWorkflowExecutions(ctx context.Context, request *shared.ListOpenWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListOpenWorkflowExecutionsResponse, error) {
	response, err := a.visibility.ListOpenWorkflowExecutions(ctx, protoListOpenWorkflowExecutionsRequest(request), opts...)
	return thriftListOpenWorkflowExecutionsResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) ListTaskListPartitions(ctx context.Context, request *shared.ListTaskListPartitionsRequest, opts ...yarpc.CallOption) (*shared.ListTaskListPartitionsResponse, error) {
	response, err := a.workflow.ListTaskListPartitions(ctx, protoListTaskListPartitionsRequest(request), opts...)
	return thriftListTaskListPartitionsResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) ListWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListWorkflowExecutionsResponse, error) {
	response, err := a.visibility.ListWorkflowExecutions(ctx, protoListWorkflowExecutionsRequest(request), opts...)
	return thriftListWorkflowExecutionsResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) PollForActivityTask(ctx context.Context, request *shared.PollForActivityTaskRequest, opts ...yarpc.CallOption) (*shared.PollForActivityTaskResponse, error) {
	response, err := a.worker.PollForActivityTask(ctx, protoPollForActivityTaskRequest(request), opts...)
	return thriftPollForActivityTaskResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) PollForDecisionTask(ctx context.Context, request *shared.PollForDecisionTaskRequest, opts ...yarpc.CallOption) (*shared.PollForDecisionTaskResponse, error) {
	response, err := a.worker.PollForDecisionTask(ctx, protoPollForDecisionTaskRequest(request), opts...)
	return thriftPollForDecisionTaskResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) QueryWorkflow(ctx context.Context, request *shared.QueryWorkflowRequest, opts ...yarpc.CallOption) (*shared.QueryWorkflowResponse, error) {
	response, err := a.workflow.QueryWorkflow(ctx, protoQueryWorkflowRequest(request), opts...)
	return thriftQueryWorkflowResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) RecordActivityTaskHeartbeat(ctx context.Context, request *shared.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	response, err := a.worker.RecordActivityTaskHeartbeat(ctx, protoRecordActivityTaskHeartbeatRequest(request), opts...)
	return thriftRecordActivityTaskHeartbeatResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) RecordActivityTaskHeartbeatByID(ctx context.Context, request *shared.RecordActivityTaskHeartbeatByIDRequest, opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	response, err := a.worker.RecordActivityTaskHeartbeatByID(ctx, protoRecordActivityTaskHeartbeatByIdRequest(request), opts...)
	return thriftRecordActivityTaskHeartbeatByIdResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) RegisterDomain(ctx context.Context, request *shared.RegisterDomainRequest, opts ...yarpc.CallOption) error {
	_, err := a.domain.RegisterDomain(ctx, protoRegisterDomainRequest(request), opts...)
	return thriftError(err)
}

func (a thrift2protoAdapter) RequestCancelWorkflowExecution(ctx context.Context, request *shared.RequestCancelWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	_, err := a.workflow.RequestCancelWorkflowExecution(ctx, protoRequestCancelWorkflowExecutionRequest(request), opts...)
	return thriftError(err)
}

func (a thrift2protoAdapter) ResetStickyTaskList(ctx context.Context, request *shared.ResetStickyTaskListRequest, opts ...yarpc.CallOption) (*shared.ResetStickyTaskListResponse, error) {
	_, err := a.worker.ResetStickyTaskList(ctx, protoResetStickyTaskListRequest(request), opts...)
	return &shared.ResetStickyTaskListResponse{}, thriftError(err)
}

func (a thrift2protoAdapter) ResetWorkflowExecution(ctx context.Context, request *shared.ResetWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.ResetWorkflowExecutionResponse, error) {
	response, err := a.workflow.ResetWorkflowExecution(ctx, protoResetWorkflowExecutionRequest(request), opts...)
	return thriftResetWorkflowExecutionResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) RespondActivityTaskCanceled(ctx context.Context, request *shared.RespondActivityTaskCanceledRequest, opts ...yarpc.CallOption) error {
	_, err := a.worker.RespondActivityTaskCanceled(ctx, protoRespondActivityTaskCanceledRequest(request), opts...)
	return thriftError(err)
}

func (a thrift2protoAdapter) RespondActivityTaskCanceledByID(ctx context.Context, request *shared.RespondActivityTaskCanceledByIDRequest, opts ...yarpc.CallOption) error {
	_, err := a.worker.RespondActivityTaskCanceledByID(ctx, protoRespondActivityTaskCanceledByIdRequest(request), opts...)
	return thriftError(err)
}

func (a thrift2protoAdapter) RespondActivityTaskCompleted(ctx context.Context, request *shared.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption) error {
	_, err := a.worker.RespondActivityTaskCompleted(ctx, protoRespondActivityTaskCompletedRequest(request), opts...)
	return thriftError(err)
}

func (a thrift2protoAdapter) RespondActivityTaskCompletedByID(ctx context.Context, request *shared.RespondActivityTaskCompletedByIDRequest, opts ...yarpc.CallOption) error {
	_, err := a.worker.RespondActivityTaskCompletedByID(ctx, protoRespondActivityTaskCompletedByIdRequest(request), opts...)
	return thriftError(err)
}

func (a thrift2protoAdapter) RespondActivityTaskFailed(ctx context.Context, request *shared.RespondActivityTaskFailedRequest, opts ...yarpc.CallOption) error {
	_, err := a.worker.RespondActivityTaskFailed(ctx, protoRespondActivityTaskFailedRequest(request), opts...)
	return thriftError(err)
}

func (a thrift2protoAdapter) RespondActivityTaskFailedByID(ctx context.Context, request *shared.RespondActivityTaskFailedByIDRequest, opts ...yarpc.CallOption) error {
	_, err := a.worker.RespondActivityTaskFailedByID(ctx, protoRespondActivityTaskFailedByIdRequest(request), opts...)
	return thriftError(err)
}

func (a thrift2protoAdapter) RespondDecisionTaskCompleted(ctx context.Context, request *shared.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption) (*shared.RespondDecisionTaskCompletedResponse, error) {
	response, err := a.worker.RespondDecisionTaskCompleted(ctx, protoRespondDecisionTaskCompletedRequest(request), opts...)
	return thriftRespondDecisionTaskCompletedResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) RespondDecisionTaskFailed(ctx context.Context, request *shared.RespondDecisionTaskFailedRequest, opts ...yarpc.CallOption) error {
	_, err := a.worker.RespondDecisionTaskFailed(ctx, protoRespondDecisionTaskFailedRequest(request), opts...)
	return thriftError(err)
}

func (a thrift2protoAdapter) RespondQueryTaskCompleted(ctx context.Context, request *shared.RespondQueryTaskCompletedRequest, opts ...yarpc.CallOption) error {
	_, err := a.worker.RespondQueryTaskCompleted(ctx, protoRespondQueryTaskCompletedRequest(request), opts...)
	return thriftError(err)
}

func (a thrift2protoAdapter) ScanWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListWorkflowExecutionsResponse, error) {
	response, err := a.visibility.ScanWorkflowExecutions(ctx, protoScanWorkflowExecutionsRequest(request), opts...)
	return thriftScanWorkflowExecutionsResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) SignalWithStartWorkflowExecution(ctx context.Context, request *shared.SignalWithStartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	response, err := a.workflow.SignalWithStartWorkflowExecution(ctx, protoSignalWithStartWorkflowExecutionRequest(request), opts...)
	return thriftSignalWithStartWorkflowExecutionResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) SignalWorkflowExecution(ctx context.Context, request *shared.SignalWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	_, err := a.workflow.SignalWorkflowExecution(ctx, protoSignalWorkflowExecutionRequest(request), opts...)
	return thriftError(err)
}

func (a thrift2protoAdapter) StartWorkflowExecution(ctx context.Context, request *shared.StartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	response, err := a.workflow.StartWorkflowExecution(ctx, protoStartWorkflowExecutionRequest(request), opts...)
	return thriftStartWorkflowExecutionResponse(response), thriftError(err)
}

func (a thrift2protoAdapter) TerminateWorkflowExecution(ctx context.Context, request *shared.TerminateWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	_, err := a.workflow.TerminateWorkflowExecution(ctx, protoTerminateWorkflowExecutionRequest(request), opts...)
	return thriftError(err)
}

func (a thrift2protoAdapter) UpdateDomain(ctx context.Context, request *shared.UpdateDomainRequest, opts ...yarpc.CallOption) (*shared.UpdateDomainResponse, error) {
	response, err := a.domain.UpdateDomain(ctx, protoUpdateDomainRequest(request), opts...)
	return thriftUpdateDomainResponse(response), thriftError(err)
}
