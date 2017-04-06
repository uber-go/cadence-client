package cadence

import (
	m "github.com/uber-go/cadence-client/.gen/go/cadence"
	s "github.com/uber-go/cadence-client/.gen/go/shared"
	"github.com/uber-go/tally"
)

type (
	// WorkflowClient is the client for starting and getting information about a workflow execution.
	WorkflowClient interface {
		// StartWorkflowExecution starts a workflow execution
		// The user can use this to start using a function or workflow type name.
		// Either by
		//     StartWorkflowExecution(options, "workflowTypeName", input)
		//     or
		//     StartWorkflowExecution(options, workflowExecuteFn, arg1, arg2, arg3)
		StartWorkflowExecution(options StartWorkflowOptions, workflow interface{}, args ...interface{}) (*WorkflowExecution, error)

		// GetHistory gets history of a particular workflow.
		GetHistory(workflowID string, runID string) (*s.History, error)

		// CompleteActivity reports activity completed. Activity Execute method can return cadence.ActivityResultPendingError to
		// indicate the activity is not completed when it's Execute method returns. In that case, this CompleteActivity() method
		// should be called when that activity is completed with the actual result and error. If err is nil, activity task
		// completed event will be reported; if err is CanceledError, activity task cancelled event will be reported; otherwise,
		// activity task failed event will be reported.
		CompleteActivity(taskToken, result []byte, err error) error

		// RecordActivityHeartbeat records heartbeat for an activity.
		RecordActivityHeartbeat(taskToken, details []byte) error
	}

	// WorkflowClientOptions are optional parameters for WorkflowClient creation.
	WorkflowClientOptions struct {
		MetricsScope tally.Scope
		Identity     string
	}

	// StartWorkflowOptions configuration parameters for starting a workflow execution.
	StartWorkflowOptions struct {
		ID                                     string
		TaskList                               string
		ExecutionStartToCloseTimeoutSeconds    int32
		DecisionTaskStartToCloseTimeoutSeconds int32
	}
)

// NewWorkflowClient creates an instance of workflow client that users can start a workflow
func NewWorkflowClient(service m.TChanWorkflowService, options *WorkflowClientOptions) WorkflowClient {
	var identity string
	if options == nil || options.Identity == "" {
		identity = getWorkerIdentity("")
	} else {
		identity = options.Identity
	}
	var metricScope tally.Scope
	if options != nil {
		metricScope = options.MetricsScope
	}
	return &workflowClient{workflowService: service, metricsScope: metricScope, identity: identity}
}
