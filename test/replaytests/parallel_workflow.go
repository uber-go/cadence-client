package replaytests

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

/**
 * This sample workflow executes multiple branches in parallel using workflow.Go() method.
 */

// sampleParallelWorkflow workflow decider
func sampleParallelWorkflow(ctx workflow.Context) error {
	waitChannel := workflow.NewChannel(ctx)

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	workflow.Go(ctx, func(ctx workflow.Context) {
		err := workflow.ExecuteActivity(ctx, sampleActivity, "branch1.1").Get(ctx, nil)
		if err != nil {
			logger.Error("Activity failed", zap.Error(err))
			waitChannel.Send(ctx, err.Error())
			return
		}
		err = workflow.ExecuteActivity(ctx, sampleActivity, "branch1.2").Get(ctx, nil)
		if err != nil {
			logger.Error("Activity failed", zap.Error(err))
			waitChannel.Send(ctx, err.Error())
			return
		}
		waitChannel.Send(ctx, "")
	})

	workflow.Go(ctx, func(ctx workflow.Context) {
		err := workflow.ExecuteActivity(ctx, sampleActivity, "branch2").Get(ctx, nil)
		if err != nil {
			logger.Error("Activity failed", zap.Error(err))
			waitChannel.Send(ctx, err.Error())
			return
		}
		waitChannel.Send(ctx, "")
	})

	// wait for both of the coroutinue to complete.
	var errMsg string
	for i := 0; i != 2; i++ {
		waitChannel.Receive(ctx, &errMsg)
		if errMsg != "" {
			err := errors.New(errMsg)
			logger.Error("Coroutine failed", zap.Error(err))
			return err
		}
	}

	logger.Info("Workflow completed.")
	return nil
}

func sampleActivity(input string) (string, error) {
	name := "sampleActivity"
	fmt.Printf("Run %s with input %v \n", name, input)
	return "Result_" + name, nil
}
