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

package debug

import "go.uber.org/atomic"

type (
	// lifeCycleImpl implements the LifeCycle interface
	lifeCycleImpl struct {
		pollerCount atomic.Int32
	}

	// runImpl implements the Run interface
	runImpl struct {
		lifeCycle *lifeCycleImpl
		stopped   atomic.Bool
	}

	// Run is a helper for simpler tracking of the go routine run
	// Deprecated: in development and very likely to change
	Run interface {
		// Stop is the method to report stats once a poller thread is done
		Stop()
	}

	// LifeCycle contains a set of methods to collect information on a running worker
	// Deprecated: in development and very likely to change
	LifeCycle interface {
		// PollerStart collects information on poller start up.
		// consumers should provide a concurrency-safe implementation.
		PollerStart(workerID string) Run
		// ReadPollerCount return the number or running pollers
		ReadPollerCount() int32
	}

	// EventMonitoring provides a set of methods that can be used to collect
	// stats on the Worker for debugging purposes.
	// Deprecated: in development and very likely to change
	EventMonitoring struct {
		LifeCycle LifeCycle
	}
)

func (lc *lifeCycleImpl) PollerStart(workerID string) Run {
	lc.pollerCount.Inc()
	return &runImpl{
		lifeCycle: lc,
		stopped:   atomic.Bool{},
	}
}

func (lc *lifeCycleImpl) ReadPollerCount() int32 {
	return lc.pollerCount.Load()
}

// NewLifeCycle creates a new LifeCycle instance
func NewLifeCycle() LifeCycle { return &lifeCycleImpl{} }

var _ LifeCycle = &lifeCycleImpl{}

func (r *runImpl) Stop() {
	// Check if Stop() has already been called
	if r.stopped.CompareAndSwap(false, true) {
		r.lifeCycle.pollerCount.Dec()
	} else {
		// Stop has already been called, do nothing
		return
	}
}

var _ Run = &runImpl{}
