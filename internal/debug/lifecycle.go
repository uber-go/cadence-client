package debug

import "go.uber.org/atomic"

type (
	// lifeCycleImpl implements the LifeCycle interface
	lifeCycleImpl struct {
		pollerCount atomic.Int32
	}

	// runImpl implements the Run interface
	runImpl struct {
		lifeCycle *lifeCycleImpl
		stopped   atomic.Bool
	}

	// Run is a helper for simpler tracking of the go routine run
	// Deprecated: in development and very likely to change
	Run interface {
		// Stop is the method to report stats once a poller thread is done
		Stop()
	}

	// LifeCycle contains a set of methods to collect information on a running worker
	// Deprecated: in development and very likely to change
	LifeCycle interface {
		// PollerStart collects information on poller start up.
		// consumers should provide a concurrency-safe implementation.
		PollerStart(workerID string) Run
		ReadPollerCount() int32
	}

	// EventMonitoring provides a set of methods that can be used to collect
	// stats on the Worker for debugging purposes.
	// Deprecated: in development and very likely to change
	EventMonitoring struct {
		LifeCycle LifeCycle
	}
)

func (lc *lifeCycleImpl) PollerStart(workerID string) Run {
	lc.pollerCount.Inc()
	return &runImpl{
		lifeCycle: lc,
		stopped:   atomic.Bool{},
	}
}

func (lc *lifeCycleImpl) ReadPollerCount() int32 {
	return lc.pollerCount.Load()
}

func NewLifeCycle() LifeCycle { return &lifeCycleImpl{} }

var _ LifeCycle = &lifeCycleImpl{}

func (r *runImpl) Stop() {
	// Check if Stop() has already been called
	if r.stopped.CompareAndSwap(false, true) {
		r.lifeCycle.pollerCount.Dec()
	} else {
		// Stop has already been called, do nothing
		return
	}
}

var _ Run = &runImpl{}
