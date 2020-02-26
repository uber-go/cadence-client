package saga

import "go.uber.org/cadence/workflow"

type (
	action interface {
		Act(workflow.Context) error
	}

	compensation interface {
		Compensate(workflow.Context) error
	}

	SagaOptions struct {
		ParallelCompensation bool
		ContinueWithError    bool
	}

	sagaImpl struct {
		options       *SagaOptions
		actions       []action
		compensations []compensation
	}
)

func NewSaga(options SagaOptions) *sagaImpl {
	return &sagaImpl{options: &options}
}

func (s *sagaImpl) Action(act action) *sagaImpl {
	s.actions = append(s.actions, act)
	return s
}

func (s *sagaImpl) WithCompensation(comp compensation) *sagaImpl {
	s.compensations = append(s.compensations, comp)
	return s
}

func (s *sagaImpl) Run(ctx workflow.Context) error {
	for _, action := range s.actions {
		err := action.Act(ctx)
		if err != nil {
			compErr := s.Compensate(ctx)
			if compErr != nil {
				return CompensationError{
					err:    compErr,
					actErr: err,
				}
			}
			return err
		}
	}
	return nil
}

func (s *sagaImpl) Compensate(ctx workflow.Context) error {
	if s.options.ParallelCompensation == true {
		for _, compensation := range s.compensations {
			go func() {
				_ = compensate(ctx, compensation, true)
			}()
		}
		return nil
	}
	for _, compensation := range s.compensations {
		err := compensate(ctx, compensation, s.options.ContinueWithError)
		if err != nil {
			return err
		}
	}
	return nil
}

func compensate(ctx workflow.Context, c compensation, continueWithCompensationError bool) error {
	err := c.Compensate(ctx)
	if err != nil && continueWithCompensationError == false {
		return err
	}
	return nil
}
