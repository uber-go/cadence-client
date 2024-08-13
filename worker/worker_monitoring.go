package worker

import "go.uber.org/cadence/internal"

type (
	// EventsMonitoringprovides a set of methods that can be used to collect
	// stats on the Worker for debugging purposes.
	EventsMonitoring = internal.EventMonitoring

	// PollerLifeCycle contains a set of methods to collect information on running pollers for a worker
	// Deprecated: in development and very likely to change
	PollerLifeCycle = internal.PollerLifeCycle

	// PollerRun contains a set of methods to collect information on
	// Deprecated: in development and very likely to change
	PollerRun = internal.PollerRun
)
