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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkflowRegistration(t *testing.T) {
	tests := []struct {
		msg               string
		register          func(r *registry)
		workflowType      string
		resolveByFunction interface{}
		resolveByAlias    string
	}{
		{
			msg:               "register workflow function",
			register:          func(r *registry) { r.RegisterWorkflow(testWorkflowFunction) },
			workflowType:      "go.uber.org/cadence/internal.testWorkflowFunction",
			resolveByFunction: testWorkflowFunction,
		},
		{
			msg: "register workflow function with short name",
			register: func(r *registry) {
				r.RegisterWorkflowWithOptions(testWorkflowFunction, RegisterWorkflowOptions{EnableShortName: true})
			},
			workflowType:      "testWorkflowFunction",
			resolveByFunction: testWorkflowFunction,
		},
		{
			msg: "register workflow function with alias",
			register: func(r *registry) {
				r.RegisterWorkflowWithOptions(testWorkflowFunction, RegisterWorkflowOptions{Name: "workflow.alias"})
			},
			workflowType:      "workflow.alias",
			resolveByFunction: testWorkflowFunction,
			resolveByAlias:    "workflow.alias",
		},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			r := newRegistry()
			tt.register(r)

			// Verify registered workflow type
			workflowType := r.getRegisteredWorkflowTypes()[0]
			require.Equal(t, tt.workflowType, workflowType)

			// Verify workflow is resolved from workflow type
			_, ok := r.getWorkflowFn(tt.workflowType)
			require.True(t, ok)

			// Verify resolving by function reference
			workflowType = getWorkflowFunctionName(r, tt.resolveByFunction)
			require.Equal(t, tt.workflowType, workflowType)

			// Verify resolving by alias
			if tt.resolveByAlias != "" {
				workflowType = getWorkflowFunctionName(r, tt.resolveByAlias)
				require.Equal(t, tt.workflowType, workflowType)
			}
		})
	}
}

func TestActivityRegistration(t *testing.T) {
	tests := []struct {
		msg               string
		register          func(r *registry)
		activityType      string
		resolveByFunction interface{}
		resolveByAlias    string
	}{
		{
			msg:               "register activity function",
			register:          func(r *registry) { r.RegisterActivity(testActivityFunction) },
			activityType:      "go.uber.org/cadence/internal.testActivityFunction",
			resolveByFunction: testActivityFunction,
		},
		{
			msg: "register activity function with short name",
			register: func(r *registry) {
				r.RegisterActivityWithOptions(testActivityFunction, RegisterActivityOptions{EnableShortName: true})
			},
			activityType:      "testActivityFunction",
			resolveByFunction: testActivityFunction,
		},
		{
			msg: "register activity function with an alias",
			register: func(r *registry) {
				r.RegisterActivityWithOptions(testActivityFunction, RegisterActivityOptions{Name: "activity.alias"})
			},
			activityType:      "activity.alias",
			resolveByFunction: testActivityFunction,
			resolveByAlias:    "activity.alias",
		},
		{
			msg:               "register activity struct",
			register:          func(r *registry) { r.RegisterActivity(&testActivityStruct{}) },
			activityType:      "go.uber.org/cadence/internal.(*testActivityStruct).Method",
			resolveByFunction: (&testActivityStruct{}).Method,
		},
		{
			msg: "register activity struct with short name",
			register: func(r *registry) {
				r.RegisterActivityWithOptions(&testActivityStruct{}, RegisterActivityOptions{EnableShortName: true})
			},
			activityType:      "Method",
			resolveByFunction: (&testActivityStruct{}).Method,
		},
		{
			msg: "register activity struct with a prefix",
			register: func(r *registry) {
				r.RegisterActivityWithOptions(&testActivityStruct{}, RegisterActivityOptions{Name: "prefix."})
			},
			activityType:      "prefix.Method",
			resolveByFunction: (&testActivityStruct{}).Method,
			resolveByAlias:    "prefix.Method",
		},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			r := newRegistry()
			tt.register(r)

			// Verify registered activity type
			activityType := r.getRegisteredActivities()[0].ActivityType().Name
			require.Equal(t, tt.activityType, activityType, "activity type")

			// Verify activity is resolved from activity type
			_, ok := r.GetActivity(tt.activityType)
			require.True(t, ok)

			// Verify resolving by function reference
			activityType = getActivityFunctionName(r, tt.resolveByFunction)
			require.Equal(t, tt.activityType, activityType, "resolve by function reference")

			// Verify resolving by alias
			if tt.resolveByAlias != "" {
				activityType = getActivityFunctionName(r, tt.resolveByAlias)
				require.Equal(t, tt.activityType, activityType, "resolve by alias")
			}
		})
	}
}

type testActivityStruct struct{}

func (ts *testActivityStruct) Method() error { return nil }

func testActivityFunction() error            { return nil }
func testWorkflowFunction(ctx Context) error { return nil }
