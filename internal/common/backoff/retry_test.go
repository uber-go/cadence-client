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

package backoff

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	t.Parallel()

	succeedOnAttemptNum := 5
	tests := []struct {
		name        string
		maxAttempts int
		maxTime     time.Duration // context timeout
		isRetryable func(error) (bool, time.Duration)

		shouldError   bool
		expectedCalls int
	}{
		{"success", 2 * succeedOnAttemptNum, time.Second, nil, false, succeedOnAttemptNum},
		{"too many tries", 3, time.Second, nil, true, 4}, // max 3 retries == 4 calls.  must be < succeedOnAttemptNum to work.
		{"success with always custom retry", 2 * succeedOnAttemptNum, time.Second, func(err error) (bool, time.Duration) {
			return true, 0 // retry on all errors, same as no custom retry
		}, false, succeedOnAttemptNum},
		{"success with never custom retry", 2 * succeedOnAttemptNum, time.Second, func(err error) (bool, time.Duration) {
			return false, 0 // never retry
		}, true, 1},

		// elapsed-time-sensitive tests below.
		// consider raising time granularity if flaky, or we could set up a more complete mock
		// to resolve flakiness for real, but that's a fair bit more complex.

		// try -> sleep(10ms) -> try -> sleep(20ms) -> try -> sleep(40ms) -> timeout == 3 calls.
		{"timed out eventually", 5, 50 * time.Millisecond, func(err error) (bool, time.Duration) {
			return true, 0
		}, true, 3},
		// try -> sleep(longer than context timeout) -> timeout == 1 call.
		{"timed out due to long minimum delay", 5, 10 * time.Millisecond, func(err error) (bool, time.Duration) {
			return true, 20 * time.Millisecond
		}, true, 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			i := 0
			op := func() error {
				i++

				if i == succeedOnAttemptNum { // prevent infinite loops, and lets max-attempts > 5 eventually succeed
					return nil
				}

				return &someError{}
			}

			policy := NewExponentialRetryPolicy(10 * time.Millisecond)
			policy.SetMaximumInterval(50 * time.Millisecond)
			policy.SetMaximumAttempts(test.maxAttempts)

			ctx, cancel := context.WithTimeout(context.Background(), test.maxTime)
			defer cancel()
			err := Retry(ctx, op, policy, test.isRetryable)
			if test.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err, "Retry count: %v", i)
			}
			assert.Equal(t, test.expectedCalls, i, "wrong number of calls")
		})
	}
}

func TestConcurrentRetrier(t *testing.T) {
	t.Parallel()
	a := assert.New(t)
	policy := NewExponentialRetryPolicy(1 * time.Millisecond)
	policy.SetMaximumInterval(10 * time.Millisecond)
	policy.SetMaximumAttempts(4)

	// Basic checks
	retrier := NewConcurrentRetrier(policy)
	retrier.Failed()
	a.Equal(int64(1), retrier.failureCount)
	retrier.Succeeded()
	a.Equal(int64(0), retrier.failureCount)
	sleepDuration := retrier.throttleInternal()
	a.Equal(done, sleepDuration)

	// Multiple count check.
	retrier.Failed()
	retrier.Failed()
	a.Equal(int64(2), retrier.failureCount)
	// Verify valid sleep times.
	ch := make(chan time.Duration, 3)
	go func() {
		for i := 0; i < 3; i++ {
			ch <- retrier.throttleInternal()
		}
	}()
	for i := 0; i < 3; i++ {
		val := <-ch
		t.Logf("Duration: %d\n", val)
		a.True(val > 0)
	}
	retrier.Succeeded()
	a.Equal(int64(0), retrier.failureCount)
	// Verify we don't have any sleep times.
	go func() {
		for i := 0; i < 3; i++ {
			ch <- retrier.throttleInternal()
		}
	}()
	for i := 0; i < 3; i++ {
		val := <-ch
		t.Logf("Duration: %d\n", val)
		a.Equal(done, val)
	}
}

type someError struct{}

func (e *someError) Error() string {
	return "Some Error"
}
