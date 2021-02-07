// Copyright (c) 2017-2020 Uber Technologies Inc.
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

package internal

import (
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/zap"
)

type (
	// WorkflowShadowerOptions configs WorkflowShadower
	WorkflowShadowerOptions struct {
		Domain string

		WorkflowQuery           string
		WorkflowTypes           []string
		WorkflowCloseStatus     []string
		WorkflowStartTimeFilter string
		SamplingRate            float64

		// ExitCondition

		logger zap.Logger
	}

	// WorkflowShadower retrieves and replays workflow history from Cadence service to determine if there's any nondeterministic changes in the workflow definition
	WorkflowShadower struct {
		service  workflowserviceclient.Interface
		config   *WorkflowShadowerOptions
		replayer *WorkflowReplayer
	}
)

// NewWorkflowShadower creates an instance of the WorkflowShadower
func NewWorkflowShadower(
	service workflowserviceclient.Interface,
	config *WorkflowShadowerOptions,
) *WorkflowShadower {
	return &WorkflowShadower{
		service:  service,
		config:   config,
		replayer: NewWorkflowReplayer(),
	}
}

// RegisterWorkflow registers workflow function to replay
func (s *WorkflowShadower) RegisterWorkflow(w interface{}) {
	s.replayer.RegisterWorkflow(w)
}

// RegisterWorkflowWithOptions registers workflow function with custom workflow name to replay
func (s *WorkflowShadower) RegisterWorkflowWithOptions(w interface{}, options RegisterWorkflowOptions) {
	s.replayer.RegisterWorkflowWithOptions(w, options)
}

func (s *WorkflowShadower) Start() {}
func (s *WorkflowShadower) Run()   {}
func (s *WorkflowShadower) Stop()  {}

func (s *WorkflowShadower) shadowWorker() {

}
