package internal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	s "go.uber.org/cadence/.gen/go/shared"
)

func TestErrRetries(t *testing.T) {
	t.Run("retryable", func(t *testing.T) {
		for _, err := range []error{
			errors.New("unrecognized"),
			&s.InternalServiceError{},
			&s.ServiceBusyError{},
			&s.LimitExceededError{},
		} {
			assert.True(t, isServiceTransientError(err), "%T should be transient", err)
		}
	})
	t.Run("terminal", func(t *testing.T) {
		for _, err := range []error{
			&s.BadRequestError{},
			&s.ClientVersionNotSupportedError{},
			&s.DomainAlreadyExistsError{},
			&s.DomainNotActiveError{},
			&s.EntityNotExistsError{},
			&s.QueryFailedError{},
			&s.WorkflowExecutionAlreadyCompletedError{},
			&s.WorkflowExecutionAlreadyStartedError{},
			errShutdown,
		} {
			assert.False(t, isServiceTransientError(err), "%T should be fatal", err)
		}
	})
}
