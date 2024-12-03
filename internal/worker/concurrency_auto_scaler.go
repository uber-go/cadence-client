package worker

import (
	"context"
	"sync"
	"time"
)

type ConcurrencyAutoScaler struct {
	ctx context.Context
	cancel context.CancelFunc
	wg *sync.WaitGroup

	concurrency ConcurrencyLimit
	tick time.Duration

	PollerPermitLastUpdate time.Time
}

type ConcurrencyAutoScalerOptions struct {
	Concurrency ConcurrencyLimit
	Tick        time.Duration // frequency of auto tuning
	Cooldown    time.Duration    // cooldown time of update

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
	go func ()  {
		defer c.wg.Done()
		for {
			select {
			case <-c.ctx.Done():
			case <-time.Tick(c.tick):

			}
		}
	}()
}

func (c *ConcurrencyAutoScaler) updatePollerPermit() {
	c.PollerPermitLastUpdate = time.Now()
}

func (c *ConcurrencyAutoScaler) Stop() {
	c.cancel()
	c.wg.Wait()
}
