// Copyright (c) 2017 Uber Technologies, Inc.
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

package replaytests

import (
	"context"
	"math/rand"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// ApplicationName is the task list for this sample
const ApplicationName = "timerGroup"

// sampleTimerWorkflow workflow decider
func sampleTimerWorkflow(ctx workflow.Context, processingTimeThreshold time.Duration) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	childCtx, cancelHandler := workflow.WithCancel(ctx)
	selector := workflow.NewSelector(ctx)

	// In this sample case, we want to demo a use case where the workflow starts a long running order processing operation
	// and in the case that the processing takes too long, we want to send out a notification email to user about the delay,
	// but we won't cancel the operation. If the operation finishes before the timer fires, then we want to cancel the timer.

	var processingDone bool
	f := workflow.ExecuteActivity(ctx, orderProcessingActivity)
	selector.AddFuture(f, func(f workflow.Future) {
		processingDone = true
		// cancel timerFuture
		cancelHandler()
	})

	// use timer future to send notification email if processing takes too long
	timerFuture := workflow.NewTimer(childCtx, processingTimeThreshold)
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		if !processingDone {
			// processing is not done yet when timer fires, send notification email
			workflow.ExecuteActivity(ctx, sendEmailActivity).Get(ctx, nil)
		}
	})

	// wait the timer or the order processing to finish
	selector.Select(ctx)

	// now either the order processing is finished, or timer is fired.
	if !processingDone {
		// processing not done yet, so the handler for timer will send out notification email.
		// we still want the order processing to finish, so wait on it.
		selector.Select(ctx)
	}

	workflow.GetLogger(ctx).Info("Workflow completed.")
	return nil
}

// sampleTimerWorkflow workflow decider
func sampleTimerWorkflow2(ctx workflow.Context, processingTimeThreshold time.Duration) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	childCtx, cancelHandler := workflow.WithCancel(ctx)
	selector := workflow.NewSelector(ctx)

	// In this sample case, we want to demo a use case where the workflow starts a long running order processing operation
	// and in the case that the processing takes too long, we want to send out a notification email to user about the delay,
	// but we won't cancel the operation. If the operation finishes before the timer fires, then we want to cancel the timer.

	var processingDone bool
	f := workflow.ExecuteActivity(ctx, orderProcessingActivity2)
	selector.AddFuture(f, func(f workflow.Future) {
		processingDone = true
		// cancel timerFuture
		cancelHandler()
	})

	// use timer future to send notification email if processing takes too long
	timerFuture := workflow.NewTimer(childCtx, processingTimeThreshold)
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		if !processingDone {
			// processing is not done yet when timer fires, send notification email
			workflow.ExecuteActivity(ctx, sendEmailActivity).Get(ctx, nil)
		}
	})

	// wait the timer or the order processing to finish
	selector.Select(ctx)

	// now either the order processing is finished, or timer is fired.
	if !processingDone {
		// processing not done yet, so the handler for timer will send out notification email.
		// we still want the order processing to finish, so wait on it.
		selector.Select(ctx)
	}

	workflow.GetLogger(ctx).Info("Workflow completed.")
	return nil
}

func orderProcessingActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("sampleActivity processing started.")
	timeNeededToProcess := time.Second * time.Duration(rand.Intn(10))
	time.Sleep(timeNeededToProcess)
	logger.Info("sampleActivity done.", zap.Duration("duration", timeNeededToProcess))
	return nil
}

func orderProcessingActivity2(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("sampleActivity processing started.")
	timeNeededToProcess := time.Second * time.Duration(60)
	time.Sleep(timeNeededToProcess)
	logger.Info("sampleActivity done.", zap.Duration("duration", timeNeededToProcess))
	return nil
}

func sendEmailActivity(ctx context.Context) error {
	activity.GetLogger(ctx).Info("sendEmailActivity sending notification email as the process takes long time.")
	return nil
}
