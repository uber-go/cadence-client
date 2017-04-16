package cadence

import (
	"errors"
	"fmt"

	"github.com/uber-go/cadence-client/.gen/go/shared"
)

type (
	// Marker functions are used to ensure that interfaces never implement each other.
	// For example without marker an implementation of ErrorWithDetails matches
	// CanceledError interface as well.

	// ErrorWithDetails to return from Workflow and activity implementations.
	ErrorWithDetails interface {
		error
		Reason() string
		Details() []byte
		errorWithDetails() // interface marker
	}

	// TimeoutError returned when activity or child workflow timed out
	TimeoutError interface {
		error
		TimeoutType() shared.TimeoutType
		Details() []byte // Present only for HEARTBEAT TimeoutType
		timeoutError()   // interface marker
	}

	// CanceledError returned when operation was canceled
	CanceledError interface {
		error
		Details() []byte
		canceledError() // interface marker
	}

	// PanicError contains information about panicked workflow
	PanicError interface {
		error
		Value() interface{} // Value passed to panic call
		StackTrace() string // Stack trace of a panicked coroutine
		panicError()        // interface marker
	}
)

var _ ErrorWithDetails = (*errorWithDetails)(nil)
var _ CanceledError = (*canceledError)(nil)
var _ TimeoutError = (*timeoutError)(nil)
var _ PanicError = (*panicError)(nil)

// ErrActivityResultPending is returned from activity's Execute method to indicate the activity is not completed when
// Execute method returns. activity will be completed asynchronously when Client.CompleteActivity() is called.
var ErrActivityResultPending = errors.New("not error: do not autocomplete, " +
	"using Client.CompleteActivity() to complete")

// NewErrorWithDetails creates ErrorWithDetails instance
// Create standard error through errors.New or fmt.Errorf if no details are provided
func NewErrorWithDetails(reason string, details []byte) ErrorWithDetails {
	return &errorWithDetails{reason: reason, details: details}
}

// NewTimeoutError creates TimeoutError instance
func NewTimeoutError(timeoutType shared.TimeoutType) TimeoutError {
	return &timeoutError{timeoutType: timeoutType}
}

// NewHeartbeatTimeoutError creates TimeoutError instance
func NewHeartbeatTimeoutError(details []byte) TimeoutError {
	return &timeoutError{timeoutType: shared.TimeoutType_HEARTBEAT, details: details}
}

// NewCanceledErrorWithDetails creates CanceledError instance
func NewCanceledErrorWithDetails(details []byte) CanceledError {
	return &canceledError{details: details}
}

// NewCanceledError creates CanceledError instance
func NewCanceledError() CanceledError {
	return NewCanceledErrorWithDetails([]byte{})
}

// errorWithDetails implements ErrorWithDetails
type errorWithDetails struct {
	reason  string
	details []byte
}

func (e *errorWithDetails) Error() string {
	return e.reason
}

// Reason is from ErrorWithDetails interface
func (e *errorWithDetails) Reason() string {
	return e.reason
}

// Details is from ErrorWithDetails interface
func (e *errorWithDetails) Details() []byte {
	return e.details
}

// errorWithDetails is from ErrorWithDetails interface
func (e *errorWithDetails) errorWithDetails() {}

// timeoutError implements TimeoutError
type timeoutError struct {
	timeoutType shared.TimeoutType
	details     []byte
}

// ErrorWithDetails from error.ErrorWithDetails
func (e *timeoutError) Error() string {
	return fmt.Sprintf("TimeoutType: %v", e.timeoutType)
}

func (e *timeoutError) TimeoutType() shared.TimeoutType {
	return e.timeoutType
}

// Details is from ErrorWithDetails interface
func (e *timeoutError) Details() []byte {
	return e.details
}

func (e *timeoutError) timeoutError() {}

type canceledError struct {
	details []byte
}

// ErrorWithDetails from error.ErrorWithDetails
func (e *canceledError) Error() string {
	return fmt.Sprintf("Details: %s", e.details)
}

// Details of the error
func (e *canceledError) Details() []byte {
	return e.details
}

func (e *canceledError) canceledError() {}

type panicError struct {
	value      interface{}
	stackTrace string
}

func newPanicError(value interface{}, stackTrace string) PanicError {
	return &panicError{value: value, stackTrace: stackTrace}
}

func (e *panicError) Error() string {
	return fmt.Sprintf("%v", e.value)
}

func (e *panicError) Value() interface{} {
	return e.value
}

func (e *panicError) StackTrace() string {
	return e.stackTrace
}

func (e *panicError) panicError() {}
