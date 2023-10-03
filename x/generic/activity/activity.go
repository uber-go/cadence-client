/*
Package activity contains "friendly" generics-using Activity APIs, and is
roughly equivalent to [activity] in its abilities.

If you find the APIs in here too restrictive, see the [bare] package instead.
That package provides similar APIs but with only truly minimal requirements,
often giving up safety or ease-of-use to achieve it.

BY USING ANY PACKAGE IN X YOU ACCEPT THAT EVERYTHING MAY CHANGE, AND YOU ARE
QUITE LIKELY TO NOT GET ANY SUPPORT AT ALL.  Use with caution.
*/
package activity

import (
	"context"
	"fmt"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/cadence/x/generic/base"
	"go.uber.org/cadence/x/generic/internal"
	gworkflow "go.uber.org/cadence/x/generic/workflow"
	"reflect"

	_ "go.uber.org/cadence/x/generic/bare"
)

/*
Register registers an activity and returns an Executable that can be used in
workflows to run this activity with relatively compile-time and run-time safety.

# Alternatives

If these requirements are too strict and you have verified that they are not
needed for your usecase: that's fine!  Check the [go.uber.org/cadence/x/generic/bare]
APIs for looser versions that are intended to only block things that are
essentially guaranteed to be incorrect.

If these requirements are too *loose* for your taste: that's fine!  Check the
[go.uber.org/cadence/x/generic/safe] APIs for versions that are much more strict
about allowed types, and perform much more runtime validation.  These assume you
are using the built-in DataConverter, which is essentially [encoding/json.Marshal]
(and Unmarshal), with carve-outs for using [apache.org/thrift] or
[googlesource.com/protobuf].  Note however that these APIs are currently VERY
unstable, and no guarantees are made AT ALL that a type that works today will
work in the next version.  Any new error-prone patterns we discover may be
blocked as we discover them.

# Requirements and how to use

The fundamental requirements are essentially:

	// define your types and funcs:
	type ArgType struct{...}    // must be a serialization-friendly struct
	type ResultType struct{...} // must be a serialization-friendly struct
	func (a *activityImpl) act(ctx context.Context, arg ArgType) (*ResultType, error) { ... }

	// at startup:
	// 1) create your worker:
	w := ...
	// 2) register your activities
	a := &activityImpl{dependencies...}
	exe := activity.Register(w, a.act, "explicit name")
	// 3) register your workflows with Executables that they use:
	wfimpl := &wf{act: exe}
	workflow.Register(w, wfimpl.Run...)

	// we encourage you to use a dependency-injection tool to make this more
	// ergonomic.

And use the returned value in workflows for better ergonomics than the
interface{}-heavy APIs:

	func (w *wf) Run(ctx workflow.Context, arg Arg) (*Result, error) {
		// ...
		result, err := w.act.Run(ctx, ArgType{Field:"yay"}) // blocks until a result
		// ...
	}

# Cautions

Unfortunately this cannot guarantee a number of potential issues at compile
time, so this function will panic when its expectations are violated, to improve
the chance of early detection of problems:

  - Arguments must be "serialization friendly".
    What this means depends on your DataConverter, but by default (and generally
    speaking) this means "structs with public fields", so fields can be added in
    the future without breaking backwards compatibility on a binary level.
  - Results must follow the same rules as arguments.
  - There must be one argument and one result.  If they are unused, they must
    still be defined; just use `struct{}` or the equivalent empty object for
    your DataConverter as a placeholder.  This helps ensure you can add fields
    in the future without breaking backwards compatibility on a binary level.
  - Your name-string must be unique (per activity type), and must not be empty.
    This helps ensure you can rename or move your types and functions around
    without breaking backwards compatibility due to the inferred name changes.

Pointers vs values generally does not matter to deserialization frameworks, but
do note that if an error is returned, ONLY the error will be saved.  Cadence
does not support BOTH value and error returns, as this does not have a clear
translation to all languages.  E.g. Java uses exceptions for error handling, and
there is no way to both throw an exception and return a value in Java.
*/
func Register[Arg base.SerializationFriendly, Result base.SerializationFriendly](w worker.Worker, fn func(ctx context.Context, arg Arg) (Result, error), name string) Executable[Arg, Result] {
	if name == "" {
		// no good reason to avoid this
		panic(fmt.Sprintf(
			"cannot register this activity without an explicit name: %T, would use generated name: %s",
			fn, "tbd", // TODO: get would-be-generated name, so they can just copy/paste
		))
	}

	// TODO: json-safety validator instead?
	argType := reflect.TypeOf(fn).In(1)
	if argType.Kind() != reflect.Struct {
		// some non-struct types are valid, but it generally requires a custom DataConverter
		panic(fmt.Sprintf(
			"registering an activity with a non-struct argument is risky, use x/generic/bare if this is intentional: %s in activity named %s",
			argType.String(), name,
		))
	}

	// can panic if the name is already registered
	w.RegisterActivityWithOptions(fn, activity.RegisterOptions{Name: name})
	return &executable[Arg, Result]{
		fn: fn,
	}
}

/*
Executable represents a generic activity that can be started, and can return a
value through a future (or optionally synchronously).
*/
type Executable[Arg base.SerializationFriendly, Result base.SerializationFriendly] interface {
	// Start starts this activity asynchronously, check the returned gworkflow.Future for results.
	Start(ctx workflow.Context, arg Arg) gworkflow.Future[Result]
	// Run starts this activity and waits for the result synchronously, returning it.
	Run(ctx workflow.Context, arg Arg) (Result, error)
}

type executable[Arg base.SerializationFriendly, Result base.SerializationFriendly] struct {
	fn func(ctx context.Context, arg Arg) (Result, error)
}

func (e *executable[Arg, Result]) Start(ctx workflow.Context, arg Arg) gworkflow.Future[Result] {
	f := workflow.ExecuteActivity(ctx, e.fn, arg)
	return &internal.Future[Result]{Fut: f}
}

func (e *executable[Arg, Result]) Run(ctx workflow.Context, arg Arg) (Result, error) {
	f := e.Start(ctx, arg)
	return f.Get(ctx)
}
