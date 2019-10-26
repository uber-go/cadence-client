package workflow

type (
    SagaOptions struct {
        ParallelCompensation bool
        ContinueWithError    bool
    }

    Saga interface {
        AddCompensation(Context, interface{}, ...interface{}) Saga
        Compensate() error
    }

    sagaImpl struct {
        options         *SagaOptions
        compensationOps []*compensationOp
    }

    compensationOp struct {
        ctx      Context
        activity interface{}
        args     []interface{}
    }
)

func NewSaga(options SagaOptions) Saga {
    return &sagaImpl{options: &options}
}

func (s *sagaImpl) AddCompensation(ctx Context, activity interface{}, args ...interface{}) Saga {
    s.compensationOps = append(s.compensationOps, &compensationOp{ctx: ctx, activity: activity, args: args})
    return s
}

func (s *sagaImpl) Compensate() error {
    // avoid compensating twice
    defer func() { s.compensationOps = nil }()

    var futures []Future
    for _, op := range s.compensationOps {
        future := ExecuteActivity(op.ctx, op.activity, op.args...)

        if s.options.ParallelCompensation {
            futures = append(futures, future)
        } else {
            err := future.Get(op.ctx, nil)
            if err != nil && !s.options.ContinueWithError {
                return err
            }
        }
    }

    // for ParallelCompensation
    for i, future := range futures {
        err := future.Get(s.compensationOps[i].ctx, nil)
        if err != nil && !s.options.ContinueWithError {
            return err
        }
    }

    return nil
}
