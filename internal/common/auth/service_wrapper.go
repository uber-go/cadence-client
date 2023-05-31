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

	"go.uber.org/yarpc"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
)

const (
	jwtHeaderName        = "cadence-authorization"
	isolationGroupHeader = "isolation-group"
)

type workflowServiceIdentityAndAuthWrapper struct {
	service                  workflowserviceclient.Interface
	authProvider             AuthorizationProvider
	isolationGroupIdentifier string
}

type AuthorizationProvider interface {
	// GetAuthToken provides the OAuth authorization token
	// It's called before every request to Cadence server, and sets the token in the request header.
	GetAuthToken() ([]byte, error)
}

type JWTClaims struct {
	Sub    string
	Name   string
	Groups string // separated by space
	Admin  bool
	Iat    int64
	TTL    int64
}

// NewWorkflowServiceWrapper creates
func NewWorkflowServiceWrapper(service workflowserviceclient.Interface, authorizationProvider AuthorizationProvider, isolationGroup string) workflowserviceclient.Interface {
	return &workflowServiceIdentityAndAuthWrapper{
		service:                  service,
		authProvider:             authorizationProvider,
		isolationGroupIdentifier: isolationGroup,
	}
}

func (w *workflowServiceIdentityAndAuthWrapper) getYarpcJWTHeader() (*yarpc.CallOption, error) {
	token, err := w.authProvider.GetAuthToken()
	if err != nil {
		return nil, err
	}
	header := yarpc.WithHeader(jwtHeaderName, string(token))
	return &header, nil
}

func (w *workflowServiceIdentityAndAuthWrapper) getIsolationGroupIdentifier() yarpc.CallOption {
	return yarpc.WithHeader(isolationGroupHeader, w.isolationGroupIdentifier)
}

func (w *workflowServiceIdentityAndAuthWrapper) DeprecateDomain(ctx context.Context, request *shared.DeprecateDomainRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	err = w.service.DeprecateDomain(ctx, request, opts...)
	return err
}

func (w *workflowServiceIdentityAndAuthWrapper) ListDomains(ctx context.Context, request *shared.ListDomainsRequest, opts ...yarpc.CallOption) (*shared.ListDomainsResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ListDomains(ctx, request, opts...)

	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) DescribeDomain(ctx context.Context, request *shared.DescribeDomainRequest, opts ...yarpc.CallOption) (*shared.DescribeDomainResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.DescribeDomain(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) DescribeWorkflowExecution(ctx context.Context, request *shared.DescribeWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.DescribeWorkflowExecutionResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.DescribeWorkflowExecution(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) GetWorkflowExecutionHistory(ctx context.Context, request *shared.GetWorkflowExecutionHistoryRequest, opts ...yarpc.CallOption) (*shared.GetWorkflowExecutionHistoryResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.GetWorkflowExecutionHistory(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) ListClosedWorkflowExecutions(ctx context.Context, request *shared.ListClosedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListClosedWorkflowExecutionsResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ListClosedWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) ListOpenWorkflowExecutions(ctx context.Context, request *shared.ListOpenWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListOpenWorkflowExecutionsResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ListOpenWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) ListWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListWorkflowExecutionsResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ListWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) ListArchivedWorkflowExecutions(ctx context.Context, request *shared.ListArchivedWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListArchivedWorkflowExecutionsResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ListArchivedWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) ScanWorkflowExecutions(ctx context.Context, request *shared.ListWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.ListWorkflowExecutionsResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ScanWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) CountWorkflowExecutions(ctx context.Context, request *shared.CountWorkflowExecutionsRequest, opts ...yarpc.CallOption) (*shared.CountWorkflowExecutionsResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.CountWorkflowExecutions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) PollForActivityTask(ctx context.Context, request *shared.PollForActivityTaskRequest, opts ...yarpc.CallOption) (*shared.PollForActivityTaskResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.PollForActivityTask(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) PollForDecisionTask(ctx context.Context, request *shared.PollForDecisionTaskRequest, opts ...yarpc.CallOption) (*shared.PollForDecisionTaskResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.PollForDecisionTask(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) RecordActivityTaskHeartbeat(ctx context.Context, request *shared.RecordActivityTaskHeartbeatRequest, opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.RecordActivityTaskHeartbeat(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) RecordActivityTaskHeartbeatByID(ctx context.Context, request *shared.RecordActivityTaskHeartbeatByIDRequest, opts ...yarpc.CallOption) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.RecordActivityTaskHeartbeatByID(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) RegisterDomain(ctx context.Context, request *shared.RegisterDomainRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	err = w.service.RegisterDomain(ctx, request, opts...)
	return err
}

func (w *workflowServiceIdentityAndAuthWrapper) RequestCancelWorkflowExecution(ctx context.Context, request *shared.RequestCancelWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	err = w.service.RequestCancelWorkflowExecution(ctx, request, opts...)
	return err
}

func (w *workflowServiceIdentityAndAuthWrapper) RespondActivityTaskCanceled(ctx context.Context, request *shared.RespondActivityTaskCanceledRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	err = w.service.RespondActivityTaskCanceled(ctx, request, opts...)
	return err
}

func (w *workflowServiceIdentityAndAuthWrapper) RespondActivityTaskCompleted(ctx context.Context, request *shared.RespondActivityTaskCompletedRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	err = w.service.RespondActivityTaskCompleted(ctx, request, opts...)
	return err
}

func (w *workflowServiceIdentityAndAuthWrapper) RespondActivityTaskFailed(ctx context.Context, request *shared.RespondActivityTaskFailedRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	err = w.service.RespondActivityTaskFailed(ctx, request, opts...)
	return err
}

func (w *workflowServiceIdentityAndAuthWrapper) RespondActivityTaskCanceledByID(ctx context.Context, request *shared.RespondActivityTaskCanceledByIDRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	err = w.service.RespondActivityTaskCanceledByID(ctx, request, opts...)
	return err
}

func (w *workflowServiceIdentityAndAuthWrapper) RespondActivityTaskCompletedByID(ctx context.Context, request *shared.RespondActivityTaskCompletedByIDRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	err = w.service.RespondActivityTaskCompletedByID(ctx, request, opts...)
	return err
}

func (w *workflowServiceIdentityAndAuthWrapper) RespondActivityTaskFailedByID(ctx context.Context, request *shared.RespondActivityTaskFailedByIDRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	err = w.service.RespondActivityTaskFailedByID(ctx, request, opts...)
	return err
}

func (w *workflowServiceIdentityAndAuthWrapper) RespondDecisionTaskCompleted(ctx context.Context, request *shared.RespondDecisionTaskCompletedRequest, opts ...yarpc.CallOption) (*shared.RespondDecisionTaskCompletedResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	response, err := w.service.RespondDecisionTaskCompleted(ctx, request, opts...)
	return response, err
}

func (w *workflowServiceIdentityAndAuthWrapper) RespondDecisionTaskFailed(ctx context.Context, request *shared.RespondDecisionTaskFailedRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	err = w.service.RespondDecisionTaskFailed(ctx, request, opts...)
	return err
}

func (w *workflowServiceIdentityAndAuthWrapper) SignalWorkflowExecution(ctx context.Context, request *shared.SignalWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	err = w.service.SignalWorkflowExecution(ctx, request, opts...)
	return err
}

func (w *workflowServiceIdentityAndAuthWrapper) SignalWithStartWorkflowExecution(ctx context.Context, request *shared.SignalWithStartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.SignalWithStartWorkflowExecution(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) StartWorkflowExecution(ctx context.Context, request *shared.StartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.StartWorkflowExecutionResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.StartWorkflowExecution(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) TerminateWorkflowExecution(ctx context.Context, request *shared.TerminateWorkflowExecutionRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	err = w.service.TerminateWorkflowExecution(ctx, request, opts...)
	return err
}

func (w *workflowServiceIdentityAndAuthWrapper) ResetWorkflowExecution(ctx context.Context, request *shared.ResetWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.ResetWorkflowExecutionResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ResetWorkflowExecution(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) UpdateDomain(ctx context.Context, request *shared.UpdateDomainRequest, opts ...yarpc.CallOption) (*shared.UpdateDomainResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.UpdateDomain(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) QueryWorkflow(ctx context.Context, request *shared.QueryWorkflowRequest, opts ...yarpc.CallOption) (*shared.QueryWorkflowResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.QueryWorkflow(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) ResetStickyTaskList(ctx context.Context, request *shared.ResetStickyTaskListRequest, opts ...yarpc.CallOption) (*shared.ResetStickyTaskListResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ResetStickyTaskList(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) DescribeTaskList(ctx context.Context, request *shared.DescribeTaskListRequest, opts ...yarpc.CallOption) (*shared.DescribeTaskListResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.DescribeTaskList(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) RespondQueryTaskCompleted(ctx context.Context, request *shared.RespondQueryTaskCompletedRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	err = w.service.RespondQueryTaskCompleted(ctx, request, opts...)
	return err
}

func (w *workflowServiceIdentityAndAuthWrapper) GetSearchAttributes(ctx context.Context, opts ...yarpc.CallOption) (*shared.GetSearchAttributesResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	result, err := w.service.GetSearchAttributes(ctx, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) ListTaskListPartitions(ctx context.Context, request *shared.ListTaskListPartitionsRequest, opts ...yarpc.CallOption) (*shared.ListTaskListPartitionsResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.ListTaskListPartitions(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) GetClusterInfo(ctx context.Context, opts ...yarpc.CallOption) (*shared.ClusterInfo, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.GetClusterInfo(ctx, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) GetTaskListsByDomain(ctx context.Context, request *shared.GetTaskListsByDomainRequest, opts ...yarpc.CallOption) (*shared.GetTaskListsByDomainResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	result, err := w.service.GetTaskListsByDomain(ctx, request, opts...)
	return result, err
}

func (w *workflowServiceIdentityAndAuthWrapper) RefreshWorkflowTasks(ctx context.Context, request *shared.RefreshWorkflowTasksRequest, opts ...yarpc.CallOption) error {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return err
	}
	opts = append(opts, *tokenHeader)
	opts = append(opts, w.getIsolationGroupIdentifier())
	err = w.service.RefreshWorkflowTasks(ctx, request, opts...)
	return err
}

func (w *workflowServiceAuthWrapper) RestartWorkflowExecution(ctx context.Context, request *shared.RestartWorkflowExecutionRequest, opts ...yarpc.CallOption) (*shared.RestartWorkflowExecutionResponse, error) {
	tokenHeader, err := w.getYarpcJWTHeader()
	if err != nil {
		return nil, err
	}
	opts = append(opts, *tokenHeader)
	return w.service.RestartWorkflowExecution(ctx, request, opts...)
}
