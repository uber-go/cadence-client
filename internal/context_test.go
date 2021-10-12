// Copyright (c) 2017-2021 Uber Technologies Inc.
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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestContextChildParentCancelRace(t *testing.T) {
	/*
		Testing previous race happened while child and parent cancelling at the same time
		While child is trying to remove itself from the parent, parent tries to iterate
		its children and cancel them at the same time.
	*/
	env := newTestWorkflowEnv(t)

	wf := func(ctx Context) error {
		parentCtx, parentCancel := WithCancel(ctx)
		defer parentCancel()

		type cancelerContext struct {
			ctx      Context
			canceler func()
		}

		children := []cancelerContext{}
		numChildren := 100

		for i := 0; i < numChildren; i++ {
			c, canceler := WithCancel(parentCtx)
			children = append(children, cancelerContext{
				ctx:      c,
				canceler: canceler,
			})
		}

		for i := 0; i < numChildren; i++ {
			go children[i].canceler()
			if i == numChildren/2 {
				go parentCancel()
			}
		}

		return nil
	}
	env.RegisterWorkflow(wf)
	env.ExecuteWorkflow(wf)
	assert.NoError(t, env.GetWorkflowError())
}

func TestContextConcurrentCancelRace(t *testing.T) {
	/*
		A race condition existed due to concurrently ending goroutines on shutdown (i.e. closing their chan without waiting
		on them to finish shutdown), which executed... quite a lot of non-concurrency-safe code in a concurrent way.  All
		decision-sensitive code is assumed to be run strictly sequentially.

		Context cancellation was one identified by a customer, and it's fairly easy to test.
		In principle this must be safe to do - contexts are supposed to be concurrency-safe.  Even if ours are not actually
		safe (for valid reasons), our execution model needs to ensure they *act* like it's safe.
	*/
	env := newTestWorkflowEnv(t)
	wf := func(ctx Context) error {
		ctx, cancel := WithCancel(ctx)
		racyCancel := func(ctx Context) {
			defer cancel() // defer is necessary as Sleep will never return due to Goexit
			_ = Sleep(ctx, time.Hour)
		}
		// start a handful to increase odds of a race being detected
		for i := 0; i < 10; i++ {
			Go(ctx, racyCancel)
		}

		_ = Sleep(ctx, time.Minute) // die early
		return nil
	}
	env.RegisterWorkflow(wf)
	env.ExecuteWorkflow(wf)
	assert.NoError(t, env.GetWorkflowError())
}

func TestContextAddChildCancelParentRace(t *testing.T) {
	/*
		It's apparently also possible to race on adding children while propagating the cancel to children.
	*/
	env := newTestWorkflowEnv(t)
	wf := func(ctx Context) error {
		ctx, cancel := WithCancel(ctx)
		racyCancel := func(ctx Context) {
			defer cancel() // defer is necessary as Sleep will never return due to Goexit
			defer func() {
				_, ccancel := WithCancel(ctx)
				cancel()
				ccancel()
			}()
			_ = Sleep(ctx, time.Hour)
		}
		// start a handful to increase odds of a race being detected
		for i := 0; i < 10; i++ {
			Go(ctx, racyCancel)
		}

		_ = Sleep(ctx, time.Minute) // die early
		return nil
	}
	env.RegisterWorkflow(wf)
	env.ExecuteWorkflow(wf)
	assert.NoError(t, env.GetWorkflowError())
}
