// Copyright (c) 2017-2021 Uber Technologies Inc.
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
	"github.com/stretchr/testify/assert"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"testing"
)

type matchReplayWithHistoryTestCase struct {
	name            string
	replayDecisions []*shared.Decision
	historyEvents   []*shared.HistoryEvent
	expectError     bool
}

func TestMatchReplayWithHistory(t *testing.T) {
	testCases := []matchReplayWithHistoryTestCase{
		{
			name: "Success matching",
			replayDecisions: []*shared.Decision{
				mockDecision(shared.DecisionTypeScheduleActivityTask),
			},
			historyEvents: []*shared.HistoryEvent{
				mockHistoryEvent(shared.EventTypeActivityTaskScheduled),
			},
			expectError: false,
		},
		{
			name: "Mismatch error",
			replayDecisions: []*shared.Decision{
				mockDecision(shared.DecisionTypeStartTimer),
			},
			historyEvents: []*shared.HistoryEvent{
				mockHistoryEvent(shared.EventTypeTimerCanceled),
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info := &WorkflowInfo{} // Assuming WorkflowInfo is a struct you have defined
			err := matchReplayWithHistory(info, tc.replayDecisions, tc.historyEvents)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func mockHistoryEvent(eventType shared.EventType) *shared.HistoryEvent {
	event := &shared.HistoryEvent{
		EventType: &eventType,
	}

	switch eventType {
	case shared.EventTypeActivityTaskScheduled:
		event.ActivityTaskScheduledEventAttributes = &shared.ActivityTaskScheduledEventAttributes{
			ActivityId:   common.StringPtr("mockActivityId"),
			ActivityType: &shared.ActivityType{Name: common.StringPtr("mockActivityType")},
			TaskList:     &shared.TaskList{Name: common.StringPtr("mockTaskList")},
			Input:        []byte("mockInput"),
		}
	case shared.EventTypeTimerCanceled: // Newly added case
		event.TimerCanceledEventAttributes = &shared.TimerCanceledEventAttributes{
			TimerId: common.StringPtr("mockTimerId"),
		}
	}

	return event
}

func mockDecision(decisionType shared.DecisionType) *shared.Decision {
	decision := &shared.Decision{
		DecisionType: &decisionType,
	}

	switch decisionType {
	case shared.DecisionTypeScheduleActivityTask:
		decision.ScheduleActivityTaskDecisionAttributes = &shared.ScheduleActivityTaskDecisionAttributes{
			ActivityId:   common.StringPtr("mockActivityId"),
			ActivityType: &shared.ActivityType{Name: common.StringPtr("mockActivityType")},
			TaskList:     &shared.TaskList{Name: common.StringPtr("mockTaskList")},
			Input:        []byte("mockInput"),
		}
	case shared.DecisionTypeStartTimer: // Newly added case
		decision.StartTimerDecisionAttributes = &shared.StartTimerDecisionAttributes{
			TimerId:                   common.StringPtr("mockTimerId"),
			StartToFireTimeoutSeconds: common.Int64Ptr(60), // Example value
		}
	}

	return decision
}
