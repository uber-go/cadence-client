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

package auth

import (
    "context"
    "go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
    "go.uber.org/cadence/.gen/go/shared"
    "go.uber.org/cadence/worker"
    "go.uber.org/yarpc"
)

const (
    jwtHeaderName = "cadence-client-jwt"
)

type workflowServiceAuthWrapper struct {
    service workflowserviceclient.Interface
    authProvider worker.AuthorizationProvider
}

// NewWorkflowServiceWrapper creates
func NewWorkflowServiceWrapper(service workflowserviceclient.Interface, authorizationProvider worker.AuthorizationProvider) workflowserviceclient.Interface {
    return &workflowServiceAuthWrapper{
        service: service,
        authProvider: authorizationProvider,
    }
}

func (w *workflowServiceAuthWrapper) getYarpcJWTHeader() yarpc.CallOption {
    token := w.authProvider.GetAuthToken()
    return yarpc.WithHeader(jwtHeaderName, string(token))
}

func (w *workflowServiceAuthWrapper) DeprecateDomain(ctx context.Context, request *shared.DeprecateDomainRequest, opts ...yarpc.CallOption) error {
    opts = append(opts, w.getYarpcJWTHeader())
    err := w.service.DeprecateDomain(ctx, request, opts...)
    return err
}

func (w workflowServiceAuthWrapper) CountWorkflowExecutions(ctx context.Context, CountRequest *shared.CountWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.CountWorkflowExecutionsResponse, error) {
    panic("implement me")
}


func (w workflowServiceAuthWrapper) DescribeDomain(ctx context.Context, DescribeRequest *shared.DescribeDomainRequest, opts ...yarpc.CallOption) (*shared.DescribeDomainResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) DescribeTaskList(ctx context.Context, Request *shared.DescribeTaskListRequest, opts ...yarpc.CallOption) (*shared.DescribeTaskListResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) DescribeWorkflowExecution(ctx context.Context, DescribeRequest *shared.DescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.DescribeWorkflowExecutionResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) GetClusterInfo(ctx context.Context, opts ...yarpc.CallOption) (*shared.ClusterInfo, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) GetSearchAttributes(ctx context.Context, opts ...yarpc.CallOption) (*shared.GetSearchAttributesResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) GetWorkflowExecutionHistory(ctx context.Context, GetRequest *shared.GetWorkflowExecutionHistoryRequest, opts ...yarpc.CallOption) (*shared.GetWorkflowExecutionHistoryResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) ListArchivedWorkflowExecutions(ctx context.Context, ListRequest *shared.ListArchivedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListArchivedWorkflowExecutionsResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) ListClosedWorkflowExecutions(ctx context.Context, ListRequest *shared.ListClosedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListClosedWorkflowExecutionsResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) ListDomains(ctx context.Context, ListRequest *shared.ListDomainsRequest, opts ...yarpc.CallOption) (*shared.ListDomainsResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) ListOpenWorkflowExecutions(ctx context.Context, ListRequest *shared.ListOpenWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListOpenWorkflowExecutionsResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) ListTaskListPartitions(ctx context.Context, Request *shared.ListTaskListPartitionsRequest, opts ...yarpc.CallOption) (*shared.ListTaskListPartitionsResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) ListWorkflowExecutions(ctx context.Context, ListRequest *shared.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListWorkflowExecutionsResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) PollForActivityTask(ctx context.Context, PollRequest *shared.PollForActivityTaskRequest, opts ...yarpc.CallOption) (*shared.PollForActivityTaskResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) PollForDecisionTask(ctx context.Context, PollRequest *shared.PollForDecisionTaskRequest, opts ...yarpc.CallOption) (*shared.PollForDecisionTaskResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) QueryWorkflow(ctx context.Context, QueryRequest *shared.QueryWorkflowRequest, opts ...yarpc.CallOption) (*shared.QueryWorkflowResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) RecordActivityTaskHeartbeat(ctx context.Context, HeartbeatRequest *shared.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) RecordActivityTaskHeartbeatByID(ctx context.Context, HeartbeatRequest *shared.RecordActivityTaskHeartbeatByIDRequest, opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) RegisterDomain(ctx context.Context, RegisterRequest *shared.RegisterDomainRequest, opts ...yarpc.CallOption) error {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) RequestCancelWorkflowExecution(ctx context.Context, CancelRequest *shared.RequestCancelWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) ResetStickyTaskList(ctx context.Context, ResetRequest *shared.ResetStickyTaskListRequest, opts ...yarpc.CallOption) (*shared.ResetStickyTaskListResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) ResetWorkflowExecution(ctx context.Context, ResetRequest *shared.ResetWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.ResetWorkflowExecutionResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) RespondActivityTaskCanceled(ctx context.Context, CanceledRequest *shared.RespondActivityTaskCanceledRequest, opts ...yarpc.CallOption) error {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) RespondActivityTaskCanceledByID(ctx context.Context, CanceledRequest *shared.RespondActivityTaskCanceledByIDRequest, opts ...yarpc.CallOption) error {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) RespondActivityTaskCompleted(ctx context.Context, CompleteRequest *shared.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption) error {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) RespondActivityTaskCompletedByID(ctx context.Context, CompleteRequest *shared.RespondActivityTaskCompletedByIDRequest, opts ...yarpc.CallOption) error {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) RespondActivityTaskFailed(ctx context.Context, FailRequest *shared.RespondActivityTaskFailedRequest, opts ...yarpc.CallOption) error {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) RespondActivityTaskFailedByID(ctx context.Context, FailRequest *shared.RespondActivityTaskFailedByIDRequest, opts ...yarpc.CallOption) error {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) RespondDecisionTaskCompleted(ctx context.Context, CompleteRequest *shared.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption) (*shared.RespondDecisionTaskCompletedResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) RespondDecisionTaskFailed(ctx context.Context, FailedRequest *shared.RespondDecisionTaskFailedRequest, opts ...yarpc.CallOption) error {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) RespondQueryTaskCompleted(ctx context.Context, CompleteRequest *shared.RespondQueryTaskCompletedRequest, opts ...yarpc.CallOption) error {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) ScanWorkflowExecutions(ctx context.Context, ListRequest *shared.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListWorkflowExecutionsResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) SignalWithStartWorkflowExecution(ctx context.Context, SignalWithStartRequest *shared.SignalWithStartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) SignalWorkflowExecution(ctx context.Context, SignalRequest *shared.SignalWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) StartWorkflowExecution(ctx context.Context, StartRequest *shared.StartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) TerminateWorkflowExecution(ctx context.Context, TerminateRequest *shared.TerminateWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
    panic("implement me")
}

func (w workflowServiceAuthWrapper) UpdateDomain(ctx context.Context, UpdateRequest *shared.UpdateDomainRequest, opts ...yarpc.CallOption) (*shared.UpdateDomainResponse, error) {
    panic("implement me")
}