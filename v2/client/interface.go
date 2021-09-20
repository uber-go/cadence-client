// Copyright (c) 2021 Uber Technologies, Inc.
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

package client

import (
	"context"
	"time"

	"go.uber.org/cadence/encoded"
	cadence "go.uber.org/cadence/v2"
)

type (
	// Client is the top level client entity for interactive with Cadence server.
	Client interface {
		// Domains returns an interface for interacting with Cadence domains.
		Domains() Domains

		// Activities returns an interface for interacting with activities (via opaque task tokens).
		Activities() Activities

		// SearchAttributes returns an interface for interacting with cluster search attributes.
		SearchAttributes() SearchAttributes
	}

	// Domains is an interface for interacting with Cadence domains.
	Domains interface {
		// Register registers a domain within cadence server.
		// Name and replication settings are required. Optional fields can be set via options.
		Register(ctx context.Context, name string, replication DomainReplicationConfig, opts ...DomainRegisterOption) (Domain, error)

		// List retrieves all domains that are registered within Cadence server.
		List(ctx context.Context, page Page, opts ...DomainListOption) ([]ListedDomain, Page, error)

		// Get selects a domain by a given name for further operations.
		Get(name string) Domain
	}

	// ListedDomain is a domain that was retrievied via Domains.List function.
	// It also contains info for the domain that was included in the List API call.
	ListedDomain interface {
		Domain
		Info() cadence.DomainInfo
	}

	DescribedDomain interface {
		Info() cadence.DomainInfo
	}

	// Domain is a Cadence way to group workflows of some application or owner.
	Domain interface {
		// Name return the name of this domain
		Name() string

		// Describe returns information about the domain.
		Describe(ctx context.Context, opts ...DomainDescribeOption) (DescribedDomain, error)

		// Update updates one or more domain fields. Use options to specify which fields to update.
		Update(ctx context.Context, opts ...DomainUpdateOption) error

		// Failover will failover the domain to another specified cluster.
		Failover(ctx context.Context, cluster string, opts ...DomainFailoverOption) error

		// Deprecate will deprecate the domain.
		Deprecate(ctx context.Context, opts ...DomainDeprecateOption) error

		// BadBinaries returns an interface for interacting with bad binaries.
		BadBinaries() BadBinaries
	}

	// BadBinaries is an interface for interacting with bad binaries.
	// A bad binary can be used to indicate a bad deployment, so that workers with it will
	// stop making progress and workflows could be reset a state before the deployment.
	BadBinaries interface {
		// List returns a list of all bad binaries registered for the domain.
		List(ctx context.Context, opts ...BadBinaryListOption) ([]cadence.BadBinary, error)

		// Add will include new bad binary with the given checksum.
		Add(ctx context.Context, checksum string, reason string, opts ...BadBinaryAddOption) error

		// Delete will delete an existing bad binary by the given checksum.
		Delete(ctx context.Context, checksum string, opts ...BadBinaryDeleteOption) error
	}

	// Workflows is an interface for interacting with Cadence workflows for the selected domain.
	Workflows interface {
		// Starts starts a new workflow execution
		// The user can use this to start using a function or workflow type name.
		Start(ctx context.Context, wfFunc interface{}, args []interface{}, taskList string, timeout time.Duration, opts ...WorkflowStartOption) (Workflow, error)

		// Count gets a number of workflow executions based on query.
		Count(ctx context.Context, query Query, opts ...WorkflowCountOption) (int64, error)

		// List returns a list of workflow executions based on query.
		List(ctx context.Context, query Query, page Page, opts ...WorkflowListOption) ([]ListedWorkflow, *Page, error)

		// Get will select a concrete workflow run by a given workflow and run ID for further operations.
		Get(workflowID, runID string) Workflow

		// GetCurrent will select the latest run of the workflow given by its workflowID for further operations.
		GetCurrent(workflowID string) Workflow
	}

	ListedWorkflow interface {
		Workflow
		Info() cadence.WorkflowInfo
	}

	DescribedWorkflow interface {
		Info() cadence.WorkflowInfo

		// PendingActivities returns information about pending activities for the workflow (if any)
		PendingActivities() []cadence.PendingActivityInfo
		// PendingChildren returns information pending child workflows (if any)
		PendingChildren() []cadence.PendingChildWorkflowInfo
		// PendingDecision return information about pending decision
		PendingDecision() *cadence.PendingDecisionInfo
	}

	Workflow interface {
		// Domain returns Domain of this workflow
		Domain() Domain
		// WorkflowID returns ID of this workflow
		WorkflowID() string
		// RunID return Run ID of this workflow
		RunID() string

		// Signal sends a signals to a workflow in execution.
		Signal(ctx context.Context, signalName string, args []interface{}, opts ...WorkflowSignalOption) error

		// Query queries a workflow's execution and returns the query result synchronously.
		// The queryType specifies the type of query you want to run.
		// By default, cadence supports "__stack_trace" as a standard query type, which will return string value
		// representing the call stack of the target workflow. The target workflow could also setup different query handler
		// to handle custom query types.
		// See comments at workflow.SetQueryHandler(ctx Context, queryType string, handler interface{}) for more details
		// on how to setup query handler within the target workflow.
		Query(ctx context.Context, queryType string, args []interface{}, opts ...WorkflowQueryOption) (encoded.Value, error)

		// Describe returns information about the workflow execution.
		Describe(ctx context.Context, opts ...WorkflowDescribeOption) (DescribedWorkflow, error)

		// Cancel cancels a workflow in execution.
		// This is different from Terminate, as it would cancel workflow context giving opportunity for it to do the cleanup and exit gracefully.
		Cancel(ctx context.Context, opts ...WorkflowCancelOption) error

		// Terminate terminates a workflow execution.
		// This is different from Cancel, as it would force workflow termination from server side.
		Terminate(ctx context.Context, reason string, opts ...WorkflowTerminateOption) error

		// Reset resets a workflow execution to the given point and returns a new execution.
		Reset(ctx context.Context, reason string, point WorkflowResetPoint, opts ...WorkflowResetOption) (Workflow, error)

		// GetResult will return workflow result if workflow execution is a success,
		// or return corresponding error. This is a blocking API.
		GetResult(ctx context.Context, opts ...WorkflowGetResultOption) (encoded.Value, error)

		// Activities returns an interface for interactive activities of this workflow.
		Activities() WorkflowActivities
	}

	// Activities is an interface for interacting with any activities.
	Activities interface {
		// Get selects an activity with the given task token.
		// This is an opaque token that can be obtained within activity via GetActivityInfo(ctx).TaskToken function.
		// It is used to interact with activity externally - to complete it or record is progress and heartbeat.
		Get(taskToken []byte) Activity
	}

	// WorkflowActivities is an interface for interacting with selected workflow activities.
	WorkflowActivities interface {
		Activities

		// GetByID will select an activity for the selected workflow by the given activityID.
		GetByID(activityID string) Activity
	}

	Activity interface {
		// Complete reports activity completed.
		// Activity Execute method can return activity.ErrResultPending to
		// indicate the activity is not completed when it's Execute method returns. In that case, this Complete method
		// should be called when that activity is completed with the actual result and error. If err is nil, activity task
		// completed event will be reported; if err is CanceledError, activity task cancelled event will be reported; otherwise,
		// activity task failed event will be reported.
		Complete(ctx context.Context, result interface{}, err error, opts ...ActivityCompleteOption) error

		// RecordHeartbeat records heartbeat for an activity.
		// details - is the progress you want to record along with heart beat for this activity.
		RecordHeartbeat(ctx context.Context, details interface{}, opts ...ActivityHeartbeatOption) error
	}

	// TaskLists is an interface for interacting with task lists.
	TaskLists interface {
		// Get selects a task list by the given name and type.
		Get(name string, taskListType cadence.TaskListType) TaskList
	}

	// TaskList is a light-weight Cadence queue that is used to deliver tasks to Cadence worker.
	TaskList interface {
		// Describe returns information about the tasklist, right now this API returns the
		// pollers which polled this tasklist in last few minutes.
		Describe(ctx context.Context) (cadence.TaskListInfo, error)
	}

	// SearchAttributes is an interface for interacting with search attributes.
	// The search attributes can be used in query of workflow List/Count APIs.
	// Adding new search attributes requires cadence server to update dynamic config ValidSearchAttributes.
	SearchAttributes interface {
		// List returns valid search attributes keys and value types.
		List(ctx context.Context) ([]cadence.SearchAttribute, error)
	}
)
