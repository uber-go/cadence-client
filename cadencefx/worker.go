// Copyright (c) 2017 Uber Technologies, Inc.
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

package cadencefx

import (
	"context"
	"fmt"

	"github.com/uber-go/tally"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/worker"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// HostSpecificWorkerConfig is used create a worker that waits on the uniquely named
// host specific task list specified by TaskList.
//
// A host specific task list is unique to the host the worker is running on, thus, scheduling a task to a host specific tasklist
// forces the activity to be processed only by workers on the host. The main use case for this feature is with workflows that
// require all subsequent activities be completed on the same node as its first activity. In other words, workflows that store
// some or all of its working state on a node and therefore must have the entirety of its run stickied to a host.
//
// A Provider function using this struct could look like:
//
//	func newHostSpecificTaskList() *HostSpecificTaskListConfig {
//		return &HostSpecificTaskListConfig{
//			Domain:"TestDomain",
//			TaskList: "TestTaskList",
//			Options: worker.Options{DisableWorkflowWorker:false},
//		}}
//	}
type HostSpecificWorkerConfig struct {
	Domain   string
	TaskList string
	Options  options
}

// ContextInjector is used to ensure that properties are properly wired to the
// context that will be used as the root context for activity invocations.
// Generally this is for activity libraries to ensure their runtime dependencies
// are wired up.
type ContextInjector func(context.Context) context.Context

// The ContextInjectorResult is sugar for ensuring your ContextInjector is
// under the appropriate named Fx group ("activities_context")
type ContextInjectorResult struct {
	fx.Out

	ContextInjector ContextInjector `group:"activities_context"`
}

// WorkerParams is used to inject the input params for the NewWorkers method.
type WorkerParams struct {
	fx.In

	ServiceClient              workflowserviceclient.Interface
	HostSpecificWorkerConfig   *HostSpecificWorkerConfig `optional:"true"`
	ActivitiesContextInjectors []ContextInjector         `group:"activities_context"`

	Config config.Provider
	Scope  tally.Scope
	Logger *zap.Logger
}

// NewWorkers instantiates Cadence workers based on passed in configuration.
func NewWorkers(p WorkerParams) ([]worker.Worker, error) {
	config, err := loadWorkerConfig(p.Config)
	if err != nil {
		return nil, err
	}

	// If a HostSpecificWorker is injected, add it to the workerConfig slice
	// so that it will get processed and instantiated below.
	if p.HostSpecificWorkerConfig != nil {
		hostSpecificWorkerConfig := workerConfig{
			TaskList: p.HostSpecificWorkerConfig.TaskList,
			Domain:   p.HostSpecificWorkerConfig.Domain,
			Options:  p.HostSpecificWorkerConfig.Options}

		config.Workers = append(config.Workers, hostSpecificWorkerConfig)
	}

	ctx := context.Background()
	for _, inject := range p.ActivitiesContextInjectors {
		ctx = inject(ctx)
	}

	var workers []worker.Worker
	for _, wConfig := range config.Workers {
		// create the worker
		opts := wConfig.Options
		workerOptions := worker.Options{
			MetricsScope:                 p.Scope,
			Logger:                       p.Logger,
			BackgroundActivityContext:    ctx,
			TaskListActivitiesPerSecond:  opts.TaskListActivitiesPerSecond,
			WorkerActivitiesPerSecond:    opts.WorkerActivitiesPerSecond,
			AutoHeartBeat:                opts.AutoHeartBeat,
			Identity:                     opts.Identity,
			EnableLoggingInReplay:        opts.EnableLoggingInReplay,
			DisableWorkflowWorker:        opts.DisableWorkflowWorker,
			DisableActivityWorker:        opts.DisableActivityWorker,
			DisableStickyExecution:       opts.DisableStickyExecution,
			StickyScheduleToStartTimeout: opts.StickyScheduleToStartTimeout,
		}

		worker := worker.New(
			p.ServiceClient,
			wConfig.Domain,
			wConfig.TaskList,
			workerOptions)

		// add to slice
		workers = append(workers, worker)
	}
	return workers, nil
}

// RegisterParams is used to inject the parameters for the RegisterWorkers method.
type RegisterParams struct {
	fx.In

	Lifecycle fx.Lifecycle
	Logger    *zap.Logger
}

// RegisterWorkers inserts hooks in the service Lifecycle that start and Stop
// the Cadence workers.
func RegisterWorkers(p RegisterParams, workers []worker.Worker) {
	for _, worker := range workers {
		w := worker
		p.Lifecycle.Append(fx.Hook{
			OnStart: func(context.Context) error {
				p.Logger.Info("Starting Worker.", zap.Any("Worker", fmt.Sprintf("%d", &w)))
				return w.Start()
			},
			OnStop: func(context.Context) error {
				p.Logger.Info("Stopping Worker.", zap.Any("Worker", fmt.Sprintf("%d", &w)))
				w.Stop()
				return nil
			},
		})
	}
}
