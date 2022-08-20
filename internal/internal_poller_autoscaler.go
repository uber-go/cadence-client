package internal

import (
	"context"
	"fmt"
	"github.com/marusama/semaphore/v2"
	"go.uber.org/atomic"
	"go.uber.org/cadence/internal/common/autoscaler"
	"go.uber.org/zap"
	"time"
)

// defaultPollerScalerCooldownInSeconds
const defaultPollerScalerCooldownInSeconds = 120

var (
	_ autoscaler.AutoScaler = (*pollerAutoScaler)(nil)
	_ autoscaler.Estimator  = (*pollerUsageEstimator)(nil)
)

type (
	pollerAutoScaler struct {
		pollerUsageEstimator

		isDryRun     bool
		cooldownTime time.Duration
		logger       *zap.Logger
		sem          semaphore.Semaphore // resizable semaphore to control number of concurrent pollers
		ctx          context.Context
		cancel       context.CancelFunc
		recommender  autoscaler.Recommender
	}

	pollerUsageEstimator struct {
		noTaskPolls *atomic.Uint64
		taskPolls   *atomic.Uint64
	}
)

func newPollerScaler(
	minPollerCount int,
	maxPollerCount int,
	targetUtilization float64,
	cooldownTime time.Duration,
	isDryRun bool) *pollerAutoScaler {
	ctx, cancel := context.WithCancel(context.Background())
	return &pollerAutoScaler{
		isDryRun:     isDryRun,
		cooldownTime: cooldownTime,
		sem:          semaphore.New(maxPollerCount),
		ctx:          ctx,
		cancel:       cancel,
		pollerUsageEstimator: pollerUsageEstimator{
			noTaskPolls: atomic.NewUint64(0),
			taskPolls:   atomic.NewUint64(0),
		},
		recommender: autoscaler.NewLinearRecommender(
			autoscaler.Resource(minPollerCount),
			autoscaler.Resource(maxPollerCount),
			autoscaler.Usages{
				autoscaler.PollerUtilizationRate: targetUtilization,
			},
		),
	}
}

// Acquire concurrent poll quota
func (p *pollerAutoScaler) Acquire(resource autoscaler.Resource) error {
	return p.sem.Acquire(p.ctx, int(resource))
}

// Release concurrent poll quota
func (p *pollerAutoScaler) Release(resource autoscaler.Resource) {
	p.sem.Release(int(resource))
}

// Start an auto-scaler go routine and returns a done to stop it
func (p *pollerAutoScaler) Start() autoscaler.DoneFunc {
	go func() {
		for {
			select {
			case <-p.ctx.Done():
			case <-time.After(p.cooldownTime):
				currentResource := autoscaler.Resource(p.sem.GetLimit())
				currentUsages := p.pollerUsageEstimator.Estimate()
				proposedResource := p.recommender.Recommend(currentResource, currentUsages)
				p.logger.Info(fmt.Sprintf("poller scaler recommendation: from %d to %d", int(currentResource), int(proposedResource)))
				if !p.isDryRun {
					p.sem.SetLimit(int(proposedResource))
				}
				p.pollerUsageEstimator.Reset()
			}
		}
	}()
	return func() {
		p.cancel()
	}
}

// Reset metrics from the start
func (m *pollerUsageEstimator) Reset() {
	m.noTaskPolls.Store(0)
	m.taskPolls.Store(0)
}

// CollectUsage counts past poll results to estimate autoscaler.Usages
func (m *pollerUsageEstimator) CollectUsage(data interface{}) {
	switch v := data.(type) {
	case *polledTask:
		if v == nil {
			m.noTaskPolls.Add(1)
		} else {
			m.taskPolls.Add(1)
		}
	}
}

// Estimate is based on past poll counts
func (m *pollerUsageEstimator) Estimate() autoscaler.Usages {
	noTask := m.noTaskPolls.Load()
	task := m.taskPolls.Load()
	return autoscaler.Usages{
		autoscaler.PollerUtilizationRate: float64(task) / float64(noTask+task),
	}
}
