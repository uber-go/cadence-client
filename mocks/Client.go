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

// Code generated by mockery v1.0.0.
// Modified manually for type alias to work correctly.
// https://github.com/vektra/mockery/issues/236
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/internal"
	"go.uber.org/cadence/workflow"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// CancelWorkflow provides a mock function with given fields: ctx, workflowID, runID
func (_m *Client) CancelWorkflow(ctx context.Context, workflowID string, runID string) error {
	ret := _m.Called(ctx, workflowID, runID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, workflowID, runID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CompleteActivity provides a mock function with given fields: ctx, taskToken, result, err
func (_m *Client) CompleteActivity(ctx context.Context, taskToken []byte, result interface{}, err error) error {
	ret := _m.Called(ctx, taskToken, result, err)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []byte, interface{}, error) error); ok {
		r0 = rf(ctx, taskToken, result, err)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CompleteActivityByID provides a mock function with given fields: ctx, domain, workflowID, runID, activityID, result, err
func (_m *Client) CompleteActivityByID(ctx context.Context, domain string, workflowID string, runID string, activityID string, result interface{}, err error) error {
	ret := _m.Called(ctx, domain, workflowID, runID, activityID, result, err)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string, interface{}, error) error); ok {
		r0 = rf(ctx, domain, workflowID, runID, activityID, result, err)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CountWorkflow provides a mock function with given fields: ctx, request
func (_m *Client) CountWorkflow(ctx context.Context, request *shared.CountWorkflowExecutionsRequest) (*shared.CountWorkflowExecutionsResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *shared.CountWorkflowExecutionsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.CountWorkflowExecutionsRequest) *shared.CountWorkflowExecutionsResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.CountWorkflowExecutionsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.CountWorkflowExecutionsRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DescribeTaskList provides a mock function with given fields: ctx, tasklist, tasklistType
func (_m *Client) DescribeTaskList(ctx context.Context, tasklist string, tasklistType shared.TaskListType) (*shared.DescribeTaskListResponse, error) {
	ret := _m.Called(ctx, tasklist, tasklistType)

	var r0 *shared.DescribeTaskListResponse
	if rf, ok := ret.Get(0).(func(context.Context, string, shared.TaskListType) *shared.DescribeTaskListResponse); ok {
		r0 = rf(ctx, tasklist, tasklistType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.DescribeTaskListResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, shared.TaskListType) error); ok {
		r1 = rf(ctx, tasklist, tasklistType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DescribeWorkflowExecution provides a mock function with given fields: ctx, workflowID, runID
func (_m *Client) DescribeWorkflowExecution(ctx context.Context, workflowID string, runID string) (*shared.DescribeWorkflowExecutionResponse, error) {
	ret := _m.Called(ctx, workflowID, runID)

	var r0 *shared.DescribeWorkflowExecutionResponse
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *shared.DescribeWorkflowExecutionResponse); ok {
		r0 = rf(ctx, workflowID, runID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.DescribeWorkflowExecutionResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, workflowID, runID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ExecuteWorkflow provides a mock function with given fields: ctx, options, workflow, args
func (_m *Client) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error) {
	var _ca []interface{}
	_ca = append(_ca, ctx, options, workflow)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 internal.WorkflowRun
	if rf, ok := ret.Get(0).(func(context.Context, internal.StartWorkflowOptions, interface{}, ...interface{}) internal.WorkflowRun); ok {
		r0 = rf(ctx, options, workflow, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(internal.WorkflowRun)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, internal.StartWorkflowOptions, interface{}, ...interface{}) error); ok {
		r1 = rf(ctx, options, workflow, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSearchAttributes provides a mock function with given fields: ctx
func (_m *Client) GetSearchAttributes(ctx context.Context) (*shared.GetSearchAttributesResponse, error) {
	ret := _m.Called(ctx)

	var r0 *shared.GetSearchAttributesResponse
	if rf, ok := ret.Get(0).(func(context.Context) *shared.GetSearchAttributesResponse); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.GetSearchAttributesResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetWorkflow provides a mock function with given fields: ctx, workflowID, runID
func (_m *Client) GetWorkflow(ctx context.Context, workflowID string, runID string) client.WorkflowRun {
	ret := _m.Called(ctx, workflowID, runID)

	var r0 client.WorkflowRun
	if rf, ok := ret.Get(0).(func(context.Context, string, string) client.WorkflowRun); ok {
		r0 = rf(ctx, workflowID, runID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(client.WorkflowRun)
		}
	}

	return r0
}

// GetWorkflowHistory provides a mock function with given fields: ctx, workflowID, runID
func (_m *Client) GetWorkflowHistory(ctx context.Context, workflowID string, runID string) client.HistoryEventIterator {
	ret := _m.Called(ctx, workflowID, runID)

	var r0 internal.HistoryEventIterator
	if rf, ok := ret.Get(0).(func(context.Context, string, string) internal.HistoryEventIterator); ok {
		r0 = rf(ctx, workflowID, runID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(internal.HistoryEventIterator)
		}
	}

	return r0
}

// PollWorkflowHistory provides a mock function with given fields: ctx, workflowID, runID and filterType
func (_m *Client) PollWorkflowHistory(ctx context.Context, workflowID string, runID string, filterType shared.HistoryEventFilterType) client.HistoryEventIterator {
	ret := _m.Called(ctx, workflowID, runID, filterType)

	var r0 internal.HistoryEventIterator
	if rf, ok := ret.Get(0).(func(context.Context, string, string, shared.HistoryEventFilterType) internal.HistoryEventIterator); ok {
		r0 = rf(ctx, workflowID, runID, filterType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(internal.HistoryEventIterator)
		}
	}

	return r0
}

// ListClosedWorkflow provides a mock function with given fields: ctx, request
func (_m *Client) ListClosedWorkflow(ctx context.Context, request *shared.ListClosedWorkflowExecutionsRequest) (*shared.ListClosedWorkflowExecutionsResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *shared.ListClosedWorkflowExecutionsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.ListClosedWorkflowExecutionsRequest) *shared.ListClosedWorkflowExecutionsResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.ListClosedWorkflowExecutionsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.ListClosedWorkflowExecutionsRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListOpenWorkflow provides a mock function with given fields: ctx, request
func (_m *Client) ListOpenWorkflow(ctx context.Context, request *shared.ListOpenWorkflowExecutionsRequest) (*shared.ListOpenWorkflowExecutionsResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *shared.ListOpenWorkflowExecutionsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.ListOpenWorkflowExecutionsRequest) *shared.ListOpenWorkflowExecutionsResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.ListOpenWorkflowExecutionsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.ListOpenWorkflowExecutionsRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListWorkflow provides a mock function with given fields: ctx, request
func (_m *Client) ListWorkflow(ctx context.Context, request *shared.ListWorkflowExecutionsRequest) (*shared.ListWorkflowExecutionsResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *shared.ListWorkflowExecutionsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.ListWorkflowExecutionsRequest) *shared.ListWorkflowExecutionsResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.ListWorkflowExecutionsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.ListWorkflowExecutionsRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListArchivedWorkflow provides a mock function with given fields: ctx, request
func (_m *Client) ListArchivedWorkflow(ctx context.Context, request *shared.ListArchivedWorkflowExecutionsRequest) (*shared.ListArchivedWorkflowExecutionsResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *shared.ListArchivedWorkflowExecutionsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.ListArchivedWorkflowExecutionsRequest) *shared.ListArchivedWorkflowExecutionsResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.ListArchivedWorkflowExecutionsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.ListArchivedWorkflowExecutionsRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// QueryWorkflow provides a mock function with given fields: ctx, workflowID, runID, queryType, args
func (_m *Client) QueryWorkflow(ctx context.Context, workflowID string, runID string, queryType string, args ...interface{}) (encoded.Value, error) {
	var _ca []interface{}
	_ca = append(_ca, ctx, workflowID, runID, queryType)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 encoded.Value
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, ...interface{}) encoded.Value); ok {
		r0 = rf(ctx, workflowID, runID, queryType, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(encoded.Value)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, ...interface{}) error); ok {
		r1 = rf(ctx, workflowID, runID, queryType, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// QueryWorkflowWithOptions provides a mock function with given fields: ctx, request
func (_m *Client) QueryWorkflowWithOptions(ctx context.Context, request *client.QueryWorkflowWithOptionsRequest) (*client.QueryWorkflowWithOptionsResponse, error) {
	var _ca []interface{}
	_ca = append(_ca, ctx, request)
	ret := _m.Called(_ca...)

	var r0 *client.QueryWorkflowWithOptionsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *client.QueryWorkflowWithOptionsRequest) *client.QueryWorkflowWithOptionsResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*client.QueryWorkflowWithOptionsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *client.QueryWorkflowWithOptionsRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RecordActivityHeartbeat provides a mock function with given fields: ctx, taskToken, details
func (_m *Client) RecordActivityHeartbeat(ctx context.Context, taskToken []byte, details ...interface{}) error {
	var _ca []interface{}
	_ca = append(_ca, ctx, taskToken)
	_ca = append(_ca, details...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []byte, ...interface{}) error); ok {
		r0 = rf(ctx, taskToken, details...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RecordActivityHeartbeatByID provides a mock function with given fields: ctx, domain, workflowID, runID, activityID, details
func (_m *Client) RecordActivityHeartbeatByID(ctx context.Context, domain string, workflowID string, runID string, activityID string, details ...interface{}) error {
	var _ca []interface{}
	_ca = append(_ca, ctx, domain, workflowID, runID, activityID)
	_ca = append(_ca, details...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string, ...interface{}) error); ok {
		r0 = rf(ctx, domain, workflowID, runID, activityID, details...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ScanWorkflow provides a mock function with given fields: ctx, request
func (_m *Client) ScanWorkflow(ctx context.Context, request *shared.ListWorkflowExecutionsRequest) (*shared.ListWorkflowExecutionsResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *shared.ListWorkflowExecutionsResponse
	if rf, ok := ret.Get(0).(func(context.Context, *shared.ListWorkflowExecutionsRequest) *shared.ListWorkflowExecutionsResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*shared.ListWorkflowExecutionsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *shared.ListWorkflowExecutionsRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SignalWithStartWorkflow provides a mock function with given fields: ctx, workflowID, signalName, signalArg, options, workflow, workflowArgs
func (_m *Client) SignalWithStartWorkflow(ctx context.Context, workflowID string, signalName string, signalArg interface{}, options client.StartWorkflowOptions, workflow interface{}, workflowArgs ...interface{}) (*workflow.Execution, error) {
	var _ca []interface{}
	_ca = append(_ca, ctx, workflowID, signalName, signalArg, options, workflow)
	_ca = append(_ca, workflowArgs...)
	ret := _m.Called(_ca...)

	var r0 *internal.WorkflowExecution
	if rf, ok := ret.Get(0).(func(context.Context, string, string, interface{}, client.StartWorkflowOptions, interface{}, ...interface{}) *internal.WorkflowExecution); ok {
		r0 = rf(ctx, workflowID, signalName, signalArg, options, workflow, workflowArgs...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*internal.WorkflowExecution)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, interface{}, client.StartWorkflowOptions, interface{}, ...interface{}) error); ok {
		r1 = rf(ctx, workflowID, signalName, signalArg, options, workflow, workflowArgs...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SignalWorkflow provides a mock function with given fields: ctx, workflowID, runID, signalName, arg
func (_m *Client) SignalWorkflow(ctx context.Context, workflowID string, runID string, signalName string, arg interface{}) error {
	ret := _m.Called(ctx, workflowID, runID, signalName, arg)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, interface{}) error); ok {
		r0 = rf(ctx, workflowID, runID, signalName, arg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StartWorkflow provides a mock function with given fields: ctx, options, workflow, args
func (_m *Client) StartWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (*workflow.Execution, error) {
	var _ca []interface{}
	_ca = append(_ca, ctx, options, workflow)
	_ca = append(_ca, args...)
	ret := _m.Called(_ca...)

	var r0 *internal.WorkflowExecution
	if rf, ok := ret.Get(0).(func(context.Context, client.StartWorkflowOptions, interface{}, ...interface{}) *internal.WorkflowExecution); ok {
		r0 = rf(ctx, options, workflow, args...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*internal.WorkflowExecution)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, client.StartWorkflowOptions, interface{}, ...interface{}) error); ok {
		r1 = rf(ctx, options, workflow, args...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TerminateWorkflow provides a mock function with given fields: ctx, workflowID, runID, reason, details
func (_m *Client) TerminateWorkflow(ctx context.Context, workflowID string, runID string, reason string, details []byte) error {
	ret := _m.Called(ctx, workflowID, runID, reason, details)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, []byte) error); ok {
		r0 = rf(ctx, workflowID, runID, reason, details)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
