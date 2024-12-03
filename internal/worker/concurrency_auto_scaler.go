package worker

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

const (
	concurrencyAutoScalerObservabilityInterval = time.Millisecond*500
	targetPollerWaitTimeInMsLog2 = 4 // 16 ms
)

type ConcurrencyAutoScaler struct {
	ctx context.Context
	cancel context.CancelFunc
	wg *sync.WaitGroup
	log *zap.Logger
	scope tally.Scope

	concurrency ConcurrencyLimit
	tick time.Duration
	cooldown    time.Duration

	// enable auto scaler on concurrency or not
	enable      atomic.Bool

	// poller
	pollerPermitInitCount int
	pollerPermitMaxCount int
	pollerPermitMinCount int
	pollerWaitTimeInMsLog2 *rollingAverage // log2(pollerWaitTimeInMs+1) for smoothing (ideal value is 0)
	pollerPermitLastUpdate time.Time
}

type ConcurrencyAutoScalerOptions struct {
	Concurrency ConcurrencyLimit
	Tick        time.Duration // frequency of auto tuning
	Cooldown    time.Duration // cooldown time of update
}

func NewConcurrencyAutoScaler(options ConcurrencyAutoScalerOptions) *ConcurrencyAutoScaler {
	ctx, cancel := context.WithCancel(context.Background())

	return &ConcurrencyAutoScaler{
		ctx: ctx,
		cancel: cancel,
		wg: &sync.WaitGroup{},
		concurrency: options.Concurrency,
		tick: options.Tick,
	}
}

func (c *ConcurrencyAutoScaler) Start() {
	c.wg.Add(1)
	go func ()  { // scaling daemon
		defer c.wg.Done()
		ticker := time.NewTicker(c.tick)
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
	go func () { // observability daemon
		defer c.wg.Done()
		ticker := time.NewTicker(concurrencyAutoScalerObservabilityInterval)
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

// CollectPollerWaitTime collects the poller wait time for auto scaling
func (c *ConcurrencyAutoScaler) CollectPollerUsage(waitTimeInMs int64) {
	c.pollerWaitTimeInMsLog2.Add(math.Log2(float64(waitTimeInMs+1)))
}

// ResetConcurrency reset to the initial value. This will be used for gracefully switching the auto scaler off to avoid workers stuck in the wrong state
func (c *ConcurrencyAutoScaler) ResetConcurrency() {
	c.concurrency.PollerPermit.SetQuota(c.pollerPermitInitCount)
}

func (c *ConcurrencyAutoScaler) emit() {
	c.scope.Gauge("poller_in_action").Update(float64(c.concurrency.PollerPermit.Quota()-c.concurrency.PollerPermit.Count()))
	c.scope.Gauge("poller_quota").Update(float64(c.concurrency.PollerPermit.Quota()))
	c.scope.Gauge("poller_wait_time").Update(math.Exp2(c.pollerWaitTimeInMsLog2.Average()))
}

func (c *ConcurrencyAutoScaler) updatePollerPermit() {
	updateTime := time.Now()
	if updateTime.Before(c.pollerPermitLastUpdate.Add(c.cooldown)) { // before cooldown
		return
	}
	currentQuota := c.concurrency.PollerPermit.Quota()
	newQuota := int(math.Round( float64(currentQuota) * c.pollerWaitTimeInMsLog2.Average() / targetPollerWaitTimeInMsLog2))
	if newQuota < c.pollerPermitMinCount {
		newQuota = c.pollerPermitMinCount
	}
	if newQuota > c.pollerPermitMaxCount {
		newQuota = c.pollerPermitMaxCount
	}
	if newQuota == currentQuota {
		return // no change
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
	mu sync.Mutex
	window []float64
	index int
	sum float64
	count int
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
	if r.count == 0 {
		return 0
	}
	return r.sum / float64(r.count)
}
