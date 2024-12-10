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

package worker

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uber-go/tally"
	"go.uber.org/zap"

	"go.uber.org/cadence/.gen/go/shared"
)

const (
	concurrencyAutoScalerUpdateTick        = time.Second
	concurrencyAutoScalerObservabilityTick = time.Millisecond * 500
	targetPollerWaitTimeInMsLog2           = 4 // 16 ms
	numberOfPollsInRollingAverage          = 20
)

type ConcurrencyAutoScaler struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
	log    *zap.Logger
	scope  tally.Scope

	concurrency *ConcurrencyLimit
	cooldown    time.Duration

	// enable auto scaler on concurrency or not
	enable atomic.Bool

	// poller
	pollerMaxCount         int
	pollerMinCount         int
	pollerWaitTimeInMsLog2 *rollingAverage // log2(pollerWaitTimeInMs+1) for smoothing (ideal value is 0)
	pollerPermitLastUpdate time.Time
}

type ConcurrencyAutoScalerInput struct {
	Concurrency    *ConcurrencyLimit
	Cooldown       time.Duration // cooldown time of update
	PollerMaxCount int
	PollerMinCount int
	Logger         *zap.Logger
	Scope          tally.Scope
}

func NewPollerAutoScaler(input ConcurrencyAutoScalerInput) *ConcurrencyAutoScaler {
	ctx, cancel := context.WithCancel(context.Background())

	return &ConcurrencyAutoScaler{
		ctx:                    ctx,
		cancel:                 cancel,
		wg:                     &sync.WaitGroup{},
		concurrency:            input.Concurrency,
		cooldown:               input.Cooldown,
		log:                    input.Logger,
		scope:                  input.Scope,
		enable:                 atomic.Bool{}, // initial value should be false and is only turned on from auto config hint
		pollerMaxCount:         input.PollerMaxCount,
		pollerMinCount:         input.PollerMinCount,
		pollerWaitTimeInMsLog2: newRollingAverage(numberOfPollsInRollingAverage),
	}
}

func (c *ConcurrencyAutoScaler) Start() {
	c.wg.Add(1)
	go func() { // scaling daemon
		defer c.wg.Done()
		ticker := time.NewTicker(concurrencyAutoScalerUpdateTick)
		for {
			select {
			case <-c.ctx.Done():
				ticker.Stop()
			case <-ticker.C:
				c.updatePollerPermit()
			}
		}
	}()
	c.wg.Add(1)
	go func() { // observability daemon
		defer c.wg.Done()
		ticker := time.NewTicker(concurrencyAutoScalerUpdateTick)
		for {
			select {
			case <-c.ctx.Done():
				ticker.Stop()
			case <-ticker.C:
				c.emit()
			}
		}
	}()
}

func (c *ConcurrencyAutoScaler) Stop() {
	c.cancel()
	c.wg.Wait()
}

// ProcessPollerHint reads the poller response hint and take actions
// 1. update poller wait time
// 2. enable/disable auto scaler
func (c *ConcurrencyAutoScaler) ProcessPollerHint(hint *shared.AutoConfigHint) {
	if hint == nil {
		return
	}
	if hint.PollerWaitTimeInMs != nil {
		waitTimeInMs := *hint.PollerWaitTimeInMs
		c.pollerWaitTimeInMsLog2.Add(math.Log2(float64(waitTimeInMs + 1)))
	}

	/*
		Atomically compare and switch the auto scaler enable flag. If auto scaler is turned off, reset the concurrency limits.
	*/
	var shouldEnable bool
	if hint.EnableAutoConfig != nil && *hint.EnableAutoConfig {
		shouldEnable = true
	}
	if switched := c.enable.CompareAndSwap(!shouldEnable, shouldEnable); switched {
		if shouldEnable {
			c.log.Sugar().Infof("auto scaler enabled")
		} else {
			c.log.Sugar().Infof("auto scaler disabled")
			c.ResetConcurrency()
		}
	}
}

// ResetConcurrency reset poller quota to the max value. This will be used for gracefully switching the auto scaler off to avoid workers stuck in the wrong state
func (c *ConcurrencyAutoScaler) ResetConcurrency() {
	c.concurrency.PollerPermit.SetQuota(c.pollerMaxCount)
}

func (c *ConcurrencyAutoScaler) emit() {
	if c.enable.Load() {
		c.scope.Counter("concurrency_auto_scaler.enabled").Inc(1)
	} else {
		c.scope.Counter("concurrency_auto_scaler.disabled").Inc(1)
	}
	c.scope.Gauge("poller_in_action").Update(float64(c.concurrency.PollerPermit.Quota() - c.concurrency.PollerPermit.Count()))
	c.scope.Gauge("poller_quota").Update(float64(c.concurrency.PollerPermit.Quota()))
	c.scope.Gauge("poller_wait_time").Update(math.Exp2(c.pollerWaitTimeInMsLog2.Average()))
}

func (c *ConcurrencyAutoScaler) updatePollerPermit() {
	updateTime := time.Now()
	if updateTime.Before(c.pollerPermitLastUpdate.Add(c.cooldown)) { // before cooldown
		return
	}
	currentQuota := c.concurrency.PollerPermit.Quota()
	newQuota := int(math.Round(float64(currentQuota) * c.pollerWaitTimeInMsLog2.Average() / targetPollerWaitTimeInMsLog2))
	if newQuota < c.pollerMinCount {
		newQuota = c.pollerMinCount
	}
	if newQuota > c.pollerMaxCount {
		newQuota = c.pollerMaxCount
	}
	if newQuota == currentQuota {
		return
	}
	enabled := c.enable.Load()
	c.log.Sugar().With("applied", enabled).Infof("update poller permit: %v -> %v", currentQuota, newQuota)
	if !c.enable.Load() {
		return
	}
	c.concurrency.PollerPermit.SetQuota(newQuota)
	c.pollerPermitLastUpdate = updateTime
}

type rollingAverage struct {
	mu     sync.Mutex
	window []float64
	index  int
	sum    float64
	count  int
}

func newRollingAverage(capacity int) *rollingAverage {
	return &rollingAverage{
		window: make([]float64, capacity),
	}
}

// Add always add positive numbers
func (r *rollingAverage) Add(value float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// replace the old value with the new value
	r.index %= len(r.window)
	r.sum += value - r.window[r.index]
	r.window[r.index] = value
	r.index++

	if r.count < len(r.window) {
		r.count++
	}
}

func (r *rollingAverage) Average() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.count == 0 {
		return 0
	}
	return r.sum / float64(r.count)
}
