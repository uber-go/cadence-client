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
	"math"
	"sync"
	"time"

	"github.com/pborman/uuid"
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

	errTooManySessionsMsg string = "too many outstanding sessions"
)

var (
	errNoOpenSession            = errors.New("no open session in the context")
	errNoSessionInfo            = errors.New("no session information found in the context")
	errFoundExistingOpenSession = errors.New("found exisiting open session in the context")
	errSessionFailed            = errors.New("session has failed")
)

// CreateSession create a session
func CreateSession(ctx Context) (Context, error) {
	options := getActivityOptions(ctx)
	baseTasklist := options.TaskListName
	if baseTasklist == "" {
		baseTasklist = options.OriginalTaskListName
	}
	return createSession(ctx, getCreationTasklist(baseTasklist), true)
}

// RecreateSession recreate a session
func RecreateSession(ctx Context, sessionInfo *SessionInfo) (Context, error) {
	return createSession(ctx, sessionInfo.tasklist, false)
}

// CompleteSession complete a session
func CompleteSession(ctx Context) error {
	sessionInfo := getSessionInfo(ctx)
	if sessionInfo == nil || sessionInfo.sessionState == sessionStateClosed {
		return errNoOpenSession
	}

	if sessionInfo.sessionState == sessionStateFailed {
		return nil
	}
	retryPolicy := &RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 1.1,
		MaximumInterval:    time.Second * 10,
		MaximumAttempts:    5,
	}
	ao := ActivityOptions{
		ScheduleToStartTimeout: time.Second * 10,
		StartToCloseTimeout:    time.Second * 10,
		RetryPolicy:            retryPolicy,
	}
	completionCtx := WithActivityOptions(ctx, ao)
	// the tasklist will be overrided to use the one stored in sessionInfo
	err := ExecuteActivity(completionCtx, sessionCompletionActivityName, sessionInfo.SessionID).Get(ctx, nil)
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

func createSession(ctx Context, creationTasklist string, retryable bool) (Context, error) {
	if prevSessionInfo := getSessionInfo(ctx); prevSessionInfo != nil && prevSessionInfo.sessionState == sessionStateOpen {
		return nil, errFoundExistingOpenSession
	}
	sessionID, err := generateSessionID(ctx)
	if err != nil {
		return nil, err
	}

	tasklistChan := GetSignalChannel(ctx, sessionID) // use sessionID as channel name
	// Retry is only needed when creating new session and the error returned is NewCustomError(errTooManySessionsMsg)
	retryPolicy := &RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 1.1,
		MaximumInterval:    time.Second * 10,
		MaximumAttempts:    10,
		NonRetriableErrorReasons: []string{
			"cadenceInternal:Panic",
			"cadenceInternal:Generic",
			"cadenceInternal:Timeout START_TO_CLOSE",
			"cadenceInternal:Timeout HEARTBEAT",
		},
	}
	// Creation activity need an infinite StartToCloseTimeout.
	scheduleToStartTimeout := time.Second * 10
	startToCloseTimeout := time.Second*math.MaxInt32 - scheduleToStartTimeout
	HeartbeatTimeout := time.Second * 20
	ao := ActivityOptions{
		TaskList:               creationTasklist,
		ScheduleToStartTimeout: scheduleToStartTimeout,
		StartToCloseTimeout:    startToCloseTimeout,
		HeartbeatTimeout:       HeartbeatTimeout,
	}
	if retryable {
		ao.RetryPolicy = retryPolicy
	}
	creationCtx := WithActivityOptions(ctx, ao)
	creationFuture := ExecuteActivity(creationCtx, sessionCreationActivityName, sessionID)

	var creationErr error
	var tasklist string
	s := NewSelector(ctx)
	s.AddReceive(tasklistChan, func(c Channel, more bool) {
		c.Receive(ctx, &tasklist)
	})
	s.AddFuture(creationFuture, func(f Future) {
		// activity stoped before signal is received, must be creation timeout.
		creationErr = f.Get(ctx, nil)
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

func generateSessionID(ctx Context) (string, error) {
	var sessionID string
	err := SideEffect(ctx, func(ctx Context) interface{} {
		return uuid.New()
	}).Get(&sessionID)
	return sessionID, err
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
		return NewCustomError(errTooManySessionsMsg)
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
				sessionEnv.closeAndDeleteDoneChan(sessionID)
				return err
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
	sessionEnv.closeAndDeleteDoneChan(sessionID)
	return nil
}

func isSessionCreationActivity(activity interface{}) bool {
	activityName, ok := activity.(string)
	return ok && activityName == sessionCreationActivityName
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

func (env *sessionEnvironment) closeAndDeleteDoneChan(sessionID string) {
	env.Lock()
	defer env.Unlock()

	if doneChan, ok := env.doneChanMap[sessionID]; ok {
		delete(env.doneChanMap, sessionID)
		close(doneChan)
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
