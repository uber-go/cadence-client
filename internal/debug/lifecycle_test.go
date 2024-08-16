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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPollerLifeCycle(t *testing.T) {
	lifeCycle := NewLifeCycle()

	// Initially, poller count should be 0
	require.Equal(t, int32(0), lifeCycle.ReadPollerCount(), "Expected initial poller count to be 0")

	// Start a poller and verify that the count increments
	run1 := lifeCycle.PollerStart("worker-1")
	require.Equal(t, int32(1), lifeCycle.ReadPollerCount(), "Expected poller count to be 1 after starting a poller")

	// Start another poller and verify that the count increments again
	run2 := lifeCycle.PollerStart("worker-2")
	require.Equal(t, int32(2), lifeCycle.ReadPollerCount(), "Expected poller count to be 2 after starting a second poller")

	// Stop the poller twice and verify idempotency
	run1.Stop()
	require.Equal(t, int32(1), lifeCycle.ReadPollerCount(), "Expected poller count to be 1")
	run1.Stop()
	assert.Equal(t, int32(1), lifeCycle.ReadPollerCount(), "Expected poller count to be 1")

	run2.Stop()
	assert.Equal(t, int32(0), lifeCycle.ReadPollerCount(), "Expected poller count to remain 0 after second stop")
}
