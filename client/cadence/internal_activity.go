package cadence

// All code in this file is private to the package.

import (
	"errors"
	"time"

	"golang.org/x/net/context"

	"fmt"
	"github.com/uber-go/cadence-client/common"
	"reflect"
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

func getValidatedActivityOptions(ctx Context) (*executeActivityParameters, error) {
	p := getActivityOptions(ctx)
	if p == nil {
		// We need task list as a compulsory parameter. This can be removed after registration
		return nil, errActivityParamsBadRequest
	}
	if p.ScheduleToStartTimeoutSeconds <= 0 {
		return nil, errors.New("missing or negative ScheduleToStartTimeoutSeconds")
	}
	if p.ScheduleToCloseTimeoutSeconds <= 0 {
		return nil, errors.New("missing or negative ScheduleToCloseTimeoutSeconds")
	}
	if p.StartToCloseTimeoutSeconds <= 0 {
		return nil, errors.New("missing or negative StartToCloseTimeoutSeconds")
	}
	return p, nil
}

func marshalFunctionArgs(fnName string, args ...interface{}) ([]byte, error) {
	s := fnSignature{FnName: fnName, Args: args}
	input, err := getHostEnvironment().Encoder().Marshal(s)
	if err != nil {
		return nil, err
	}
	return input, nil
}

func validateFunctionArgs(f interface{}, args []interface{}) error {
	fType := reflect.TypeOf(f)
	if fType.Kind() != reflect.Func {
		return fmt.Errorf("Provided type: %v is not a function type", f)
	}
	fnName := getFunctionName(f)

	fnArgIndex := 0
	// Skip Context function argument.
	if fType.NumIn() > 0 && (isWorkflowContext(fType.In(0)) || isActivityContext(fType.In(0))) {
		fnArgIndex++
	}

	// Validate provided args match with function order match.
	if fType.NumIn()-fnArgIndex != len(args) {
		return fmt.Errorf(
			"expected %d args for function: %v but found %v",
			fType.NumIn()-fnArgIndex, fnName, len(args))
	}

	for i := 0; fnArgIndex < fType.NumIn(); fnArgIndex, i = fnArgIndex+1, i+1 {
		fnArgType := fType.In(fnArgIndex)
		argType := reflect.TypeOf(args[i])
		if !argType.AssignableTo(fnArgType) {
			return fmt.Errorf(
				"cannot assign function argument: %d from type: %s to type: %s",
				fnArgIndex+1, argType, fnArgType,
			)
		}
	}

	return nil
}

func getValidatedActivityFunction(f interface{}, args []interface{}) (ActivityType, []byte, error) {
	fnName := ""
	fType := reflect.TypeOf(f)
	switch fType.Kind() {
	case reflect.String:
		fnName = reflect.ValueOf(f).String()

	case reflect.Func:
		if err := validateFunctionArgs(f, args); err != nil {
			return ActivityType{}, nil, err
		}
		fnName = getFunctionName(f)

	default:
		return ActivityType{}, nil, fmt.Errorf(
			"Invalid type 'f' parameter provided, it can be either activity function (or) name of the activity: %v", f)
	}

	input, err := marshalFunctionArgs(fnName, args)
	if err != nil {
		return ActivityType{}, nil, err
	}
	return ActivityType{Name: fnName}, input, nil
}

func isActivityContext(inType reflect.Type) bool {
	contextElem := reflect.TypeOf((*context.Context)(nil)).Elem()
	return inType.Implements(contextElem)
}

type fnReturnSignature struct {
	Ret interface{}
}

func validateFunctionAndGetResults(f interface{}, values []reflect.Value) (result []byte, err error) {
	fnName := getFunctionName(f)
	resultSize := len(values)

	if resultSize < 1 || resultSize > 2 {
		return nil, fmt.Errorf(
			"The function: %v execution returned %d results where as it is expecting to return either error (or) result, error",
			fnName, resultSize)
	}
	result = nil
	err = nil

	// Parse result
	if resultSize > 1 {
		r := values[0].Interface()
		if err := getHostEnvironment().Encoder().Register(r); err != nil {
			return nil, err
		}
		fr := fnReturnSignature{Ret: r}
		result, err = getHostEnvironment().Encoder().Marshal(fr)
		if err != nil {
			return nil, err
		}
	}

	// Parse error.
	if v, ok := values[resultSize-1].Interface().(error); ok {
		err = v
	}
	return result, err
}

func deSerializeFunctionResult(f interface{}, result []byte) (interface{}, error) {
	fnType := reflect.TypeOf(f)

	switch fnType.Kind() {
	case reflect.Func:
		// We already validated that it either have (result, error) (or) just error.
		if fnType.NumOut() <= 1 {
			return nil, nil
		} else if fnType.NumOut() == 2 {
			var fr fnReturnSignature
			if err := getHostEnvironment().Encoder().Unmarshal(result, &fr); err != nil {
				return nil, err
			}
			return fr.Ret, nil
		}
	}
	// For everything we return result.
	return result, nil
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
	taskListName                  *string
	scheduleToCloseTimeoutSeconds *int32
	scheduleToStartTimeoutSeconds *int32
	startToCloseTimeoutSeconds    *int32
	heartbeatTimeoutSeconds       *int32
	waitForCancellation           *bool
}

// WithTaskList sets the task list name for this Context.
func (ab *activityOptions) WithTaskList(name string) ActivityOptions {
	ab.taskListName = common.StringPtr(name)
	return ab
}

// WithScheduleToCloseTimeout sets timeout for this Context.
func (ab *activityOptions) WithScheduleToCloseTimeout(d time.Duration) ActivityOptions {
	ab.scheduleToCloseTimeoutSeconds = common.Int32Ptr(int32(d.Seconds()))
	return ab
}

// WithScheduleToStartTimeout sets timeout for this Context.
func (ab *activityOptions) WithScheduleToStartTimeout(d time.Duration) ActivityOptions {
	ab.scheduleToStartTimeoutSeconds = common.Int32Ptr(int32(d.Seconds()))
	return ab
}

// WithStartToCloseTimeout sets timeout for this Context.
func (ab *activityOptions) WithStartToCloseTimeout(d time.Duration) ActivityOptions {
	ab.startToCloseTimeoutSeconds = common.Int32Ptr(int32(d.Seconds()))
	return ab
}

// WithHeartbeatTimeout sets timeout for this Context.
func (ab *activityOptions) WithHeartbeatTimeout(d time.Duration) ActivityOptions {
	ab.heartbeatTimeoutSeconds = common.Int32Ptr(int32(d.Seconds()))
	return ab
}

// WithWaitForCancellation sets timeout for this Context.
func (ab *activityOptions) WithWaitForCancellation(wait bool) ActivityOptions {
	ab.waitForCancellation = &wait
	return ab
}

// WithActivityID sets the activity task list ID for this Context.
// NOTE: We don't expose configuring Activity ID to the user, This is something will be done in future
// so they have end to end scenario of how to use this ID to complete and fail an activity(business use case).
func (ab *activityOptions) WithActivityID(activityID string) ActivityOptions {
	ab.activityID = common.StringPtr(activityID)
	return ab
}
