package replaytests

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type replayerSampleMessage struct {
	Message string
}

// replayerHelloWorldWorkflow is a sample workflow that runs replayerHelloWorldActivity.
// In the previous version it was running the activity twice, sequentially.
// History of a past execution is in sequential.json
// Corresponding unit test covers the scenario that new workflow's history records is subset of previous version's history.
//
//	v1: wf started -> call activity -> call activity -> wf complete
//	v2:  wf started -> call activity -> wf complete
//
// The v2 clearly has determinism issues and should be considered as non-determism error for replay tests.
func replayerHelloWorldWorkflow(ctx workflow.Context, inputMsg *replayerSampleMessage) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	logger := workflow.GetLogger(ctx)
	logger.Info("executing replayerHelloWorldWorkflow")
	ctx = workflow.WithActivityOptions(ctx, ao)

	count := 1
	for i := 0; i < count; i++ {
		var greeting string
		err := workflow.ExecuteActivity(ctx, replayerHelloWorldActivity, inputMsg).Get(ctx, &greeting)
		if err != nil {
			logger.Error("replayerHelloWorldActivity is broken ", zap.Error(err))
			return err
		}

		logger.Sugar().Infof("replayerHelloWorldWorkflow is greeting to you %dth time -> ", i, greeting)
	}

	return nil
}

// replayerHelloWorldActivity takes a sample message input and return a greeting to the caller.
func replayerHelloWorldActivity(ctx context.Context, inputMsg *replayerSampleMessage) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("executing replayerHelloWorldActivity")

	return fmt.Sprintf("Hello, %s!", inputMsg.Message), nil
}
