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
			&s.CurrentBranchChangedError{},      // unsure, but second attempt may get the latest branch?
			&s.RetryTaskV2Error{},               // unsure but added here to make it explicit
			&s.RemoteSyncMatchedError{},         // retry may get a new match
			&s.InternalDataInconsistencyError{}, // seems unlikely to work, but not zero, server may detect and handle

			&s.InternalServiceError{},  // fallback error type from server, "unknown" = "retry might work"
			errors.New("unrecognized"), // fallback error type elsewhere, "unknown" = "retry might work"
		} {
			assert.True(t, isServiceTransientError(err), "%T should be transient", err)
		}
	})
	t.Run("terminal", func(t *testing.T) {
		for _, err := range []error{
			&s.AccessDeniedError{},                      // access won't be granted immediately
			&s.BadRequestError{},                        // bad requests don't become good
			&s.CancellationAlreadyRequestedError{},      // can only cancel once
			&s.ClientVersionNotSupportedError{},         // clients don't magically upgrade
			&s.DomainAlreadyExistsError{},               // re-creating again won't work
			&s.DomainNotActiveError{},                   // may actually succeed, but very unlikely to work immediately
			&s.EntityNotExistsError{},                   // may actually succeed, but very unlikely to work immediately
			&s.FeatureNotEnabledError{},                 // features won't magically enable
			&s.LimitExceededError{},                     // adding +1 to a value does not move it below its limit
			&s.QueryFailedError{},                       // arguable, this could be considered "unknown" and retryable
			&s.ServiceBusyError{},                       // intentionally not retried as that's a retry storm
			&s.WorkflowExecutionAlreadyCompletedError{}, // completed workflows won't uncomplete
			&s.WorkflowExecutionAlreadyStartedError{},   // started workflows could complete quickly, but re-starting may not be desirable

			errShutdown, // shutdowns can't be stopped
		} {
			assert.False(t, isServiceTransientError(err), "%T should be fatal", err)
		}
	})
}
