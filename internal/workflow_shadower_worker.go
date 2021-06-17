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
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/pborman/uuid"
	"go.uber.org/cadence/v2/.gen/go/shadower"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
	"go.uber.org/cadence/v2/internal/api"
	"go.uber.org/cadence/v2/internal/common"
	"go.uber.org/cadence/v2/internal/common/backoff"
	"go.uber.org/zap"
)

type (
	shadowWorker struct {
		activityWorker *activityWorker

		service      api.Interface
		domain       string
		taskList     string
		options      ShadowOptions
		logger       *zap.Logger
		featureFlags FeatureFlags
	}
)

func newShadowWorker(
	service api.Interface,
	domain string,
	shadowOptions ShadowOptions,
	params workerExecutionParameters,
	registry *registry,
) *shadowWorker {
	registry.RegisterActivityWithOptions(scanWorkflowActivity, RegisterActivityOptions{
		Name: shadower.ScanWorkflowActivityName,
	})
	registry.RegisterActivityWithOptions(replayWorkflowActivity, RegisterActivityOptions{
		Name:                shadower.ReplayWorkflowActivityName,
		EnableAutoHeartbeat: true,
	})

	replayer := NewWorkflowReplayerWithOptions(ReplayOptions{
		DataConverter:                     params.DataConverter,
		ContextPropagators:                params.ContextPropagators,
		WorkflowInterceptorChainFactories: params.WorkflowInterceptors,
		Tracer:                            params.Tracer,
		FeatureFlags:                      params.FeatureFlags,
	})
	replayer.registry = registry

	ensureRequiredParams(&params)
	if len(params.TaskList) != 0 {
		// include domain name in tasklist to avoid confliction
		// since all shadow workflow will be run in a single system domain
		params.TaskList = generateShadowTaskList(domain, params.TaskList)
		params.MetricsScope = tagScope(params.MetricsScope, tagTaskList, params.TaskList)
		params.Logger = params.Logger.With(zap.String(tagTaskList, params.TaskList))
	}

	params.UserContext = context.WithValue(params.UserContext, serviceClientContextKey, service)
	params.UserContext = context.WithValue(params.UserContext, workflowReplayerContextKey, replayer)

	// data converter, interceptors, context propagators, tracers provided by user is for replay
	// for the actual shadowing workflow use default values.
	// this is required for data converter, for the other three, it should be ok to still use
	// the value provided by user.
	params.DataConverter = getDefaultDataConverter()
	params.WorkflowInterceptors = []WorkflowInterceptorFactory{}
	params.ContextPropagators = []ContextPropagator{}
	params.Tracer = opentracing.NoopTracer{}

	activityWorker := newActivityWorker(
		service,
		shadower.LocalDomainName, // note: this is the system domain for all shadow workflows
		params,
		nil,
		registry,
		nil,
	)
	return &shadowWorker{
		activityWorker: activityWorker,

		service:      service,
		domain:       domain,
		taskList:     params.TaskList,
		options:      shadowOptions,
		logger:       params.Logger,
		featureFlags: params.FeatureFlags,
	}
}

func (sw *shadowWorker) Start() error {
	if err := sw.options.validateAndPopulateFields(); err != nil {
		return err
	}

	if err := verifyDomainExist(sw.service, sw.domain, sw.logger, sw.featureFlags); err != nil {
		return err
	}

	if len(sw.taskList) == 0 {
		return errTaskListNotSet
	}

	if err := sw.startShadowWorkflow(); err != nil {
		return err
	}

	return sw.activityWorker.Start()
}

func (sw *shadowWorker) Stop() {
	sw.activityWorker.Stop()
}

func (sw *shadowWorker) startShadowWorkflow() error {
	workflowParams := shadower.WorkflowParams{
		Domain:        common.StringPtr(sw.domain),
		TaskList:      common.StringPtr(sw.taskList),
		WorkflowQuery: common.StringPtr(sw.options.WorkflowQuery),
		SamplingRate:  common.Float64Ptr(sw.options.SamplingRate),
		ShadowMode:    sw.options.Mode.toThriftPtr(),
		ExitCondition: sw.options.ExitCondition.toThriftPtr(),
		Concurrency:   common.Int32Ptr(int32(sw.options.Concurrency)),
	}

	ctx := context.Background()

	workflowType, input, err := getValidatedWorkflowFunction(shadower.WorkflowName, []interface{}{workflowParams}, getDefaultDataConverter(), nil)
	if err != nil {
		return err
	}

	startWorkflowRequest := &apiv1.StartWorkflowExecutionRequest{
		Domain:       shadower.LocalDomainName,
		WorkflowId:   sw.domain + shadower.WorkflowIDSuffix,
		WorkflowType: &apiv1.WorkflowType{Name: workflowType.Name},
		TaskList: &apiv1.TaskList{
			Name: shadower.TaskList,
		},
		Input:                               &apiv1.Payload{Data: input},
		ExecutionStartToCloseTimeout:        api.SecondsToProto(864000),
		TaskStartToCloseTimeout:             api.SecondsToProto(60),
		RequestId:                           uuid.New(),
		WorkflowIdReusePolicy:               apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}

	startWorkflowOp := func() error {
		tchCtx, cancel, opt := newChannelContext(ctx, sw.featureFlags)
		defer cancel()
		_, err := sw.service.StartWorkflowExecution(tchCtx, startWorkflowRequest, opt...)
		err = api.ConvertError(err)
		if err != nil {
			if _, ok := err.(*api.WorkflowExecutionAlreadyStartedError); ok {
				return nil
			}
		}

		return err
	}

	return backoff.Retry(ctx, startWorkflowOp, createDynamicServiceRetryPolicy(ctx), isServiceTransientError)
}

func generateShadowTaskList(domain, taskList string) string {
	return domain + "-" + taskList
}
