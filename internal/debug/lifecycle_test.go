package debug

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPollerLifeCycle(t *testing.T) {
	lifeCycle := NewLifeCycle()

	// Initially, poller count should be 0
	require.Equal(t, int32(0), lifeCycle.ReadPollerCount(), "Expected initial poller count to be 0")

	// Start a poller and verify that the count increments
	run1 := lifeCycle.PollerStart("worker-1")
	require.Equal(t, int32(1), lifeCycle.ReadPollerCount(), "Expected poller count to be 1 after starting a poller")

	// Start another poller and verify that the count increments again
	run2 := lifeCycle.PollerStart("worker-2")
	require.Equal(t, int32(2), lifeCycle.ReadPollerCount(), "Expected poller count to be 2 after starting a second poller")

	// Stop the poller twice and verify idempotency
	run1.Stop()
	require.Equal(t, int32(1), lifeCycle.ReadPollerCount(), "Expected poller count to be 1")
	run1.Stop()
	assert.Equal(t, int32(1), lifeCycle.ReadPollerCount(), "Expected poller count to be 1")

	run2.Stop()
	assert.Equal(t, int32(0), lifeCycle.ReadPollerCount(), "Expected poller count to remain 0 after second stop")
}
