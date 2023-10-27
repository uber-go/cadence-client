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
	"fmt"
	"time"

	"go.uber.org/zap"

	"go.uber.org/cadence/workflow"
)

/**
 * This sample workflow Execute one of many code paths based on the result of an activity.
 */

const (
	orderChoiceApple  = "apple"
	orderChoiceBanana = "banana"
	orderChoiceCherry = "cherry"
)

// exclusiveChoiceWorkflow executes main.getOrderActivity and executes either cherry or banana activity depends on what main.getOrderActivity returns
func exclusiveChoiceWorkflow(ctx workflow.Context) error {
	// Get order.
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var orderChoice string
	err := workflow.ExecuteActivity(ctx, "main.getOrderActivity").Get(ctx, &orderChoice)
	if err != nil {
		return err
	}

	logger := workflow.GetLogger(ctx)

	// choose next activity based on order result
	switch orderChoice {
	case orderChoiceBanana:
		workflow.ExecuteActivity(ctx, "main.orderBananaActivity", orderChoice)
	case orderChoiceCherry:
		workflow.ExecuteActivity(ctx, "main.orderCherryActivity", orderChoice)
	default:
		logger.Error("Unexpected order", zap.String("Choice", orderChoice))
	}

	logger.Info("Workflow completed.")
	return nil
}

// exclusiveChoiceWorkflow executes main.getOrderActivity and executes either cherry or banana activity depends on what main.getOrderActivity returns
func exclusiveChoiceWorkflowAlwaysCherry(ctx workflow.Context) error {
	// Get order.
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var orderChoice string
	err := workflow.ExecuteActivity(ctx, "main.getOrderActivity").Get(ctx, &orderChoice)
	if err != nil {
		return err
	}

	logger := workflow.GetLogger(ctx)
	logger.Sugar().Infof("Got order for %s but will ignore and order cherry!!", orderChoice)

	workflow.ExecuteActivity(ctx, "main.orderCherryActivity", orderChoice)

	logger.Info("Workflow completed.")
	return nil
}

func getBananaOrderActivity() (string, error) {
	fmt.Printf("Order is for Apple")
	return "apple", nil
}

func orderBananaActivity(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func orderCherryActivity(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}
