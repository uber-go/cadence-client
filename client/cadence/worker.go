package cadence

import (
	"context"

	m "github.com/uber-go/cadence-client/.gen/go/cadence"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

type (
	// Worker represents objects that can be started and stopped.
	Worker interface {
		Stop()
		Start() error
	}

	// WorkerOptions is to configure a worker instance,
	// for example (1) the logger or any specific metrics.
	// 	       (2) Whether to heart beat for activities automatically.
	// Use NewWorkerOptions function to create an instance.
	WorkerOptions struct {
		// Optional: To set the maximum concurrent activity executions this host can have.
		// The zero value of this uses the default value.
		// default: defaultMaxConcurrentActivityExecutionSize(10k)
		MaxConcurrentActivityExecutionSize int

		// Optional: Sets the rate limiting on number of activities that can be executed per second.
		// This can be used to protect down stream services from flooding.
		// The zero value of this uses the default value.
		// default: defaultMaxActivityExecutionRate(100k)
		MaxActivityExecutionRate float32

		// Optional: if the activities need auto heart beating for those activities
		// by the framework
		// default: false not to heartbeat.
		AutoHeartBeat bool

		// Optional: Sets an identify that can be used to track this host for debugging.
		// default: default identity that include hostname, groupName and process ID.
		Identity string

		// Optional: Metrics to be reported.
		// default: no metrics.
		MetricsScope tally.Scope

		// Optional: Logger framework can use to log.
		// default: default logger provided.
		Logger *zap.Logger

		// Optional: Enable logging in replay
		// default: false
		EnableLoggingInReplay bool

		// Optional: Disable running workflow workers.
		// default: false
		DisableWorkflowWorker bool

		// Optional: Disable running activity workers.
		// default: false
		DisableActivityWorker bool

		// Optional: sets context for activity
		ActivityContext context.Context

		// Optional: This is only internal cadence use, users doesn't need to set this.
		TestTags map[string]map[string]string
	}
)

// NewWorker creates an instance of worker for managing workflow and activity executions.
// service 	- thrift connection to the cadence server.
// domain - the name of the cadence domain.
// groupName 	- is the name you use to identify your client worker, also
// 		  identifies group of workflow and activity implementations that are hosted by a single worker process.
// options 	-  configure any worker specific options like logger, metrics, identity.
func NewWorker(
	service m.TChanWorkflowService,
	domain string,
	groupName string,
	options WorkerOptions,
) Worker {
	return newAggregatedWorker(service, domain, groupName, options)
}
