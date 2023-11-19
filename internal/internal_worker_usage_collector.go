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

	oncePerHost interface {
		Do(func())
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
	w.wg.Add(1)
	go func() {
		defer func() {
			if p := recover(); p != nil {
				w.metricsScope.Counter(metrics.WorkerUsageCollectorPanic).Inc(1)
				topLine := fmt.Sprintf("WorkerUsageCollector panic for workertype: %v", w.workerType)
				st := getStackTraceRaw(topLine, 7, 0)
				w.logger.Error("WorkerUsageCollector panic.",
					zap.String(tagPanicError, fmt.Sprintf("%v", p)),
					zap.String(tagPanicStack, st))
			}
		}()
		defer w.wg.Done()
		ticker := time.NewTicker(w.cooldownTime)
		defer ticker.Stop()

		w.wg.Add(1)
		go w.runHardwareCollector(ticker)

	}()
	return
}

func (w *workerUsageCollector) Stop() {
	w.cancel()
	close(w.shutdownCh)
	w.wg.Wait()
}

func (w *workerUsageCollector) runHardwareCollector(tick *time.Ticker) {
	defer w.wg.Done()
	w.emitOncePerHost.Do(func() {
		for {
			select {
			case <-w.shutdownCh:
				return
			case <-tick.C:
				hardwareUsageData := w.collectHardwareUsage()
				if w.metricsScope != nil {
					w.emitHardwareUsage(hardwareUsageData)
				}
			}
		}
	})
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
