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

// Code generated by thriftrw-plugin-yarpc
// @generated

package workflowserviceserver

import (
	"context"
	"go.uber.org/cadence/.gen/go/cadence"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/thriftrw/wire"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/encoding/thrift"
)

// Interface is the server-side interface for the WorkflowService service.
type Interface interface {
	DeprecateDomain(
		ctx context.Context,
		DeprecateRequest *shared.DeprecateDomainRequest,
	) error

	DescribeDomain(
		ctx context.Context,
		DescribeRequest *shared.DescribeDomainRequest,
	) (*shared.DescribeDomainResponse, error)

	DescribeTaskList(
		ctx context.Context,
		Request *shared.DescribeTaskListRequest,
	) (*shared.DescribeTaskListResponse, error)

	DescribeWorkflowExecution(
		ctx context.Context,
		DescribeRequest *shared.DescribeWorkflowExecutionRequest,
	) (*shared.DescribeWorkflowExecutionResponse, error)

	GetWorkflowExecutionHistory(
		ctx context.Context,
		GetRequest *shared.GetWorkflowExecutionHistoryRequest,
	) (*shared.GetWorkflowExecutionHistoryResponse, error)

	ListClosedWorkflowExecutions(
		ctx context.Context,
		ListRequest *shared.ListClosedWorkflowExecutionsRequest,
	) (*shared.ListClosedWorkflowExecutionsResponse, error)

	ListOpenWorkflowExecutions(
		ctx context.Context,
		ListRequest *shared.ListOpenWorkflowExecutionsRequest,
	) (*shared.ListOpenWorkflowExecutionsResponse, error)

	PollForActivityTask(
		ctx context.Context,
		PollRequest *shared.PollForActivityTaskRequest,
	) (*shared.PollForActivityTaskResponse, error)

	PollForDecisionTask(
		ctx context.Context,
		PollRequest *shared.PollForDecisionTaskRequest,
	) (*shared.PollForDecisionTaskResponse, error)

	QueryTaskList(
		ctx context.Context,
		QueryRequest *shared.QueryTaskListRequest,
	) (*shared.QueryTaskListResponse, error)

	QueryWorkflow(
		ctx context.Context,
		QueryRequest *shared.QueryWorkflowRequest,
	) (*shared.QueryWorkflowResponse, error)

	RecordActivityTaskHeartbeat(
		ctx context.Context,
		HeartbeatRequest *shared.RecordActivityTaskHeartbeatRequest,
	) (*shared.RecordActivityTaskHeartbeatResponse, error)

	RecordActivityTaskHeartbeatByID(
		ctx context.Context,
		HeartbeatRequest *shared.RecordActivityTaskHeartbeatByIDRequest,
	) (*shared.RecordActivityTaskHeartbeatResponse, error)

	RegisterDomain(
		ctx context.Context,
		RegisterRequest *shared.RegisterDomainRequest,
	) error

	RequestCancelWorkflowExecution(
		ctx context.Context,
		CancelRequest *shared.RequestCancelWorkflowExecutionRequest,
	) error

	ResetStickyTaskList(
		ctx context.Context,
		ResetRequest *shared.ResetStickyTaskListRequest,
	) (*shared.ResetStickyTaskListResponse, error)

	RespondActivityTaskCanceled(
		ctx context.Context,
		CanceledRequest *shared.RespondActivityTaskCanceledRequest,
	) error

	RespondActivityTaskCanceledByID(
		ctx context.Context,
		CanceledRequest *shared.RespondActivityTaskCanceledByIDRequest,
	) error

	RespondActivityTaskCompleted(
		ctx context.Context,
		CompleteRequest *shared.RespondActivityTaskCompletedRequest,
	) error

	RespondActivityTaskCompletedByID(
		ctx context.Context,
		CompleteRequest *shared.RespondActivityTaskCompletedByIDRequest,
	) error

	RespondActivityTaskFailed(
		ctx context.Context,
		FailRequest *shared.RespondActivityTaskFailedRequest,
	) error

	RespondActivityTaskFailedByID(
		ctx context.Context,
		FailRequest *shared.RespondActivityTaskFailedByIDRequest,
	) error

	RespondDecisionTaskCompleted(
		ctx context.Context,
		CompleteRequest *shared.RespondDecisionTaskCompletedRequest,
	) (*shared.RespondDecisionTaskCompletedResponse, error)

	RespondDecisionTaskFailed(
		ctx context.Context,
		FailedRequest *shared.RespondDecisionTaskFailedRequest,
	) error

	RespondQueryTaskCompleted(
		ctx context.Context,
		CompleteRequest *shared.RespondQueryTaskCompletedRequest,
	) error

	SignalWithStartWorkflowExecution(
		ctx context.Context,
		SignalWithStartRequest *shared.SignalWithStartWorkflowExecutionRequest,
	) (*shared.StartWorkflowExecutionResponse, error)

	SignalWorkflowExecution(
		ctx context.Context,
		SignalRequest *shared.SignalWorkflowExecutionRequest,
	) error

	StartWorkflowExecution(
		ctx context.Context,
		StartRequest *shared.StartWorkflowExecutionRequest,
	) (*shared.StartWorkflowExecutionResponse, error)

	TerminateWorkflowExecution(
		ctx context.Context,
		TerminateRequest *shared.TerminateWorkflowExecutionRequest,
	) error

	UpdateDomain(
		ctx context.Context,
		UpdateRequest *shared.UpdateDomainRequest,
	) (*shared.UpdateDomainResponse, error)
}

// New prepares an implementation of the WorkflowService service for
// registration.
//
// 	handler := WorkflowServiceHandler{}
// 	dispatcher.Register(workflowserviceserver.New(handler))
func New(impl Interface, opts ...thrift.RegisterOption) []transport.Procedure {
	h := handler{impl}
	service := thrift.Service{
		Name: "WorkflowService",
		Methods: []thrift.Method{

			thrift.Method{
				Name: "DeprecateDomain",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.DeprecateDomain),
				},
				Signature:    "DeprecateDomain(DeprecateRequest *shared.DeprecateDomainRequest)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "DescribeDomain",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.DescribeDomain),
				},
				Signature:    "DescribeDomain(DescribeRequest *shared.DescribeDomainRequest) (*shared.DescribeDomainResponse)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "DescribeTaskList",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.DescribeTaskList),
				},
				Signature:    "DescribeTaskList(Request *shared.DescribeTaskListRequest) (*shared.DescribeTaskListResponse)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "DescribeWorkflowExecution",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.DescribeWorkflowExecution),
				},
				Signature:    "DescribeWorkflowExecution(DescribeRequest *shared.DescribeWorkflowExecutionRequest) (*shared.DescribeWorkflowExecutionResponse)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "GetWorkflowExecutionHistory",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.GetWorkflowExecutionHistory),
				},
				Signature:    "GetWorkflowExecutionHistory(GetRequest *shared.GetWorkflowExecutionHistoryRequest) (*shared.GetWorkflowExecutionHistoryResponse)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "ListClosedWorkflowExecutions",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.ListClosedWorkflowExecutions),
				},
				Signature:    "ListClosedWorkflowExecutions(ListRequest *shared.ListClosedWorkflowExecutionsRequest) (*shared.ListClosedWorkflowExecutionsResponse)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "ListOpenWorkflowExecutions",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.ListOpenWorkflowExecutions),
				},
				Signature:    "ListOpenWorkflowExecutions(ListRequest *shared.ListOpenWorkflowExecutionsRequest) (*shared.ListOpenWorkflowExecutionsResponse)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "PollForActivityTask",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.PollForActivityTask),
				},
				Signature:    "PollForActivityTask(PollRequest *shared.PollForActivityTaskRequest) (*shared.PollForActivityTaskResponse)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "PollForDecisionTask",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.PollForDecisionTask),
				},
				Signature:    "PollForDecisionTask(PollRequest *shared.PollForDecisionTaskRequest) (*shared.PollForDecisionTaskResponse)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "QueryTaskList",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.QueryTaskList),
				},
				Signature:    "QueryTaskList(QueryRequest *shared.QueryTaskListRequest) (*shared.QueryTaskListResponse)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "QueryWorkflow",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.QueryWorkflow),
				},
				Signature:    "QueryWorkflow(QueryRequest *shared.QueryWorkflowRequest) (*shared.QueryWorkflowResponse)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "RecordActivityTaskHeartbeat",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.RecordActivityTaskHeartbeat),
				},
				Signature:    "RecordActivityTaskHeartbeat(HeartbeatRequest *shared.RecordActivityTaskHeartbeatRequest) (*shared.RecordActivityTaskHeartbeatResponse)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "RecordActivityTaskHeartbeatByID",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.RecordActivityTaskHeartbeatByID),
				},
				Signature:    "RecordActivityTaskHeartbeatByID(HeartbeatRequest *shared.RecordActivityTaskHeartbeatByIDRequest) (*shared.RecordActivityTaskHeartbeatResponse)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "RegisterDomain",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.RegisterDomain),
				},
				Signature:    "RegisterDomain(RegisterRequest *shared.RegisterDomainRequest)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "RequestCancelWorkflowExecution",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.RequestCancelWorkflowExecution),
				},
				Signature:    "RequestCancelWorkflowExecution(CancelRequest *shared.RequestCancelWorkflowExecutionRequest)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "ResetStickyTaskList",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.ResetStickyTaskList),
				},
				Signature:    "ResetStickyTaskList(ResetRequest *shared.ResetStickyTaskListRequest) (*shared.ResetStickyTaskListResponse)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "RespondActivityTaskCanceled",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.RespondActivityTaskCanceled),
				},
				Signature:    "RespondActivityTaskCanceled(CanceledRequest *shared.RespondActivityTaskCanceledRequest)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "RespondActivityTaskCanceledByID",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.RespondActivityTaskCanceledByID),
				},
				Signature:    "RespondActivityTaskCanceledByID(CanceledRequest *shared.RespondActivityTaskCanceledByIDRequest)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "RespondActivityTaskCompleted",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.RespondActivityTaskCompleted),
				},
				Signature:    "RespondActivityTaskCompleted(CompleteRequest *shared.RespondActivityTaskCompletedRequest)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "RespondActivityTaskCompletedByID",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.RespondActivityTaskCompletedByID),
				},
				Signature:    "RespondActivityTaskCompletedByID(CompleteRequest *shared.RespondActivityTaskCompletedByIDRequest)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "RespondActivityTaskFailed",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.RespondActivityTaskFailed),
				},
				Signature:    "RespondActivityTaskFailed(FailRequest *shared.RespondActivityTaskFailedRequest)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "RespondActivityTaskFailedByID",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.RespondActivityTaskFailedByID),
				},
				Signature:    "RespondActivityTaskFailedByID(FailRequest *shared.RespondActivityTaskFailedByIDRequest)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "RespondDecisionTaskCompleted",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.RespondDecisionTaskCompleted),
				},
				Signature:    "RespondDecisionTaskCompleted(CompleteRequest *shared.RespondDecisionTaskCompletedRequest) (*shared.RespondDecisionTaskCompletedResponse)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "RespondDecisionTaskFailed",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.RespondDecisionTaskFailed),
				},
				Signature:    "RespondDecisionTaskFailed(FailedRequest *shared.RespondDecisionTaskFailedRequest)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "RespondQueryTaskCompleted",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.RespondQueryTaskCompleted),
				},
				Signature:    "RespondQueryTaskCompleted(CompleteRequest *shared.RespondQueryTaskCompletedRequest)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "SignalWithStartWorkflowExecution",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.SignalWithStartWorkflowExecution),
				},
				Signature:    "SignalWithStartWorkflowExecution(SignalWithStartRequest *shared.SignalWithStartWorkflowExecutionRequest) (*shared.StartWorkflowExecutionResponse)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "SignalWorkflowExecution",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.SignalWorkflowExecution),
				},
				Signature:    "SignalWorkflowExecution(SignalRequest *shared.SignalWorkflowExecutionRequest)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "StartWorkflowExecution",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.StartWorkflowExecution),
				},
				Signature:    "StartWorkflowExecution(StartRequest *shared.StartWorkflowExecutionRequest) (*shared.StartWorkflowExecutionResponse)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "TerminateWorkflowExecution",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.TerminateWorkflowExecution),
				},
				Signature:    "TerminateWorkflowExecution(TerminateRequest *shared.TerminateWorkflowExecutionRequest)",
				ThriftModule: cadence.ThriftModule,
			},

			thrift.Method{
				Name: "UpdateDomain",
				HandlerSpec: thrift.HandlerSpec{

					Type:  transport.Unary,
					Unary: thrift.UnaryHandler(h.UpdateDomain),
				},
				Signature:    "UpdateDomain(UpdateRequest *shared.UpdateDomainRequest) (*shared.UpdateDomainResponse)",
				ThriftModule: cadence.ThriftModule,
			},
		},
	}

	procedures := make([]transport.Procedure, 0, 30)
	procedures = append(procedures, thrift.BuildProcedures(service, opts...)...)
	return procedures
}

type handler struct{ impl Interface }

func (h handler) DeprecateDomain(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_DeprecateDomain_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	err := h.impl.DeprecateDomain(ctx, args.DeprecateRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_DeprecateDomain_Helper.WrapResponse(err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) DescribeDomain(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_DescribeDomain_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.DescribeDomain(ctx, args.DescribeRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_DescribeDomain_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) DescribeTaskList(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_DescribeTaskList_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.DescribeTaskList(ctx, args.Request)

	hadError := err != nil
	result, err := cadence.WorkflowService_DescribeTaskList_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) DescribeWorkflowExecution(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_DescribeWorkflowExecution_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.DescribeWorkflowExecution(ctx, args.DescribeRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_DescribeWorkflowExecution_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) GetWorkflowExecutionHistory(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_GetWorkflowExecutionHistory_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.GetWorkflowExecutionHistory(ctx, args.GetRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_GetWorkflowExecutionHistory_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) ListClosedWorkflowExecutions(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_ListClosedWorkflowExecutions_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.ListClosedWorkflowExecutions(ctx, args.ListRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_ListClosedWorkflowExecutions_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) ListOpenWorkflowExecutions(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_ListOpenWorkflowExecutions_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.ListOpenWorkflowExecutions(ctx, args.ListRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_ListOpenWorkflowExecutions_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) PollForActivityTask(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_PollForActivityTask_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.PollForActivityTask(ctx, args.PollRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_PollForActivityTask_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) PollForDecisionTask(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_PollForDecisionTask_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.PollForDecisionTask(ctx, args.PollRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_PollForDecisionTask_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) QueryTaskList(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_QueryTaskList_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.QueryTaskList(ctx, args.QueryRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_QueryTaskList_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) QueryWorkflow(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_QueryWorkflow_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.QueryWorkflow(ctx, args.QueryRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_QueryWorkflow_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) RecordActivityTaskHeartbeat(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_RecordActivityTaskHeartbeat_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.RecordActivityTaskHeartbeat(ctx, args.HeartbeatRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_RecordActivityTaskHeartbeat_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) RecordActivityTaskHeartbeatByID(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_RecordActivityTaskHeartbeatByID_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.RecordActivityTaskHeartbeatByID(ctx, args.HeartbeatRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_RecordActivityTaskHeartbeatByID_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) RegisterDomain(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_RegisterDomain_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	err := h.impl.RegisterDomain(ctx, args.RegisterRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_RegisterDomain_Helper.WrapResponse(err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) RequestCancelWorkflowExecution(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_RequestCancelWorkflowExecution_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	err := h.impl.RequestCancelWorkflowExecution(ctx, args.CancelRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_RequestCancelWorkflowExecution_Helper.WrapResponse(err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) ResetStickyTaskList(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_ResetStickyTaskList_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.ResetStickyTaskList(ctx, args.ResetRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_ResetStickyTaskList_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) RespondActivityTaskCanceled(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_RespondActivityTaskCanceled_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	err := h.impl.RespondActivityTaskCanceled(ctx, args.CanceledRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_RespondActivityTaskCanceled_Helper.WrapResponse(err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) RespondActivityTaskCanceledByID(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_RespondActivityTaskCanceledByID_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	err := h.impl.RespondActivityTaskCanceledByID(ctx, args.CanceledRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_RespondActivityTaskCanceledByID_Helper.WrapResponse(err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) RespondActivityTaskCompleted(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_RespondActivityTaskCompleted_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	err := h.impl.RespondActivityTaskCompleted(ctx, args.CompleteRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_RespondActivityTaskCompleted_Helper.WrapResponse(err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) RespondActivityTaskCompletedByID(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_RespondActivityTaskCompletedByID_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	err := h.impl.RespondActivityTaskCompletedByID(ctx, args.CompleteRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_RespondActivityTaskCompletedByID_Helper.WrapResponse(err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) RespondActivityTaskFailed(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_RespondActivityTaskFailed_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	err := h.impl.RespondActivityTaskFailed(ctx, args.FailRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_RespondActivityTaskFailed_Helper.WrapResponse(err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) RespondActivityTaskFailedByID(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_RespondActivityTaskFailedByID_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	err := h.impl.RespondActivityTaskFailedByID(ctx, args.FailRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_RespondActivityTaskFailedByID_Helper.WrapResponse(err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) RespondDecisionTaskCompleted(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_RespondDecisionTaskCompleted_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.RespondDecisionTaskCompleted(ctx, args.CompleteRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_RespondDecisionTaskCompleted_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) RespondDecisionTaskFailed(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_RespondDecisionTaskFailed_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	err := h.impl.RespondDecisionTaskFailed(ctx, args.FailedRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_RespondDecisionTaskFailed_Helper.WrapResponse(err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) RespondQueryTaskCompleted(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_RespondQueryTaskCompleted_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	err := h.impl.RespondQueryTaskCompleted(ctx, args.CompleteRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_RespondQueryTaskCompleted_Helper.WrapResponse(err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) SignalWithStartWorkflowExecution(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_SignalWithStartWorkflowExecution_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.SignalWithStartWorkflowExecution(ctx, args.SignalWithStartRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_SignalWithStartWorkflowExecution_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) SignalWorkflowExecution(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_SignalWorkflowExecution_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	err := h.impl.SignalWorkflowExecution(ctx, args.SignalRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_SignalWorkflowExecution_Helper.WrapResponse(err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) StartWorkflowExecution(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_StartWorkflowExecution_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.StartWorkflowExecution(ctx, args.StartRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_StartWorkflowExecution_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) TerminateWorkflowExecution(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_TerminateWorkflowExecution_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	err := h.impl.TerminateWorkflowExecution(ctx, args.TerminateRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_TerminateWorkflowExecution_Helper.WrapResponse(err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}

func (h handler) UpdateDomain(ctx context.Context, body wire.Value) (thrift.Response, error) {
	var args cadence.WorkflowService_UpdateDomain_Args
	if err := args.FromWire(body); err != nil {
		return thrift.Response{}, err
	}

	success, err := h.impl.UpdateDomain(ctx, args.UpdateRequest)

	hadError := err != nil
	result, err := cadence.WorkflowService_UpdateDomain_Helper.WrapResponse(success, err)

	var response thrift.Response
	if err == nil {
		response.IsApplicationError = hadError
		response.Body = result
	}
	return response, err
}
