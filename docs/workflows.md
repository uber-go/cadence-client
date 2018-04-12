# Workflows

A workflow is the implementation of coordination logic. Its sole purpose is to orchestrate activity executions.

Workflows are implemented as functions. Startup data can be passed to a workflow via function parameters. The parameters can be either basic types or structs, with the only requirement being that the parameters need to be serializable. The first parameter of a workflow function is of type `workflow.Context`. The function must return an **error** value, and can optionally return a result value. The result value can be either a basic type or a struct with the only requirement being that it is serializable.

Workflow functions need to execute deterministically. Therefore, here is a list of rules that workflow code should obey:
* Use `workflow.Context` everywhere.
* Don’t use range over `map`.
* Use `workflow.SideEffect` to call rand and similar nondeterministic functions like UUID generator.
* Use `workflow.Now` to get the current time. Use `workflow.NewTimer` or `workflow.Sleep` instead of standard Go functions.
* Don’t use native channel and select. Use `workflow.Channel` and `workflow.Selector`.
* Don’t use go func(...). Use `workflow.Go(func(...))`.
* Don’t use non-constant global variables as multiple instances of a workflow function can be executing in parallel.
* Don’t use any blocking functions besides belonging to `Channel`, `Selector` or `Future`.
* Don’t use any synchronization primitives as they can cause blockage and there is no possibility of races when running under dispatcher.
* Don't just change workflow code when there are open workflows. Always update code using `workflow.GetVersion`.
* Don’t perform any IO or service calls as they are not usually deterministic. Use activities for that.
* Don’t access configuration APIs directly from a workflow as changes in configuration will affect the workflow execution path. Either return configuration from an activity or use `workflow.SideEffect` to load it.

In order to make the workflow visible to the worker process hosting it, the workflow needs to be registered via a call to **workflow.Register**.

```go
package simple

import (
	"time"


	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

func init() {
	workflow.Register(SimpleWorkflow)
}

// SimpleWorkflow is a sample Cadence workflow that accepts one parameter and
// executes an activity to which it passes the aforementioned parameter.
func SimpleWorkflow(ctx workflow.Context, value string) error {
	options := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Second * 60,
		StartToCloseTimeout:    time.Second * 60,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var result string
	err := workflow.ExecuteActivity(ctx, activity.SimpleActivity, value).Get(ctx, &result)
	if err != nil {
		return err
	}
	workflow.GetLogger(ctx).Info(
		"SimpleActivity returned successfully!", zap.String("Result", result))

	workflow.GetLogger(ctx).Info("SimpleWorkflow completed!")
	return nil
}
```
