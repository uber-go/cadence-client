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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
	"go.uber.org/yarpc/encoding/protobuf"
	"go.uber.org/yarpc/yarpcerrors"
)

func Test_ConvertError(t *testing.T) {
	tests := []struct{
		transportErr error
		sdkError error
	} {
		{
			transportErr: nil,
			sdkError: nil,
		},
		{
			transportErr: protobuf.NewError(yarpcerrors.CodeOK, ""),
			sdkError: nil,
		},
		{
			transportErr: protobuf.NewError(yarpcerrors.CodePermissionDenied, "message"),
			sdkError: &AccessDeniedError{Message: "message"},
		},
		{
			transportErr: protobuf.NewError(yarpcerrors.CodeInternal, "message"),
			sdkError: &InternalServiceError{Message: "message"},
		},
		{
			transportErr: protobuf.NewError(yarpcerrors.CodeNotFound, "message", protobuf.WithErrorDetails(
				&apiv1.EntityNotExistsError{
					CurrentCluster: "current",
					ActiveCluster:  "active",
				})),
			sdkError: &EntityNotExistsError{
				Message: "message",
				CurrentCluster: "current",
				ActiveCluster: "active",
			},
		},
		{
			transportErr: protobuf.NewError(yarpcerrors.CodeNotFound, "message", protobuf.WithErrorDetails(
				&apiv1.WorkflowExecutionAlreadyCompletedError{})),
			sdkError: &WorkflowExecutionAlreadyCompletedError{Message: "message"},
		},
		{
			transportErr: protobuf.NewError(yarpcerrors.CodeInvalidArgument, "message"),
			sdkError: &BadRequestError{Message: "message"},
		},
		{
			transportErr: protobuf.NewError(yarpcerrors.CodeInvalidArgument, "message", protobuf.WithErrorDetails(
				&apiv1.QueryFailedError{})),
			sdkError: &QueryFailedError{Message: "message"},
		},
		{
			transportErr: protobuf.NewError(yarpcerrors.CodeAlreadyExists, "message", protobuf.WithErrorDetails(
				&apiv1.CancellationAlreadyRequestedError{})),
			sdkError: &CancellationAlreadyRequestedError{Message: "message"},
		},
		{
			transportErr: protobuf.NewError(yarpcerrors.CodeAlreadyExists, "message", protobuf.WithErrorDetails(
				&apiv1.DomainAlreadyExistsError{})),
			sdkError: &DomainAlreadyExistsError{Message: "message"},
		},
		{
			transportErr: protobuf.NewError(yarpcerrors.CodeAlreadyExists, "message", protobuf.WithErrorDetails(
				&apiv1.WorkflowExecutionAlreadyStartedError{
					StartRequestId: "startRequestId",
					RunId: "runId",
				})),
			sdkError: &WorkflowExecutionAlreadyStartedError{
				Message: "message",
				StartRequestID: "startRequestId",
				RunID: "runId",
			},
		},
		{
			transportErr: protobuf.NewError(yarpcerrors.CodeFailedPrecondition, "message", protobuf.WithErrorDetails(
				&apiv1.ClientVersionNotSupportedError{
					FeatureVersion: "featureVersion",
					ClientImpl: "clientImpl",
					SupportedVersions: "supportedVersions",
				})),
			sdkError: &ClientVersionNotSupportedError{
				FeatureVersion: "featureVersion",
				ClientImpl: "clientImpl",
				SupportedVersions: "supportedVersions",
			},
		},
		{
			transportErr: protobuf.NewError(yarpcerrors.CodeFailedPrecondition, "message", protobuf.WithErrorDetails(
				&apiv1.DomainNotActiveError{
					Domain: "domain",
					CurrentCluster: "current",
					ActiveCluster: "active",
				})),
			sdkError: &DomainNotActiveError{
				Message: "message",
				Domain: "domain",
				CurrentCluster: "current",
				ActiveCluster: "active",
			},
		},
		{
			transportErr: protobuf.NewError(yarpcerrors.CodeResourceExhausted, "message", protobuf.WithErrorDetails(
				&apiv1.LimitExceededError{})),
			sdkError: &LimitExceededError{Message: "message"},
		},
		{
			transportErr: protobuf.NewError(yarpcerrors.CodeResourceExhausted, "message", protobuf.WithErrorDetails(
				&apiv1.ServiceBusyError{})),
			sdkError: &ServiceBusyError{Message: "message"},
		},
		{
			transportErr: protobuf.NewError(yarpcerrors.CodeUnknown, "message"),
			sdkError: protobuf.NewError(yarpcerrors.CodeUnknown, "message"),
		},
		{
			transportErr: errors.New("non-proto error"),
			sdkError: errors.New("non-proto error"),
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.sdkError, ConvertError(tt.transportErr))
	}
}