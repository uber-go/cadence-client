// Copyright (c) 2017 Uber Technologies, Inc.
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
	"sync"
	"time"
)

type (
	// Operation to retry
	Operation func() error

	// RetryAfter handler can be used to exclude certain errors during retry,
	// and define how long to wait at a minimum before trying again.
	// delays must be 0 or larger.
	RetryAfter func(error) (isRetryable bool, retryAfter time.Duration)

	// ConcurrentRetrier is used for client-side throttling. It determines whether to
	// throttle outgoing traffic in case downstream backend server rejects
	// requests due to out-of-quota or server busy errors.
	ConcurrentRetrier struct {
		sync.Mutex
		retrier      Retrier // Backoff retrier
		failureCount int64   // Number of consecutive failures seen
	}
)

// Throttle Sleep if there were failures since the last success call.
func (c *ConcurrentRetrier) Throttle() {
	c.throttleInternal()
}

func (c *ConcurrentRetrier) throttleInternal() time.Duration {
	next := done

	// Check if we have failure count.
	c.Lock()
	if c.failureCount > 0 {
		next = c.retrier.NextBackOff()
	}
	c.Unlock()

	if next != done {
		time.Sleep(next)
	}

	return next
}

// Succeeded marks client request succeeded.
func (c *ConcurrentRetrier) Succeeded() {
	defer c.Unlock()
	c.Lock()
	c.failureCount = 0
	c.retrier.Reset()
}

// Failed marks client request failed because backend is busy.
func (c *ConcurrentRetrier) Failed() {
	defer c.Unlock()
	c.Lock()
	c.failureCount++
}

// NewConcurrentRetrier returns an instance of concurrent backoff retrier.
func NewConcurrentRetrier(retryPolicy RetryPolicy) *ConcurrentRetrier {
	retrier := NewRetrier(retryPolicy, SystemClock)
	return &ConcurrentRetrier{retrier: retrier}
}

// Retry function can be used to wrap any call with retry logic using the passed in policy
func Retry(ctx context.Context, operation Operation, policy RetryPolicy, retryAfter RetryAfter) error {
	var err error
	var next time.Duration

	r := NewRetrier(policy, SystemClock)
Retry_Loop:
	for {
		// operation completed successfully.  No need to retry.
		if err = operation(); err == nil {
			return nil
		}

		if next = r.NextBackOff(); next == done {
			return err
		}

		// Check if the error is retryable
		if retryAfter != nil {
			retryable, minNext := retryAfter(err)
			// fail on non-retryable errors
			if !retryable {
				return err
			}
			// update the time to wait until the next attempt.
			// as this is a *minimum*, just add it to the current pending time.
			// this way repeated service busy errors will take increasing amounts of time before retrying,
			// and when the request is not limited it does not further increase the time until trying again.
			next += minNext
		}

		// check if ctx is done
		if ctx.Err() != nil {
			return err
		}

		// wait for the next retry period (or context timeout)
		if ctxDone := ctx.Done(); ctxDone != nil {
			// we could check if this is longer than context deadline and immediately fail...
			// ...but wasting time prevents higher-level retries from trying too early.
			// this is particularly useful for service-busy, but seems valid for essentially all retried errors.
			timer := time.NewTimer(next)
			select {
			case <-ctxDone:
				timer.Stop()
				return err
			case <-timer.C:
				continue Retry_Loop
			}
		}

		// ctx is not cancellable
		time.Sleep(next)
	}
}
