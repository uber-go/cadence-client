// Package pahlimiter contains a PollAndHistoryLimiter, used to share resources between polls and history loading,
// to prevent flooding the server with history requests that will not complete in a reasonable time.
package pahlimiter

import (
	"context"
	"errors"
	"sync"
)

type (
	// PollAndHistoryLimiter defines an interface used to share request resources between pollers and history iterator
	// funcs, to prevent unsustainable growth of history-loading requests.
	//
	// this is intended to be used with other poller limiters and retry backoffs, not on its own.
	//
	// implementations include:
	// - NewUnlimited (historical behavior, a noop)
	// - NewHistoryLimited (limits history requests, does not limit polls)
	// - NewWeighted (history requests "consume" poll requests, and can reduce or stop polls)
	PollAndHistoryLimiter interface {
		// Poll will try to acquire a poll resource,
		// blocking until it succeeds or the context is canceled.
		//
		// The done func will release the resource - it will always be returned and can be called multiple times,
		// only the first will have an effect.
		// TODO: see if this is necessary... but it's easy and safe.
		Poll(context.Context) (ok bool, done func())
		// GetHistory will try to acquire a history-downloading resource,
		// blocking until it succeeds or the context is canceled.
		//
		// The done func will release the resource - it will always be returned and can be called multiple times,
		// only the first will have an effect.
		// TODO: see if this is necessary... but it's easy and safe.
		GetHistory(context.Context) (ok bool, done func())

		// Close will clean up any resources, call at worker shutdown.
		// This blocks until they are cleaned up.
		Close()
	}
	unlimited struct{}
	history   struct {
		tokens chan struct{} // sized at startup
	}
	weighted struct {
		stopOnce sync.Once

		// close to clean up resources
		stop chan struct{}
		// closed when cleaned up
		stopped chan struct{}

		// used to signal history requests starting and stopping
		historyStart, historyDone chan struct{}
		// used to signal poll requests starting and stopping
		pollStart, pollDone chan struct{}
	}
)

var _ PollAndHistoryLimiter = (*unlimited)(nil)
var _ PollAndHistoryLimiter = (*history)(nil)
var _ PollAndHistoryLimiter = (*weighted)(nil)

// NewUnlimited creates a new "unlimited" poll-and-history limiter, which does not constrain either operation.
// This is the default, historical behavior.
func NewUnlimited() (PollAndHistoryLimiter, error) {
	return (*unlimited)(nil), nil
}

func (*unlimited) Poll(_ context.Context) (ok bool, done func())       { return true, func() {} }
func (*unlimited) GetHistory(_ context.Context) (ok bool, done func()) { return true, func() {} }
func (*unlimited) Close()                                              {}

// NewHistoryLimited creates a simple limiter, which allows a specified number of concurrent history requests,
// and does not limit polls at all.
//
// This implementation is NOT expected to be used widely, but it exists as a trivially-safe fallback implementation
// that will still behave better than the historical default.
//
// This is very simple and should be sufficient to stop request floods during rate-limiting with many pending decision
// tasks, but seems likely to allow too many workflows to *attempt* to make progress on a host, starving progress
// when the sticky cache is higher than this size and leading to interrupted or timed out decision tasks.
func NewHistoryLimited(concurrentHistoryRequests int) (PollAndHistoryLimiter, error) {
	l := &history{
		tokens: make(chan struct{}, concurrentHistoryRequests),
	}
	// fill the token buffer
	for i := 0; i < concurrentHistoryRequests; i++ {
		l.tokens <- struct{}{}
	}
	return l, nil
}

func (p *history) Poll(_ context.Context) (ok bool, done func()) { return true, func() {} }
func (p *history) Close()                                        {}
func (p *history) GetHistory(ctx context.Context) (ok bool, done func()) {
	select {
	case <-p.tokens:
		var once sync.Once
		return true, func() {
			once.Do(func() {
				p.tokens <- struct{}{}
			})
		}
	case <-ctx.Done():
		return false, func() {} // canceled, nothing to release
	}
}

// NewWeighted creates a new "weighted" poll-and-handler limiter, which shares resources between history requests
// and polls.
//
// Each running poll or history request consumes its weight in total available (capped at max) resources, and one
// request type is allowed to reduce resources for or starve the other completely.
//
// Since this runs "inside" other poller limiting, having equal or lesser poll-resources than the poller limiter
// will allow history requests to block polls... and if history weights are lower, they can perpetually starve polls
// by not releasing enough resources.
//
// **This is intended behavior**, as it can be used to cause a heavily-history-loading worker to stop pulling more
// workflows that may also need their history loaded, until some resources free up.
//
// ---
//
// The reverse situation, where history resources cannot prevent polls, may lead to some undesirable behavior.
// Continually adding workflows while not allowing them to pull history degrades to NewHistoryLimited behavior:
// it is easily possible to have hundreds or thousands of workflows trying to load history, but few or none of them
// are allowed through this limiter to actually perform that request.
//
// In this situation it will still limit the number of actual concurrent requests to load history, but with a very
// large increase in complexity.  If you want this, strongly consider just using NewHistoryLimited.
//
// ---
//
// All that said: this is NOT built to be a reliable blocker of polls for at least two reasons:
//   - History iterators do not hold their resources between loading (and executing) pages of history, causing a gap
//     where a poller could claim resources despite the service being "too busy" loading history from a human's view.
//   - History iterators race with polls.  If enough resources are available and both possibilities can be satisfied,
//     Go chooses fairly between them.
//
// To reduce the chance of this happening, keep history weights relatively small compared to polls, so many concurrent
// workflows loading history will be unlikely to free up enough resources for a poll to occur.
func NewWeighted(pollRequestWeight, historyRequestWeight, maxResources int) (PollAndHistoryLimiter, error) {
	if historyRequestWeight > maxResources || pollRequestWeight > maxResources {
		return nil, errors.New("weights must be less than max resources, or no requests can be sent")
	}

	l := &weighted{
		stopOnce:     sync.Once{},
		stop:         make(chan struct{}),
		stopped:      make(chan struct{}),
		historyStart: make(chan struct{}),
		historyDone:  make(chan struct{}),
		pollStart:    make(chan struct{}),
		pollDone:     make(chan struct{}),
	}
	l.init(pollRequestWeight, historyRequestWeight, maxResources)
	return l, nil
}

func (p *weighted) init(pollRequestWeight, historyRequestWeight, maxResources int) {
	// mutated only by the actor goroutine
	available := maxResources

	// start an actor-goroutine to simplify concurrency logic with many possibilities at any time.
	// all logic is decided single-threaded, run by this goroutine, and every operation (except stop) is blocking.
	//
	// this actor only sends to history/poll channels.
	// modifying functions only read from them.
	// both read from "stop" and "stopped".
	//
	// - by reading from a channel, the caller has successfully acquired or released resources, and it can immediately proceed.
	// - by sending on a channel, this actor has observed that resources are changed, and it must update its state.
	// - by closing `p.stop`, this limiter will stop reading from channels.
	//   - ALL channel operations (except stop) will block forever.
	//   - this means "xDone" resource-releasing must also read from `p.stop`.
	// - because `p.stop` races with other channel operations, stop DOES NOT guarantee no further polls will start,
	//   even on the same goroutine, until `Close()` returns.
	//   - this is one reason why `Close()` waits for the actor to exit.  without it, you do not have sequential
	//     logic guarantees.
	//   - you can `Close()` any number of times from any goroutines, all calls will wait for the actor to stop.
	//
	// all operations are "fast", and it must remain this way.
	// callers block while waiting on this actor, including when releasing resources.
	go func() {
		defer func() { close(p.stopped) }()
		for {
			// every branch must:
			// 1. read from `p.stop`, so this limiter can be stopped.
			// 2. write to "done" chans, so resources can be freed.
			// 3. optionally write to "start" chans, so resources can be acquired
			//
			// doing otherwise for any reason risks deadlocks or invalid resource values.

			if available >= pollRequestWeight && available >= historyRequestWeight {
				// resources available for either == wait for either
				select {
				case <-p.stop:
					return

				case p.historyStart <- struct{}{}:
					available -= historyRequestWeight
				case p.pollStart <- struct{}{}:
					available -= pollRequestWeight

				case p.historyDone <- struct{}{}:
					available += historyRequestWeight
				case p.pollDone <- struct{}{}:
					available += pollRequestWeight
				}
			} else if available >= pollRequestWeight && available < historyRequestWeight {
				// only poll resources available
				select {
				case <-p.stop:
					return

				// case p.historyStart <- struct{}{}: // insufficient resources
				case p.pollStart <- struct{}{}:
					available -= pollRequestWeight

				case p.historyDone <- struct{}{}:
					available += historyRequestWeight
				case p.pollDone <- struct{}{}:
					available += pollRequestWeight
				}
			} else if available < pollRequestWeight && available >= historyRequestWeight {
				// only history resources available
				select {
				case <-p.stop:
					return

				case p.historyStart <- struct{}{}:
					available -= historyRequestWeight
				// case p.pollStart <- struct{}{}: // insufficient resources

				case p.historyDone <- struct{}{}:
					available += historyRequestWeight
				case p.pollDone <- struct{}{}:
					available += pollRequestWeight
				}
			} else {
				// no resources for either, wait for something to finish
				select {
				case <-p.stop:
					return

				// case p.historyStart <- struct{}{}: // insufficient resources
				// case p.pollStart <- struct{}{}:    // insufficient resources

				case p.historyDone <- struct{}{}:
					available += historyRequestWeight
				case p.pollDone <- struct{}{}:
					available += pollRequestWeight
				}
			}
		}
	}()
}

func (p *weighted) Close() {
	p.stopOnce.Do(func() {
		close(p.stop)
	})
	<-p.stopped
}

func (p *weighted) Poll(ctx context.Context) (ok bool, done func()) {
	select {
	case <-ctx.Done():
		return false, func() {} // canceled
	case <-p.stop:
		return false, func() {} // shutting down
	case <-p.pollStart:
		// resource acquired
		var once sync.Once
		return true, func() {
			once.Do(func() {
				select {
				case <-p.pollDone: // released
				case <-p.stop: // shutting down
				}
			})
		}
	}
}

func (p *weighted) GetHistory(ctx context.Context) (ok bool, done func()) {
	select {
	case <-ctx.Done():
		return false, func() {} // canceled
	case <-p.stop:
		return false, func() {} // shutting down
	case <-p.historyStart:
		// resource acquired
		var once sync.Once
		return true, func() {
			once.Do(func() {
				select {
				case <-p.historyDone: // released
				case <-p.stop: // shutting down
				}
			})
		}
	}
}
