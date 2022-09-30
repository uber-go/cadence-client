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
	"errors"
	"sync"
	"time"

	s "go.uber.org/cadence/.gen/go/shared"
)

type (
	// Operation to retry
	Operation func() error

	// IsRetryable handler can be used to exclude certain errors during retry
	IsRetryable func(error) bool

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
func Retry(ctx context.Context, operation Operation, policy RetryPolicy, isRetriable IsRetryable) error {
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

		if !isRetriable(err) {
			return err
		}

		retryAfter := ErrRetryableAfter(err)
		// update the time to wait until the next attempt.
		// as this is a *minimum*, just add it to the current delay time.
		//
		// this could be changed to clamp to retryAfter as a minimum.
		// this is intentionally *not* done here, so repeated service-busy errors are guaranteed
		// to generate *increasing* amount of time between requests, and not just send N in a row
		// with 1 second of delay.  duplicates imply "still overloaded", so this will hopefully
		// help reduce the odds of snowballing.
		// this is a pretty minor thing though, and it should not cause problems if we change it
		// to make behavior more predictable.
		next += retryAfter

		// check if ctx is done
		if ctx.Err() != nil {
			return err
		}

		// wait for the next retry period (or context timeout)
		if ctxDone := ctx.Done(); ctxDone != nil {
			// we could check if this is longer than context deadline and immediately fail...
			// ...but wasting time prevents higher-level retries from trying too early.
			// this is particularly useful for service-busy, but seems valid for essentially all retried errors.
			//
			// this could probably be changed if we get requests for it, but for now it better-protects
			// the server by preventing "external" retry storms.
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

// ErrRetryableAfter returns a minimum delay until the next attempt.
//
// for most errors this will be 0, and normal backoff logic will determine
// the full retry period, but e.g. service busy errors (or any case where the
// server knows a "time until it is not useful to retry") are safe to assume
// that a literally immediate retry is *not* going to be useful.
//
// note that this is only a minimum, however.  longer delays are assumed to
// be equally valid.
func ErrRetryableAfter(err error) (retryAfter time.Duration) {
	if target := (*s.ServiceBusyError)(nil); errors.As(err, &target) {
		// eventually: return a time-until-retry from the server.
		// for now though, just ensure at least one second before the next attempt.
		return time.Second
	}
	return 0
}
