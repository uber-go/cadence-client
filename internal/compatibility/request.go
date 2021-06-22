// Copyright (c) 2021 Uber Technologies, Inc.
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

package compatibility

import (
	"go.uber.org/cadence/v2/.gen/go/shared"
	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
)

func protoCountWorkflowExecutionsRequest(t *shared.CountWorkflowExecutionsRequest) *apiv1.CountWorkflowExecutionsRequest {
	if t == nil {
		return nil
	}
	return &apiv1.CountWorkflowExecutionsRequest{
		Domain: t.GetDomain(),
		Query:  t.GetQuery(),
	}
}

func protoDeprecateDomainRequest(t *shared.DeprecateDomainRequest) *apiv1.DeprecateDomainRequest {
	if t == nil {
		return nil
	}
	return &apiv1.DeprecateDomainRequest{
		Name:          t.GetName(),
		SecurityToken: t.GetSecurityToken(),
	}
}

func protoDescribeDomainRequest(t *shared.DescribeDomainRequest) *apiv1.DescribeDomainRequest {
	if t == nil {
		return nil
	}
	if t.UUID != nil {
		return &apiv1.DescribeDomainRequest{DescribeBy: &apiv1.DescribeDomainRequest_Id{Id: *t.UUID}}
	}
	if t.Name != nil {
		return &apiv1.DescribeDomainRequest{DescribeBy: &apiv1.DescribeDomainRequest_Name{Name: *t.Name}}
	}
	panic("neither oneof field is set for DescribeDomainRequest")
}

func protoDescribeTaskListRequest(t *shared.DescribeTaskListRequest) *apiv1.DescribeTaskListRequest {
	if t == nil {
		return nil
	}
	return &apiv1.DescribeTaskListRequest{
		Domain:                t.GetDomain(),
		TaskList:              protoTaskList(t.TaskList),
		TaskListType:          protoTaskListType(t.TaskListType),
		IncludeTaskListStatus: t.GetIncludeTaskListStatus(),
	}
}

func protoDescribeWorkflowExecutionRequest(t *shared.DescribeWorkflowExecutionRequest) *apiv1.DescribeWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &apiv1.DescribeWorkflowExecutionRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: protoWorkflowExecution(t.Execution),
	}
}

func protoGetWorkflowExecutionHistoryRequest(t *shared.GetWorkflowExecutionHistoryRequest) *apiv1.GetWorkflowExecutionHistoryRequest {
	if t == nil {
		return nil
	}
	return &apiv1.GetWorkflowExecutionHistoryRequest{
		Domain:                 t.GetDomain(),
		WorkflowExecution:      protoWorkflowExecution(t.Execution),
		PageSize:               t.GetMaximumPageSize(),
		NextPageToken:          t.GetNextPageToken(),
		WaitForNewEvent:        t.GetWaitForNewEvent(),
		HistoryEventFilterType: protoEventFilterType(t.HistoryEventFilterType),
		SkipArchival:           t.GetSkipArchival(),
	}
}

func protoListArchivedWorkflowExecutionsRequest(t *shared.ListArchivedWorkflowExecutionsRequest) *apiv1.ListArchivedWorkflowExecutionsRequest {
	if t == nil {
		return nil
	}
	return &apiv1.ListArchivedWorkflowExecutionsRequest{
		Domain:        t.GetDomain(),
		PageSize:      t.GetPageSize(),
		NextPageToken: t.GetNextPageToken(),
		Query:         t.GetQuery(),
	}
}

func protoListDomainsRequest(t *shared.ListDomainsRequest) *apiv1.ListDomainsRequest {
	if t == nil {
		return nil
	}
	return &apiv1.ListDomainsRequest{
		PageSize:      t.GetPageSize(),
		NextPageToken: t.NextPageToken,
	}
}

func protoListTaskListPartitionsRequest(t *shared.ListTaskListPartitionsRequest) *apiv1.ListTaskListPartitionsRequest {
	if t == nil {
		return nil
	}
	return &apiv1.ListTaskListPartitionsRequest{
		Domain:   t.GetDomain(),
		TaskList: protoTaskList(t.TaskList),
	}
}

func protoListWorkflowExecutionsRequest(t *shared.ListWorkflowExecutionsRequest) *apiv1.ListWorkflowExecutionsRequest {
	if t == nil {
		return nil
	}
	return &apiv1.ListWorkflowExecutionsRequest{
		Domain:        t.GetDomain(),
		PageSize:      t.GetPageSize(),
		NextPageToken: t.NextPageToken,
		Query:         t.GetQuery(),
	}
}

func protoPollForActivityTaskRequest(t *shared.PollForActivityTaskRequest) *apiv1.PollForActivityTaskRequest {
	if t == nil {
		return nil
	}
	return &apiv1.PollForActivityTaskRequest{
		Domain:           t.GetDomain(),
		TaskList:         protoTaskList(t.TaskList),
		Identity:         t.GetIdentity(),
		TaskListMetadata: protoTaskListMetadata(t.TaskListMetadata),
	}
}

func protoPollForDecisionTaskRequest(t *shared.PollForDecisionTaskRequest) *apiv1.PollForDecisionTaskRequest {
	if t == nil {
		return nil
	}
	return &apiv1.PollForDecisionTaskRequest{
		Domain:         t.GetDomain(),
		TaskList:       protoTaskList(t.TaskList),
		Identity:       t.GetIdentity(),
		BinaryChecksum: t.GetBinaryChecksum(),
	}
}


func protoQueryWorkflowRequest(t *shared.QueryWorkflowRequest) *apiv1.QueryWorkflowRequest {
	if t == nil {
		return nil
	}
	return &apiv1.QueryWorkflowRequest{
		Domain:                t.GetDomain(),
		WorkflowExecution:     protoWorkflowExecution(t.Execution),
		Query:                 protoWorkflowQuery(t.Query),
		QueryRejectCondition:  protoQueryRejectCondition(t.QueryRejectCondition),
		QueryConsistencyLevel: protoQueryConsistencyLevel(t.QueryConsistencyLevel),
	}
}

func protoRecordActivityTaskHeartbeatByIdRequest(t *shared.RecordActivityTaskHeartbeatByIDRequest) *apiv1.RecordActivityTaskHeartbeatByIDRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RecordActivityTaskHeartbeatByIDRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: protoWorkflowRunPair(t.GetWorkflowID(), t.GetRunID()),
		ActivityId:        t.GetActivityID(),
		Details:           protoPayload(t.Details),
		Identity:          t.GetIdentity(),
	}
}

func protoRecordActivityTaskHeartbeatRequest(t *shared.RecordActivityTaskHeartbeatRequest) *apiv1.RecordActivityTaskHeartbeatRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RecordActivityTaskHeartbeatRequest{
		TaskToken: t.TaskToken,
		Details:   protoPayload(t.Details),
		Identity:  t.GetIdentity(),
	}
}

func protoRegisterDomainRequest(t *shared.RegisterDomainRequest) *apiv1.RegisterDomainRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RegisterDomainRequest{
		Name:                             t.GetName(),
		Description:                      t.GetDescription(),
		OwnerEmail:                       t.GetOwnerEmail(),
		WorkflowExecutionRetentionPeriod: daysToDuration(t.WorkflowExecutionRetentionPeriodInDays),
		Clusters:                         protoClusterReplicationConfigurationArray(t.Clusters),
		ActiveClusterName:                t.GetActiveClusterName(),
		Data:                             t.Data,
		SecurityToken:                    t.GetSecurityToken(),
		IsGlobalDomain:                   t.GetIsGlobalDomain(),
		HistoryArchivalStatus:            protoArchivalStatus(t.HistoryArchivalStatus),
		HistoryArchivalUri:               t.GetHistoryArchivalURI(),
		VisibilityArchivalStatus:         protoArchivalStatus(t.VisibilityArchivalStatus),
		VisibilityArchivalUri:            t.GetVisibilityArchivalURI(),
	}
}

func protoRequestCancelWorkflowExecutionRequest(t *shared.RequestCancelWorkflowExecutionRequest) *apiv1.RequestCancelWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RequestCancelWorkflowExecutionRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: protoWorkflowExecution(t.WorkflowExecution),
		Identity:          t.GetIdentity(),
		RequestId:         t.GetRequestId(),
	}
}

func protoResetStickyTaskListRequest(t *shared.ResetStickyTaskListRequest) *apiv1.ResetStickyTaskListRequest {
	if t == nil {
		return nil
	}
	return &apiv1.ResetStickyTaskListRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: protoWorkflowExecution(t.Execution),
	}
}

func protoResetWorkflowExecutionRequest(t *shared.ResetWorkflowExecutionRequest) *apiv1.ResetWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &apiv1.ResetWorkflowExecutionRequest{
		Domain:                t.GetDomain(),
		WorkflowExecution:     protoWorkflowExecution(t.WorkflowExecution),
		Reason:                t.GetReason(),
		DecisionFinishEventId: t.GetDecisionFinishEventId(),
		RequestId:             t.GetRequestId(),
		SkipSignalReapply:     t.GetSkipSignalReapply(),
	}
}

func protoRespondActivityTaskCanceledByIdRequest(t *shared.RespondActivityTaskCanceledByIDRequest) *apiv1.RespondActivityTaskCanceledByIDRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondActivityTaskCanceledByIDRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: protoWorkflowRunPair(t.GetWorkflowID(), t.GetRunID()),
		ActivityId:        t.GetActivityID(),
		Details:           protoPayload(t.Details),
		Identity:          t.GetIdentity(),
	}
}

func protoRespondActivityTaskCanceledRequest(t *shared.RespondActivityTaskCanceledRequest) *apiv1.RespondActivityTaskCanceledRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondActivityTaskCanceledRequest{
		TaskToken: t.TaskToken,
		Details:   protoPayload(t.Details),
		Identity:  t.GetIdentity(),
	}
}

func protoRespondActivityTaskCompletedByIdRequest(t *shared.RespondActivityTaskCompletedByIDRequest) *apiv1.RespondActivityTaskCompletedByIDRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondActivityTaskCompletedByIDRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: protoWorkflowRunPair(t.GetWorkflowID(), t.GetRunID()),
		ActivityId:        t.GetActivityID(),
		Result:            protoPayload(t.Result),
		Identity:          t.GetIdentity(),
	}
}

func protoRespondActivityTaskCompletedRequest(t *shared.RespondActivityTaskCompletedRequest) *apiv1.RespondActivityTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondActivityTaskCompletedRequest{
		TaskToken: t.TaskToken,
		Result:    protoPayload(t.Result),
		Identity:  t.GetIdentity(),
	}
}

func protoRespondActivityTaskFailedByIdRequest(t *shared.RespondActivityTaskFailedByIDRequest) *apiv1.RespondActivityTaskFailedByIDRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondActivityTaskFailedByIDRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: protoWorkflowRunPair(t.GetWorkflowID(), t.GetRunID()),
		ActivityId:        t.GetActivityID(),
		Failure:           protoFailure(t.Reason, t.Details),
		Identity:          t.GetIdentity(),
	}
}

func protoRespondActivityTaskFailedRequest(t *shared.RespondActivityTaskFailedRequest) *apiv1.RespondActivityTaskFailedRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondActivityTaskFailedRequest{
		TaskToken: t.TaskToken,
		Failure:   protoFailure(t.Reason, t.Details),
		Identity:  t.GetIdentity(),
	}
}

func protoRespondDecisionTaskCompletedRequest(t *shared.RespondDecisionTaskCompletedRequest) *apiv1.RespondDecisionTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondDecisionTaskCompletedRequest{
		TaskToken:                  t.TaskToken,
		Decisions:                  protoDecisionArray(t.Decisions),
		ExecutionContext:           t.ExecutionContext,
		Identity:                   t.GetIdentity(),
		StickyAttributes:           protoStickyExecutionAttributes(t.StickyAttributes),
		ReturnNewDecisionTask:      t.GetReturnNewDecisionTask(),
		ForceCreateNewDecisionTask: t.GetForceCreateNewDecisionTask(),
		BinaryChecksum:             t.GetBinaryChecksum(),
		QueryResults:               protoWorkflowQueryResultMap(t.QueryResults),
	}
}

func protoRespondDecisionTaskFailedRequest(t *shared.RespondDecisionTaskFailedRequest) *apiv1.RespondDecisionTaskFailedRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondDecisionTaskFailedRequest{
		TaskToken:      t.TaskToken,
		Cause:          protoDecisionTaskFailedCause(t.Cause),
		Details:        protoPayload(t.Details),
		Identity:       t.GetIdentity(),
		BinaryChecksum: t.GetBinaryChecksum(),
	}
}

func protoRespondQueryTaskCompletedRequest(t *shared.RespondQueryTaskCompletedRequest) *apiv1.RespondQueryTaskCompletedRequest {
	if t == nil {
		return nil
	}
	return &apiv1.RespondQueryTaskCompletedRequest{
		TaskToken: t.TaskToken,
		Result: &apiv1.WorkflowQueryResult{
			ResultType:   protoQueryTaskCompletedType(t.CompletedType),
			Answer:       protoPayload(t.QueryResult),
			ErrorMessage: t.GetErrorMessage(),
		},
		WorkerVersionInfo: protoWorkerVersionInfo(t.WorkerVersionInfo),
	}
}

func protoScanWorkflowExecutionsRequest(t *shared.ListWorkflowExecutionsRequest) *apiv1.ScanWorkflowExecutionsRequest {
	if t == nil {
		return nil
	}
	return &apiv1.ScanWorkflowExecutionsRequest{
		Domain:        t.GetDomain(),
		PageSize:      t.GetPageSize(),
		NextPageToken: t.NextPageToken,
		Query:         t.GetQuery(),
	}
}

func protoSignalWithStartWorkflowExecutionRequest(t *shared.SignalWithStartWorkflowExecutionRequest) *apiv1.SignalWithStartWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &apiv1.SignalWithStartWorkflowExecutionRequest{
		StartRequest: &apiv1.StartWorkflowExecutionRequest{
			Domain:                       t.GetDomain(),
			WorkflowId:                   t.GetWorkflowId(),
			WorkflowType:                 protoWorkflowType(t.WorkflowType),
			TaskList:                     protoTaskList(t.TaskList),
			Input:                        protoPayload(t.Input),
			ExecutionStartToCloseTimeout: secondsToDuration(t.ExecutionStartToCloseTimeoutSeconds),
			TaskStartToCloseTimeout:      secondsToDuration(t.TaskStartToCloseTimeoutSeconds),
			Identity:                     t.GetIdentity(),
			RequestId:                    t.GetRequestId(),
			WorkflowIdReusePolicy:        protoWorkflowIdReusePolicy(t.WorkflowIdReusePolicy),
			RetryPolicy:                  protoRetryPolicy(t.RetryPolicy),
			CronSchedule:                 t.GetCronSchedule(),
			Memo:                         protoMemo(t.Memo),
			SearchAttributes:             protoSearchAttributes(t.SearchAttributes),
			Header:                       protoHeader(t.Header),
			DelayStart:                   secondsToDuration(t.DelayStartSeconds),
		},
		SignalName:  t.GetSignalName(),
		SignalInput: protoPayload(t.SignalInput),
		Control:     t.Control,
	}
}

func protoSignalWorkflowExecutionRequest(t *shared.SignalWorkflowExecutionRequest) *apiv1.SignalWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &apiv1.SignalWorkflowExecutionRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: protoWorkflowExecution(t.WorkflowExecution),
		SignalName:        t.GetSignalName(),
		SignalInput:       protoPayload(t.Input),
		Identity:          t.GetIdentity(),
		RequestId:         t.GetRequestId(),
		Control:           t.Control,
	}
}

func protoStartWorkflowExecutionRequest(t *shared.StartWorkflowExecutionRequest) *apiv1.StartWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &apiv1.StartWorkflowExecutionRequest{
		Domain:                       t.GetDomain(),
		WorkflowId:                   t.GetWorkflowId(),
		WorkflowType:                 protoWorkflowType(t.WorkflowType),
		TaskList:                     protoTaskList(t.TaskList),
		Input:                        protoPayload(t.Input),
		ExecutionStartToCloseTimeout: secondsToDuration(t.ExecutionStartToCloseTimeoutSeconds),
		TaskStartToCloseTimeout:      secondsToDuration(t.TaskStartToCloseTimeoutSeconds),
		Identity:                     t.GetIdentity(),
		RequestId:                    t.GetRequestId(),
		WorkflowIdReusePolicy:        protoWorkflowIdReusePolicy(t.WorkflowIdReusePolicy),
		RetryPolicy:                  protoRetryPolicy(t.RetryPolicy),
		CronSchedule:                 t.GetCronSchedule(),
		Memo:                         protoMemo(t.Memo),
		SearchAttributes:             protoSearchAttributes(t.SearchAttributes),
		Header:                       protoHeader(t.Header),
		DelayStart:                   secondsToDuration(t.DelayStartSeconds),
	}
}

func protoTerminateWorkflowExecutionRequest(t *shared.TerminateWorkflowExecutionRequest) *apiv1.TerminateWorkflowExecutionRequest {
	if t == nil {
		return nil
	}
	return &apiv1.TerminateWorkflowExecutionRequest{
		Domain:            t.GetDomain(),
		WorkflowExecution: protoWorkflowExecution(t.WorkflowExecution),
		Reason:            t.GetReason(),
		Details:           protoPayload(t.Details),
		Identity:          t.GetIdentity(),
	}
}

const (
	DomainUpdateDescriptionField              = "description"
	DomainUpdateOwnerEmailField               = "owner_email"
	DomainUpdateDataField                     = "data"
	DomainUpdateRetentionPeriodField          = "workflow_execution_retention_period"
	DomainUpdateBadBinariesField              = "bad_binaries"
	DomainUpdateHistoryArchivalStatusField    = "history_archival_status"
	DomainUpdateHistoryArchivalURIField       = "history_archival_uri"
	DomainUpdateVisibilityArchivalStatusField = "visibility_archival_status"
	DomainUpdateVisibilityArchivalURIField    = "visibility_archival_uri"
	DomainUpdateActiveClusterNameField        = "active_cluster_name"
	DomainUpdateClustersField                 = "clusters"
	DomainUpdateDeleteBadBinaryField          = "delete_bad_binary"
	DomainUpdateFailoverTimeoutField          = "failover_timeout"
)

func protoUpdateDomainRequest(t *shared.UpdateDomainRequest) *apiv1.UpdateDomainRequest {
	if t == nil {
		return nil
	}
	request := apiv1.UpdateDomainRequest{
		Name:          t.GetName(),
		SecurityToken: t.GetSecurityToken(),
	}
	fields := []string{}

	if description := t.GetUpdatedInfo().Description; description != nil {
		request.Description = *description
		fields = append(fields, DomainUpdateDescriptionField)
	}
	if email := t.GetUpdatedInfo().OwnerEmail; email != nil {
		request.OwnerEmail = *email
		fields = append(fields, DomainUpdateOwnerEmailField)
	}
	if data := t.GetUpdatedInfo().Data; data != nil {
		request.Data = data
		fields = append(fields, DomainUpdateDataField)
	}
	if retention := t.GetConfiguration().WorkflowExecutionRetentionPeriodInDays; retention != nil {
		request.WorkflowExecutionRetentionPeriod = daysToDuration(retention)
		fields = append(fields, DomainUpdateRetentionPeriodField)
	}
	//if t.EmitMetric != nil {} - DEPRECATED
	if badBinaries := t.GetConfiguration().BadBinaries; badBinaries != nil {
		request.BadBinaries = protoBadBinaries(badBinaries)
		fields = append(fields, DomainUpdateBadBinariesField)
	}
	if historyArchivalStatus := t.GetConfiguration().HistoryArchivalStatus; historyArchivalStatus != nil {
		request.HistoryArchivalStatus = protoArchivalStatus(historyArchivalStatus)
		fields = append(fields, DomainUpdateHistoryArchivalStatusField)
	}
	if historyArchivalURI := t.GetConfiguration().HistoryArchivalURI; historyArchivalURI != nil {
		request.HistoryArchivalUri = *historyArchivalURI
		fields = append(fields, DomainUpdateHistoryArchivalURIField)
	}
	if visibilityArchivalStatus := t.GetConfiguration().VisibilityArchivalStatus; visibilityArchivalStatus != nil {
		request.VisibilityArchivalStatus = protoArchivalStatus(visibilityArchivalStatus)
		fields = append(fields, DomainUpdateVisibilityArchivalStatusField)
	}
	if visibilityArchivalURI := t.GetConfiguration().VisibilityArchivalURI; visibilityArchivalURI != nil {
		request.VisibilityArchivalUri = *visibilityArchivalURI
		fields = append(fields, DomainUpdateVisibilityArchivalURIField)
	}
	if activeClusterName := t.GetReplicationConfiguration().ActiveClusterName; activeClusterName != nil {
		request.ActiveClusterName = *activeClusterName
		fields = append(fields, DomainUpdateActiveClusterNameField)
	}
	if clusters := t.GetReplicationConfiguration().Clusters; clusters != nil {
		request.Clusters = protoClusterReplicationConfigurationArray(clusters)
		fields = append(fields, DomainUpdateClustersField)
	}
	if t.DeleteBadBinary != nil {
		request.DeleteBadBinary = *t.DeleteBadBinary
		fields = append(fields, DomainUpdateDeleteBadBinaryField)
	}
	if t.FailoverTimeoutInSeconds != nil {
		request.FailoverTimeout = secondsToDuration(t.FailoverTimeoutInSeconds)
		fields = append(fields, DomainUpdateFailoverTimeoutField)
	}

	request.UpdateMask = newFieldMask(fields)

	return &request
}

func protoListClosedWorkflowExecutionsRequest(r *shared.ListClosedWorkflowExecutionsRequest) *apiv1.ListClosedWorkflowExecutionsRequest {
	if r == nil {
		return nil
	}
	request := apiv1.ListClosedWorkflowExecutionsRequest{
		Domain:          r.GetDomain(),
		PageSize:        r.GetMaximumPageSize(),
		NextPageToken:   r.NextPageToken,
		StartTimeFilter: protoStartTimeFilter(r.StartTimeFilter),
	}
	if r.ExecutionFilter != nil {
		request.Filters = &apiv1.ListClosedWorkflowExecutionsRequest_ExecutionFilter{
			ExecutionFilter: protoWorkflowExecutionFilter(r.ExecutionFilter),
		}
	}
	if r.TypeFilter != nil {
		request.Filters = &apiv1.ListClosedWorkflowExecutionsRequest_TypeFilter{
			TypeFilter: protoWorkflowTypeFilter(r.TypeFilter),
		}
	}
	if r.StatusFilter != nil {
		request.Filters = &apiv1.ListClosedWorkflowExecutionsRequest_StatusFilter{
			StatusFilter: protoStatusFilter(r.StatusFilter),
		}
	}

	return &request
}

func protoListOpenWorkflowExecutionsRequest(r *shared.ListOpenWorkflowExecutionsRequest) *apiv1.ListOpenWorkflowExecutionsRequest {
	if r == nil {
		return nil
	}
	request := apiv1.ListOpenWorkflowExecutionsRequest{
		Domain:          r.GetDomain(),
		PageSize:        r.GetMaximumPageSize(),
		NextPageToken:   r.NextPageToken,
		StartTimeFilter: protoStartTimeFilter(r.StartTimeFilter),
	}
	if r.ExecutionFilter != nil {
		request.Filters = &apiv1.ListOpenWorkflowExecutionsRequest_ExecutionFilter{
			ExecutionFilter: protoWorkflowExecutionFilter(r.ExecutionFilter),
		}
	}
	if r.TypeFilter != nil {
		request.Filters = &apiv1.ListOpenWorkflowExecutionsRequest_TypeFilter{
			TypeFilter: protoWorkflowTypeFilter(r.TypeFilter),
		}
	}

	return &request
}


func protoPayload(data []byte) *apiv1.Payload {
	if data == nil {
		return nil
	}
	return &apiv1.Payload{
		Data: data,
	}
}

func protoFailure(reason *string, details []byte) *apiv1.Failure {
	if reason == nil {
		return nil
	}
	return &apiv1.Failure{
		Reason:  *reason,
		Details: details,
	}
}

func protoWorkflowExecution(t *shared.WorkflowExecution) *apiv1.WorkflowExecution {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecution{
		WorkflowId: t.GetWorkflowId(),
		RunId:      t.GetRunId(),
	}
}

func protoWorkflowRunPair(workflowId, runId string) *apiv1.WorkflowExecution {
	return &apiv1.WorkflowExecution{
		WorkflowId: workflowId,
		RunId:      runId,
	}
}

func protoActivityType(t *shared.ActivityType) *apiv1.ActivityType {
	if t == nil {
		return nil
	}
	return &apiv1.ActivityType{
		Name: t.GetName(),
	}
}

func protoWorkflowType(t *shared.WorkflowType) *apiv1.WorkflowType {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowType{
		Name: t.GetName(),
	}
}

func protoTaskList(t *shared.TaskList) *apiv1.TaskList {
	if t == nil {
		return nil
	}
	return &apiv1.TaskList{
		Name: t.GetName(),
		Kind: protoTaskListKind(t.Kind),
	}
}

func protoTaskListMetadata(t *shared.TaskListMetadata) *apiv1.TaskListMetadata {
	if t == nil {
		return nil
	}
	return &apiv1.TaskListMetadata{
		MaxTasksPerSecond: fromDoubleValue(t.MaxTasksPerSecond),
	}
}

func protoRetryPolicy(t *shared.RetryPolicy) *apiv1.RetryPolicy {
	if t == nil {
		return nil
	}
	return &apiv1.RetryPolicy{
		InitialInterval:          secondsToDuration(t.InitialIntervalInSeconds),
		BackoffCoefficient:       t.GetBackoffCoefficient(),
		MaximumInterval:          secondsToDuration(t.MaximumIntervalInSeconds),
		MaximumAttempts:          t.GetMaximumAttempts(),
		NonRetryableErrorReasons: t.NonRetriableErrorReasons,
		ExpirationInterval:       secondsToDuration(t.ExpirationIntervalInSeconds),
	}
}

func protoHeader(t *shared.Header) *apiv1.Header {
	if t == nil {
		return nil
	}
	return &apiv1.Header{
		Fields: protoPayloadMap(t.Fields),
	}
}

func protoMemo(t *shared.Memo) *apiv1.Memo {
	if t == nil {
		return nil
	}
	return &apiv1.Memo{
		Fields: protoPayloadMap(t.Fields),
	}
}

func protoSearchAttributes(t *shared.SearchAttributes) *apiv1.SearchAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.SearchAttributes{
		IndexedFields: protoPayloadMap(t.IndexedFields),
	}
}


func protoBadBinaries(t *shared.BadBinaries) *apiv1.BadBinaries {
	if t == nil {
		return nil
	}
	return &apiv1.BadBinaries{
		Binaries: protoBadBinaryInfoMap(t.Binaries),
	}
}

func protoBadBinaryInfo(t *shared.BadBinaryInfo) *apiv1.BadBinaryInfo {
	if t == nil {
		return nil
	}
	return &apiv1.BadBinaryInfo{
		Reason:      t.GetReason(),
		Operator:    t.GetOperator(),
		CreatedTime: unixNanoToTime(t.CreatedTimeNano),
	}
}

func protoClusterReplicationConfiguration(t *shared.ClusterReplicationConfiguration) *apiv1.ClusterReplicationConfiguration {
	if t == nil {
		return nil
	}
	return &apiv1.ClusterReplicationConfiguration{
		ClusterName: t.GetClusterName(),
	}
}

func protoWorkflowQuery(t *shared.WorkflowQuery) *apiv1.WorkflowQuery {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowQuery{
		QueryType: t.GetQueryType(),
		QueryArgs: protoPayload(t.QueryArgs),
	}
}

func protoWorkflowQueryResult(t *shared.WorkflowQueryResult) *apiv1.WorkflowQueryResult {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowQueryResult{
		ResultType:   protoQueryResultType(t.ResultType),
		Answer:       protoPayload(t.Answer),
		ErrorMessage: t.GetErrorMessage(),
	}
}


func protoStickyExecutionAttributes(t *shared.StickyExecutionAttributes) *apiv1.StickyExecutionAttributes {
	if t == nil {
		return nil
	}
	return &apiv1.StickyExecutionAttributes{
		WorkerTaskList:         protoTaskList(t.WorkerTaskList),
		ScheduleToStartTimeout: secondsToDuration(t.ScheduleToStartTimeoutSeconds),
	}
}

func protoWorkerVersionInfo(t *shared.WorkerVersionInfo) *apiv1.WorkerVersionInfo {
	if t == nil {
		return nil
	}
	return &apiv1.WorkerVersionInfo{
		Impl:           t.GetImpl(),
		FeatureVersion: t.GetFeatureVersion(),
	}
}

func protoStartTimeFilter(t *shared.StartTimeFilter) *apiv1.StartTimeFilter {
	if t == nil {
		return nil
	}
	return &apiv1.StartTimeFilter{
		EarliestTime: unixNanoToTime(t.EarliestTime),
		LatestTime:   unixNanoToTime(t.LatestTime),
	}
}

func protoWorkflowExecutionFilter(t *shared.WorkflowExecutionFilter) *apiv1.WorkflowExecutionFilter {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowExecutionFilter{
		WorkflowId: t.GetWorkflowId(),
		RunId:      t.GetRunId(),
	}
}

func protoWorkflowTypeFilter(t *shared.WorkflowTypeFilter) *apiv1.WorkflowTypeFilter {
	if t == nil {
		return nil
	}
	return &apiv1.WorkflowTypeFilter{
		Name: t.GetName(),
	}
}

func protoStatusFilter(t *shared.WorkflowExecutionCloseStatus) *apiv1.StatusFilter {
	if t == nil {
		return nil
	}
	return &apiv1.StatusFilter{
		Status: protoWorkflowExecutionCloseStatus(t),
	}
}

func protoDecision(d *shared.Decision) *apiv1.Decision {
	if d == nil {
		return nil
	}
	decision := apiv1.Decision{}
	switch *d.DecisionType {
	case shared.DecisionTypeScheduleActivityTask:
		attr := d.ScheduleActivityTaskDecisionAttributes
		decision.Attributes = &apiv1.Decision_ScheduleActivityTaskDecisionAttributes{
			ScheduleActivityTaskDecisionAttributes: &apiv1.ScheduleActivityTaskDecisionAttributes{
				ActivityId:             attr.GetActivityId(),
				ActivityType:           protoActivityType(attr.ActivityType),
				Domain:                 attr.GetDomain(),
				TaskList:               protoTaskList(attr.TaskList),
				Input:                  protoPayload(attr.Input),
				ScheduleToCloseTimeout: secondsToDuration(attr.ScheduleToCloseTimeoutSeconds),
				ScheduleToStartTimeout: secondsToDuration(attr.ScheduleToStartTimeoutSeconds),
				StartToCloseTimeout:    secondsToDuration(attr.StartToCloseTimeoutSeconds),
				HeartbeatTimeout:       secondsToDuration(attr.HeartbeatTimeoutSeconds),
				RetryPolicy:            protoRetryPolicy(attr.RetryPolicy),
				Header:                 protoHeader(attr.Header),
				RequestLocalDispatch:   attr.GetRequestLocalDispatch(),
			},
		}
	case shared.DecisionTypeRequestCancelActivityTask:
		attr := d.RequestCancelActivityTaskDecisionAttributes
		decision.Attributes = &apiv1.Decision_RequestCancelActivityTaskDecisionAttributes{
			RequestCancelActivityTaskDecisionAttributes: &apiv1.RequestCancelActivityTaskDecisionAttributes{
				ActivityId: attr.GetActivityId(),
			},
		}
	case shared.DecisionTypeStartTimer:
		attr := d.StartTimerDecisionAttributes
		decision.Attributes = &apiv1.Decision_StartTimerDecisionAttributes{
			StartTimerDecisionAttributes: &apiv1.StartTimerDecisionAttributes{
				TimerId:            attr.GetTimerId(),
				StartToFireTimeout: secondsToDuration(int64To32(attr.StartToFireTimeoutSeconds)),
			},
		}
	case shared.DecisionTypeCompleteWorkflowExecution:
		attr := d.CompleteWorkflowExecutionDecisionAttributes
		decision.Attributes = &apiv1.Decision_CompleteWorkflowExecutionDecisionAttributes{
			CompleteWorkflowExecutionDecisionAttributes: &apiv1.CompleteWorkflowExecutionDecisionAttributes{
				Result: protoPayload(attr.Result),
			},
		}
	case shared.DecisionTypeFailWorkflowExecution:
		attr := d.FailWorkflowExecutionDecisionAttributes
		decision.Attributes = &apiv1.Decision_FailWorkflowExecutionDecisionAttributes{
			FailWorkflowExecutionDecisionAttributes: &apiv1.FailWorkflowExecutionDecisionAttributes{
				Failure: protoFailure(attr.Reason, attr.Details),
			},
		}
	case shared.DecisionTypeCancelTimer:
		attr := d.CancelTimerDecisionAttributes
		decision.Attributes = &apiv1.Decision_CancelTimerDecisionAttributes{
			CancelTimerDecisionAttributes: &apiv1.CancelTimerDecisionAttributes{
				TimerId: attr.GetTimerId(),
			},
		}
	case shared.DecisionTypeCancelWorkflowExecution:
		attr := d.CancelWorkflowExecutionDecisionAttributes
		decision.Attributes = &apiv1.Decision_CancelWorkflowExecutionDecisionAttributes{
			CancelWorkflowExecutionDecisionAttributes: &apiv1.CancelWorkflowExecutionDecisionAttributes{
				Details: protoPayload(attr.Details),
			},
		}
	case shared.DecisionTypeRequestCancelExternalWorkflowExecution:
		attr := d.RequestCancelExternalWorkflowExecutionDecisionAttributes
		decision.Attributes = &apiv1.Decision_RequestCancelExternalWorkflowExecutionDecisionAttributes{
			RequestCancelExternalWorkflowExecutionDecisionAttributes: &apiv1.RequestCancelExternalWorkflowExecutionDecisionAttributes{
				Domain:            attr.GetDomain(),
				WorkflowExecution: protoWorkflowRunPair(attr.GetWorkflowId(), attr.GetRunId()),
				Control:           attr.Control,
				ChildWorkflowOnly: attr.GetChildWorkflowOnly(),
			},
		}
	case shared.DecisionTypeRecordMarker:
		attr := d.RecordMarkerDecisionAttributes
		decision.Attributes = &apiv1.Decision_RecordMarkerDecisionAttributes{
			RecordMarkerDecisionAttributes: &apiv1.RecordMarkerDecisionAttributes{
				MarkerName: attr.GetMarkerName(),
				Details:    protoPayload(attr.Details),
				Header:     protoHeader(attr.Header),
			},
		}
	case shared.DecisionTypeContinueAsNewWorkflowExecution:
		attr := d.ContinueAsNewWorkflowExecutionDecisionAttributes
		decision.Attributes = &apiv1.Decision_ContinueAsNewWorkflowExecutionDecisionAttributes{
			ContinueAsNewWorkflowExecutionDecisionAttributes: &apiv1.ContinueAsNewWorkflowExecutionDecisionAttributes{
				WorkflowType:                 protoWorkflowType(attr.WorkflowType),
				TaskList:                     protoTaskList(attr.TaskList),
				Input:                        protoPayload(attr.Input),
				ExecutionStartToCloseTimeout: secondsToDuration(attr.ExecutionStartToCloseTimeoutSeconds),
				TaskStartToCloseTimeout:      secondsToDuration(attr.TaskStartToCloseTimeoutSeconds),
				BackoffStartInterval:         secondsToDuration(attr.BackoffStartIntervalInSeconds),
				RetryPolicy:                  protoRetryPolicy(attr.RetryPolicy),
				Initiator:                    protoContinueAsNewInitiator(attr.Initiator),
				Failure:                      protoFailure(attr.FailureReason, attr.FailureDetails),
				LastCompletionResult:         protoPayload(attr.LastCompletionResult),
				CronSchedule:                 attr.GetCronSchedule(),
				Header:                       protoHeader(attr.Header),
				Memo:                         protoMemo(attr.Memo),
				SearchAttributes:             protoSearchAttributes(attr.SearchAttributes),
			},
		}
	case shared.DecisionTypeStartChildWorkflowExecution:
		attr := d.StartChildWorkflowExecutionDecisionAttributes
		decision.Attributes = &apiv1.Decision_StartChildWorkflowExecutionDecisionAttributes{
			StartChildWorkflowExecutionDecisionAttributes: &apiv1.StartChildWorkflowExecutionDecisionAttributes{
				Domain:                       attr.GetDomain(),
				WorkflowId:                   attr.GetWorkflowId(),
				WorkflowType:                 protoWorkflowType(attr.WorkflowType),
				TaskList:                     protoTaskList(attr.TaskList),
				Input:                        protoPayload(attr.Input),
				ExecutionStartToCloseTimeout: secondsToDuration(attr.ExecutionStartToCloseTimeoutSeconds),
				TaskStartToCloseTimeout:      secondsToDuration(attr.TaskStartToCloseTimeoutSeconds),
				ParentClosePolicy:            protoParentClosePolicy(attr.ParentClosePolicy),
				Control:                      attr.Control,
				WorkflowIdReusePolicy:        protoWorkflowIdReusePolicy(attr.WorkflowIdReusePolicy),
				RetryPolicy:                  protoRetryPolicy(attr.RetryPolicy),
				CronSchedule:                 attr.GetCronSchedule(),
				Header:                       protoHeader(attr.Header),
				Memo:                         protoMemo(attr.Memo),
				SearchAttributes:             protoSearchAttributes(attr.SearchAttributes),
			},
		}
	case shared.DecisionTypeSignalExternalWorkflowExecution:
		attr := d.SignalExternalWorkflowExecutionDecisionAttributes
		decision.Attributes = &apiv1.Decision_SignalExternalWorkflowExecutionDecisionAttributes{
			SignalExternalWorkflowExecutionDecisionAttributes: &apiv1.SignalExternalWorkflowExecutionDecisionAttributes{
				Domain:            attr.GetDomain(),
				WorkflowExecution: protoWorkflowExecution(attr.Execution),
				SignalName:        attr.GetSignalName(),
				Input:             protoPayload(attr.Input),
				Control:           attr.Control,
				ChildWorkflowOnly: attr.GetChildWorkflowOnly(),
			},
		}
	case shared.DecisionTypeUpsertWorkflowSearchAttributes:
		attr := d.UpsertWorkflowSearchAttributesDecisionAttributes
		decision.Attributes = &apiv1.Decision_UpsertWorkflowSearchAttributesDecisionAttributes{
			UpsertWorkflowSearchAttributesDecisionAttributes: &apiv1.UpsertWorkflowSearchAttributesDecisionAttributes{
				SearchAttributes: protoSearchAttributes(attr.SearchAttributes),
			},
		}
	}
	return &decision
}

func protoPayloadMap(t map[string][]byte) map[string]*apiv1.Payload {
	if t == nil {
		return nil
	}
	v := make(map[string]*apiv1.Payload, len(t))
	for key := range t {
		v[key] = protoPayload(t[key])
	}
	return v
}

func protoBadBinaryInfoMap(t map[string]*shared.BadBinaryInfo) map[string]*apiv1.BadBinaryInfo {
	if t == nil {
		return nil
	}
	v := make(map[string]*apiv1.BadBinaryInfo, len(t))
	for key := range t {
		v[key] = protoBadBinaryInfo(t[key])
	}
	return v
}

func protoClusterReplicationConfigurationArray(t []*shared.ClusterReplicationConfiguration) []*apiv1.ClusterReplicationConfiguration {
	if t == nil {
		return nil
	}
	v := make([]*apiv1.ClusterReplicationConfiguration, len(t))
	for i := range t {
		v[i] = protoClusterReplicationConfiguration(t[i])
	}
	return v
}

func protoWorkflowQueryResultMap(t map[string]*shared.WorkflowQueryResult) map[string]*apiv1.WorkflowQueryResult {
	if t == nil {
		return nil
	}
	v := make(map[string]*apiv1.WorkflowQueryResult, len(t))
	for key := range t {
		v[key] = protoWorkflowQueryResult(t[key])
	}
	return v
}

func protoDecisionArray(t []*shared.Decision) []*apiv1.Decision {
	if t == nil {
		return nil
	}
	v := make([]*apiv1.Decision, len(t))
	for i := range t {
		v[i] = protoDecision(t[i])
	}
	return v
}

func protoTaskListKind(t *shared.TaskListKind) apiv1.TaskListKind {
	if t == nil {
		return apiv1.TaskListKind_TASK_LIST_KIND_INVALID
	}
	switch *t {
	case shared.TaskListKindNormal:
		return apiv1.TaskListKind_TASK_LIST_KIND_NORMAL
	case shared.TaskListKindSticky:
		return apiv1.TaskListKind_TASK_LIST_KIND_STICKY
	}
	panic("unexpected enum value")
}

func protoTaskListType(t *shared.TaskListType) apiv1.TaskListType {
	if t == nil {
		return apiv1.TaskListType_TASK_LIST_TYPE_INVALID
	}
	switch *t {
	case shared.TaskListTypeDecision:
		return apiv1.TaskListType_TASK_LIST_TYPE_DECISION
	case shared.TaskListTypeActivity:
		return apiv1.TaskListType_TASK_LIST_TYPE_ACTIVITY
	}
	panic("unexpected enum value")
}

func protoEventFilterType(t *shared.HistoryEventFilterType) apiv1.EventFilterType {
	if t == nil {
		return apiv1.EventFilterType_EVENT_FILTER_TYPE_INVALID
	}
	switch *t {
	case shared.HistoryEventFilterTypeAllEvent:
		return apiv1.EventFilterType_EVENT_FILTER_TYPE_ALL_EVENT
	case shared.HistoryEventFilterTypeCloseEvent:
		return apiv1.EventFilterType_EVENT_FILTER_TYPE_CLOSE_EVENT
	}
	panic("unexpected enum value")
}

func protoQueryRejectCondition(t *shared.QueryRejectCondition) apiv1.QueryRejectCondition {
	if t == nil {
		return apiv1.QueryRejectCondition_QUERY_REJECT_CONDITION_INVALID
	}
	switch *t {
	case shared.QueryRejectConditionNotOpen:
		return apiv1.QueryRejectCondition_QUERY_REJECT_CONDITION_NOT_OPEN
	case shared.QueryRejectConditionNotCompletedCleanly:
		return apiv1.QueryRejectCondition_QUERY_REJECT_CONDITION_NOT_COMPLETED_CLEANLY
	}
	panic("unexpected enum value")
}

func protoQueryConsistencyLevel(t *shared.QueryConsistencyLevel) apiv1.QueryConsistencyLevel {
	if t == nil {
		return apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_INVALID
	}
	switch *t {
	case shared.QueryConsistencyLevelEventual:
		return apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_EVENTUAL
	case shared.QueryConsistencyLevelStrong:
		return apiv1.QueryConsistencyLevel_QUERY_CONSISTENCY_LEVEL_STRONG
	}
	panic("unexpected enum value")
}

func protoContinueAsNewInitiator(t *shared.ContinueAsNewInitiator) apiv1.ContinueAsNewInitiator {
	if t == nil {
		return apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_INVALID
	}
	switch *t {
	case shared.ContinueAsNewInitiatorDecider:
		return apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_DECIDER
	case shared.ContinueAsNewInitiatorRetryPolicy:
		return apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_RETRY_POLICY
	case shared.ContinueAsNewInitiatorCronSchedule:
		return apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_CRON_SCHEDULE
	}
	panic("unexpected enum value")
}

func protoWorkflowIdReusePolicy(t *shared.WorkflowIdReusePolicy) apiv1.WorkflowIdReusePolicy {
	if t == nil {
		return apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_INVALID
	}
	switch *t {
	case shared.WorkflowIdReusePolicyAllowDuplicateFailedOnly:
		return apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY
	case shared.WorkflowIdReusePolicyAllowDuplicate:
		return apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE
	case shared.WorkflowIdReusePolicyRejectDuplicate:
		return apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE
	case shared.WorkflowIdReusePolicyTerminateIfRunning:
		return apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING
	}
	panic("unexpected enum value")
}

func protoQueryResultType(t *shared.QueryResultType) apiv1.QueryResultType {
	if t == nil {
		return apiv1.QueryResultType_QUERY_RESULT_TYPE_INVALID
	}
	switch *t {
	case shared.QueryResultTypeAnswered:
		return apiv1.QueryResultType_QUERY_RESULT_TYPE_ANSWERED
	case shared.QueryResultTypeFailed:
		return apiv1.QueryResultType_QUERY_RESULT_TYPE_FAILED
	}
	panic("unexpected enum value")
}

func protoArchivalStatus(t *shared.ArchivalStatus) apiv1.ArchivalStatus {
	if t == nil {
		return apiv1.ArchivalStatus_ARCHIVAL_STATUS_INVALID
	}
	switch *t {
	case shared.ArchivalStatusDisabled:
		return apiv1.ArchivalStatus_ARCHIVAL_STATUS_DISABLED
	case shared.ArchivalStatusEnabled:
		return apiv1.ArchivalStatus_ARCHIVAL_STATUS_ENABLED
	}
	panic("unexpected enum value")
}

func protoParentClosePolicy(t *shared.ParentClosePolicy) apiv1.ParentClosePolicy {
	if t == nil {
		return apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_INVALID
	}
	switch *t {
	case shared.ParentClosePolicyAbandon:
		return apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_ABANDON
	case shared.ParentClosePolicyRequestCancel:
		return apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_REQUEST_CANCEL
	case shared.ParentClosePolicyTerminate:
		return apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_TERMINATE
	}
	panic("unexpected enum value")
}

func protoDecisionTaskFailedCause(t *shared.DecisionTaskFailedCause) apiv1.DecisionTaskFailedCause {
	if t == nil {
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_INVALID
	}
	switch *t {
	case shared.DecisionTaskFailedCauseUnhandledDecision:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_UNHANDLED_DECISION
	case shared.DecisionTaskFailedCauseBadScheduleActivityAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SCHEDULE_ACTIVITY_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadRequestCancelActivityAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_REQUEST_CANCEL_ACTIVITY_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadStartTimerAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_START_TIMER_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadCancelTimerAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CANCEL_TIMER_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadRecordMarkerAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_RECORD_MARKER_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadCompleteWorkflowExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_COMPLETE_WORKFLOW_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadFailWorkflowExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_FAIL_WORKFLOW_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadCancelWorkflowExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CANCEL_WORKFLOW_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadRequestCancelExternalWorkflowExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_REQUEST_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadContinueAsNewAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CONTINUE_AS_NEW_ATTRIBUTES
	case shared.DecisionTaskFailedCauseStartTimerDuplicateID:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_START_TIMER_DUPLICATE_ID
	case shared.DecisionTaskFailedCauseResetStickyTasklist:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_RESET_STICKY_TASK_LIST
	case shared.DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_WORKFLOW_WORKER_UNHANDLED_FAILURE
	case shared.DecisionTaskFailedCauseBadSignalWorkflowExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SIGNAL_WORKFLOW_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseBadStartChildExecutionAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_START_CHILD_EXECUTION_ATTRIBUTES
	case shared.DecisionTaskFailedCauseForceCloseDecision:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_FORCE_CLOSE_DECISION
	case shared.DecisionTaskFailedCauseFailoverCloseDecision:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_FAILOVER_CLOSE_DECISION
	case shared.DecisionTaskFailedCauseBadSignalInputSize:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SIGNAL_INPUT_SIZE
	case shared.DecisionTaskFailedCauseResetWorkflow:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_RESET_WORKFLOW
	case shared.DecisionTaskFailedCauseBadBinary:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_BINARY
	case shared.DecisionTaskFailedCauseScheduleActivityDuplicateID:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_SCHEDULE_ACTIVITY_DUPLICATE_ID
	case shared.DecisionTaskFailedCauseBadSearchAttributes:
		return apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SEARCH_ATTRIBUTES
	}
	panic("unexpected enum value")
}

func protoWorkflowExecutionCloseStatus(t *shared.WorkflowExecutionCloseStatus) apiv1.WorkflowExecutionCloseStatus {
	if t == nil {
		return apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_INVALID
	}
	switch *t {
	case shared.WorkflowExecutionCloseStatusCompleted:
		return apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_COMPLETED
	case shared.WorkflowExecutionCloseStatusFailed:
		return apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_FAILED
	case shared.WorkflowExecutionCloseStatusCanceled:
		return apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CANCELED
	case shared.WorkflowExecutionCloseStatusTerminated:
		return apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TERMINATED
	case shared.WorkflowExecutionCloseStatusContinuedAsNew:
		return apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CONTINUED_AS_NEW
	case shared.WorkflowExecutionCloseStatusTimedOut:
		return apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TIMED_OUT
	}
	panic("unexpected enum value")
}

func protoQueryTaskCompletedType(t *shared.QueryTaskCompletedType) apiv1.QueryResultType {
	if t == nil {
		return apiv1.QueryResultType_QUERY_RESULT_TYPE_INVALID
	}
	switch *t {
	case shared.QueryTaskCompletedTypeCompleted:
		return apiv1.QueryResultType_QUERY_RESULT_TYPE_ANSWERED
	case shared.QueryTaskCompletedTypeFailed:
		return apiv1.QueryResultType_QUERY_RESULT_TYPE_FAILED
	}
	panic("unexpected enum value")
}