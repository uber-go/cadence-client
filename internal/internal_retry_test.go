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
			// service-busy means "retry later", which is still transient/retryable.
			// callers with retries MUST detect this separately and delay before retrying,
			// and ideally we'll return a minimum time-to-wait in errors in the future.
			&s.ServiceBusyError{},

			&s.InternalServiceError{},  // fallback error type from server, "unknown" = "retry might work"
			errors.New("unrecognized"), // fallback error type elsewhere, "unknown" = "retry might work"

			// below are all server-side internal errors and should never be exposed, they're just included
			// to be explicit about current behavior (without intentionally deciding on any of them).
			// retry by default, like any other unknown error.  it shouldn't lead to incorrect behavior.
			&s.CurrentBranchChangedError{},
			&s.RetryTaskV2Error{},
			&s.RemoteSyncMatchedError{},
			&s.InternalDataInconsistencyError{},
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
			&s.WorkflowExecutionAlreadyCompletedError{}, // completed workflows won't uncomplete
			&s.WorkflowExecutionAlreadyStartedError{},   // started workflows could complete quickly, but re-starting may not be desirable

			errShutdown, // shutdowns can't be stopped
		} {
			assert.False(t, isServiceTransientError(err), "%T should be fatal", err)
		}
	})
}
