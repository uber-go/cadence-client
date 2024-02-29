// Copyright (c) 2017-2020 Uber Technologies Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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
	"context"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSetMemoOnStart(t *testing.T) {
	t.Parallel()
	env := newTestWorkflowEnv(t)

	memo := map[string]interface{}{
		"key": make(chan int),
	}
	err := env.SetMemoOnStart(memo)
	require.Error(t, err)

	memo = map[string]interface{}{
		"memoKey": "memo",
	}
	require.Nil(t, env.impl.workflowInfo.Memo)
	err = env.SetMemoOnStart(memo)
	require.NoError(t, err)
	require.NotNil(t, env.impl.workflowInfo.Memo)
}

func TestSetSearchAttributesOnStart(t *testing.T) {
	t.Parallel()
	env := newTestWorkflowEnv(t)

	invalidSearchAttr := map[string]interface{}{
		"key": make(chan int),
	}
	err := env.SetSearchAttributesOnStart(invalidSearchAttr)
	require.Error(t, err)

	searchAttr := map[string]interface{}{
		"CustomIntField": 1,
	}
	err = env.SetSearchAttributesOnStart(searchAttr)
	require.NoError(t, err)
	require.NotNil(t, env.impl.workflowInfo.SearchAttributes)
}

func TestUnregisteredActivity(t *testing.T) {
	t.Parallel()
	env := newTestWorkflowEnv(t)
	workflow := func(ctx Context) error {
		ctx = WithActivityOptions(ctx, ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
		})
		return ExecuteActivity(ctx, "unregistered").Get(ctx, nil)
	}
	env.RegisterWorkflow(workflow)
	env.ExecuteWorkflow(workflow)
	require.Error(t, env.GetWorkflowError())
	ee := env.GetWorkflowError()
	require.NotNil(t, ee)
	require.True(t, strings.HasPrefix(ee.Error(), "unable to find activityType=unregistered"), ee.Error())
}

func TestNoExplicitRegistrationRequired(t *testing.T) {
	env := newTestWorkflowEnv(t)
	activity := func(ctx context.Context, arg string) (string, error) { return arg + " World!", nil }
	env.RegisterActivity(activity)
	env.ExecuteWorkflow(func(ctx Context, arg1 string) (string, error) {
		ctx = WithActivityOptions(ctx, ActivityOptions{
			ScheduleToCloseTimeout: time.Hour,
			StartToCloseTimeout:    time.Hour,
			ScheduleToStartTimeout: time.Hour,
		})
		var result string
		err := ExecuteActivity(ctx, activity, arg1).Get(ctx, &result)
		if err != nil {
			return "", err
		}
		return result, nil
	}, "Hello")
	require.NoError(t, env.GetWorkflowError())
	var result string
	err := env.GetWorkflowResult(&result)
	require.NoError(t, err)
	require.Equal(t, "Hello World!", result)
}

func TestWorkflowReturnNil(t *testing.T) {
	env := newTestWorkflowEnv(t)

	var isExecuted bool
	testWF := func(ctx Context) error {
		isExecuted = true
		return nil
	}
	env.ExecuteWorkflow(testWF)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.True(t, isExecuted)

	var r struct{}
	err := env.GetWorkflowResult(&r)
	require.NoError(t, err)
}

type InterceptorTestSuite struct {
	suite.Suite
	WorkflowTestSuite

	env         *TestWorkflowEnvironment
	testFactory InterceptorFactory
}

type InterceptorFactory struct {
	workflowInterceptorInvocationCounter      int
	childWorkflowInterceptorInvocationCounter int
}

type Interceptor struct {
	WorkflowInterceptorBase
	workflowInterceptorInvocationCounter      *int
	childWorkflowInterceptorInvocationCounter *int
}

func (i *Interceptor) ExecuteWorkflow(ctx Context, workflowType string, args ...interface{}) []interface{} {
	*i.workflowInterceptorInvocationCounter += 1
	return i.Next.ExecuteWorkflow(ctx, workflowType, args...)
}
func (i *Interceptor) ExecuteChildWorkflow(ctx Context, workflowType string, args ...interface{}) ChildWorkflowFuture {
	*i.childWorkflowInterceptorInvocationCounter += 1
	return i.Next.ExecuteChildWorkflow(ctx, workflowType, args...)
}

func (f *InterceptorFactory) NewInterceptor(_ *WorkflowInfo, next WorkflowInterceptor) WorkflowInterceptor {
	return &Interceptor{
		WorkflowInterceptorBase: WorkflowInterceptorBase{
			Next: next,
		},
		workflowInterceptorInvocationCounter:      &f.workflowInterceptorInvocationCounter,
		childWorkflowInterceptorInvocationCounter: &f.childWorkflowInterceptorInvocationCounter,
	}
}

func (s *InterceptorTestSuite) SetupTest() {
	// Create a test workflow environment with the trace interceptor configured.
	s.env = s.NewTestWorkflowEnvironment()
	s.testFactory = InterceptorFactory{}
	s.env.SetWorkerOptions(WorkerOptions{
		WorkflowInterceptorChainFactories: []WorkflowInterceptorFactory{
			&s.testFactory,
		},
	})
}

func TestInterceptorTestSuite(t *testing.T) {
	suite.Run(t, new(InterceptorTestSuite))
}

func (s *InterceptorTestSuite) Test_GeneralInterceptor_IsExecutedOnChildren() {
	r := s.Require()
	childWf := func(ctx Context) error {
		return nil
	}
	s.env.RegisterWorkflowWithOptions(childWf, RegisterWorkflowOptions{Name: "child"})
	wf := func(ctx Context) error {
		return ExecuteChildWorkflow(ctx, childWf).Get(ctx, nil)
	}
	s.env.RegisterWorkflowWithOptions(wf, RegisterWorkflowOptions{Name: "parent"})
	s.env.ExecuteWorkflow(wf)
	r.True(s.env.IsWorkflowCompleted())
	r.NoError(s.env.GetWorkflowError())
	r.Equal(s.testFactory.workflowInterceptorInvocationCounter, 2)
	r.Equal(s.testFactory.childWorkflowInterceptorInvocationCounter, 1)
}
