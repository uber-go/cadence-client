// Copyright (c) 2017-2020 Uber Technologies Inc.
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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/cadence/workflowservicetest"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/internal/common"
	"go.uber.org/cadence/internal/common/serializer"
	"go.uber.org/zap"
)

// WorkflowReplayer is used to replay workflow code from an event history
type WorkflowReplayer struct {
	registry *registry
}

// NewWorkflowReplayer creates an instance of the WorkflowReplayer
func NewWorkflowReplayer() *WorkflowReplayer {
	return &WorkflowReplayer{registry: newRegistry()}
}

// RegisterWorkflow registers workflow function to replay
func (r *WorkflowReplayer) RegisterWorkflow(w interface{}) {
	r.registry.RegisterWorkflow(w)
}

// RegisterWorkflowWithOptions registers workflow function with custom workflow name to replay
func (r *WorkflowReplayer) RegisterWorkflowWithOptions(w interface{}, options RegisterWorkflowOptions) {
	r.registry.RegisterWorkflowWithOptions(w, options)
}

// ReplayWorkflowHistory executes a single decision task for the given history.
// Use for testing backwards compatibility of code changes and troubleshooting workflows in a debugger.
// The logger is an optional parameter. Defaults to the noop logger.
func (r *WorkflowReplayer) ReplayWorkflowHistory(logger *zap.Logger, history *shared.History) error {
	if logger == nil {
		logger = zap.NewNop()
	}

	testReporter := logger.Sugar()
	controller := gomock.NewController(testReporter)
	service := workflowservicetest.NewMockClient(controller)

	return r.replayWorkflowHistory(logger, service, ReplayDomainName, history)
}

// ReplayWorkflowHistoryFromJSONFile executes a single decision task for the given json history file.
// Use for testing the backwards compatibility of code changes and troubleshooting workflows in a debugger.
// The logger is an optional parameter. Defaults to the noop logger.
func (r *WorkflowReplayer) ReplayWorkflowHistoryFromJSONFile(logger *zap.Logger, jsonfileName string) error {
	return r.ReplayPartialWorkflowHistoryFromJSONFile(logger, jsonfileName, 0)
}

// ReplayPartialWorkflowHistoryFromJSONFile executes a single decision task for the given json history file upto provided
// lastEventID(inclusive).
// Use for testing backwards compatibility of code changes and troubleshooting workflows in a debugger.
// The logger is an optional parameter. Defaults to the noop logger.
func (r *WorkflowReplayer) ReplayPartialWorkflowHistoryFromJSONFile(logger *zap.Logger, jsonfileName string, lastEventID int64) error {
	history, err := extractHistoryFromFile(jsonfileName, lastEventID)

	if err != nil {
		return err
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	testReporter := logger.Sugar()
	controller := gomock.NewController(testReporter)
	service := workflowservicetest.NewMockClient(controller)

	return r.replayWorkflowHistory(logger, service, ReplayDomainName, history)
}

// ReplayWorkflowExecution replays workflow execution loading it from Temporal service.
func (r *WorkflowReplayer) ReplayWorkflowExecution(
	ctx context.Context,
	service workflowserviceclient.Interface,
	logger *zap.Logger,
	domain string,
	execution WorkflowExecution,
) error {
	sharedExecution := &shared.WorkflowExecution{
		RunId:      common.StringPtr(execution.RunID),
		WorkflowId: common.StringPtr(execution.ID),
	}
	request := &shared.GetWorkflowExecutionHistoryRequest{
		Domain:    common.StringPtr(domain),
		Execution: sharedExecution,
	}
	hResponse, err := service.GetWorkflowExecutionHistory(ctx, request)
	if err != nil {
		return err
	}

	if hResponse.RawHistory != nil {
		history, err := serializer.DeserializeBlobDataToHistoryEvents(hResponse.RawHistory, shared.HistoryEventFilterTypeAllEvent)
		if err != nil {
			return err
		}

		hResponse.History = history
	}

	return r.replayWorkflowHistory(logger, service, domain, hResponse.History)
}

func (r *WorkflowReplayer) replayWorkflowHistory(
	logger *zap.Logger,
	service workflowserviceclient.Interface,
	domain string,
	history *shared.History,
) error {
	taskList := "ReplayTaskList"
	events := history.Events
	if events == nil {
		return errors.New("empty events")
	}
	if len(events) < 3 {
		return errors.New("at least 3 events expected in the history")
	}
	first := events[0]
	if first.GetEventType() != shared.EventTypeWorkflowExecutionStarted {
		return errors.New("first event is not WorkflowExecutionStarted")
	}
	last := events[len(events)-1]

	attr := first.WorkflowExecutionStartedEventAttributes
	if attr == nil {
		return errors.New("corrupted WorkflowExecutionStarted")
	}
	workflowType := attr.WorkflowType
	execution := &shared.WorkflowExecution{
		RunId:      common.StringPtr(uuid.NewRandom().String()),
		WorkflowId: common.StringPtr("ReplayId"),
	}
	if first.WorkflowExecutionStartedEventAttributes.GetOriginalExecutionRunId() != "" {
		execution.RunId = common.StringPtr(first.WorkflowExecutionStartedEventAttributes.GetOriginalExecutionRunId())
	}

	task := &shared.PollForDecisionTaskResponse{
		Attempt:                common.Int64Ptr(0),
		TaskToken:              []byte("ReplayTaskToken"),
		WorkflowType:           workflowType,
		WorkflowExecution:      execution,
		History:                history,
		PreviousStartedEventId: common.Int64Ptr(math.MaxInt64),
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	metricScope := tally.NoopScope
	iterator := &historyIteratorImpl{
		nextPageToken: task.NextPageToken,
		execution:     task.WorkflowExecution,
		domain:        ReplayDomainName,
		service:       service,
		metricsScope:  metricScope,
		maxEventID:    task.GetStartedEventId(),
	}
	params := workerExecutionParameters{
		TaskList: taskList,
		Identity: "replayID",
		Logger:   logger,
	}
	taskHandler := newWorkflowTaskHandler(domain, params, nil, r.registry)
	resp, err := taskHandler.ProcessWorkflowTask(&workflowTask{task: task, historyIterator: iterator}, nil)
	if err != nil {
		return err
	}

	if last.GetEventType() != shared.EventTypeWorkflowExecutionCompleted && last.GetEventType() != shared.EventTypeWorkflowExecutionContinuedAsNew {
		return nil
	}
	err = fmt.Errorf("replay workflow doesn't return the same result as the last event, resp: %v, last: %v", resp, last)
	if resp != nil {
		completeReq, ok := resp.(*shared.RespondDecisionTaskCompletedRequest)
		if ok {
			for _, d := range completeReq.Decisions {
				if d.GetDecisionType() == shared.DecisionTypeContinueAsNewWorkflowExecution {
					if last.GetEventType() == shared.EventTypeWorkflowExecutionContinuedAsNew {
						inputA := d.ContinueAsNewWorkflowExecutionDecisionAttributes.Input
						inputB := last.WorkflowExecutionContinuedAsNewEventAttributes.Input
						if bytes.Compare(inputA, inputB) == 0 {
							return nil
						}
					}
				}
				if d.GetDecisionType() == shared.DecisionTypeCompleteWorkflowExecution {
					if last.GetEventType() == shared.EventTypeWorkflowExecutionCompleted {
						resultA := last.WorkflowExecutionCompletedEventAttributes.Result
						resultB := d.CompleteWorkflowExecutionDecisionAttributes.Result
						if bytes.Compare(resultA, resultB) == 0 {
							return nil
						}
					}
				}
			}
		}
	}
	return err
}

func extractHistoryFromFile(jsonfileName string, lastEventID int64) (*shared.History, error) {
	raw, err := ioutil.ReadFile(jsonfileName)
	if err != nil {
		return nil, err
	}

	var deserializedEvents []*shared.HistoryEvent
	err = json.Unmarshal(raw, &deserializedEvents)

	if err != nil {
		return nil, err
	}

	if lastEventID <= 0 {
		return &shared.History{Events: deserializedEvents}, nil
	}

	// Caller is potentially asking for subset of history instead of all history events
	var events []*shared.HistoryEvent
	for _, event := range deserializedEvents {
		events = append(events, event)
		if event.GetEventId() == lastEventID {
			// Copy history upto last event (inclusive)
			break
		}
	}

	return &shared.History{Events: events}, nil
}
