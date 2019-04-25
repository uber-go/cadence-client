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

type sessionState string

// SessionInfo contains information for a created session
// we may also want to expose resourceID
type SessionInfo struct {
	SessionID    string
	tasklist     string // tasklist is resource specific
	sessionState sessionState
}

const (
	sessionStateOpen   sessionState = "open"
	sessionStateFailed sessionState = "failed"
	sessionStateClosed sessionState = "closed"

	sessionInfoContextKey contextKey = "sessionInfo"

	sessionCreationActivityName   string = "internalSessionCreationActivity"
	sessionCompletionActivityName string = "internalSessionCompletionActivity"
)

// CreateSession create a session
func CreateSession(ctx Context) (Context, error) {
	return createSession(ctx, generateSessionID(ctx), getCreationTasklist(getActivityOptions(ctx).TaskListName))
}

// RecreateSession recreate a session
func RecreateSession(ctx Context, sessionInfo *SessionInfo) (Context, error) {
	return createSession(ctx, sessionInfo.SessionID, sessionInfo.tasklist)
}

// CompleteSession complete a session
func CompleteSession(ctx Context) error {
	sessionInfo := getSessionInfo(ctx)
	if sessionInfo == nil || sessionInfo.sessionState != sessionStateOpen {
		return errors.New("no open session in the context")
	}

	// the tasklist will be overrided to use the one stored in sessionInfo
	return ExecuteActivity(ctx, sessionCompletionActivityName, sessionInfo.SessionID).Get(ctx, nil)
}

// GetSessionInfo returns session information
func GetSessionInfo(ctx Context) (SessionInfo, error) {
	info := getSessionInfo(ctx)
	if info == nil {
		return SessionInfo{}, errors.New("no session information found in the context")
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
		return nil, errors.New("found exisiting open session in the context")
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
		creationErr = f.Get(ctx, nil)
		// ycyang TODO: provide better error message
		creationErr = errors.New("failed to create session: " + creationErr.Error())
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
			return
		}
		sessionInfo.sessionState = sessionStateClosed
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
	return identity + "_" + resourceID
}

const sessionWorkerInfoContextKey contextKey = "sessionWorkerInfo"

type sessionWorkerInfo struct {
	doneChanMap              *sessionDoneChanMap
	resourceID               string
	resourceSpecificTasklist string
}

type sessionDoneChanMap struct {
	sync.Mutex
	doneChanMap map[string]chan struct{}
}

func newSessionDoneChanMap() *sessionDoneChanMap {
	m := &sessionDoneChanMap{}
	m.doneChanMap = make(map[string]chan struct{})
	return m
}

func (d *sessionDoneChanMap) getAndDelete(sessionID string) chan struct{} {
	d.Lock()
	defer d.Unlock()
	doneChan, ok := d.doneChanMap[sessionID]
	if !ok {
		return nil
	}
	delete(d.doneChanMap, sessionID)
	return doneChan
}

func (d *sessionDoneChanMap) put(sessionID string, doneChan chan struct{}) {
	d.Lock()
	defer d.Unlock()
	d.doneChanMap[sessionID] = doneChan
}

func sessionCreationActivity(ctx context.Context, sessionID string) error {
	workerInfo, ok := ctx.Value(sessionWorkerInfoContextKey).(*sessionWorkerInfo)
	if !ok {
		return errors.New("no session worker info in user context")
	}

	doneChan := make(chan struct{})
	workerInfo.doneChanMap.put(sessionID, doneChan)

	activityEnv := getActivityEnv(ctx)
	invoker := activityEnv.serviceInvoker

	data, err := encodeArg(getDataConverterFromActivityCtx(ctx), workerInfo.resourceSpecificTasklist)
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
	workerInfo, ok := ctx.Value(sessionWorkerInfoContextKey).(*sessionWorkerInfo)
	if !ok {
		return errors.New("no session worker info in user context")
	}
	doneChan := workerInfo.doneChanMap.getAndDelete(sessionID)
	if doneChan != nil {
		close(doneChan)
	}
	return nil
}
