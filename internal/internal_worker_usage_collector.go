package internal

import (
	"context"
	"fmt"
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
		workerType      string
		cooldownTime    time.Duration
		logger          *zap.Logger
		ctx             context.Context
		shutdownCh      chan struct{}
		wg              *sync.WaitGroup // graceful stop
		cancel          context.CancelFunc
		metricsScope    tally.Scope
		emitOncePerHost oncePerHost
	}

	workerUsageCollectorOptions struct {
		Enabled      bool
		Cooldown     time.Duration
		MetricsScope tally.Scope
		WorkerType   string
		EmitOnce     oncePerHost
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
		workerType:      options.WorkerType,
		cooldownTime:    options.Cooldown,
		metricsScope:    options.MetricsScope,
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
		wg:              &sync.WaitGroup{},
		emitOncePerHost: options.EmitOnce,
		shutdownCh:      make(chan struct{}),
	}
}

func (w *workerUsageCollector) Start() {

	if w.emitOncePerHost != nil {
		w.emitOncePerHost.Do(
			func() {
				w.wg.Add(1)
				w.logger.Info(fmt.Sprintf("Going to start hardware collector for workertype: %v", w.workerType))
				go w.runHardwareCollector()
			})
	}

}

func (w *workerUsageCollector) Stop() {
	close(w.shutdownCh)
	w.wg.Wait()
	w.cancel()

}

func (w *workerUsageCollector) runHardwareCollector() {
	defer w.wg.Done()
	ticker := time.NewTicker(w.cooldownTime)
	defer ticker.Stop()
	w.logger.Info(fmt.Sprintf("Started worker usage collector for workertype: %v", w.workerType))
	for {
		select {
		case <-w.shutdownCh:
			return
		case <-ticker.C:
			hardwareUsageData := w.collectHardwareUsage()
			if w.metricsScope != nil {
				w.emitHardwareUsage(hardwareUsageData)
			}
		}
	}
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
	w.metricsScope.Gauge(metrics.NumCPUCores).Update(float64(usage.NumCPUCores))
	w.metricsScope.Gauge(metrics.CPUPercentage).Update(usage.CPUPercent)
	w.metricsScope.Gauge(metrics.NumGoRoutines).Update(float64(usage.NumGoRoutines))
	w.metricsScope.Gauge(metrics.TotalMemory).Update(float64(usage.TotalMemory))
	w.metricsScope.Gauge(metrics.MemoryUsedHeap).Update(float64(usage.MemoryUsedHeap))
	w.metricsScope.Gauge(metrics.MemoryUsedStack).Update(float64(usage.MemoryUsedStack))
}
