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

package internal

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/encoded"
)

type SessionTestSuite struct {
	*require.Assertions
	suite.Suite
	WorkflowTestSuite
}

func (s *SessionTestSuite) SetupSuite() {
	RegisterActivityWithOptions(testSessionActivity, RegisterActivityOptions{Name: "testSessionActivity"})
}

func (s *SessionTestSuite) SetupTest() {
	s.Assertions = require.New(s.T())
}

func TestSessionTestSuite(t *testing.T) {
	suite.Run(t, new(SessionTestSuite))
}

func (s *SessionTestSuite) TestCreationCompletion() {
	workflowFn := func(ctx Context) error {
		ao := ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
			HeartbeatTimeout:       time.Second * 20,
		}
		ctx = WithActivityOptions(ctx, ao)
		sessionCtx, err := CreateSession(ctx)
		if err != nil {
			return err
		}
		info, err := GetSessionInfo(sessionCtx)
		if err != nil || info.sessionState != sessionStateOpen {
			return errors.New("session state should be open after creation")
		}

		err = CompleteSession(sessionCtx)
		if err != nil {
			return err
		}
		info, err = GetSessionInfo(sessionCtx)
		if err != nil || info.sessionState != sessionStateClosed {
			return errors.New("session state should be closed after completion")
		}
		return nil
	}

	RegisterWorkflow(workflowFn)
	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *SessionTestSuite) TestCreationWithOpenSessionContext() {
	workflowFn := func(ctx Context) error {
		sessionCtx := setSessionInfo(ctx, &SessionInfo{
			SessionID:    "some random sessionID",
			tasklist:     "some random tasklist",
			sessionState: sessionStateOpen,
		})
		_, err := CreateSession(sessionCtx)
		return err
	}

	RegisterWorkflow(workflowFn)
	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.Equal(errFoundExistingOpenSession.Error(), env.GetWorkflowError().Error())
}

func (s *SessionTestSuite) TestCreationWithClosedSessionContext() {
	workflowFn := func(ctx Context) error {
		ao := ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
			HeartbeatTimeout:       time.Second * 20,
		}
		ctx = WithActivityOptions(ctx, ao)
		sessionCtx := setSessionInfo(ctx, &SessionInfo{
			SessionID:    "some random sessionID",
			tasklist:     "some random tasklist",
			sessionState: sessionStateClosed,
		})

		sessionCtx, err := CreateSession(sessionCtx)
		if err != nil {
			return err
		}
		return CompleteSession(sessionCtx)
	}

	RegisterWorkflow(workflowFn)
	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *SessionTestSuite) TestCreationWithFailedSessionContext() {
	workflowFn := func(ctx Context) error {
		ao := ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
			HeartbeatTimeout:       time.Second * 20,
		}
		ctx = WithActivityOptions(ctx, ao)
		sessionCtx := setSessionInfo(ctx, &SessionInfo{
			SessionID:    "some random sessionID",
			tasklist:     "some random tasklist",
			sessionState: sessionStateFailed,
		})

		sessionCtx, err := CreateSession(sessionCtx)
		if err != nil {
			return err
		}
		return CompleteSession(sessionCtx)
	}

	RegisterWorkflow(workflowFn)
	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *SessionTestSuite) TestCompletionWithClosedSessionContext() {
	workflowFn := func(ctx Context) error {
		sessionCtx := setSessionInfo(ctx, &SessionInfo{
			SessionID:    "some random sessionID",
			tasklist:     "some random tasklist",
			sessionState: sessionStateClosed,
		})
		return CompleteSession(sessionCtx)
	}

	RegisterWorkflow(workflowFn)
	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.Equal(errNoOpenSession.Error(), env.GetWorkflowError().Error())
}

func (s *SessionTestSuite) TestCompletionWithFailedSessionContext() {
	workflowFn := func(ctx Context) error {
		sessionCtx := setSessionInfo(ctx, &SessionInfo{
			SessionID:    "some random sessionID",
			tasklist:     "some random tasklist",
			sessionState: sessionStateFailed,
		})
		return CompleteSession(sessionCtx)
	}

	RegisterWorkflow(workflowFn)
	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *SessionTestSuite) TestGetSessionInfo() {
	workflowFn := func(ctx Context) error {
		_, err := GetSessionInfo(ctx)
		if err == nil {
			return errors.New("GetSessionInfo should return error when there's no such info")
		}

		sessionCtx := setSessionInfo(ctx, &SessionInfo{
			SessionID:    "some random sessionID",
			tasklist:     "some random tasklist",
			sessionState: sessionStateFailed,
		})
		_, err = GetSessionInfo(sessionCtx)
		if err != nil {
			return err
		}

		newSessionInfo := &SessionInfo{
			SessionID:    "another sessionID",
			tasklist:     "another tasklist",
			sessionState: sessionStateClosed,
		}
		sessionCtx = setSessionInfo(ctx, newSessionInfo)
		info, err := GetSessionInfo(sessionCtx)
		if err != nil {
			return err
		}
		if info != *newSessionInfo {
			return errors.New("GetSessionInfo should return info for the most recent session in the context")
		}
		return nil
	}

	RegisterWorkflow(workflowFn)
	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *SessionTestSuite) TestRecreation() {
	workflowFn := func(ctx Context) error {
		ao := ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
			HeartbeatTimeout:       time.Second * 20,
		}
		ctx = WithActivityOptions(ctx, ao)
		sessionInfo := &SessionInfo{
			SessionID:    "some random sessionID",
			tasklist:     "some random tasklist",
			sessionState: sessionStateFailed,
		}

		sessionCtx, err := RecreateSession(ctx, sessionInfo)
		if err != nil {
			return err
		}
		return CompleteSession(sessionCtx)
	}

	RegisterWorkflow(workflowFn)
	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *SessionTestSuite) TestMaxConcurrentSession_CreationOnly() {
	maxConCurrentSessionExecutionSize := 3
	workflowFn := func(ctx Context) error {
		ao := ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
			HeartbeatTimeout:       time.Second * 20,
		}
		ctx = WithActivityOptions(ctx, ao)
		for i := 0; i != maxConCurrentSessionExecutionSize+1; i++ {
			if _, err := s.createSessionWithoutRetry(ctx); err != nil {
				return err
			}
		}
		return nil
	}

	RegisterWorkflow(workflowFn)
	env := s.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(WorkerOptions{
		MaxConCurrentSessionExecutionSize: maxConCurrentSessionExecutionSize,
	})
	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.Equal(errTooManySessionsMsg, env.GetWorkflowError().Error())
}

func (s *SessionTestSuite) TestMaxConcurrentSession_WithRecreation() {
	maxConCurrentSessionExecutionSize := 3
	workflowFn := func(ctx Context) error {
		ao := ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
			HeartbeatTimeout:       time.Second * 20,
		}
		ctx = WithActivityOptions(ctx, ao)
		sessionCtx, err := CreateSession(ctx)
		if err != nil {
			return err
		}
		sessionInfo, err := GetSessionInfo(sessionCtx)
		if err != nil {
			return nil
		}

		for i := 0; i != maxConCurrentSessionExecutionSize; i++ {
			if i%2 == 0 {
				_, err = s.createSessionWithoutRetry(ctx)
			} else {
				_, err = RecreateSession(ctx, &sessionInfo)
			}
			if err != nil {
				return err
			}
		}
		return nil
	}

	RegisterWorkflow(workflowFn)
	env := s.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(WorkerOptions{
		MaxConCurrentSessionExecutionSize: maxConCurrentSessionExecutionSize,
	})
	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.Equal(errTooManySessionsMsg, env.GetWorkflowError().Error())
}

func (s *SessionTestSuite) TestSessionTaskList() {
	numActivities := 3
	workflowFn := func(ctx Context) error {
		ao := ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
			HeartbeatTimeout:       time.Second * 20,
		}
		ctx = WithActivityOptions(ctx, ao)
		sessionCtx, err := CreateSession(ctx)
		if err != nil {
			return err
		}

		fmt.Println("session created")
		for i := 0; i != numActivities; i++ {
			if err := ExecuteActivity(sessionCtx, testSessionActivity, "a random name").Get(sessionCtx, nil); err != nil {
				return err
			}
		}

		return CompleteSession(sessionCtx)
	}

	RegisterWorkflow(workflowFn)
	env := s.NewTestWorkflowEnvironment()
	taskListUsed := []string{}
	env.SetOnActivityStartedListener(func(activityInfo *ActivityInfo, ctx context.Context, args encoded.Values) {
		taskListUsed = append(taskListUsed, activityInfo.TaskList)
	})
	workerIdentity := "testWorker"
	resourceID := "testResourceID"
	env.SetWorkerOptions(WorkerOptions{
		Identity:          workerIdentity,
		SessionResourceID: resourceID,
	})
	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.Equal(numActivities+2, len(taskListUsed))
	s.Equal(getCreationTasklist(defaultTestTaskList), taskListUsed[0])
	expectedTaskList := getResourceSpecificTasklist(workerIdentity, resourceID)
	for _, taskList := range taskListUsed[1:] {
		s.Equal(expectedTaskList, taskList)
	}
}

func (s *SessionTestSuite) TestSessionRecreationTaskList() {
	numActivities := 3
	workerIdentity := "testWorker"
	resourceID := "testResourceID"
	resourceSpecificTaskList := getResourceSpecificTasklist(workerIdentity, resourceID)
	workflowFn := func(ctx Context) error {
		ao := ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
			HeartbeatTimeout:       time.Second * 20,
		}
		ctx = WithActivityOptions(ctx, ao)

		sessionInfo := &SessionInfo{
			SessionID:    "testSessionID",
			tasklist:     resourceSpecificTaskList,
			sessionState: sessionStateClosed,
		}
		sessionCtx, err := RecreateSession(ctx, sessionInfo)
		if err != nil {
			return err
		}

		for i := 0; i != numActivities; i++ {
			if err := ExecuteActivity(sessionCtx, testSessionActivity, "a random name").Get(sessionCtx, nil); err != nil {
				return err
			}
		}

		return CompleteSession(sessionCtx)
	}

	RegisterWorkflow(workflowFn)
	env := s.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(WorkerOptions{
		Identity:          workerIdentity,
		SessionResourceID: resourceID,
	})
	taskListUsed := []string{}
	env.SetOnActivityStartedListener(func(activityInfo *ActivityInfo, ctx context.Context, args encoded.Values) {
		taskListUsed = append(taskListUsed, activityInfo.TaskList)
	})
	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.Equal(numActivities+2, len(taskListUsed))
	for _, taskList := range taskListUsed {
		s.Equal(resourceSpecificTaskList, taskList)
	}
}

func (s *SessionTestSuite) TestExecuteActivityInFailedSession() {
	workflowFn := func(ctx Context) error {
		ao := ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
			HeartbeatTimeout:       time.Second * 20,
		}
		ctx = WithActivityOptions(ctx, ao)
		sessionCtx := setSessionInfo(ctx, &SessionInfo{
			SessionID:    "random sessionID",
			tasklist:     "random tasklist",
			sessionState: sessionStateFailed,
		})

		return ExecuteActivity(sessionCtx, testSessionActivity, "a random name").Get(sessionCtx, nil)
	}

	RegisterWorkflow(workflowFn)
	env := s.NewTestWorkflowEnvironment()

	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.Equal(errSessionFailed.Error(), env.GetWorkflowError().Error())
}

func (s *SessionTestSuite) TestExecuteActivityInClosedSession() {
	workflowFn := func(ctx Context) error {
		ao := ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
			HeartbeatTimeout:       time.Second * 20,
		}
		ctx = WithActivityOptions(ctx, ao)
		sessionCtx := setSessionInfo(ctx, &SessionInfo{
			SessionID:    "random sessionID",
			tasklist:     "random tasklist",
			sessionState: sessionStateClosed,
		})

		return ExecuteActivity(sessionCtx, testActivityHello, "some random message").Get(sessionCtx, nil)
	}

	RegisterWorkflow(workflowFn)
	env := s.NewTestWorkflowEnvironment()
	var taskListUsed string
	env.SetOnActivityStartedListener(func(activityInfo *ActivityInfo, ctx context.Context, args encoded.Values) {
		taskListUsed = activityInfo.TaskList
	})

	env.ExecuteWorkflow(workflowFn)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.Equal(defaultTestTaskList, taskListUsed)
}

func (s *SessionTestSuite) createSessionWithoutRetry(ctx Context) (Context, error) {
	options := getActivityOptions(ctx)
	baseTasklist := options.TaskListName
	if baseTasklist == "" {
		baseTasklist = options.OriginalTaskListName
	}
	sessionID, err := generateSessionID(ctx)
	if err != nil {
		return nil, err
	}
	return createSession(ctx, sessionID, getCreationTasklist(baseTasklist), false)
}

func testSessionActivity(ctx context.Context, name string) (string, error) {
	return "Hello" + name + "!", nil
}
