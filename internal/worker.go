// Copyright (c) 2017-2020 Uber Technologies Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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
	"fmt"
	"time"

	"go.uber.org/cadence/internal/common/debug"

	"github.com/opentracing/opentracing-go"
	"github.com/uber-go/tally"
	"go.uber.org/zap"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common/auth"
)

type (
	// WorkerOptions is used to configure a worker instance.
	// The current timeout resolution implementation is in seconds and uses math.Ceil(d.Seconds()) as the duration. But is
	// subjected to change in the future.
	WorkerOptions struct {
		// Optional: To set the maximum concurrent activity executions this worker can have.
		// The zero value of this uses the default value.
		// default: defaultMaxConcurrentActivityExecutionSize(1k)
		MaxConcurrentActivityExecutionSize int

		// Optional: Sets the rate limiting on number of activities that can be executed per second per
		// worker. This can be used to limit resources used by the worker.
		// Notice that the number is represented in float, so that you can set it to less than
		// 1 if needed. For example, set the number to 0.1 means you want your activity to be executed
		// once for every 10 seconds. This can be used to protect down stream services from flooding.
		// The zero value of this uses the default value. Default: 100k
		WorkerActivitiesPerSecond float64

		// Optional: To set the maximum concurrent local activity executions this worker can have.
		// The zero value of this uses the default value.
		// default: 1k
		MaxConcurrentLocalActivityExecutionSize int

		// Optional: Sets the rate limiting on number of local activities that can be executed per second per
		// worker. This can be used to limit resources used by the worker.
		// Notice that the number is represented in float, so that you can set it to less than
		// 1 if needed. For example, set the number to 0.1 means you want your local activity to be executed
		// once for every 10 seconds. This can be used to protect down stream services from flooding.
		// The zero value of this uses the default value. Default: 100k
		WorkerLocalActivitiesPerSecond float64

		// Optional: Sets the rate limiting on number of activities that can be executed per second.
		// This is managed by the server and controls activities per second for your entire tasklist
		// whereas WorkerActivityTasksPerSecond controls activities only per worker.
		// Notice that the number is represented in float, so that you can set it to less than
		// 1 if needed. For example, set the number to 0.1 means you want your activity to be executed
		// once for every 10 seconds. This can be used to protect down stream services from flooding.
		// The zero value of this uses the default value. Default: 100k
		TaskListActivitiesPerSecond float64

		// optional: Sets the maximum number of goroutines that will concurrently poll the
		// cadence-server to retrieve activity tasks. Changing this value will affect the
		// rate at which the worker is able to consume tasks from a task list.
		// Default value is 2
		MaxConcurrentActivityTaskPollers int

		// optional: Sets the minimum number of goroutines that will concurrently poll the
		// cadence-server to retrieve activity tasks. Changing this value will NOT affect the
		// rate at which the worker is able to consume tasks from a task list,
		// unless FeatureFlags.PollerAutoScalerEnabled is set to true.
		// Default value is 1
		MinConcurrentActivityTaskPollers int

		// Optional: To set the maximum concurrent decision task executions this worker can have.
		// The zero value of this uses the default value.
		// default: defaultMaxConcurrentTaskExecutionSize(1k)
		MaxConcurrentDecisionTaskExecutionSize int

		// Optional: Sets the rate limiting on number of decision tasks that can be executed per second per
		// worker. This can be used to limit resources used by the worker.
		// The zero value of this uses the default value. Default: 100k
		WorkerDecisionTasksPerSecond float64

		// optional: Sets the maximum number of goroutines that will concurrently poll the
		// cadence-server to retrieve decision tasks. Changing this value will affect the
		// rate at which the worker is able to consume tasks from a task list.
		// Default value is 2
		MaxConcurrentDecisionTaskPollers int

		// optional: Sets the minimum number of goroutines that will concurrently poll the
		// cadence-server to retrieve decision tasks. Unless FeatureFlags.PollerAutoScalerEnabled is set to true,
		// changing this value will NOT affect the rate at which the worker is able to consume tasks from a task list.
		// Default value is 2
		MinConcurrentDecisionTaskPollers int

		// optional: Sets the interval of poller autoscaling, between which poller autoscaler changes the poller count
		// based on poll result. It takes effect if FeatureFlags.PollerAutoScalerEnabled is set to true.
		// Default value is 1 min
		PollerAutoScalerCooldown time.Duration

		// optional: Sets the target utilization rate between [0,1].
		// Utilization Rate = pollResultWithTask / (pollResultWithTask + pollResultWithNoTask)
		// It takes effect if FeatureFlags.PollerAutoScalerEnabled is set to true.
		// Default value is 0.6
		PollerAutoScalerTargetUtilization float64

		// optional: Sets whether to start dry run mode of autoscaler.
		// Default value is false
		PollerAutoScalerDryRun bool

		// Optional: Sets an identify that can be used to track this host for debugging.
		// default: default identity that include hostname, groupName and process ID.
		Identity string

		// Optional: Defines the 'zone' or the failure group that the worker belongs to
		IsolationGroup string

		// Optional: Metrics to be reported. Metrics emitted by the cadence client are not prometheus compatible by
		// default. To ensure metrics are compatible with prometheus make sure to create tally scope with sanitizer
		// options set.
		// var (
		// _safeCharacters = []rune{'_'}
		// _sanitizeOptions = tally.SanitizeOptions{
		// 	NameCharacters: tally.ValidCharacters{
		// 		Ranges:     tally.AlphanumericRange,
		// 		Characters: _safeCharacters,
		// 	},
		// 		KeyCharacters: tally.ValidCharacters{
		// 			Ranges:     tally.AlphanumericRange,
		// 			Characters: _safeCharacters,
		// 		},
		// 		ValueCharacters: tally.ValidCharacters{
		// 			Ranges:     tally.AlphanumericRange,
		// 			Characters: _safeCharacters,
		// 		},
		// 		ReplacementCharacter: tally.DefaultReplacementCharacter,
		// 	}
		// )
		// opts := tally.ScopeOptions{
		// 	Reporter:        reporter,
		// 	SanitizeOptions: &_sanitizeOptions,
		// }
		// scope, _ := tally.NewRootScope(opts, time.Second)
		// default: no metrics.
		MetricsScope tally.Scope

		// Optional: Logger framework can use to log.
		// default: default logger provided.
		Logger *zap.Logger

		// Optional: Enable logging in replay.
		// In the workflow code you can use workflow.GetLogger(ctx) to write logs. By default, the logger will skip log
		// entry during replay mode so you won't see duplicate logs. This option will enable the logging in replay mode.
		// This is only useful for debugging purpose.
		// default: false
		EnableLoggingInReplay bool

		// Optional: Disable running workflow workers.
		// default: false
		DisableWorkflowWorker bool

		// Optional: Disable running activity workers.
		// default: false
		DisableActivityWorker bool

		// Optional: Disable sticky execution.
		// default: false
		// Sticky Execution is to run the decision tasks for one workflow execution on same worker host. This is an
		// optimization for workflow execution. When sticky execution is enabled, worker keeps the workflow state in
		// memory. New decision task contains the new history events will be dispatched to the same worker. If this
		// worker crashes, the sticky decision task will timeout after StickyScheduleToStartTimeout, and cadence server
		// will clear the stickiness for that workflow execution and automatically reschedule a new decision task that
		// is available for any worker to pick up and resume the progress.
		DisableStickyExecution bool

		// Optional: Sticky schedule to start timeout.
		// default: 5s
		// The resolution is seconds. See details about StickyExecution on the comments for DisableStickyExecution.
		StickyScheduleToStartTimeout time.Duration

		// Optional: sets context for activity. The context can be used to pass any configuration to activity
		// like common logger for all activities.
		BackgroundActivityContext context.Context

		// Optional: Sets how decision worker deals with non-deterministic history events
		// (presumably arising from non-deterministic workflow definitions or non-backward compatible workflow definition changes).
		// default: NonDeterministicWorkflowPolicyBlockWorkflow, which just logs error but reply nothing back to server
		NonDeterministicWorkflowPolicy NonDeterministicWorkflowPolicy

		// Optional: Sets DataConverter to customize serialization/deserialization of arguments in Cadence
		// default: defaultDataConverter, an combination of thriftEncoder and jsonEncoder
		DataConverter DataConverter

		// Optional: worker graceful shutdown timeout
		// default: 0s
		WorkerStopTimeout time.Duration

		// Optional: Enable running session workers.
		// Session workers is for activities within a session.
		// Enable this option to allow worker to process sessions.
		// default: false
		EnableSessionWorker bool

		// Uncomment this option when we support automatic reestablish failed sessions.
		// Optional: The identifier of the resource consumed by sessions.
		// It's the user's responsibility to ensure there's only one worker using this resourceID.
		// For now, if user doesn't specify one, a new uuid will be used as the resourceID.
		// SessionResourceID string

		// Optional: Sets the maximum number of concurrently running sessions the resource support.
		// default: 1000
		MaxConcurrentSessionExecutionSize int

		// Optional: Specifies factories used to instantiate workflow interceptor chain
		// The chain is instantiated per each replay of a workflow execution
		WorkflowInterceptorChainFactories []WorkflowInterceptorFactory

		// Optional: Sets ContextPropagators that allows users to control the context information passed through a workflow
		// default: no ContextPropagators
		ContextPropagators []ContextPropagator

		// Optional: Sets opentracing Tracer that is to be used to emit tracing information
		// default: no tracer - opentracing.NoopTracer
		Tracer opentracing.Tracer

		// Optional: Enable worker for running shadowing workflows to replay existing workflows
		// If set to true:
		// 1. Worker will run in shadow mode and all other workers (decision, activity, session)
		// will be disabled to prevent them from updating existing workflow states.
		// 2. DataConverter, WorkflowInterceptorChainFactories, ContextPropagators, Tracer will be
		// used as ReplayOptions and forwarded to the underlying WorkflowReplayer.
		// The actual shadower activity worker will not use them.
		// 3. TaskList will become Domain-TaskList, to prevent conflict across domains as there's
		// only one shadowing domain which is responsible for shadowing workflows for all domains.
		// default: false
		EnableShadowWorker bool

		// Optional: Configures shadowing workflow
		// default: please check the documentation for ShadowOptions for default options
		ShadowOptions ShadowOptions

		// Optional: Flags to turn on/off some server side options
		// default: all the features in the struct are turned off
		FeatureFlags FeatureFlags

		// Optional: Authorization interface to get the Auth Token
		// default: No provider
		Authorization auth.AuthorizationProvider

		// Optional: See WorkerBugPorts for more details
		//
		// Deprecated: All bugports are always deprecated and may be removed at any time.
		WorkerBugPorts WorkerBugPorts

		// Optional: WorkerStats provides a set of methods that can be used to collect
		// stats on the Worker for debugging purposes.
		// default: noop implementation provided
		// Deprecated: in development and very likely to change
		WorkerStats debug.WorkerStats
	}

	// WorkerBugPorts allows opt-in enabling of older, possibly buggy behavior, primarily intended to allow temporarily
	// emulating old behavior until a fix is deployed.
	// By default, bugs (especially rarely-occurring ones) are fixed and all users are opted into the new behavior.
	// Back-ported buggy behavior *may* be available via these flags.
	//
	// Bugports are always deprecated and may be removed in future versions.
	// Generally speaking they will *likely* remain in place for one minor version, and then they may be removed to
	// allow cleaning up the additional code complexity that they cause.
	// Deprecated: All bugports are always deprecated and may be removed at any time
	WorkerBugPorts struct {
		// Optional: Disable strict non-determinism checks for workflow.
		// There are some non-determinism cases which are missed by original implementation and a fix is on the way.
		// The fix will be toggleable by this parameter.
		// Default: false, which means strict non-determinism checks are enabled.
		//
		// Deprecated: All bugports are always deprecated and may be removed at any time
		DisableStrictNonDeterminismCheck bool
	}
)

// NonDeterministicWorkflowPolicy is an enum for configuring how client's decision task handler deals with
// mismatched history events (presumably arising from non-deterministic workflow definitions).
type NonDeterministicWorkflowPolicy int

const (
	// NonDeterministicWorkflowPolicyBlockWorkflow is the default policy for handling detected non-determinism.
	// This option simply logs to console with an error message that non-determinism is detected, but
	// does *NOT* reply anything back to the server.
	// It is chosen as default for backward compatibility reasons because it preserves the old behavior
	// for handling non-determinism that we had before NonDeterministicWorkflowPolicy type was added to
	// allow more configurability.
	NonDeterministicWorkflowPolicyBlockWorkflow NonDeterministicWorkflowPolicy = iota
	// NonDeterministicWorkflowPolicyFailWorkflow behaves exactly the same as Ignore, up until the very
	// end of processing a decision task.
	// Whereas default does *NOT* reply anything back to the server, fail workflow replies back with a request
	// to fail the workflow execution.
	NonDeterministicWorkflowPolicyFailWorkflow
)

// NewWorker creates an instance of worker for managing workflow and activity executions.
// service 	- thrift connection to the cadence server.
// domain - the name of the cadence domain.
// taskList 	- is the task list name you use to identify your client worker, also
//
//	identifies group of workflow and activity implementations that are hosted by a single worker process.
//
// options 	-  configure any worker specific options like logger, metrics, identity.
func NewWorker(
	service workflowserviceclient.Interface,
	domain string,
	taskList string,
	options WorkerOptions,
) (*aggregatedWorker, error) {
	return newAggregatedWorker(service, domain, taskList, options)
}

// ReplayWorkflowExecution loads a workflow execution history from the Cadence service and executes a single decision task for it.
// Use for testing backwards compatibility of code changes and troubleshooting workflows in a debugger.
// The logger is the only optional parameter. Defaults to the noop logger.
// Deprecated: Global workflow replay methods are replaced by equivalent WorkflowReplayer instance methods.
// This method is kept to maintain backward compatibility and should not be used.
func ReplayWorkflowExecution(
	ctx context.Context,
	service workflowserviceclient.Interface,
	logger *zap.Logger,
	domain string,
	execution WorkflowExecution,
) error {
	r := NewWorkflowReplayer()
	return r.ReplayWorkflowExecution(ctx, service, logger, domain, execution)
}

// ReplayWorkflowHistory executes a single decision task for the given history.
// Use for testing the backwards compatibility of code changes and troubleshooting workflows in a debugger.
// The logger is an optional parameter. Defaults to the noop logger.
// Deprecated: Global workflow replay methods are replaced by equivalent WorkflowReplayer instance methods.
// This method is kept to maintain backward compatibility and should not be used.
func ReplayWorkflowHistory(logger *zap.Logger, history *shared.History) error {
	r := NewWorkflowReplayer()
	return r.ReplayWorkflowHistory(logger, history)
}

// ReplayWorkflowHistoryFromJSONFile executes a single decision task for the given json history file.
// Use for testing backwards compatibility of code changes and troubleshooting workflows in a debugger.
// The logger is an optional parameter. Defaults to the noop logger.
// Deprecated: Global workflow replay methods are replaced by equivalent WorkflowReplayer instance methods.
// This method is kept to maintain backward compatibility and should not be used.
func ReplayWorkflowHistoryFromJSONFile(logger *zap.Logger, jsonfileName string) error {
	r := NewWorkflowReplayer()
	return r.ReplayWorkflowHistoryFromJSONFile(logger, jsonfileName)
}

// ReplayPartialWorkflowHistoryFromJSONFile executes a single decision task for the given json history file upto provided
// lastEventID(inclusive).
// Use for testing backwards compatibility of code changes and troubleshooting workflows in a debugger.
// The logger is an optional parameter. Defaults to the noop logger.
// Deprecated: Global workflow replay methods are replaced by equivalent WorkflowReplayer instance methods.
// This method is kept to maintain backward compatibility and should not be used.
func ReplayPartialWorkflowHistoryFromJSONFile(logger *zap.Logger, jsonfileName string, lastEventID int64) error {
	r := NewWorkflowReplayer()
	return r.ReplayPartialWorkflowHistoryFromJSONFile(logger, jsonfileName, lastEventID)
}

// Validate sanity validation of WorkerOptions
func (o WorkerOptions) Validate() error {
	// decision task pollers must be >= 2 or unset if sticky tasklist is enabled https://github.com/uber-go/cadence-client/issues/1369
	if !o.DisableStickyExecution && (o.MaxConcurrentDecisionTaskPollers == 1 || o.MinConcurrentDecisionTaskPollers == 1) {
		return fmt.Errorf("DecisionTaskPollers must be >= 2 or use default value")
	}
	return nil
}
