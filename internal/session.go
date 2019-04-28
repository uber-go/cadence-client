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
	"sync"
	"time"
)

func init() {
	RegisterActivityWithOptions(sessionCreationActivity, RegisterActivityOptions{
		Name: sessionCreationActivityName,
	})
	RegisterActivityWithOptions(sessionCompletionActivity, RegisterActivityOptions{
		Name: sessionCompletionActivityName,
	})
}

type (
	sessionState string

	// SessionInfo contains information for a created session
	SessionInfo struct {
		SessionID    string
		tasklist     string // tasklist is resource specific
		sessionState sessionState
	}

	sessionToken struct {
		*sync.Cond
		availableToken int
	}

	sessionEnvironment struct {
		*sync.Mutex
		doneChanMap              map[string]chan struct{}
		resourceID               string
		resourceSpecificTasklist string
		sessionToken             *sessionToken
		testEnv                  *testWorkflowEnvironmentImpl
	}
)

const (
	sessionStateOpen   sessionState = "open"
	sessionStateFailed sessionState = "failed"
	sessionStateClosed sessionState = "closed"

	sessionInfoContextKey        contextKey = "sessionInfo"
	sessionEnvironmentContextKey contextKey = "sessionEnvironment"

	sessionCreationActivityName   string = "internalSessionCreationActivity"
	sessionCompletionActivityName string = "internalSessionCompletionActivity"
)

var (
	errNoOpenSession            = errors.New("no open session in the context")
	errNoSessionInfo            = errors.New("no session information found in the context")
	errFoundExistingOpenSession = errors.New("found exisiting open session in the context")
	errTooManySessions          = errors.New("too many outstanding sessions")
)

// CreateSession create a session
func CreateSession(ctx Context) (Context, error) {
	options := getActivityOptions(ctx)
	baseTasklist := options.TaskListName
	if baseTasklist == "" {
		baseTasklist = options.OriginalTaskListName
	}
	return createSession(ctx, generateSessionID(ctx), getCreationTasklist(baseTasklist))
}

// RecreateSession recreate a session
func RecreateSession(ctx Context, sessionInfo *SessionInfo) (Context, error) {
	return createSession(ctx, sessionInfo.SessionID, sessionInfo.tasklist)
}

// CompleteSession complete a session
func CompleteSession(ctx Context) error {
	sessionInfo := getSessionInfo(ctx)
	if sessionInfo == nil || sessionInfo.sessionState != sessionStateOpen {
		return errNoOpenSession
	}

	// the tasklist will be overrided to use the one stored in sessionInfo
	err := ExecuteActivity(ctx, sessionCompletionActivityName, sessionInfo.SessionID).Get(ctx, nil)
	if err == nil {
		sessionInfo.sessionState = sessionStateClosed
	} else {
		sessionInfo.sessionState = sessionStateFailed
	}
	return err
}

// GetSessionInfo returns session information
func GetSessionInfo(ctx Context) (SessionInfo, error) {
	info := getSessionInfo(ctx)
	if info == nil {
		return SessionInfo{}, errNoSessionInfo
	}
	return *info, nil
}

func getSessionInfo(ctx Context) *SessionInfo {
	info := ctx.Value(sessionInfoContextKey)
	if info == nil {
		return nil
	}
	return info.(*SessionInfo)
}

func setSessionInfo(ctx Context, sessionInfo *SessionInfo) Context {
	return WithValue(ctx, sessionInfoContextKey, sessionInfo)
}

func createSession(ctx Context, sessionID string, creationTasklist string) (Context, error) {
	if prevSessionInfo := getSessionInfo(ctx); prevSessionInfo != nil && prevSessionInfo.sessionState == sessionStateOpen {
		return nil, errFoundExistingOpenSession
	}

	tasklistChan := GetSignalChannel(ctx, sessionID) // use sessionID as channel name
	childCtx := WithTaskList(ctx, creationTasklist)
	creationFuture := ExecuteActivity(childCtx, sessionCreationActivityName, sessionID)

	var creationErr error
	var tasklist string
	s := NewSelector(ctx)
	s.AddReceive(tasklistChan, func(c Channel, more bool) {
		c.Receive(ctx, &tasklist)
	})
	s.AddFuture(creationFuture, func(f Future) {
		// activity stoped before signal is received, must be creation timeout.
		creationErr = errors.New("failed to create session: " + f.Get(ctx, nil).Error())
	})
	s.Select(ctx)

	if creationErr != nil {
		return nil, creationErr
	}

	sessionInfo := &SessionInfo{
		SessionID:    sessionID,
		tasklist:     tasklist,
		sessionState: sessionStateOpen,
	}

	Go(ctx, func(ctx Context) {
		if creationFuture.Get(ctx, nil) != nil {
			sessionInfo.sessionState = sessionStateFailed
		}
	})

	return setSessionInfo(ctx, sessionInfo), nil
}

func generateSessionID(ctx Context) string {
	// ycyang TODO: figure out how to generate sessionID which should be deterministic and unique
	// we can also use the side effect API
	return getWorkflowEnvironment(ctx).WorkflowInfo().WorkflowExecution.RunID + Now(ctx).String()
}

func getCreationTasklist(base string) string {
	return base + "__internal_session_creation"
}

func getResourceSpecificTasklist(identity, resourceID string) string {
	return resourceID + "@" + identity
}

func sessionCreationActivity(ctx context.Context, sessionID string) error {
	sessionEnv, ok := ctx.Value(sessionEnvironmentContextKey).(*sessionEnvironment)
	if !ok {
		panic("no session environment in context")
	}
	sessionEnv.Lock()
	if sessionEnv.sessionToken.availableToken == 0 {
		sessionEnv.Unlock()
		return errTooManySessions
	}

	sessionEnv.sessionToken.availableToken--
	doneChan := make(chan struct{})
	sessionEnv.doneChanMap[sessionID] = doneChan
	sessionEnv.Unlock()

	defer func() {
		sessionEnv.sessionToken.addToken()
	}()

	activityEnv := getActivityEnv(ctx)

	if sessionEnv.testEnv == nil {
		invoker := activityEnv.serviceInvoker

		data, err := encodeArg(getDataConverterFromActivityCtx(ctx), sessionEnv.resourceSpecificTasklist)
		if err != nil {
			return err
		}

		err = invoker.SignalWorkflow(activityEnv.workflowDomain,
			activityEnv.workflowExecution.ID,
			activityEnv.workflowExecution.RunID,
			sessionID,
			data)
		if err != nil {
			return err
		}
	} else {
		sessionEnv.testEnv.signalWorkflow(sessionID, sessionEnv.resourceSpecificTasklist)
	}

	ticker := time.NewTicker(activityEnv.heartbeatTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := activityEnv.serviceInvoker.Heartbeat([]byte{})
			if err != nil {
				return nil
			}
		case <-doneChan:
			return nil
		}
	}
}

func sessionCompletionActivity(ctx context.Context, sessionID string) error {
	sessionEnv, ok := ctx.Value(sessionEnvironmentContextKey).(*sessionEnvironment)
	if !ok {
		panic("no session environment in context")
	}

	sessionEnv.Lock()
	defer sessionEnv.Unlock()

	if doneChan, ok := sessionEnv.doneChanMap[sessionID]; ok {
		delete(sessionEnv.doneChanMap, sessionID)
		close(doneChan)
	}
	return nil
}

func (t *sessionToken) waitForAvailableToken() {
	t.L.Lock()
	for t.availableToken == 0 {
		t.Wait()
	}
	t.L.Unlock()
}

func (t *sessionToken) addToken() {
	t.L.Lock()
	t.availableToken++
	t.L.Unlock()
	t.Signal()
}

func newSessionEnvironment(identity, resourceID string, concurrentSessionExecutionSize int) *sessionEnvironment {
	resourceSpecificTasklist := getResourceSpecificTasklist(identity, resourceID)
	sessionMutex := &sync.Mutex{}
	sessionToken := &sessionToken{
		Cond:           sync.NewCond(sessionMutex),
		availableToken: concurrentSessionExecutionSize,
	}
	return &sessionEnvironment{
		Mutex:                    sessionMutex,
		doneChanMap:              make(map[string]chan struct{}),
		resourceID:               resourceID,
		resourceSpecificTasklist: resourceSpecificTasklist,
		sessionToken:             sessionToken,
	}
}

func getTestSessionEnvironment(params *workerExecutionParameters, concurrentSessionExecutionSize int) *sessionEnvironment {
	resourceID := params.SessionResourceID
	if resourceID == "" {
		resourceID = "testResourceID"
	}
	if concurrentSessionExecutionSize == 0 {
		concurrentSessionExecutionSize = defaultMaxConcurrentSeesionExecutionSize
	}

	return newSessionEnvironment(params.Identity, resourceID, concurrentSessionExecutionSize)
}
