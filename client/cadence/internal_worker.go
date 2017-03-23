package cadence

// All code in this file is private to the package.

import (
	"github.com/Sirupsen/logrus"
	"github.com/uber-common/bark"
	m "github.com/uber-go/cadence-client/.gen/go/cadence"
	"github.com/uber-go/tally"
)

// Assert that structs do indeed implement the interfaces
var _ taskHost = (*taskHostImpl)(nil)
var _ HostOptions = (*hostOptions)(nil)

type (
	// WorkflowWorker wraps the code for hosting workflow types.
	// And worker is mapped 1:1 with task list. If the user want's to poll multiple
	// task list names they might have to manage 'n' workers for 'n' task lists.
	workflowWorker struct {
		executionParameters WorkerExecutionParameters
		workflowService     m.TChanWorkflowService
		poller              taskPoller // taskPoller to poll the tasks.
		worker              *baseWorker
		identity            string
	}

	// activityRegistry collection of activity implementations
	activityRegistry map[string]Activity

	// ActivityWorker wraps the code for hosting activity types.
	// TODO: Worker doing heartbeating automatically while activity task is running
	activityWorker struct {
		executionParameters WorkerExecutionParameters
		activityRegistry    activityRegistry
		workflowService     m.TChanWorkflowService
		poller              *activityTaskPoller
		worker              *baseWorker
		identity            string
	}

	// Worker overrides.
	workerOverrides struct {
		workflowTaskHander  WorkflowTaskHandler
		activityTaskHandler ActivityTaskHandler
	}
)

// NewWorkflowTaskWorker returns an instance of a workflow task handler worker.
// To be used by framework level code that requires access to the original workflow task.
func NewWorkflowTaskWorker(
	taskHandler WorkflowTaskHandler,
	service m.TChanWorkflowService,
	params WorkerExecutionParameters,
) (worker Lifecycle) {
	return newWorkflowTaskWorkerInternal(taskHandler, service, params)
}

// newWorkflowWorker returns an instance of the workflow worker.
func newWorkflowWorker(
	factory workflowDefinitionFactory,
	service m.TChanWorkflowService,
	params WorkerExecutionParameters,
	ppMgr pressurePointMgr,
) Lifecycle {
	return newWorkflowWorkerInternal(factory, service, params, ppMgr, nil)
}

func ensureRequiredParams(params *WorkerExecutionParameters) {
	if params.Identity == "" {
		params.Identity = getWorkerIdentity(params.TaskList)
	}
	if params.Logger == nil {
		log := logrus.New()
		params.Logger = bark.NewLoggerFromLogrus(log)
		params.Logger.Info("No logger configured for cadence worker. Created default one.")
	}
}

func newWorkflowWorkerInternal(
	factory workflowDefinitionFactory,
	service m.TChanWorkflowService,
	params WorkerExecutionParameters,
	ppMgr pressurePointMgr,
	overrides *workerOverrides,
) Lifecycle {
	// Get a workflow task handler.
	ensureRequiredParams(&params)
	var taskHandler WorkflowTaskHandler
	if overrides != nil && overrides.workflowTaskHander != nil {
		taskHandler = overrides.workflowTaskHander
	} else {
		taskHandler = newWorkflowTaskHandler(factory, params, ppMgr)
	}
	return newWorkflowTaskWorkerInternal(taskHandler, service, params)
}

func newWorkflowTaskWorkerInternal(
	taskHandler WorkflowTaskHandler,
	service m.TChanWorkflowService,
	params WorkerExecutionParameters,
) Lifecycle {
	ensureRequiredParams(&params)
	poller := newWorkflowTaskPoller(
		taskHandler,
		service,
		params,
	)
	worker := newBaseWorker(baseWorkerOptions{
		routineCount:    params.ConcurrentPollRoutineSize,
		taskPoller:      poller,
		workflowService: service,
		identity:        params.Identity,
		workerType:      "DecisionWorker"},
		params.Logger)

	return &workflowWorker{
		executionParameters: params,
		workflowService:     service,
		poller:              poller,
		worker:              worker,
		identity:            params.Identity,
	}
}

// Start the worker.
func (ww *workflowWorker) Start() error {
	ww.worker.Start()
	return nil // TODO: propagate error
}

// Shutdown the worker.
func (ww *workflowWorker) Stop() {
	ww.worker.Stop()
}

func newActivityWorkerInternal(
	activities []Activity,
	service m.TChanWorkflowService,
	params WorkerExecutionParameters,
	overrides *workerOverrides,
) Lifecycle {
	ensureRequiredParams(&params)
	// Get a activity task handler.
	var taskHandler ActivityTaskHandler
	if overrides != nil && overrides.activityTaskHandler != nil {
		taskHandler = overrides.activityTaskHandler
	} else {
		taskHandler = newActivityTaskHandler(activities, service, params)
	}
	return NewActivityTaskWorker(taskHandler, service, params)
}

// NewActivityTaskWorker returns instance of an activity task handler worker.
// To be used by framework level code that requires access to the original workflow task.
func NewActivityTaskWorker(
	taskHandler ActivityTaskHandler,
	service m.TChanWorkflowService,
	params WorkerExecutionParameters,
) Lifecycle {
	ensureRequiredParams(&params)

	poller := newActivityTaskPoller(
		taskHandler,
		service,
		params,
	)
	worker := newBaseWorker(baseWorkerOptions{
		routineCount:    params.ConcurrentPollRoutineSize,
		taskPoller:      poller,
		workflowService: service,
		identity:        params.Identity,
		workerType:      "ActivityWorker"},
		params.Logger)

	return &activityWorker{
		executionParameters: params,
		activityRegistry:    make(map[string]Activity),
		workflowService:     service,
		worker:              worker,
		poller:              poller,
		identity:            params.Identity,
	}
}

// Start the worker.
func (aw *activityWorker) Start() error {
	aw.worker.Start()
	return nil // TODO: propagate errors
}

// Shutdown the worker.
func (aw *activityWorker) Stop() {
	aw.worker.Stop()
}

const hostEnvContextKey = "hostEnv"

func getHostEnvironment(ctx Context) *taskHostImpl {
	eap := ctx.Value(hostEnvContextKey)
	if eap == nil {
		return nil
	}
	return eap.(*taskHostImpl)
}

func setHostEnvironment(ctx Context, service m.TChanWorkflowService) Context {
	if valCtx := getHostEnvironment(ctx); valCtx == nil {
		return WithValue(ctx, hostEnvContextKey, &taskHostImpl{})
	}
	return ctx
}

// hostOptions stores all host-specific parameters that cadence can use to run the workflows
// and activities and if they need any reate limiting.
type hostOptions struct {
	// TODO
}

// SetMaxConcurrentActivityExecutionSize sets the maximum concurrent activity executions this host can have.
func (ho *hostOptions) SetMaxConcurrentActivityExecutionSize(size int) HostOptions {
	// TODO:
	return ho
}

// SetActivityExecutionRate sets the rate limiting on number of activities that can be executed.
func (ho *hostOptions) SetActivityExecutionRate(size int) HostOptions {
	// TODO:
	return ho
}

// SetIdentity identifies the host for debugging.
func (ho *hostOptions) SetIdentity(identity string) HostOptions {
	// TODO:
	return ho
}

// SetMetrics is the metrics that the client can use to report.
func (ho *hostOptions) SetMetrics(metricsScope tally.Scope) HostOptions {
	// TODO:
	return ho
}

// SetLogger sets the logger for the framework.
func (ho *hostOptions) SetLogger(logger bark.Logger) HostOptions {
	// TODO:
	return ho
}

// taskHost stores all worker-specific parameters that will
// be stored inside of a context.
type taskHost interface {
	Lifecycle
	RegisterWorkflow(taskListName string, factory WorkflowFactory)
	RegisterActivity(taskListName string, activities []Activity, autoHeartBeat bool)
	SetOptions(options hostOptions)
}

// taskHostImpl implementation of taskHost
type taskHostImpl struct {
	// TODO:
}

func (th *taskHostImpl) RegisterWorkflow(taskListName string, factory WorkflowFactory) {
	// TODO:
}

func (th *taskHostImpl) RegisterActivity(taskListName string, activities []Activity, autoHeartBeat bool) {
	// TODO:
}

func (th *taskHostImpl) SetOptions(options hostOptions) {
	// TODO:
}

func (th *taskHostImpl) Start() error {
	// TODO:
	return nil
}

func (th *taskHostImpl) Stop() {
	// TODO:
}


