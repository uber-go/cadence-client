package cadence

import (
	"github.com/pborman/uuid"
	m "github.com/uber-go/cadence-client/.gen/go/cadence"
	s "github.com/uber-go/cadence-client/.gen/go/shared"
	"github.com/uber-go/cadence-client/common"
	"github.com/uber-go/cadence-client/common/backoff"
	"github.com/uber-go/cadence-client/common/metrics"
	"github.com/uber-go/tally"
)

type (
	// WorkflowClient is the client facing for starting a workflow.
	WorkflowClient struct {
		workflowExecution WorkflowExecution
		workflowService   m.TChanWorkflowService
		metricsScope      tally.Scope
		identity          string
	}

	// StartWorkflowOptions configuration parameters for starting a workflow
	StartWorkflowOptions struct {
		ID                                     string
		TaskList                               string
		ExecutionStartToCloseTimeoutSeconds    int32
		DecisionTaskStartToCloseTimeoutSeconds int32
		Identity                               string
	}
)

// NewWorkflowClient creates an instance of workflow client that users can start a workflow
func NewWorkflowClient(service m.TChanWorkflowService, metricsScope tally.Scope, identity string) *WorkflowClient {
	if identity == "" {
		identity = getWorkerIdentity("")
	}
	return &WorkflowClient{workflowService: service, metricsScope: metricsScope, identity: identity}
}

// StartWorkflowExecution starts a workflow execution
// The user can use this to start using a functor like.
// Either by
//     StartWorkflowExecution(options, "workflowTypeName", input)
//     or
//     StartWorkflowExecution(options, workflowExecuteFn, arg1, arg2, arg3)
func (wc *WorkflowClient) StartWorkflowExecution(
	options StartWorkflowOptions,
	workflowFunc interface{},
	args ...interface{},
) (*WorkflowExecution, error) {
	// Get an identity.
	identity := options.Identity
	if identity == "" {
		identity = getWorkerIdentity(options.TaskList)
	}
	workflowID := options.ID
	if workflowID == "" {
		workflowID = uuid.NewRandom().String()
	}

	// Validate type and its arguments.
	workflowType, input, err := getValidatedWorkerFunction(workflowFunc, args)
	if err != nil {
		return nil, err
	}

	startRequest := &s.StartWorkflowExecutionRequest{
		RequestId:    common.StringPtr(uuid.New()),
		WorkflowId:   common.StringPtr(workflowID),
		WorkflowType: workflowTypePtr(*workflowType),
		TaskList:     common.TaskListPtr(s.TaskList{Name: common.StringPtr(options.TaskList)}),
		Input:        input,
		ExecutionStartToCloseTimeoutSeconds: common.Int32Ptr(options.ExecutionStartToCloseTimeoutSeconds),
		TaskStartToCloseTimeoutSeconds:      common.Int32Ptr(options.DecisionTaskStartToCloseTimeoutSeconds),
		Identity:                            common.StringPtr(identity)}

	var response *s.StartWorkflowExecutionResponse

	// Start creating workflow request.
	err = backoff.Retry(
		func() error {
			ctx, cancel := common.NewTChannelContext(respondTaskServiceTimeOut, common.RetryDefaultOptions)
			defer cancel()

			var err1 error
			response, err1 = wc.workflowService.StartWorkflowExecution(ctx, startRequest)
			return err1
		}, serviceOperationRetryPolicy, isServiceTransientError)

	if err != nil {
		return nil, err
	}

	if wc.metricsScope != nil {
		wc.metricsScope.Counter(metrics.WorkflowsStartTotalCounter).Inc(1)
	}

	executionInfo := &WorkflowExecution{
		ID:    options.ID,
		RunID: response.GetRunId()}
	return executionInfo, nil
}

// GetHistory gets history of a particular workflow.
func (wc *WorkflowClient) GetHistory(workflowID string, runID string) (*s.History, error) {
	request := &s.GetWorkflowExecutionHistoryRequest{
		Execution: &s.WorkflowExecution{
			WorkflowId: common.StringPtr(workflowID),
			RunId:      common.StringPtr(runID),
		},
	}

	var response *s.GetWorkflowExecutionHistoryResponse
	err := backoff.Retry(
		func() error {
			var err1 error
			ctx, cancel := common.NewTChannelContext(respondTaskServiceTimeOut, common.RetryDefaultOptions)
			defer cancel()
			response, err1 = wc.workflowService.GetWorkflowExecutionHistory(ctx, request)
			return err1
		}, serviceOperationRetryPolicy, isServiceTransientError)
	return response.GetHistory(), err
}

// CompleteActivity reports activity completed. Activity Execute method can return cadence.ActivityResultPendingError to
// indicate the activity is not completed when it's Execute method returns. In that case, this CompleteActivity() method
// should be called when that activity is completed with the actual result and error. If err is nil, activity task
// completed event will be reported; if err is CanceledError, activity task cancelled event will be reported; otherwise,
// activity task failed event will be reported.
func (wc *WorkflowClient) CompleteActivity(taskToken, result []byte, err error) error {
	request := convertActivityResultToRespondRequest(wc.identity, taskToken, result, err)
	return reportActivityComplete(wc.workflowService, request)
}

// RecordActivityHeartbeat records heartbeat for an activity.
func (wc *WorkflowClient) RecordActivityHeartbeat(taskToken, details []byte) error {
	return recordActivityHeartbeat(wc.workflowService, wc.identity, taskToken, details)
}
