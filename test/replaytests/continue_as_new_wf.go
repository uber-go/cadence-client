package replaytests

import (
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// ContinueAsNewWorkflow is a sample Cadence workflows that can receive a signal
func ContinueAsNewWorkflow(ctx workflow.Context) error {
	selector := workflow.NewSelector(ctx)
	var signalResult string
	signalName := "helloWorldSignal"
	for {
		signalChan := workflow.GetSignalChannel(ctx, signalName)
		selector.AddReceive(signalChan, func(c workflow.Channel, more bool) {
			c.Receive(ctx, &signalResult)
			workflow.GetLogger(ctx).Info("Received age signalResult from signal!", zap.String("signal", signalName), zap.String("value", signalResult))
		})
		workflow.GetLogger(ctx).Info("Waiting for signal on channel.. " + signalName)
		// Wait for signal
		selector.Select(ctx)
		if signalResult == "kill" {
			return nil
		}
	}
}
