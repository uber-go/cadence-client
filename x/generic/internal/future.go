package internal

import (
	"go.uber.org/cadence/workflow"
	"go.uber.org/cadence/x/generic/base"
	gworkflow "go.uber.org/cadence/x/generic/workflow"
)

var _ gworkflow.Future[struct{ f any }] = (*Future[struct{ f any }])(nil)

type Future[Result base.SerializationFriendly] struct {
	Fut workflow.Future
}

func (f *Future[Result]) Get(ctx workflow.Context) (Result, error) {
	var out Result
	err := f.Fut.Get(ctx, &out)
	return out, err
}

func (f *Future[Result]) IsReady() bool {
	return f.Fut.IsReady()
}
