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

package workflow

import (
	"go.uber.org/cadence/internal"
)

// SessionInfo contains information of a created session. For now, the only exported
// field is SessionID which has string type. SessionID is a uuid generated when
// CreateSession() or RecreateSession() is called and can be used to uniquely identify
// a session.
type SessionInfo = internal.SessionInfo

// Note: Worker should be configured to process session. To do this, set the following
// fields in WorkerOptions:
//     EnableSessionWorker: true
//     SessionResourceID: The identifier of the resource consumed by sessions.
//         It's the user's responsibility to ensure there's only one worker using this resourceID.
//     MaxConCurrentSessionExecutionSize: the maximum number of concurrently sessions the resource
//         support. By default, 1000 is used.

// CreateSession creates a session and return a new context which contains information
// of the created session. The session will be created on the tasklist user specified in
// ActivityOptions. If none is specified, the default one will be used.
//
// CreationSession will fail in the following situations:
//     1. The context passed in already contains a session which is still open
//        (not closed and failed).
//     2. All the workers are busy (number of sessions currently running on all the workers have reached
//        MaxConCurrentSessionExecutionSize, which is specified when starting the workers) and session
//        cannot be created within a specified timeout.
//
// If an activity is executed using the returned context, it's regarded as part of the
// session and all activities within a session will be executed by the same worker.
// User still needs to handle the error returned when executing an activity. Session will
// not be marked as failed if an activity within it returns an error. Only when the worker
// executing the session is down, that session will be marked as failed. Executing an activity
// within a failed session will return an session failed error immediately without even scheduling the activity.
//
// If user wants to end a session since some activity returns an error, use the CompleteSession API below,
// and new session can be created if necessary to retry the whole session.
//
// Example:
//    sessionCtx, err := CreateSession(ctx)
//    if err != nil {
//		    // Creation failed. Wrong ctx or too many outstanding sessions.
//    }
//    err = ExecuteActivity(sessionCtx, someActivityFunc, activityInput).Get(sessionCtx, nil)
//    if err != nil {
//        // Session has failed or activity itself failed.
//    }
//    ... // execute more activities using sessionCtx
//    err = CompleteSession(sessionCtx)
//    if err != nil {
//        // Wrong ctx is used or failed to release session resource.
//    }
func CreateSession(ctx Context) (Context, error) {
	return internal.CreateSession(ctx)
}

// CreateSessionForResourceID recreate a session based on the sessionInfo passed in. Activities executed within
// the recreated session will be executed by the same worker as the previous session. CreateSessionForResourceID()
// returns an error under the same situation as CreateSession() and has the same usage as CreateSession().
// It will not check the state of the session described by the sessionInfo passed in, so user can recreate
// a session based on a failed or completed session.
//
// The main usage of CreateSessionForResourceID is for long sessions that are splited into multiple runs. At the end of
// one run, complete the current session, get sessionInfo from the context and pass the information to the
// next run. In the new run, the session can be recreated on the information.
func CreateSessionForResourceID(ctx Context, sessionInfo *SessionInfo) (Context, error) {
	return internal.CreateSessionForResourceID(ctx, sessionInfo)
}

// CompleteSession completes a session. It releases worker resources, so other sessions can be created.
//
// After a session is completed, user can continue to use the context, but the activities will be scheduled
// on the normal taskList (as user specified in ActivityOptions) and may be picked up by another worker since
// it's not in a session.
func CompleteSession(ctx Context) {
	internal.CompleteSession(ctx)
}

// GetSessionInfo returns the sessionInfo stored in the context. If there are multiple sessions in the context,
// (for example, the same context is used to create, complete, create another session. Then user found that the
// session has failed, and created a new one on it), the most recent sessionInfo will be returned.
//
// This API will return nil if there's no sessionInfo in the context.
func GetSessionInfo(ctx Context) *SessionInfo {
	return internal.GetSessionInfo(ctx)
}
