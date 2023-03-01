package pahlimiter

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// something "quick" but long enough to not succumb to test noise.
// non-limited tests should be many, many times faster than this.
// all tests must be parallel so the total time for the suite is kept low
var fastTest = 100 * time.Millisecond

func countWhileLeaking(method func(context.Context) (bool, func()), timeout time.Duration) (started int) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for i := 0; i < 100; i++ {
		ok, _ := method(ctx) // intentionally leaking close-resources
		if ok {
			started += 1
		}
	}
	return started
}

func TestUnlimited(t *testing.T) {
	t.Parallel()

	// sharing one across all tests to exercise it further
	l, err := NewUnlimited()
	require.NoError(t, err, "should be impossible")
	defer l.Close()

	t.Run("should allow many history requests", func(t *testing.T) {
		t.Parallel()
		started := countWhileLeaking(l.GetHistory, fastTest)
		assert.Equal(t, 100, started, "all requests should have been allowed")
	})
	t.Run("should allow many poll requests", func(t *testing.T) {
		t.Parallel()
		started := countWhileLeaking(l.Poll, fastTest)
		assert.Equal(t, 100, started, "all requests should have been allowed")
	})
}

func TestHistoryLimited(t *testing.T) {
	t.Parallel()

	t.Run("should not limit poll requests", func(t *testing.T) {
		t.Parallel()
		l, err := NewHistoryLimited(5)
		require.NoError(t, err, "should be impossible")
		defer l.Close()

		started := countWhileLeaking(l.Poll, fastTest)
		assert.Equal(t, 100, started, "all requests should have been allowed")
	})
	t.Run("should limit history requests", func(t *testing.T) {
		t.Parallel()
		l, err := NewHistoryLimited(5)
		require.NoError(t, err, "should be impossible")
		defer l.Close()

		started := countWhileLeaking(l.GetHistory, fastTest)
		assert.Equal(t, 5, started, "history should have been limited at 5 concurrent requests")
	})
	t.Run("limited history should not limit polls", func(t *testing.T) {
		t.Parallel()

		l, err := NewHistoryLimited(5)
		require.NoError(t, err, "should be impossible")
		defer l.Close()

		assert.Equal(t, 100, countWhileLeaking(l.Poll, fastTest), "pre-polls should have all succeeded")
		assert.Equal(t, 5, countWhileLeaking(l.GetHistory, fastTest), "history should have been limited at 5 concurrent requests")
		assert.Equal(t, 100, countWhileLeaking(l.Poll, fastTest), "post-polls should have all succeeded")
	})
	t.Run("concurrent thrashing", func(t *testing.T) {
		t.Parallel()
		const maxConcurrentHistory = 5
		l, err := NewHistoryLimited(maxConcurrentHistory)
		require.NoError(t, err, "should be impossible")
		defer l.Close()

		// do a bunch of requests in parallel, assert that successes are within bounds

		var wg sync.WaitGroup
		wg.Add(100) // polls
		wg.Add(20)  // histories
		var polls int64
		var histories int64

		// fire up a lot of polls, randomly spread within < fasttest
		pollCtx, cancel := context.WithTimeout(context.Background(), fastTest)
		defer cancel()
		for i := 0; i < 100; i++ {
			go func() {
				// start at random times
				time.Sleep(time.Duration(rand.Int63n(int64(fastTest))))
				ok, _ := l.Poll(pollCtx) // intentionally leaking resources
				if ok {
					atomic.AddInt64(&polls, 1)
				}
				wg.Done()
			}()
		}

		// fire up a handful of gethistories, hold resources for a bit, make sure some but not all succeed
		historyCtx, cancel := context.WithTimeout(context.Background(), fastTest)
		defer cancel()
		for i := 0; i < 20; i++ {
			go func() {
				// everything competes at the start.
				// 5 will succeed, the rest will be random
				ok, release := l.GetHistory(historyCtx)
				if ok {
					atomic.AddInt64(&histories, 1)
					// each one holds the resource for 50% to ~100% of the time limit
					time.Sleep(fastTest / 2)                                    // 1/2
					time.Sleep(time.Duration(rand.Int63n(int64(fastTest / 2)))) // plus <1/2
					// then release
					release()
				}
				wg.Done()
			}()
		}

		wg.Wait()
		finalPolls := atomic.LoadInt64(&polls)
		finalHistories := atomic.LoadInt64(&histories)
		// all polls should have succeeded
		assert.Equal(t, int64(100), finalPolls)
		// should have had more than max-concurrent successes == used some freed resources
		assert.Greater(t, finalHistories, int64(maxConcurrentHistory))
		// should have had less than N total successes == had contention for resources.
		// in practice this is heavily biased towards 10 (never 11, no resource is held 3 times as that would have to start after the timeout),
		// as resources are acquired at least once and *usually* twice.
		assert.Less(t, finalHistories, int64((maxConcurrentHistory*2)+1))
		t.Log("histories:", finalHistories) // to also show successful values with -v
	})
}

func TestWeighted(t *testing.T) {
	t.Parallel()

	t.Run("sequential works", func(t *testing.T) {
		t.Parallel()

		l, err := NewWeighted(1, 1, 10)
		require.NoError(t, err, "should be impossible")
		defer l.Close()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		ok, release := l.Poll(ctx)
		assert.True(t, ok)
		release()

		ok, release = l.GetHistory(ctx)
		assert.True(t, ok)
		release()
	})
	t.Run("fair thrashing should be roughly fair", func(t *testing.T) {
		t.Parallel()
		// this is the actual test being done: thrash on a balanced-weight limiter,
		// and make sure the behavior is balanced.
		//
		// since "balanced" is hard to determine with random channel behavior,
		// this is just run multiple times, and checked for excessive skew.
		test := func() (int64, int64) {
			l, err := NewWeighted(1, 1, 10)
			require.NoError(t, err, "should be impossible")
			defer l.Close()

			return thrashForFairness(l)
		}

		// run N times in parallel to reduce noise.
		// at least one of them must be fair, and none may have zeros.
		var wg sync.WaitGroup
		type result struct {
			polls, histories int64
		}
		results := make([]result, 5)
		wg.Add(len(results))
		for i := 0; i < len(results); i++ {
			i := i
			go func() {
				polls, histories := test()

				// this should never happen
				assert.NotZero(t, polls)
				assert.NotZero(t, histories)

				results[i] = result{
					polls:     polls,
					histories: histories,
				}
				wg.Done()
			}()
		}
		wg.Wait()

		t.Log("results:", results)
		for _, r := range results {
			// must be within 1/3 as much, found experimentally.
			// out of 1000+ tries locally, only a couple have checked the third value.
			if r.polls < (r.histories/3)*2 {
				t.Log("low polls, checking next")
				continue
			}
			if r.histories < (r.polls/3)*2 {
				t.Log("low histories, checking next")
				continue
			}

			t.Log("fair enough!")
			break // no need to check the rest
		}
	})
	t.Run("imbalanced thrashing should starve the higher weight", func(t *testing.T) {
		t.Parallel()
		// similar to fair-thrashing, but in this case we expect the lower-weight history
		// requests to starve polls most of the time (as two need to release for a single poll to run).
		test := func() (int64, int64) {
			l, err := NewWeighted(2, 1, 10)
			require.NoError(t, err, "should be impossible")
			defer l.Close()

			return thrashForFairness(l)
		}

		// run N times in parallel to reduce noise.
		// none may have zeros, and it should be heavily skewed in favor of history requests
		var wg sync.WaitGroup
		type result struct {
			polls, histories int64
		}
		results := make([]result, 5)
		wg.Add(len(results))
		for i := 0; i < len(results); i++ {
			i := i
			go func() {
				polls, histories := test()

				// this should never happen.
				// polls, however, do sometimes go to zero.
				assert.NotZero(t, histories)

				results[i] = result{
					polls:     polls,
					histories: histories,
				}
				wg.Done()
			}()
		}
		wg.Wait()

		t.Log("results:", results)
		for _, r := range results {
			// histories must be favored 5x or better, and polls must sometimes be used, found experimentally.
			//
			// this combination ensures we do not succeed when completely starving polls (== 0),
			// as that would be unexpected, and likely a sign of a bad implementation.
			//
			// out of 1000+ tries locally, around 1% have checked the third value.
			if r.polls > 0 && r.polls > (r.histories/5) {
				t.Log("high polls, checking next")
				continue
			}

			t.Log("found expected imbalance")
			break // no need to check the rest
		}
	})
}

func thrashForFairness(l PollAndHistoryLimiter) (polls, histories int64) {
	var wg sync.WaitGroup
	wg.Add(100) // polls
	wg.Add(100) // histories

	ctx, cancel := context.WithTimeout(context.Background(), fastTest)
	defer cancel()

	useResource := func() {
		// use the resource for 10% to 20% of the test time.
		// actual amount doesn't matter, just target middling-ish consumption to avoid noise.
		// this value was found experimentally, just run with -v and check the result logs.
		time.Sleep(time.Duration(int64(fastTest/10) + rand.Int63n(int64(fastTest/10))))
	}

	var p, h int64
	for i := 0; i < 100; i++ {
		go func() {
			time.Sleep(time.Duration(rand.Int63n(int64(fastTest))))
			ok, release := l.Poll(ctx)
			if ok {
				atomic.AddInt64(&p, 1)
				useResource()
				release()
			}
			wg.Done()
		}()
		go func() {
			// randomly start through the test
			time.Sleep(time.Duration(rand.Int63n(int64(fastTest))))
			ok, release := l.GetHistory(ctx)
			if ok {
				atomic.AddInt64(&h, 1)
				useResource()
				release()
			}
			wg.Done()
		}()
	}

	wg.Wait()

	return atomic.LoadInt64(&p), atomic.LoadInt64(&h)
}
