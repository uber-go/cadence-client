// Copyright (c) 2017-2021 Uber Technologies Inc.
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

package api

import (
	"fmt"
	"strings"

	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
	"go.uber.org/yarpc/encoding/protobuf"
	"go.uber.org/yarpc/yarpcerrors"
)

func ConvertError(err error) error {
	status := yarpcerrors.FromError(err)
	if status == nil || status.Code() == yarpcerrors.CodeOK {
		return nil
	}

	switch status.Code() {
	case yarpcerrors.CodePermissionDenied:
		return &AccessDeniedError{
			Message: status.Message(),
		}
	case yarpcerrors.CodeInternal:
		return &InternalServiceError{
			Message: status.Message(),
		}
	case yarpcerrors.CodeNotFound:
		switch details := getErrorDetails(err).(type) {
		case *apiv1.EntityNotExistsError:
			return &EntityNotExistsError{
				Message:        status.Message(),
				CurrentCluster: details.CurrentCluster,
				ActiveCluster:  details.ActiveCluster,
			}
		case *apiv1.WorkflowExecutionAlreadyCompletedError:
			return &WorkflowExecutionAlreadyCompletedError{
				Message: status.Message(),
			}
		}
	case yarpcerrors.CodeInvalidArgument:
		switch getErrorDetails(err).(type) {
		case nil:
			return &BadRequestError{
				Message: status.Message(),
			}
		case *apiv1.QueryFailedError:
			return &QueryFailedError{
				Message: status.Message(),
			}
		}
	case yarpcerrors.CodeAlreadyExists:
		switch details := getErrorDetails(err).(type) {
		case *apiv1.CancellationAlreadyRequestedError:
			return &CancellationAlreadyRequestedError{
				Message: status.Message(),
			}
		case *apiv1.DomainAlreadyExistsError:
			return &DomainAlreadyExistsError{
				Message: status.Message(),
			}
		case *apiv1.WorkflowExecutionAlreadyStartedError:
			return &WorkflowExecutionAlreadyStartedError{
				Message:        status.Message(),
				StartRequestID: details.StartRequestId,
				RunID:          details.RunId,
			}
		}
	case yarpcerrors.CodeFailedPrecondition:
		switch details := getErrorDetails(err).(type) {
		case *apiv1.ClientVersionNotSupportedError:
			return &ClientVersionNotSupportedError{
				FeatureVersion:    details.FeatureVersion,
				ClientImpl:        details.ClientImpl,
				SupportedVersions: details.SupportedVersions,
			}
		case *apiv1.DomainNotActiveError:
			return &DomainNotActiveError{
				Message:        status.Message(),
				Domain:         details.Domain,
				CurrentCluster: details.CurrentCluster,
				ActiveCluster:  details.ActiveCluster,
			}
		}
	case yarpcerrors.CodeResourceExhausted:
		switch getErrorDetails(err).(type) {
		case *apiv1.LimitExceededError:
			return &LimitExceededError{
				Message: status.Message(),
			}
		case *apiv1.ServiceBusyError:
			return &ServiceBusyError{
				Message: status.Message(),
			}
		}
	}

	return err
}

type AccessDeniedError struct {
	Message string
}

type BadRequestError struct {
	Message string
}

type CancellationAlreadyRequestedError struct {
	Message string
}

type ClientVersionNotSupportedError struct {
	FeatureVersion    string
	ClientImpl        string
	SupportedVersions string
}

type DomainAlreadyExistsError struct {
	Message string
}

type DomainNotActiveError struct {
	Message        string
	Domain         string
	CurrentCluster string
	ActiveCluster  string
}

type EntityNotExistsError struct {
	Message        string
	CurrentCluster string
	ActiveCluster  string
}

type WorkflowExecutionAlreadyCompletedError struct {
	Message string
}

type InternalServiceError struct {
	Message string
}

type LimitExceededError struct {
	Message string
}

type QueryFailedError struct {
	Message string
}

type ServiceBusyError struct {
	Message string
}

type WorkflowExecutionAlreadyStartedError struct {
	Message        string
	StartRequestID string
	RunID          string
}

type EventAlreadyStartedError struct {
	Message string
}

func (err AccessDeniedError) Error() string {
	return fmt.Sprintf("AccessDeniedError{Message: %v}", err.Message)
}

func (err BadRequestError) Error() string {
	return fmt.Sprintf("BadRequestError{Message: %v}", err.Message)
}

func (err CancellationAlreadyRequestedError) Error() string {
	return fmt.Sprintf("CancellationAlreadyRequestedError{Message: %v}", err.Message)
}

func (err ClientVersionNotSupportedError) Error() string {
	return fmt.Sprintf("ClientVersionNotSupportedError{FeatureVersion: %v, ClientImpl: %v, SupportedVersions: %v}",
		err.FeatureVersion,
		err.ClientImpl,
		err.SupportedVersions)
}

func (err DomainAlreadyExistsError) Error() string {
	return fmt.Sprintf("DomainAlreadyExistsError{Message: %v}", err.Message)
}

func (err DomainNotActiveError) Error() string {
	return fmt.Sprintf("DomainNotActiveError{Message: %v, Domain: %v, CurrentCluster: %v, ActiveCluster: %v}",
		err.Message,
		err.Domain,
		err.CurrentCluster,
		err.ActiveCluster,
	)
}

func (err EntityNotExistsError) Error() string {
	sb := &strings.Builder{}
	printField(sb, "Message", err.Message)
	if err.CurrentCluster != "" {
		printField(sb, "CurrentCluster", err.CurrentCluster)
	}
	if err.ActiveCluster != "" {
		printField(sb, "ActiveCluster", err.ActiveCluster)
	}
	return fmt.Sprintf("EntityNotExistsError{%s}", sb.String())
}

func (err WorkflowExecutionAlreadyCompletedError) Error() string {
	sb := &strings.Builder{}
	printField(sb, "Message", err.Message)
	return fmt.Sprintf("WorkflowExecutionAlreadyCompletedError{%s}", sb.String())
}

func (err InternalServiceError) Error() string {
	return fmt.Sprintf("InternalServiceError{Message: %v}", err.Message)
}

func (err LimitExceededError) Error() string {
	return fmt.Sprintf("LimitExceededError{Message: %v}", err.Message)
}

func (err QueryFailedError) Error() string {
	return fmt.Sprintf("QueryFailedError{Message: %v}", err.Message)
}

func (err ServiceBusyError) Error() string {
	return fmt.Sprintf("ServiceBusyError{Message: %v}", err.Message)
}

func (err WorkflowExecutionAlreadyStartedError) Error() string {
	sb := &strings.Builder{}
	printField(sb, "Message", err.Message)
	printField(sb, "StartRequestID", err.StartRequestID)
	printField(sb, "RunID", err.RunID)
	return fmt.Sprintf("WorkflowExecutionAlreadyStartedError{%s}", sb.String())
}

func (err EventAlreadyStartedError) Error() string {
	return fmt.Sprintf("EventAlreadyStartedError{Message: %v}", err.Message)
}

func printField(sb *strings.Builder, field string, value interface{}) {
	if sb.Len() > 0 {
		fmt.Fprintf(sb, ", ")
	}
	fmt.Fprintf(sb, "%s: %v", field, value)
}

func getErrorDetails(err error) interface{} {
	details := protobuf.GetErrorDetails(err)
	if len(details) > 0 {
		return details[0]
	}
	return nil
}