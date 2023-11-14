package internal

import (
	"context"
	"github.com/shirou/gopsutil/cpu"
	"github.com/uber-go/tally"
	"go.uber.org/cadence/internal/common/metrics"
	"go.uber.org/zap"
	"runtime"
	"sync"
	"time"
)

type (
	workerUsageCollector struct {
		cooldownTime time.Duration
		logger       *zap.Logger
		ctx          context.Context
		wg           *sync.WaitGroup // graceful stop
		cancel       context.CancelFunc
		metricsScope tally.Scope
		host         string
	}

	workerUsageCollectorOptions struct {
		Enabled      bool
		Cooldown     time.Duration
		Host         string
		MetricsScope tally.Scope
	}

	hardwareUsage struct {
		NumCPUCores     int
		CPUPercent      float64
		NumGoRoutines   int
		TotalMemory     float64
		MemoryUsedHeap  float64
		MemoryUsedStack float64
	}
)

func newWorkerUsageCollector(
	options workerUsageCollectorOptions,
	logger *zap.Logger,
) *workerUsageCollector {
	if !options.Enabled {
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &workerUsageCollector{
		cooldownTime: options.Cooldown,
		host:         options.Host,
		metricsScope: options.MetricsScope,
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
		wg:           &sync.WaitGroup{},
	}
}

func (w *workerUsageCollector) Start() {
	w.wg.Add(1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				w.logger.Error("Unhandled panic in workerUsageCollector.")
			}
		}()
		defer w.wg.Done()
		collectHardwareUsageOnce.Do(
			func() {
				ticker := time.NewTicker(w.cooldownTime)
				for {
					select {
					case <-w.ctx.Done():
						return
					case <-ticker.C:
						hardwareUsageData := w.collectHardwareUsage()
						w.emitHardwareUsage(hardwareUsageData)

					}
				}
			})
	}()
	return
}

func (w *workerUsageCollector) Stop() {
	w.cancel()
	w.wg.Wait()
}

func (w *workerUsageCollector) collectHardwareUsage() hardwareUsage {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		w.logger.Warn("Failed to get cpu percent", zap.Error(err))
	}
	cpuCores, err := cpu.Counts(false)
	if err != nil {
		w.logger.Warn("Failed to get number of cpu cores", zap.Error(err))
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return hardwareUsage{
		NumCPUCores:     cpuCores,
		CPUPercent:      cpuPercent[0],
		NumGoRoutines:   runtime.NumGoroutine(),
		TotalMemory:     float64(memStats.Sys),
		MemoryUsedHeap:  float64(memStats.HeapAlloc),
		MemoryUsedStack: float64(memStats.StackInuse),
	}
}

// emitHardwareUsage emits collected hardware usage metrics to metrics scope
func (w *workerUsageCollector) emitHardwareUsage(usage hardwareUsage) {
	scope := w.metricsScope.Tagged(map[string]string{clientHostTag: w.host})
	scope.Gauge(metrics.NumCPUCores).Update(float64(usage.NumCPUCores))
	scope.Gauge(metrics.CPUPercentage).Update(usage.CPUPercent)
	scope.Gauge(metrics.NumGoRoutines).Update(float64(usage.NumGoRoutines))
	scope.Gauge(metrics.TotalMemory).Update(float64(usage.TotalMemory))
	scope.Gauge(metrics.MemoryUsedHeap).Update(float64(usage.MemoryUsedHeap))
	scope.Gauge(metrics.MemoryUsedStack).Update(float64(usage.MemoryUsedStack))
}
