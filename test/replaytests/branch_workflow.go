package replaytests

import (
	"fmt"
	"go.uber.org/cadence/workflow"
	"time"
)

/**
 * This sample workflow executes multiple branches in parallel. The number of branches is controlled by passed in parameter.
 */

// sampleBranchWorkflow workflow decider
func sampleBranchWorkflow(ctx workflow.Context) error {
	var futures []workflow.Future
	// starts activities in parallel
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for i := 1; i <= 3; i++ {
		activityInput := fmt.Sprintf("branch %d of 3", i)
		future := workflow.ExecuteActivity(ctx, sampleActivity, activityInput)
		futures = append(futures, future)
	}

	// wait until all futures are done
	for _, future := range futures {
		if err := future.Get(ctx, nil); err != nil {
			return err
		}
	}

	workflow.GetLogger(ctx).Info("Workflow completed.")

	return nil
}

// SampleBranchWorkflow2 run a workflow with different number of branch.
//If the number of expected branches is changed the replayer should catch it as a non deterministic error.
func sampleBranchWorkflow2(ctx workflow.Context) error {
	var futures []workflow.Future
	// starts activities in parallel
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for i := 1; i <= 4; i++ {
		activityInput := fmt.Sprintf("branch %d of 4", i)
		future := workflow.ExecuteActivity(ctx, sampleActivity, activityInput)
		futures = append(futures, future)
	}

	// wait until all futures are done
	for _, future := range futures {
		if err := future.Get(ctx, nil); err != nil {
			return err
		}
	}

	workflow.GetLogger(ctx).Info("Workflow completed.")

	return nil
}

func sampleActivity(input string) (string, error) {
	name := "sampleActivity"
	fmt.Printf("Run %s with input %v \n", name, input)

	return "Result_" + name, nil
}
