package cadence

// WARNING! WARNING! WARNING! WARNING! WARNING! WARNING! WARNING! WARNING! WARNING!
// Any of the APIs in this file are not supported for application level developers
// and are subject to change without any notice.
//
// APIs that are needed to test or integrate Cadence with external systems like Catalyst.
// They are internal to Cadence system developers and are public from Go sense only due
// to need to access them from other packages.
//

import (
	"errors"

	"github.com/uber-common/bark"
	m "github.com/uber-go/cadence-client/.gen/go/cadence"
	s "github.com/uber-go/cadence-client/.gen/go/shared"
)

type (
	// WorkflowTaskHandler represents workflow task handlers.
	WorkflowTaskHandler interface {
		// Process the workflow task
		ProcessWorkflowTask(task *s.PollForDecisionTaskResponse, emitStack bool) (response *s.RespondDecisionTaskCompletedRequest, stackTrace string, err error)
	}

	// ActivityTaskHandler represents activity task handlers.
	ActivityTaskHandler interface {
		// Execute the activity task
		// The return interface{} can have three requests, use switch to find the type of it.
		// - RespondActivityTaskCompletedRequest
		// - RespondActivityTaskFailedRequest
		// - RespondActivityTaskCancelRequest
		Execute(task *s.PollForActivityTaskResponse) (interface{}, error)
	}

	// DecisionTaskHandler handles decision task by replaying corresponding workflow function.
	DecisionTaskHandler struct {
		workflowDefFactory workflowDefinitionFactory
		logger             bark.Logger
		task               *s.PollForDecisionTaskResponse
		decisions          *s.RespondDecisionTaskCompletedRequest
		stackTrace         string
	}
)

// NewWorkerOptionsInternal creates an instance of worker options with default values.
func NewWorkerOptionsInternal(testTags map[string]map[string]string) WorkerOptions {
	return &workerOptions{
		maxConcurrentActivityExecutionSize: defaultMaxConcurrentActivityExecutionSize,
		maxActivityExecutionRate:           defaultMaxActivityExecutionRate,
		autoHeartBeatForActivities:         false,
		testTags:                           testTags,
		// Defaults for metrics, identity, logger is filled in by the WorkflowWorker APIs.
	}
}

// NewWorkflowTaskWorker returns an instance of a workflow task handler worker.
// To be used by framework level code that requires access to the original workflow task.
func NewWorkflowTaskWorker(
	taskHandler WorkflowTaskHandler,
	service m.TChanWorkflowService,
	params workerExecutionParameters,
) (worker Worker) {
	return newWorkflowTaskWorkerInternal(taskHandler, service, params)
}

// NewActivityTaskWorker returns instance of an activity task handler worker.
// To be used by framework level code that requires access to the original workflow task.
func NewActivityTaskWorker(
	taskHandler ActivityTaskHandler,
	service m.TChanWorkflowService,
	params workerExecutionParameters,
) Worker {
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

// NewDecisionTaskHandler creates an instance of DecisionTaskHandler from a decision poll response
// using workflow functions registered through RegisterWorkflow
// To be used to replay a workflow in a debugger.
func NewDecisionTaskHandler(task *s.PollForDecisionTaskResponse, logger bark.Logger) *DecisionTaskHandler {
	return &DecisionTaskHandler{
		workflowDefFactory: getWorkflowDefinitionFactory(newRegisteredWorkflowFactory()),
		logger:             logger,
		task:               task,
	}
}

// Process replays the history.
func (wr *DecisionTaskHandler) Process(emitStack bool) (err error) {
	history := wr.task.GetHistory()
	if history == nil {
		return errors.New("nil history")
	}
	event := history.Events[0]
	if history == nil {
		return errors.New("nil first history event")
	}
	attributes := event.GetWorkflowExecutionStartedEventAttributes()
	if attributes == nil {
		return errors.New("first history event is not WorkflowExecutionStarted")
	}
	taskList := attributes.GetTaskList()
	if taskList == nil {
		return errors.New("nil taskList in WorkflowExecutionStarted event")
	}
	params := workerExecutionParameters{
		TaskList: taskList.GetName(),
		Identity: getWorkerIdentity(taskList.GetName()),
		Logger:   wr.logger,
	}
	taskHandler := newWorkflowTaskHandler(
		wr.workflowDefFactory,
		params,
		nil)
	wr.decisions, wr.stackTrace, err = taskHandler.ProcessWorkflowTask(wr.task, emitStack)
	return err
}

// StackTrace returns the stack trace dump of all current workflow goroutines
func (wr *DecisionTaskHandler) StackTrace() string {
	return wr.stackTrace
}

// Decisions that are result of a decision task.
func (wr *DecisionTaskHandler) Decisions() *s.RespondDecisionTaskCompletedRequest {
	return wr.decisions
}
