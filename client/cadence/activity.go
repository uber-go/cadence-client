package cadence

import (
	"context"
	"time"

	"github.com/uber-go/cadence-client/.gen/go/shared"
	"go.uber.org/zap"
)

type (
	// ActivityType identifies a activity type.
	ActivityType struct {
		Name string
	}

	// ActivityInfo contains information about currently executing activity.
	ActivityInfo struct {
		TaskToken         []byte
		WorkflowExecution WorkflowExecution
		ActivityID        string
		ActivityType      ActivityType
	}
)

// RegisterActivity - register a activity function with the framework.
// A activity takes a context and input and returns a (result, error) or just error.
// Examples:
//	func sampleActivity(ctx context.Context, input []byte) (result []byte, err error)
//	func sampleActivity(ctx context.Context, arg1 int, arg2 string) (result *customerStruct, err error)
//	func sampleActivity(ctx context.Context) (err error)
//	func sampleActivity() (result string, err error)
//	func sampleActivity(arg1 bool) (result int, err error)
//	func sampleActivity(arg1 bool) (err error)
// Serialization of all primitive types, structures is supported ... except channels, functions, variadic, unsafe pointer.
// This method calls panic if activityFunc doesn't comply with the expected format.
func RegisterActivity(activityFunc interface{}) {
	thImpl := getHostEnvironment()
	err := thImpl.RegisterActivity(activityFunc)
	if err != nil {
		panic(err)
	}
}

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

// GetActivityLogger returns a logger that can be used in activity
func GetActivityLogger(ctx context.Context) *zap.Logger {
	env := getActivityEnv(ctx)
	return env.logger
}

// RecordActivityHeartbeat sends heartbeat for the currently executing activity
// If the activity is either cancelled (or) workflow/activity doesn't exist then we would cancel
// the context with error context.Canceled.
// 	TODO: we don't have a way to distinguish between the two cases when context is cancelled because
// 	context doesn't support overriding value of ctx.Error.
// 	TODO: Implement automatic heartbeating with cancellation through ctx.
// details - the details that you provided here can be seen in the worflow when it receives TimeoutError, you
//	can check error TimeOutType()/Details().
func RecordActivityHeartbeat(ctx context.Context, details ...interface{}) {
	data, err := getHostEnvironment().encodeArgs(details)
	if err != nil {
		panic(err)
	}
	env := getActivityEnv(ctx)
	err = env.serviceInvoker.Heartbeat(data)
	if err != nil {
		log := GetActivityLogger(ctx)
		log.Debug("RecordActivityHeartbeat With Error:", zap.Error(err))
	}
}

// ServiceInvoker abstracts calls to the Cadence service from an activity implementation.
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
	logger *zap.Logger,
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
		logger: logger,
	})
}

// ActivityOptions stores all activity-specific parameters that will
// be stored inside of a context.
type ActivityOptions struct {
	// TaskList that the activity needs to be scheduled on.
	// optional: The default task list with the same name as the workflow task list.
	TaskList *string

	// ScheduleToCloseTimeout - The end to end time out for the activity needed.
	// The zero value of this uses default value.
	// Optional: The default value is the sum of ScheduleToStartTimeout and StartToCloseTimeout
	ScheduleToCloseTimeout time.Duration

	// ScheduleToStartTimeout - The queue time out before the activity starts executed.
	// Mandatory: No default.
	ScheduleToStartTimeout time.Duration

	// StartToCloseTimeout - The time out from the start of execution to end of it.
	// Mandatory: No default.
	StartToCloseTimeout time.Duration

	// HeartbeatTimeout - The periodic timeout while the activity is in execution. This is
	// the max interval the server needs to hear at-least one ping from the activity.
	// Optional: Default zero, means no heart beating is needed.
	HeartbeatTimeout time.Duration

	// WaitForCancellation - Whether to wait for cancelled activity to be completed(
	// activity can be failed, completed, cancel accepted)
	// Optional: default nil(taken as false)
	WaitForCancellation *bool

	// ActivityID - Business level activity ID, this is not needed for most of the cases if you have
	// to specify this then talk to cadence team. This is something will be done in future.
	// Optional: default nil
	ActivityID *string
}

// WithActivityOptions adds all options to the context.
func WithActivityOptions(ctx Context, options ActivityOptions) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	eap := getActivityOptions(ctx1)
	if options.TaskList != nil && *options.TaskList != "" {
		eap.TaskListName = *options.TaskList
	}
	if options.ScheduleToCloseTimeout != 0 {
		eap.ScheduleToCloseTimeoutSeconds = int32(options.ScheduleToCloseTimeout.Seconds())
	}
	if options.StartToCloseTimeout != 0 {
		eap.StartToCloseTimeoutSeconds = int32(options.StartToCloseTimeout.Seconds())
	}
	if options.ScheduleToStartTimeout != 0 {
		eap.ScheduleToStartTimeoutSeconds = int32(options.ScheduleToStartTimeout.Seconds())
	}
	if options.HeartbeatTimeout != 0 {
		eap.HeartbeatTimeoutSeconds = int32(options.HeartbeatTimeout.Seconds())
	}
	if options.WaitForCancellation != nil {
		eap.WaitForCancellation = *options.WaitForCancellation
	}
	if options.ActivityID != nil {
		eap.ActivityID = options.ActivityID
	}
	return ctx1
}

// WithTaskList adds a task list to the context.
func WithTaskList(ctx Context, name string) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).TaskListName = name
	return ctx1
}

// WithScheduleToCloseTimeout adds a timeout to the context.
func WithScheduleToCloseTimeout(ctx Context, d time.Duration) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).ScheduleToCloseTimeoutSeconds = int32(d.Seconds())
	return ctx1
}

// WithScheduleToStartTimeout adds a timeout to the context.
func WithScheduleToStartTimeout(ctx Context, d time.Duration) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).ScheduleToStartTimeoutSeconds = int32(d.Seconds())
	return ctx1
}

// WithStartToCloseTimeout adds a timeout to the context.
func WithStartToCloseTimeout(ctx Context, d time.Duration) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).StartToCloseTimeoutSeconds = int32(d.Seconds())
	return ctx1
}

// WithHeartbeatTimeout adds a timeout to the context.
func WithHeartbeatTimeout(ctx Context, d time.Duration) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).HeartbeatTimeoutSeconds = int32(d.Seconds())
	return ctx1
}

// WithWaitForCancellation adds wait for the cacellation to the context.
func WithWaitForCancellation(ctx Context, wait bool) Context {
	ctx1 := setActivityParametersIfNotExist(ctx)
	getActivityOptions(ctx1).WaitForCancellation = wait
	return ctx1
}
