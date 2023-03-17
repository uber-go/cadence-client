package replaytests

import (
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"time"
)

// greetingsWorkflowActivity demonstrates the case in replayer where if an activity name is changed in a workflow or if a new activity is introdced in a workflow then the shdower fails.
// There is also no other way to register the activity except for changing the json file so far.
// CDNC-2267 addresses a fix for this.
func greetingsWorkflowActivity(ctx workflow.Context) error {
	// Get Greeting.
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	var greetResult int64
	err := workflow.ExecuteActivity(ctx, getGreetingActivitytest).Get(ctx, &greetResult)
	if err != nil {
		logger.Error("Get greeting failed.", zap.Error(err))
		return err
	}

	// Get Name.
	var nameResult string
	err = workflow.ExecuteActivity(ctx, getNameActivity).Get(ctx, &nameResult)
	if err != nil {
		logger.Error("Get name failed.", zap.Error(err))
		return err
	}

	// Say Greeting.
	var sayResult string
	err = workflow.ExecuteActivity(ctx, sayGreetingActivity, greetResult, nameResult).Get(ctx, &sayResult)
	if err != nil {
		logger.Error("Marshalling failed with error.", zap.Error(err))
		return err
	}

	logger.Info("Workflow completed.", zap.String("Result", sayResult))
	return nil
}

// Get Greeting Activity.
func getGreetingActivitytest() (string, error) {
	return "Hello", nil
}
