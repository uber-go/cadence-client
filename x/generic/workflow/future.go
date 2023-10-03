package workflow

import (
	"go.uber.org/cadence/workflow"
	"go.uber.org/cadence/x/generic/base"
)

/*
Future defines a *mostly* type-safe version of workflow.Future.

As this is operating on serialized data, not in-process types, this CANNOT
truly guarantee that the return type can be deserialized.  It is sufficient
in most simple cases though, where the serialized type matches the runtime
type, and it is a type that is safe to (de)serialize.

Concretely: the type MUST be supported by your DataConverter, and it MUST be
compatible with the serialized data, or Get may fail.
*/
type Future[Result base.SerializationFriendly] interface {
	Get(ctx workflow.Context) (Result, error)

	// IsReady will return true Get is guaranteed to not block.
	IsReady() bool
}
