package saga

type CompensationError struct {
	err    error
	actErr error
}

func (c CompensationError) ActionError() error {
	return c.actErr
}

func (c CompensationError) Error() string {
	return c.err.Error()
}
