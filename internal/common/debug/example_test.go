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

import (
	"fmt"

	"go.uber.org/atomic"
)

type (
	// pollerTrackerImpl implements the PollerTracker interface
	pollerTrackerImpl struct {
		pollerCount atomic.Int32
	}

	// stopperImpl implements the Stopper interface
	stopperImpl struct {
		pollerTracker *pollerTrackerImpl
	}
)

func (p *pollerTrackerImpl) Start() Stopper {
	p.pollerCount.Inc()
	return &stopperImpl{
		pollerTracker: p,
	}
}

func (p *pollerTrackerImpl) Stats() int32 {
	return p.pollerCount.Load()
}

func (s *stopperImpl) Stop() {
	s.pollerTracker.pollerCount.Dec()
}

func Example() {
	var pollerTracker PollerTracker
	pollerTracker = &pollerTrackerImpl{}

	// Initially, poller count should be 0
	fmt.Println(fmt.Sprintf("stats: %d", pollerTracker.Stats()))

	// Start a poller and verify that the count increments
	stopper1 := pollerTracker.Start()
	fmt.Println(fmt.Sprintf("stats: %d", pollerTracker.Stats()))

	// Start another poller and verify that the count increments again
	stopper2 := pollerTracker.Start()
	fmt.Println(fmt.Sprintf("stats: %d", pollerTracker.Stats()))

	// Stop the pollers and verify the counter
	stopper1.Stop()
	stopper2.Stop()
	fmt.Println(fmt.Sprintf("stats: %d", pollerTracker.Stats()))

	// Output:
	// stats: 0
	// stats: 1
	// stats: 2
	// stats: 0
}
