package cadence

// All code in this file is private to the package.

import (
	"golang.org/x/net/context"
	"github.com/uber-go/cadence-client/common"
)

// Assert that structs do indeed implement the interfaces
var _ ActivityOptions = (*activityOptions)(nil)

type (
	activityInfo struct {
		activityID string
	}

	// executeActivityParameters configuration parameters for scheduling an activity
	executeActivityParameters struct {
		ActivityID                    *string // Users can choose IDs but our framework makes it optional to decrease the crust.
		ActivityType                  ActivityType
		TaskListName                  string
		Input                         []byte
		ScheduleToCloseTimeoutSeconds int32
		ScheduleToStartTimeoutSeconds int32
		StartToCloseTimeoutSeconds    int32
		HeartbeatTimeoutSeconds       int32
		WaitForCancellation           bool
	}

	// asyncActivityClient for requesting activity execution
	asyncActivityClient interface {
		// The ExecuteActivity schedules an activity with a callback handler.
		// If the activity failed to complete the callback error would indicate the failure
		// and it can be one of ActivityTaskFailedError, ActivityTaskTimeoutError, ActivityTaskCanceledError
		ExecuteActivity(parameters executeActivityParameters, callback resultHandler) *activityInfo

		// This only initiates cancel request for activity. if the activity is configured to not waitForCancellation then
		// it would invoke the callback handler immediately with error code ActivityTaskCanceledError.
		// If the activity is not running(either scheduled or started) then it is a no-operation.
		RequestCancelActivity(activityID string)
	}

	activityEnvironment struct {
		taskToken         []byte
		workflowExecution WorkflowExecution
		activityID        string
		activityType      ActivityType
		serviceInvoker    ServiceInvoker
	}
)

const activityEnvContextKey = "activityEnv"

func getActivityEnv(ctx context.Context) *activityEnvironment {
	env := ctx.Value(activityEnvContextKey)
	if env == nil {
		panic("getActivityEnv: Not an activity context")
	}
	return env.(*activityEnvironment)
}

const activityOptionsContextKey = "activityOptions"

func getActivityOptions(ctx Context) *executeActivityParameters {
	eap := ctx.Value(activityOptionsContextKey)
	if eap == nil {
		return nil
	}
	return eap.(*executeActivityParameters)
}

func setActivityParametersIfNotExist(ctx Context) Context {
	if valCtx := getActivityOptions(ctx); valCtx == nil {
		return WithValue(ctx, activityOptionsContextKey, &executeActivityParameters{})
	}
	return ctx
}

// activityOptions stores all activity-specific parameters that will
// be stored inside of a context.
type activityOptions struct {
	activityID                    *string
	taskListName                  string
	scheduleToCloseTimeoutSeconds int32
	scheduleToStartTimeoutSeconds int32
	startToCloseTimeoutSeconds    int32
	heartbeatTimeoutSeconds       int32
	waitForCancellation           bool
}

// WithTaskList sets the task list name for this Context.
func (ab *activityOptions) WithTaskList(name string) ActivityOptions {
	ab.taskListName = name
	return ab
}

// WithScheduleToCloseTimeout sets timeout for this Context.
func (ab *activityOptions) WithScheduleToCloseTimeout(timeout int32) ActivityOptions {
	ab.scheduleToCloseTimeoutSeconds = timeout
	return ab
}

// WithScheduleToStartTimeout sets timeout for this Context.
func (ab *activityOptions) WithScheduleToStartTimeout(timeout int32) ActivityOptions {
	ab.scheduleToStartTimeoutSeconds = timeout
	return ab
}

// WithStartToCloseTimeout sets timeout for this Context.
func (ab *activityOptions) WithStartToCloseTimeout(timeout int32) ActivityOptions {
	ab.startToCloseTimeoutSeconds = timeout
	return ab
}

// WithHeartbeatTimeout sets timeout for this Context.
func (ab *activityOptions) WithHeartbeatTimeout(timeout int32) ActivityOptions {
	ab.heartbeatTimeoutSeconds = timeout
	return ab
}

// WithWaitForCancellation sets timeout for this Context.
func (ab *activityOptions) WithWaitForCancellation(wait bool) ActivityOptions {
	ab.waitForCancellation = wait
	return ab
}

// WithActivityID sets the activity task list ID for this Context.
// NOTE: This is now used for internal testing(mocks).
func (ab *activityOptions) WithActivityID(activityID string) ActivityOptions {
	ab.activityID = common.StringPtr(activityID)
	return ab
}

