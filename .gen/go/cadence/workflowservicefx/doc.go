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

// Code generated by thriftrw-plugin-yarpc
// @generated

// Package workflowservicefx provides better integration for Fx for services
// implementing or calling WorkflowService.
//
// # Clients
//
// If you are making requests to WorkflowService, use the Client function to inject a
// WorkflowService client into your container.
//
//	fx.Provide(workflowservicefx.Client("..."))
//
// # Servers
//
// If you are implementing WorkflowService, provide a workflowserviceserver.Interface into
// the container and use the Server function.
//
// Given,
//
//	func NewWorkflowServiceHandler() workflowserviceserver.Interface
//
// You can do the following to have the procedures of WorkflowService made available
// to an Fx application.
//
//	fx.Provide(
//		NewWorkflowServiceHandler,
//		workflowservicefx.Server(),
//	)
package workflowservicefx
