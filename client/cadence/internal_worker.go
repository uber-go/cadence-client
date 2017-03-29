package cadence

// All code in this file is private to the package.

import (
	"github.com/Sirupsen/logrus"
	"github.com/uber-common/bark"
	m "github.com/uber-go/cadence-client/.gen/go/cadence"
	"github.com/uber-go/tally"
)

// Assert that structs do indeed implement the interfaces
var _ WorkerOptions = (*workerOptions)(nil)
var _ WorkflowTask = (*workflowTaskImpl)(nil)
var _ ActivityTask = (*activityTaskImpl)(nil)

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

// workerOptions stores all host-specific parameters that cadence can use to run the workflows
// and activities and if they need any rate limiting.
type workerOptions struct {
	// TODO
}

// SetIdentity identifies the host for debugging.
func (wo *workerOptions) SetIdentity(identity string) WorkerOptions {
	// TODO:
	return wo
}

// SetMetrics is the metrics that the client can use to report.
func (wo *workerOptions) SetMetrics(metricsScope tally.Scope) WorkerOptions{
	// TODO:
	return wo
}

// SetLogger sets the logger for the framework.
func (wo *workerOptions) SetLogger(logger bark.Logger) WorkerOptions {
	// TODO:
	return wo
}

// taskHost stores all worker-specific parameters that will
// be stored inside of a context.
type taskHost interface {
	RegisterWorkflow(taskListName string, factory WorkflowFactory)
	RegisterActivity(taskListName string, activities []Activity)
}

type taskHostImpl struct {
}

func (th *taskHostImpl) RegisterWorkflow(taskListName string, factory WorkflowFactory) {
	// TODO:
}
func (th *taskHostImpl) RegisterActivity(taskListName string, activities []Activity) {
	// TODO:
}

var thImpl *taskHostImpl

func getTaskHostEnvironment() taskHost {
	if thImpl == nil {
		thImpl = &taskHostImpl{}
	}
	return thImpl
}

type workflowTaskImpl struct {
}

type activityTaskImpl struct {
}

// SetMaxConcurrentActivityExecutionSize sets the maximum concurrent activity executions this host can have.
func (at *activityTaskImpl) SetMaxConcurrentActivityExecutionSize(size int) ActivityTask {
	// TODO:
	return at
}

// SetActivityExecutionRate sets the rate limiting on number of activities that can be executed.
func (at *activityTaskImpl) SetActivityExecutionRate(size int) ActivityTask {
	// TODO:
	return at
}

func (at *activityTaskImpl) SetAutoHeartBeat(auto bool) ActivityTask {
	// TODO:
	return at
}


// aggregatedWorker combines management of both workflowWorker and activityWorker worker lifecycle.
type aggregatedWorker struct {
	// TODO:
}

func (aw *aggregatedWorker) Start() error {
	// TODO:
	return nil
}

func (aw *aggregatedWorker) Stop() {
	// TODO:
}

// aggregatedWorker returns an instance to manage the workers.
func newAggregatedWorker(
	service m.TChanWorkflowService,
	workflowTasks []WorkflowTask,
	activityTasks []ActivityTask,
	options WorkerOptions,
) (worker Lifecycle) {
	return &aggregatedWorker{}
}


