# Activities

An activity is the implementation of a particular task in the business logic.

Activities are implemented as functions. Data can be passed directly to an activity via function parameters. The parameters can be either basic types or structs, with the only requirement being that the parameters need to be serializable. Even though it is not required, we recommand that the first parameter of an activity function is of type `context.Context`, in order to allow the activity to interact with other framework methods. The function must return an `error` value, and can optionally return a result value. The result value can be either a basic type or a struct with the only requirement being that it is serializable.

The values passed to activities through invocation parameters or returned through the result value is recorded in the execution history. The entire execution history is transfered from the Cadence service to workflow workers with every event that the workflow logic needs to process. A large execution history can thus adversily impact the performance of your workflow. Therefore be mindful of the amount of data you transfer via activity invocation parameters or return values. Other than that no additional limitations exist on activity implementations.

In order to make the activity visible to the worker process hosting it, the activity needs to be registered via a call to `activity.Register`.

```go
package simple

import (
	"context"

    "go.uber.org/cadence/activity"
    "go.uber.org/zap"
)

func init() {
	activity.Register(SimpleActivity)
}

// SimpleActivity is a sample Cadence activity function that takes one parameter and
// returns a string containing the parameter value.
func SimpleActivity(ctx context.Context, value string) (string, error) {
	activity.GetLogger(ctx).Info("SimpleActivity called.", zap.String("Value", value))
	return "Processed: " + value, nil
}
```
