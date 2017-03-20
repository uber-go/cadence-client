package cadence

import (
	"context"

	"github.com/uber-go/cadence-client/.gen/go/shared"
	"github.com/uber-go/cadence-client/common"
)

type (
	// ActivityType identifies a activity type.
	ActivityType struct {
		Name string
	}

	// Activity is an interface of an activity implementation.
	Activity interface {
		Execute(ctx context.Context, input []byte) ([]byte, error)
		ActivityType() ActivityType
	}

	// ActivityInfo contains information about currently executing activity.
	ActivityInfo struct {
		TaskToken         []byte
		WorkflowExecution WorkflowExecution
		ActivityID        string
		ActivityType      ActivityType
	}
)

// GetActivityInfo returns information about currently executing activity.
func GetActivityInfo(ctx context.Context) ActivityInfo {
	env := getActivityEnv(ctx)
	return ActivityInfo{
		ActivityID:        env.activityID,
		ActivityType:      env.activityType,
		TaskToken:         env.taskToken,
		WorkflowExecution: env.workflowExecution,
	}
}

// RecordActivityHeartbeat sends heartbeat for the currently executing activity
// TODO: Implement automatic heartbeating with cancellation through ctx.
func RecordActivityHeartbeat(ctx context.Context, details []byte) error {
	env := getActivityEnv(ctx)
	return env.serviceInvoker.Heartbeat(details)
}

// ServiceInvoker abstracts calls to the Cadence service from an Activity implementation.
// Implement to unit test activities.
type ServiceInvoker interface {
	// Returns ActivityTaskCanceledError if activity is cancelled
	Heartbeat(details []byte) error
}

// WithActivityTask adds activity specific information into context.
// Use this method to unit test activity implementations that use context extractor methodshared.
func WithActivityTask(
	ctx context.Context,
	task *shared.PollForActivityTaskResponse,
	invoker ServiceInvoker,
) context.Context {
	// TODO: Add activity start to close timeout to activity task and use it as the deadline
	return context.WithValue(ctx, activityEnvContextKey, &activityEnvironment{
		taskToken:      task.TaskToken,
		serviceInvoker: invoker,
		activityType:   ActivityType{Name: *task.ActivityType.Name},
		activityID:     *task.ActivityId,
		workflowExecution: WorkflowExecution{
			RunID: *task.WorkflowExecution.RunId,
			ID:    *task.WorkflowExecution.WorkflowId},
	})
}

const activityOptionsContextKey = "activityOptions"

func getActivityOptions(ctx Context) *ExecuteActivityParameters {
	eap := ctx.Value(activityOptionsContextKey)
	if eap == nil {
		return nil
	}
	return eap.(*ExecuteActivityParameters)
}

func setActivityParametersIfNotExist(ctx Context) Context {
	if valCtx := getActivityOptions(ctx); valCtx == nil {
		return WithValue(ctx, activityOptionsContextKey, &ExecuteActivityParameters{})
	}
	return ctx
}

// ActivityOptionsBuilder stores all activity-specific parameters that will
// be stored inside of a context.
type ActivityOptionsBuilder struct {
	parentCtx                     Context
	activityID                    *string
	taskListName                  string
	scheduleToCloseTimeoutSeconds int32
	scheduleToStartTimeoutSeconds int32
	startToCloseTimeoutSeconds    int32
	heartbeatTimeoutSeconds       int32
	waitForCancellation           bool
}

// NewActivityOptionsBuilder returns a builder that can be used to create a Context.
func NewActivityOptionsBuilder(ctx Context) *ActivityOptionsBuilder {
	return &ActivityOptionsBuilder{parentCtx: ctx}
}

// SetTaskList sets the task list name for this Context.
func (ab *ActivityOptionsBuilder) SetTaskList(name string) *ActivityOptionsBuilder {
	ab.taskListName = name
	return ab
}

// SetScheduleToCloseTimeout sets timeout for this Context.
func (ab *ActivityOptionsBuilder) SetScheduleToCloseTimeout(timeout int32) *ActivityOptionsBuilder {
	ab.scheduleToCloseTimeoutSeconds = timeout
	return ab
}

// SetScheduleToStartTimeout sets timeout for this Context.
func (ab *ActivityOptionsBuilder) SetScheduleToStartTimeout(timeout int32) *ActivityOptionsBuilder {
	ab.scheduleToStartTimeoutSeconds = timeout
	return ab
}

// SetStartToCloseTimeout sets timeout for this Context.
func (ab *ActivityOptionsBuilder) SetStartToCloseTimeout(timeout int32) *ActivityOptionsBuilder {
	ab.startToCloseTimeoutSeconds = timeout
	return ab
}

// SetHeartbeatTimeout sets timeout for this Context.
func (ab *ActivityOptionsBuilder) SetHeartbeatTimeout(timeout int32) *ActivityOptionsBuilder {
	ab.heartbeatTimeoutSeconds = timeout
	return ab
}

// SetWaitForCancellation sets timeout for this Context.
func (ab *ActivityOptionsBuilder) SetWaitForCancellation(wait bool) *ActivityOptionsBuilder {
	ab.waitForCancellation = wait
	return ab
}

// activityID sets the activity task list ID for this Context.
func (ab *ActivityOptionsBuilder) SetActivityID(activityID string) *ActivityOptionsBuilder {
	ab.activityID = common.StringPtr(activityID)
	return ab
}

// Build returns a Context that can be used to make calls.
func (ab *ActivityOptionsBuilder) Build() Context {
	ctx := setActivityParametersIfNotExist(ab.parentCtx)
	getActivityOptions(ctx).TaskListName = ab.taskListName
	getActivityOptions(ctx).ScheduleToCloseTimeoutSeconds = ab.scheduleToCloseTimeoutSeconds
	getActivityOptions(ctx).StartToCloseTimeoutSeconds = ab.startToCloseTimeoutSeconds
	getActivityOptions(ctx).ScheduleToStartTimeoutSeconds = ab.scheduleToStartTimeoutSeconds
	getActivityOptions(ctx).HeartbeatTimeoutSeconds = ab.heartbeatTimeoutSeconds
	getActivityOptions(ctx).WaitForCancellation = ab.waitForCancellation
	getActivityOptions(ctx).ActivityID = ab.activityID
	return ctx
}

// WithTaskList adds a task list to the context.
func WithTaskList(ctx Context, name string) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).TaskListName = name
	return ctx1
}

// WithScheduleToCloseTimeout adds a timeout to the context.
func WithScheduleToCloseTimeout(ctx Context, timeout int32) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).ScheduleToCloseTimeoutSeconds = timeout
	return ctx1
}

// WithScheduleToStartTimeout adds a timeout to the context.
func WithScheduleToStartTimeout(ctx Context, timeout int32) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).ScheduleToStartTimeoutSeconds = timeout
	return ctx1
}

// WithStartToCloseTimeout adds a timeout to the context.
func WithStartToCloseTimeout(ctx Context, timeout int32) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).StartToCloseTimeoutSeconds = timeout
	return ctx1
}

// WithHeartbeatTimeout adds a timeout to the context.
func WithHeartbeatTimeout(ctx Context, timeout int32) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).HeartbeatTimeoutSeconds = timeout
	return ctx1
}

// WithWaitForCancellation adds wait for the cacellation to the context.
func WithWaitForCancellation(ctx Context, wait bool) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).WaitForCancellation = wait
	return ctx1
}
