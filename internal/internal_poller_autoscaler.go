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

package internal

import (
	"context"
	"errors"
	"github.com/marusama/semaphore/v2"
	"go.uber.org/atomic"
	"go.uber.org/cadence/internal/common/autoscaler"
	"go.uber.org/zap"
	"time"
)

// defaultPollerScalerCooldownInSeconds
const (
	defaultPollerScalerCooldownInSeconds = 120
)

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
		// This single atomic variable stores two variables:
		// left 32 bits is noTaskCounts, right 32 bits is taskCounts.
		// This avoids unnecessary usage of CompareAndSwap
		atomicBits *atomic.Uint64
	}
)

func newPollerScaler(
	initialPollerCount int,
	minPollerCount int,
	maxPollerCount int,
	targetUtilizationInMilli uint64,
	isDryRun bool,
	cooldownTime time.Duration,
	logger *zap.Logger) *pollerAutoScaler {
	ctx, cancel := context.WithCancel(context.Background())
	return &pollerAutoScaler{
		isDryRun:             isDryRun,
		cooldownTime:         cooldownTime,
		logger:               logger,
		sem:                  semaphore.New(initialPollerCount),
		ctx:                  ctx,
		cancel:               cancel,
		pollerUsageEstimator: pollerUsageEstimator{atomicBits: atomic.NewUint64(0)},
		recommender: autoscaler.NewLinearRecommender(
			autoscaler.Resource(minPollerCount),
			autoscaler.Resource(maxPollerCount),
			autoscaler.Usages{
				autoscaler.PollerUtilizationRate: autoscaler.UsageInMilli(targetUtilizationInMilli),
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

// GetCurrent poll quota
func (p *pollerAutoScaler) GetCurrent() autoscaler.Resource {
	return autoscaler.Resource(p.sem.GetLimit())
}

// Start an auto-scaler go routine and returns a done to stop it
func (p *pollerAutoScaler) Start() autoscaler.DoneFunc {
	logger := p.logger.Sugar()
	go func() {
		for {
			select {
			case <-p.ctx.Done():
			case <-time.After(p.cooldownTime):
				currentResource := autoscaler.Resource(p.sem.GetLimit())
				currentUsages, err := p.pollerUsageEstimator.Estimate()
				if err != nil {
					logger.Warnw("poller autoscaler skip due to estimator error", "error", err)
					continue
				}
				proposedResource := p.recommender.Recommend(currentResource, currentUsages)
				logger.Debugw("poller autoscaler recommendation",
					"currentUsage", currentUsages,
					"current", uint64(currentResource),
					"recommend", uint64(proposedResource),
					"isDryRun", p.isDryRun)
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
	m.atomicBits.Store(0)
}

// CollectUsage counts past poll results to estimate autoscaler.Usages
func (m *pollerUsageEstimator) CollectUsage(data interface{}) error {
	switch v := data.(type) {
	case *polledTask:
		if v == nil { // no-task poll
			m.atomicBits.Add(1 << 32)
		} else {
			m.atomicBits.Add(1)
		}
	}
	return nil
}

// Estimate is based on past poll counts
func (m *pollerUsageEstimator) Estimate() (autoscaler.Usages, error) {
	bits := m.atomicBits.Load()
	noTaskCounts := bits >> 32           // left 32 bits
	taskCounts := bits & ((1 << 32) - 1) // right 32 bits
	if noTaskCounts+taskCounts == 0 {
		return nil, errors.New("autoscaler.Estimator::Estimate error: not enough data")
	}

	return autoscaler.Usages{
		autoscaler.PollerUtilizationRate: autoscaler.UsageInMilli(taskCounts * 1000 / (noTaskCounts + taskCounts)),
	}, nil
}