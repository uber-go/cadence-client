package cadence

import (
	"errors"

	"github.com/uber-common/bark"
	s "github.com/uber-go/cadence-client/.gen/go/shared"
)

// DecisionTaskHandler handles decision task by replaying corresponding workflow function.
type DecisionTaskHandler struct {
	workflowDefFactory workflowDefinitionFactory
	logger             bark.Logger
	task               *s.PollForDecisionTaskResponse
	decisions          *s.RespondDecisionTaskCompletedRequest
	stackTrace         string
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

