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

/*
Package activity contains functions and types used to implement Cadence activities.

The activity is an implementation of a task to be performed as part of a larger workflow. There is no limitation of
what an activity can do. In the context of a workflow, it is in the activities where all operations that affect the
desired results must be implemented.

# Overview

The client library for Cadence does all the heavy lifting of handling the async communication between the Cadence
managed service and the worker running the activity. As such, the implementation of the activity can, for the most
part, focus on the business logic. The sample code below shows the implementation of a simple activity that accepts a
string parameter, appends a word to it and then returns the result.

	import (
		"context"

		"go.uber.org/cadence/activity"
		"go.uber.org/zap"
	)

	func init() {
		activity.Register(SimpleActivity)
	}

	func SimpleActivity(ctx context.Context, value string) (string, error) {
		activity.GetLogger(ctx).Info("SimpleActivity called.", zap.String("Value", value))
		return "Processed: ” + value, nil
	}

The following sections explore the elements of the above code.

# Declaration

In the Cadence programing model, an activity is implemented with a function. The function declaration specifies the
parameters the activity accepts as well as any values it might return. An activity function can take zero or many
activity specific parameters and can return one or two values. It must always at least return an error value. The
activity function can accept as parameters and return as results any serializable type.

	func SimpleActivity(ctx context.Context, value string) (string, error)

The first parameter to the function is context.Context. This is an optional parameter and can be omitted. This
parameter is the standard Go context.

The second string parameter is a custom activity-specific parameter that can be used to pass in data into the activity
on start. An activity can have one or more such parameters. All parameters to an activity function must be
serializable, which essentially means that params can’t be channels, functions, variadic, or unsafe pointer.
Exact details will depend on your DataConverter, but by default they must work with encoding/json.Marshal (and
Unmarshal on the receiving side, which has the same limitations plus generally cannot deserialize into an interface).

This activity declares two return values: (string, error). The string return value is used to return the result of the
activity, and can be retrieved in the workflow with this activity's Future.
The error return value is used to indicate an error was encountered during execution.
Results must be serializable, like parameters, but only a single result value is allowed (i.e. you cannot return
(string, string, error)).

# Implementation

There is nothing special about activity code. You can write activity implementation code the same way you would any
other Go service code. You can use the usual loggers and metrics collectors. You can use the standard Go concurrency
constructs.

# Failing the activity

To mark an activity as failed, return an error from your activity function via the error return value.
Note that failed activities do not record the non-error return's value: you cannot usefully return both a
value and an error, only the error will be recorded.

# Activity Heartbeating

For long running activities, Cadence provides an API for the activity code to report both liveness and progress back to
the Cadence managed service.

	progress := 0
	for hasWork {
	    // send heartbeat message to the server
	    activity.RecordHeartbeat(ctx, progress)
	    // do some work
	    ...
	    progress++
	}

When the activity times out due to a missed heartbeat, the last value of the details (progress in the above sample) is
returned from the workflow.ExecuteActivity function as the details field of TimeoutError with TimeoutType_HEARTBEAT.

It is also possible to heartbeat an activity from an external source:

	// instantiate a Cadence service Client
	client.Client client = client.NewClient(...)

	// record heartbeat
	err := client.RecordActivityHeartbeat(ctx, taskToken, details)

It expects an additional parameter, "taskToken", which is the value of the binary "TaskToken" field of the
"ActivityInfo" struct retrieved inside the activity (GetActivityInfo(ctx).TaskToken). "details" is the serializable
payload containing progress information.

# Activity Cancellation

When an activity is cancelled (or its workflow execution is completed or failed) the context passed into its function
is cancelled which closes its Done() channel. So an activity can use that to perform any necessary cleanup
and abort its execution.

Currently, cancellation is delivered only to activities that call RecordHeartbeat.  If heartbeating is not performed,
the activity will continue to run normally, but fail to record its result when it completes.

# Async and Manual Activity Completion

In certain scenarios completing an activity upon completion of its function is not possible or desirable.

One example would be the UberEATS order processing workflow that gets kicked off once an eater pushes the “Place Order”
button. Here is how that workflow could be implemented using Cadence and the “async activity completion”:

  - Activity 1: send order to restaurant
  - Activity 2: wait for restaurant to accept order
  - Activity 3: schedule pickup of order
  - Activity 4: wait for courier to pick up order
  - Activity 5: send driver location updates to eater
  - Activity 6: complete order

Activities 2 & 4 in the above flow require someone in the restaurant to push a button in the Uber app to complete the
activity. The activities could be implemented with some sort of polling mechanism. However, they can be implemented
much simpler and much less resource intensive as a Cadence activity that is completed asynchronously.

There are 2 parts to implementing an asynchronously completed activity. The first part is for the activity to provide
the information necessary to be able to be completed from an external system and notify the Cadence service that it is
waiting for that outside callback:

	// retrieve activity information needed to complete activity asynchronously
	activityInfo := activity.GetInfo(ctx)
	taskToken := activityInfo.TaskToken

	// send the taskToken to external service that will complete the activity
	...

	// return from activity function indicating the Cadence should wait for an async completion message
	return "", activity.ErrResultPending

The second part is then for the external service to call the Cadence service to complete the activity. To complete the
activity successfully you would do the following:

	// instantiate a Cadence service Client
	// the same client can be used complete or fail any number of activities
	client.Client client = client.NewClient(...)

	// complete the activity
	client.CompleteActivity(taskToken, result, nil)

And here is how you would fail the activity:

	// fail the activity
	client.CompleteActivity(taskToken, nil, err)

The parameters of the CompleteActivity function are:

  - taskToken: This is the value of the binary “TaskToken” field of the
    “ActivityInfo” struct retrieved inside the activity.
  - result: This is the return value that should be recorded for the activity.
    The type of this value needs to match the type of the return value
    declared by the activity function.
  - err: The error code to return if the activity should terminate with an
    error.

If error is not null the value of the result field is ignored.

For a full example of implementing this pattern see the Expense sample.

# Registration

In order for a workflow to be able to execute an activity type, the worker process needs to be aware of
all the implementations it has access to. An activity is registered with the following call:

	activity.Register(SimpleActivity)

This call essentially creates an in-memory mapping inside the worker process between the fully qualified function name
and the implementation. Unlike in Amazon SWF, workflow and activity types are not registered with the managed service.
If the worker receives a request to start an activity execution for an activity type it does not know it will fail that
request.
*/
package activity
