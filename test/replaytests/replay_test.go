// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package replaytests

import (
	"strings"
	"testing"

	"go.uber.org/cadence/activity"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
)

// Basic happy paths
func TestReplayWorkflowHistoryFromFile(t *testing.T) {
	for _, testFile := range []string{"basic.json", "basic_new.json", "version.json", "version_new.json"} {
		t.Run("replay_"+strings.Split(testFile, ".")[0], func(t *testing.T) {
			replayer := worker.NewWorkflowReplayer()
			replayer.RegisterWorkflow(Workflow)
			replayer.RegisterWorkflow(Workflow2)

			err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), testFile)
			require.NoError(t, err)
		})
	}
}

// Child workflow happy path
func TestReplayChildWorkflowBugBackport(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(childWorkflow, workflow.RegisterOptions{Name: "child"})
	replayer.RegisterWorkflowWithOptions(childWorkflowBug, workflow.RegisterOptions{Name: "parent"})

	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "child_bug.json")
	require.NoError(t, err)
}

// Gives a non-deterministic-error because the getGreetingActivitytest was not registered on the replayer.
func TestGreetingsWorkflowforActivity(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(greetingsWorkflowActivity, workflow.RegisterOptions{Name: "greetings"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "greetings.json")
	assert.ErrorContains(t, err, "nondeterministic workflow: mismatching history event and replay decision found")
}

// Simple greeting workflow with 3 activities executed sequentially: getGreetingsActivity, getNameActivity, sayGreetingsActivity
func TestGreetingsWorkflow(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(greetingsWorkflow, workflow.RegisterOptions{Name: "greetings"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "greetings.json")
	require.NoError(t, err)
}

// Return types of activity change is not considered non-determinism (at least for now) so this test doesn't find non-determinism error
func TestGreetingsWorkflow3(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterActivityWithOptions(getNameActivity3, activity.RegisterOptions{Name: "main.getNameActivity", DisableAlreadyRegisteredCheck: true})
	replayer.RegisterWorkflowWithOptions(greetingsWorkflow3, workflow.RegisterOptions{Name: "greetings"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "greetings.json")
	require.NoError(t, err)
}

// The recorded history has following activities in this order:  main.getOrderActivity, main.orderBananaActivity
// This test runs a version of choice workflow which does the exact same thing so no errors expected.
func TestExclusiveChoiceWorkflowSuccess(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(exclusiveChoiceWorkflow, workflow.RegisterOptions{Name: "choice"})
	replayer.RegisterActivityWithOptions(getBananaOrderActivity, activity.RegisterOptions{Name: "main.getOrderActivity"})
	replayer.RegisterActivityWithOptions(orderBananaActivity, activity.RegisterOptions{Name: "main.orderBananaActivity"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "choice.json")
	require.NoError(t, err)
}

// The recorded history has following activities in this order:  main.getOrderActivity, main.orderBananaActivity
// This test runs a version of choice workflow which does the exact same thing but the activities are not registered.
// It doesn't matter for replayer so no exceptions expected.
// The reason is that activity result decoding logic just passes the result back to the given pointer
func TestExclusiveChoiceWorkflowActivitiyRegistrationMissing(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(exclusiveChoiceWorkflow, workflow.RegisterOptions{Name: "choice"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "choice.json")
	require.NoError(t, err)
}

// The recorded history has following activities in this order:  main.getOrderActivity, main.orderBananaActivity
// This test runs a version of choice workflow which registers a single return parameter function for main.getOrderActivity
// - Original main.getOrderActivity signature: func() (string, error)
// - New main.getOrderActivity signature: func() error
//
// In this case result of main.getOrderActivity from history is not passed back to the given pointer by the workflow.
// Compared to the activity registration missing scenario (above case) this is a little bit weird behavior.
// The workflow code continues with orderChoice="" instead of "banana". Therefore it doesn't invoke 2nd activity main.getOrderActivity.
// This means history has more events then replay decisions which causes non-determinism error
func TestExclusiveChoiceWorkflowWithActivitySignatureChange(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(exclusiveChoiceWorkflow, workflow.RegisterOptions{Name: "choice"})
	replayer.RegisterActivityWithOptions(func() error { return nil }, activity.RegisterOptions{Name: "main.getOrderActivity"})
	replayer.RegisterActivityWithOptions(orderBananaActivity, activity.RegisterOptions{Name: "main.orderBananaActivity"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "choice.json")
	assert.ErrorContains(t, err, "nondeterministic workflow: missing replay decision")
}

// The recorded history has following activities in this order:   main.getOrderActivity, main.orderBananaActivity
// This test runs a version of choice workflow which calls main.getOrderActivity and then calls the main.orderCherryActivity.
// The replayer will find non-determinism because of mismatch between replay decision and history (banana vs cherry)
func TestExclusiveChoiceWorkflowWithMismatchingActivity(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(exclusiveChoiceWorkflowAlwaysCherry, workflow.RegisterOptions{Name: "choice"})
	replayer.RegisterActivityWithOptions(getBananaOrderActivity, activity.RegisterOptions{Name: "main.getOrderActivity"})
	replayer.RegisterActivityWithOptions(orderCherryActivity, activity.RegisterOptions{Name: "main.orderCherryActivity"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "choice.json")
	assert.ErrorContains(t, err, "nondeterministic workflow: mismatching history event and replay decision found")
}

// Branch workflow happy case.
// It branches out to 3 open activities and then they complete.
func TestBranchWorkflow(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(sampleBranchWorkflow, workflow.RegisterOptions{Name: "branch"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "branch.json")
	require.NoError(t, err)
}

// Branch workflow normal history file is replayed against modified workflow code which
// has 2 branches only.  This causes nondetereministic error.
func TestBranchWorkflowWithExtraBranch(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(sampleBranchWorkflow2, workflow.RegisterOptions{Name: "branch"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "branch.json")
	assert.ErrorContains(t, err, "nondeterministic workflow: missing replay decision")
}

// TestSequentialStepsWorkflow replays a history with 2 sequential non overlapping activity calls (one completes before the other is scheduled)
// and runs it against new version of the workflow code which only calls 1 activity.
// This is considered as non-determinism error.
func TestSequentialStepsWorkflow(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(replayerHelloWorldWorkflow, workflow.RegisterOptions{Name: "fx.ReplayerHelloWorldWorkflow"})
	replayer.RegisterActivityWithOptions(replayerHelloWorldActivity, activity.RegisterOptions{Name: "replayerhello"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "sequential.json")
	assert.ErrorContains(t, err, "nondeterministic workflow: missing replay decision")
}

// Runs simpleParallelWorkflow which starts two workflow.Go routines that executes 1 and 2 activities respectively.
func TestSimpleParallelWorkflow(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(sampleParallelWorkflow, workflow.RegisterOptions{Name: "branch2"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "branch2.json")
	require.NoError(t, err)
}

// Runs modified version of simpleParallelWorkflow which starts 1 less activity in the second workflow-gouroutine.
// This is considered as non-determinism error.
func TestSimpleParallelWorkflowWithMissingActivityCall(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(sampleParallelWorkflow2, workflow.RegisterOptions{Name: "branch2"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "branch2.json")
	assert.ErrorContains(t, err, "nondeterministic workflow: missing replay decision")
}

// Runs a history which ends with WorkflowExecutionContinuedAsNew. Replay fails because of the additional checks done
// for continue as new case by replayWorkflowHistory().
// This should not have any error because it's a valid continue as new case.
func TestContinueAsNew(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflowWithOptions(ContinueAsNewWorkflow, workflow.RegisterOptions{Name: "fx.SimpleSignalWorkflow"})
	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "continue_as_new.json")
	assert.ErrorContains(t, err, "replay workflow doesn't return the same result as the last event")
}
