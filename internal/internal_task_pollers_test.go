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

package internal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestLocalActivityPanic(t *testing.T) {
	// regression: panics in local activities should not terminate the process
	s := WorkflowTestSuite{logger: zaptest.NewLogger(t)}
	env := s.NewTestWorkflowEnvironment()

	wf := "panicky_local_activity"
	env.RegisterWorkflowWithOptions(func(ctx Context) error {
		ctx = WithLocalActivityOptions(ctx, LocalActivityOptions{
			ScheduleToCloseTimeout: time.Second,
		})
		return ExecuteLocalActivity(ctx, func(ctx context.Context) error {
			panic("should not kill process")
		}).Get(ctx, nil)
	}, RegisterWorkflowOptions{Name: wf})

	env.ExecuteWorkflow(wf)
	err := env.GetWorkflowError()
	require.Error(t, err)
	var perr *PanicError
	require.True(t, errors.As(err, &perr), "error should be a panic error")
	assert.Contains(t, perr.StackTrace(), "panic")
	assert.Contains(t, perr.StackTrace(), t.Name(), "should mention the source location of the local activity that panicked")
}
