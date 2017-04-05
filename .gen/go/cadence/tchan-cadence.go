// @generated Code generated by thrift-gen. Do not modify.

// Package cadence is generated code used to make or handle TChannel calls using Thrift.
package cadence

import (
	"fmt"

	athrift "github.com/apache/thrift/lib/go/thrift"
	"github.com/uber/tchannel-go/thrift"

	"github.com/uber-go/cadence-client/.gen/go/shared"
)

var _ = shared.GoUnusedProtection__

// Interfaces for the service and client for the services defined in the IDL.

// TChanWorkflowService is the interface that defines the server handler and client interface.
type TChanWorkflowService interface {
	DeprecateDomain(ctx thrift.Context, deprecateRequest *shared.DeprecateDomainRequest) error
	DescribeDomain(ctx thrift.Context, describeRequest *shared.DescribeDomainRequest) (*shared.DescribeDomainResponse, error)
	GetWorkflowExecutionHistory(ctx thrift.Context, getRequest *shared.GetWorkflowExecutionHistoryRequest) (*shared.GetWorkflowExecutionHistoryResponse, error)
	PollForActivityTask(ctx thrift.Context, pollRequest *shared.PollForActivityTaskRequest) (*shared.PollForActivityTaskResponse, error)
	PollForDecisionTask(ctx thrift.Context, pollRequest *shared.PollForDecisionTaskRequest) (*shared.PollForDecisionTaskResponse, error)
	RecordActivityTaskHeartbeat(ctx thrift.Context, heartbeatRequest *shared.RecordActivityTaskHeartbeatRequest) (*shared.RecordActivityTaskHeartbeatResponse, error)
	RegisterDomain(ctx thrift.Context, registerRequest *shared.RegisterDomainRequest) error
	RespondActivityTaskCanceled(ctx thrift.Context, canceledRequest *shared.RespondActivityTaskCanceledRequest) error
	RespondActivityTaskCompleted(ctx thrift.Context, completeRequest *shared.RespondActivityTaskCompletedRequest) error
	RespondActivityTaskFailed(ctx thrift.Context, failRequest *shared.RespondActivityTaskFailedRequest) error
	RespondDecisionTaskCompleted(ctx thrift.Context, completeRequest *shared.RespondDecisionTaskCompletedRequest) error
	StartWorkflowExecution(ctx thrift.Context, startRequest *shared.StartWorkflowExecutionRequest) (*shared.StartWorkflowExecutionResponse, error)
	UpdateDomain(ctx thrift.Context, updateRequest *shared.UpdateDomainRequest) (*shared.UpdateDomainResponse, error)
}

// Implementation of a client and service handler.

type tchanWorkflowServiceClient struct {
	thriftService string
	client        thrift.TChanClient
}

func NewTChanWorkflowServiceInheritedClient(thriftService string, client thrift.TChanClient) *tchanWorkflowServiceClient {
	return &tchanWorkflowServiceClient{
		thriftService,
		client,
	}
}

// NewTChanWorkflowServiceClient creates a client that can be used to make remote calls.
func NewTChanWorkflowServiceClient(client thrift.TChanClient) TChanWorkflowService {
	return NewTChanWorkflowServiceInheritedClient("WorkflowService", client)
}

func (c *tchanWorkflowServiceClient) DeprecateDomain(ctx thrift.Context, deprecateRequest *shared.DeprecateDomainRequest) error {
	var resp WorkflowServiceDeprecateDomainResult
	args := WorkflowServiceDeprecateDomainArgs{
		DeprecateRequest: deprecateRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "DeprecateDomain", &args, &resp)
	if err == nil && !success {
		switch {
		case resp.BadRequestError != nil:
			err = resp.BadRequestError
		case resp.InternalServiceError != nil:
			err = resp.InternalServiceError
		case resp.EntityNotExistError != nil:
			err = resp.EntityNotExistError
		default:
			err = fmt.Errorf("received no result or unknown exception for DeprecateDomain")
		}
	}

	return err
}

func (c *tchanWorkflowServiceClient) DescribeDomain(ctx thrift.Context, describeRequest *shared.DescribeDomainRequest) (*shared.DescribeDomainResponse, error) {
	var resp WorkflowServiceDescribeDomainResult
	args := WorkflowServiceDescribeDomainArgs{
		DescribeRequest: describeRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "DescribeDomain", &args, &resp)
	if err == nil && !success {
		switch {
		case resp.BadRequestError != nil:
			err = resp.BadRequestError
		case resp.InternalServiceError != nil:
			err = resp.InternalServiceError
		case resp.EntityNotExistError != nil:
			err = resp.EntityNotExistError
		default:
			err = fmt.Errorf("received no result or unknown exception for DescribeDomain")
		}
	}

	return resp.GetSuccess(), err
}

func (c *tchanWorkflowServiceClient) GetWorkflowExecutionHistory(ctx thrift.Context, getRequest *shared.GetWorkflowExecutionHistoryRequest) (*shared.GetWorkflowExecutionHistoryResponse, error) {
	var resp WorkflowServiceGetWorkflowExecutionHistoryResult
	args := WorkflowServiceGetWorkflowExecutionHistoryArgs{
		GetRequest: getRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "GetWorkflowExecutionHistory", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
		if e := resp.EntityNotExistError; e != nil {
			err = e
		}
	}

	return resp.GetSuccess(), err
}

func (c *tchanWorkflowServiceClient) PollForActivityTask(ctx thrift.Context, pollRequest *shared.PollForActivityTaskRequest) (*shared.PollForActivityTaskResponse, error) {
	var resp WorkflowServicePollForActivityTaskResult
	args := WorkflowServicePollForActivityTaskArgs{
		PollRequest: pollRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "PollForActivityTask", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
	}

	return resp.GetSuccess(), err
}

func (c *tchanWorkflowServiceClient) PollForDecisionTask(ctx thrift.Context, pollRequest *shared.PollForDecisionTaskRequest) (*shared.PollForDecisionTaskResponse, error) {
	var resp WorkflowServicePollForDecisionTaskResult
	args := WorkflowServicePollForDecisionTaskArgs{
		PollRequest: pollRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "PollForDecisionTask", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
	}

	return resp.GetSuccess(), err
}

func (c *tchanWorkflowServiceClient) RecordActivityTaskHeartbeat(ctx thrift.Context, heartbeatRequest *shared.RecordActivityTaskHeartbeatRequest) (*shared.RecordActivityTaskHeartbeatResponse, error) {
	var resp WorkflowServiceRecordActivityTaskHeartbeatResult
	args := WorkflowServiceRecordActivityTaskHeartbeatArgs{
		HeartbeatRequest: heartbeatRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "RecordActivityTaskHeartbeat", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
		if e := resp.EntityNotExistError; e != nil {
			err = e
		}
	}

	return resp.GetSuccess(), err
}

func (c *tchanWorkflowServiceClient) RegisterDomain(ctx thrift.Context, registerRequest *shared.RegisterDomainRequest) error {
	var resp WorkflowServiceRegisterDomainResult
	args := WorkflowServiceRegisterDomainArgs{
		RegisterRequest: registerRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "RegisterDomain", &args, &resp)
	if err == nil && !success {
		switch {
		case resp.BadRequestError != nil:
			err = resp.BadRequestError
		case resp.InternalServiceError != nil:
			err = resp.InternalServiceError
		case resp.DomainExistsError != nil:
			err = resp.DomainExistsError
		default:
			err = fmt.Errorf("received no result or unknown exception for RegisterDomain")
		}
	}

	return err
}

func (c *tchanWorkflowServiceClient) RespondActivityTaskCanceled(ctx thrift.Context, canceledRequest *shared.RespondActivityTaskCanceledRequest) error {
	var resp WorkflowServiceRespondActivityTaskCanceledResult
	args := WorkflowServiceRespondActivityTaskCanceledArgs{
		CanceledRequest: canceledRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "RespondActivityTaskCanceled", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
		if e := resp.EntityNotExistError; e != nil {
			err = e
		}
	}

	return err
}

func (c *tchanWorkflowServiceClient) RespondActivityTaskCompleted(ctx thrift.Context, completeRequest *shared.RespondActivityTaskCompletedRequest) error {
	var resp WorkflowServiceRespondActivityTaskCompletedResult
	args := WorkflowServiceRespondActivityTaskCompletedArgs{
		CompleteRequest: completeRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "RespondActivityTaskCompleted", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
		if e := resp.EntityNotExistError; e != nil {
			err = e
		}
	}

	return err
}

func (c *tchanWorkflowServiceClient) RespondActivityTaskFailed(ctx thrift.Context, failRequest *shared.RespondActivityTaskFailedRequest) error {
	var resp WorkflowServiceRespondActivityTaskFailedResult
	args := WorkflowServiceRespondActivityTaskFailedArgs{
		FailRequest: failRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "RespondActivityTaskFailed", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
		if e := resp.EntityNotExistError; e != nil {
			err = e
		}
	}

	return err
}

func (c *tchanWorkflowServiceClient) RespondDecisionTaskCompleted(ctx thrift.Context, completeRequest *shared.RespondDecisionTaskCompletedRequest) error {
	var resp WorkflowServiceRespondDecisionTaskCompletedResult
	args := WorkflowServiceRespondDecisionTaskCompletedArgs{
		CompleteRequest: completeRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "RespondDecisionTaskCompleted", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
		if e := resp.EntityNotExistError; e != nil {
			err = e
		}
	}

	return err
}

func (c *tchanWorkflowServiceClient) StartWorkflowExecution(ctx thrift.Context, startRequest *shared.StartWorkflowExecutionRequest) (*shared.StartWorkflowExecutionResponse, error) {
	var resp WorkflowServiceStartWorkflowExecutionResult
	args := WorkflowServiceStartWorkflowExecutionArgs{
		StartRequest: startRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "StartWorkflowExecution", &args, &resp)
	if err == nil && !success {
		if e := resp.BadRequestError; e != nil {
			err = e
		}
		if e := resp.InternalServiceError; e != nil {
			err = e
		}
		if e := resp.SessionAlreadyExistError; e != nil {
			err = e
		}
	}

	return resp.GetSuccess(), err
}

func (c *tchanWorkflowServiceClient) UpdateDomain(ctx thrift.Context, updateRequest *shared.UpdateDomainRequest) (*shared.UpdateDomainResponse, error) {
	var resp WorkflowServiceUpdateDomainResult
	args := WorkflowServiceUpdateDomainArgs{
		UpdateRequest: updateRequest,
	}
	success, err := c.client.Call(ctx, c.thriftService, "UpdateDomain", &args, &resp)
	if err == nil && !success {
		switch {
		case resp.BadRequestError != nil:
			err = resp.BadRequestError
		case resp.InternalServiceError != nil:
			err = resp.InternalServiceError
		case resp.EntityNotExistError != nil:
			err = resp.EntityNotExistError
		default:
			err = fmt.Errorf("received no result or unknown exception for UpdateDomain")
		}
	}

	return resp.GetSuccess(), err
}

type tchanWorkflowServiceServer struct {
	handler TChanWorkflowService
}

// NewTChanWorkflowServiceServer wraps a handler for TChanWorkflowService so it can be
// registered with a thrift.Server.
func NewTChanWorkflowServiceServer(handler TChanWorkflowService) thrift.TChanServer {
	return &tchanWorkflowServiceServer{
		handler,
	}
}

func (s *tchanWorkflowServiceServer) Service() string {
	return "WorkflowService"
}

func (s *tchanWorkflowServiceServer) Methods() []string {
	return []string{
		"DeprecateDomain",
		"DescribeDomain",
		"GetWorkflowExecutionHistory",
		"PollForActivityTask",
		"PollForDecisionTask",
		"RecordActivityTaskHeartbeat",
		"RegisterDomain",
		"RespondActivityTaskCanceled",
		"RespondActivityTaskCompleted",
		"RespondActivityTaskFailed",
		"RespondDecisionTaskCompleted",
		"StartWorkflowExecution",
		"UpdateDomain",
	}
}

func (s *tchanWorkflowServiceServer) Handle(ctx thrift.Context, methodName string, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	switch methodName {
	case "DeprecateDomain":
		return s.handleDeprecateDomain(ctx, protocol)
	case "DescribeDomain":
		return s.handleDescribeDomain(ctx, protocol)
	case "GetWorkflowExecutionHistory":
		return s.handleGetWorkflowExecutionHistory(ctx, protocol)
	case "PollForActivityTask":
		return s.handlePollForActivityTask(ctx, protocol)
	case "PollForDecisionTask":
		return s.handlePollForDecisionTask(ctx, protocol)
	case "RecordActivityTaskHeartbeat":
		return s.handleRecordActivityTaskHeartbeat(ctx, protocol)
	case "RegisterDomain":
		return s.handleRegisterDomain(ctx, protocol)
	case "RespondActivityTaskCanceled":
		return s.handleRespondActivityTaskCanceled(ctx, protocol)
	case "RespondActivityTaskCompleted":
		return s.handleRespondActivityTaskCompleted(ctx, protocol)
	case "RespondActivityTaskFailed":
		return s.handleRespondActivityTaskFailed(ctx, protocol)
	case "RespondDecisionTaskCompleted":
		return s.handleRespondDecisionTaskCompleted(ctx, protocol)
	case "StartWorkflowExecution":
		return s.handleStartWorkflowExecution(ctx, protocol)
	case "UpdateDomain":
		return s.handleUpdateDomain(ctx, protocol)

	default:
		return false, nil, fmt.Errorf("method %v not found in service %v", methodName, s.Service())
	}
}

func (s *tchanWorkflowServiceServer) handleDeprecateDomain(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req WorkflowServiceDeprecateDomainArgs
	var res WorkflowServiceDeprecateDomainResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	err :=
		s.handler.DeprecateDomain(ctx, req.DeprecateRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
	}

	return err == nil, &res, nil
}

func (s *tchanWorkflowServiceServer) handleDescribeDomain(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req WorkflowServiceDescribeDomainArgs
	var res WorkflowServiceDescribeDomainResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	r, err :=
		s.handler.DescribeDomain(ctx, req.DescribeRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
		res.Success = r
	}

	return err == nil, &res, nil
}

func (s *tchanWorkflowServiceServer) handleGetWorkflowExecutionHistory(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req WorkflowServiceGetWorkflowExecutionHistoryArgs
	var res WorkflowServiceGetWorkflowExecutionHistoryResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	r, err :=
		s.handler.GetWorkflowExecutionHistory(ctx, req.GetRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
		res.Success = r
	}

	return err == nil, &res, nil
}

func (s *tchanWorkflowServiceServer) handlePollForActivityTask(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req WorkflowServicePollForActivityTaskArgs
	var res WorkflowServicePollForActivityTaskResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	r, err :=
		s.handler.PollForActivityTask(ctx, req.PollRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		default:
			return false, nil, err
		}
	} else {
		res.Success = r
	}

	return err == nil, &res, nil
}

func (s *tchanWorkflowServiceServer) handlePollForDecisionTask(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req WorkflowServicePollForDecisionTaskArgs
	var res WorkflowServicePollForDecisionTaskResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	r, err :=
		s.handler.PollForDecisionTask(ctx, req.PollRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		default:
			return false, nil, err
		}
	} else {
		res.Success = r
	}

	return err == nil, &res, nil
}

func (s *tchanWorkflowServiceServer) handleRecordActivityTaskHeartbeat(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req WorkflowServiceRecordActivityTaskHeartbeatArgs
	var res WorkflowServiceRecordActivityTaskHeartbeatResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	r, err :=
		s.handler.RecordActivityTaskHeartbeat(ctx, req.HeartbeatRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
		res.Success = r
	}

	return err == nil, &res, nil
}

func (s *tchanWorkflowServiceServer) handleRegisterDomain(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req WorkflowServiceRegisterDomainArgs
	var res WorkflowServiceRegisterDomainResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	err :=
		s.handler.RegisterDomain(ctx, req.RegisterRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.DomainAlreadyExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for domainExistsError returned non-nil error type *shared.DomainAlreadyExistsError but nil value")
			}
			res.DomainExistsError = v
		default:
			return false, nil, err
		}
	} else {
	}

	return err == nil, &res, nil
}

func (s *tchanWorkflowServiceServer) handleRespondActivityTaskCanceled(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req WorkflowServiceRespondActivityTaskCanceledArgs
	var res WorkflowServiceRespondActivityTaskCanceledResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	err :=
		s.handler.RespondActivityTaskCanceled(ctx, req.CanceledRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
	}

	return err == nil, &res, nil
}

func (s *tchanWorkflowServiceServer) handleRespondActivityTaskCompleted(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req WorkflowServiceRespondActivityTaskCompletedArgs
	var res WorkflowServiceRespondActivityTaskCompletedResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	err :=
		s.handler.RespondActivityTaskCompleted(ctx, req.CompleteRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
	}

	return err == nil, &res, nil
}

func (s *tchanWorkflowServiceServer) handleRespondActivityTaskFailed(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req WorkflowServiceRespondActivityTaskFailedArgs
	var res WorkflowServiceRespondActivityTaskFailedResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	err :=
		s.handler.RespondActivityTaskFailed(ctx, req.FailRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
	}

	return err == nil, &res, nil
}

func (s *tchanWorkflowServiceServer) handleRespondDecisionTaskCompleted(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req WorkflowServiceRespondDecisionTaskCompletedArgs
	var res WorkflowServiceRespondDecisionTaskCompletedResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	err :=
		s.handler.RespondDecisionTaskCompleted(ctx, req.CompleteRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
	}

	return err == nil, &res, nil
}

func (s *tchanWorkflowServiceServer) handleStartWorkflowExecution(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req WorkflowServiceStartWorkflowExecutionArgs
	var res WorkflowServiceStartWorkflowExecutionResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	r, err :=
		s.handler.StartWorkflowExecution(ctx, req.StartRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.WorkflowExecutionAlreadyStartedError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for sessionAlreadyExistError returned non-nil error type *shared.WorkflowExecutionAlreadyStartedError but nil value")
			}
			res.SessionAlreadyExistError = v
		default:
			return false, nil, err
		}
	} else {
		res.Success = r
	}

	return err == nil, &res, nil
}

func (s *tchanWorkflowServiceServer) handleUpdateDomain(ctx thrift.Context, protocol athrift.TProtocol) (bool, athrift.TStruct, error) {
	var req WorkflowServiceUpdateDomainArgs
	var res WorkflowServiceUpdateDomainResult

	if err := req.Read(protocol); err != nil {
		return false, nil, err
	}

	r, err :=
		s.handler.UpdateDomain(ctx, req.UpdateRequest)

	if err != nil {
		switch v := err.(type) {
		case *shared.BadRequestError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for badRequestError returned non-nil error type *shared.BadRequestError but nil value")
			}
			res.BadRequestError = v
		case *shared.InternalServiceError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for internalServiceError returned non-nil error type *shared.InternalServiceError but nil value")
			}
			res.InternalServiceError = v
		case *shared.EntityNotExistsError:
			if v == nil {
				return false, nil, fmt.Errorf("Handler for entityNotExistError returned non-nil error type *shared.EntityNotExistsError but nil value")
			}
			res.EntityNotExistError = v
		default:
			return false, nil, err
		}
	} else {
		res.Success = r
	}

	return err == nil, &res, nil
}
