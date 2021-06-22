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
	"go.uber.org/cadence/v2/internal/common"
)

func thriftCountWorkflowExecutionsResponse(t *apiv1.CountWorkflowExecutionsResponse) *shared.CountWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &shared.CountWorkflowExecutionsResponse{
		Count: &t.Count,
	}
}

func thriftDescribeDomainResponse(t *apiv1.DescribeDomainResponse) *shared.DescribeDomainResponse {
	if t == nil || t.Domain == nil {
		return nil
	}
	return &shared.DescribeDomainResponse{
		DomainInfo: &shared.DomainInfo{
			Name:        &t.Domain.Name,
			Status:      thriftDomainStatus(t.Domain.Status),
			Description: &t.Domain.Description,
			OwnerEmail:  &t.Domain.OwnerEmail,
			Data:        t.Domain.Data,
			UUID:        &t.Domain.Id,
		},
		Configuration: &shared.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: durationToDays(t.Domain.WorkflowExecutionRetentionPeriod),
			EmitMetric:                             boolPtr(true),
			BadBinaries:                            thriftBadBinaries(t.Domain.BadBinaries),
			HistoryArchivalStatus:                  thriftArchivalStatus(t.Domain.HistoryArchivalStatus),
			HistoryArchivalURI:                     &t.Domain.HistoryArchivalUri,
			VisibilityArchivalStatus:               thriftArchivalStatus(t.Domain.VisibilityArchivalStatus),
			VisibilityArchivalURI:                  &t.Domain.VisibilityArchivalUri,
		},
		ReplicationConfiguration: &shared.DomainReplicationConfiguration{
			ActiveClusterName: &t.Domain.ActiveClusterName,
			Clusters:          thriftClusterReplicationConfigurationArray(t.Domain.Clusters),
		},
		FailoverVersion: &t.Domain.FailoverVersion,
		IsGlobalDomain:  &t.Domain.IsGlobalDomain,
	}
}

func thriftDescribeTaskListResponse(t *apiv1.DescribeTaskListResponse) *shared.DescribeTaskListResponse {
	if t == nil {
		return nil
	}
	return &shared.DescribeTaskListResponse{
		Pollers:        thriftPollerInfoArray(t.Pollers),
		TaskListStatus: thriftTaskListStatus(t.TaskListStatus),
	}
}

func thriftDescribeWorkflowExecutionResponse(t *apiv1.DescribeWorkflowExecutionResponse) *shared.DescribeWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &shared.DescribeWorkflowExecutionResponse{
		ExecutionConfiguration: thriftWorkflowExecutionConfiguration(t.ExecutionConfiguration),
		WorkflowExecutionInfo:  thriftWorkflowExecutionInfo(t.WorkflowExecutionInfo),
		PendingActivities:      thriftPendingActivityInfoArray(t.PendingActivities),
		PendingChildren:        thriftPendingChildExecutionInfoArray(t.PendingChildren),
		PendingDecision:        thriftPendingDecisionInfo(t.PendingDecision),
	}
}

func thriftGetClusterInfoResponse(t *apiv1.GetClusterInfoResponse) *shared.ClusterInfo {
	if t == nil {
		return nil
	}
	return &shared.ClusterInfo{
		SupportedClientVersions: thriftSupportedClientVersions(t.SupportedClientVersions),
	}
}

func thriftGetSearchAttributesResponse(t *apiv1.GetSearchAttributesResponse) *shared.GetSearchAttributesResponse {
	if t == nil {
		return nil
	}
	return &shared.GetSearchAttributesResponse{
		Keys: thriftIndexedValueTypeMap(t.Keys),
	}
}

func thriftGetWorkflowExecutionHistoryResponse(t *apiv1.GetWorkflowExecutionHistoryResponse) *shared.GetWorkflowExecutionHistoryResponse {
	if t == nil {
		return nil
	}
	return &shared.GetWorkflowExecutionHistoryResponse{
		History:       thriftHistory(t.History),
		RawHistory:    thriftDataBlobArray(t.RawHistory),
		NextPageToken: t.NextPageToken,
		Archived:      &t.Archived,
	}
}

func thriftListArchivedWorkflowExecutionsResponse(t *apiv1.ListArchivedWorkflowExecutionsResponse) *shared.ListArchivedWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &shared.ListArchivedWorkflowExecutionsResponse{
		Executions:    thriftWorkflowExecutionInfoArray(t.Executions),
		NextPageToken: t.NextPageToken,
	}
}

func thriftListClosedWorkflowExecutionsResponse(t *apiv1.ListClosedWorkflowExecutionsResponse) *shared.ListClosedWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &shared.ListClosedWorkflowExecutionsResponse{
		Executions:    thriftWorkflowExecutionInfoArray(t.Executions),
		NextPageToken: t.NextPageToken,
	}
}

func thriftListDomainsResponse(t *apiv1.ListDomainsResponse) *shared.ListDomainsResponse {
	if t == nil {
		return nil
	}
	return &shared.ListDomainsResponse{
		Domains:       thriftDescribeDomainResponseArray(t.Domains),
		NextPageToken: t.NextPageToken,
	}
}

func thriftListOpenWorkflowExecutionsResponse(t *apiv1.ListOpenWorkflowExecutionsResponse) *shared.ListOpenWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &shared.ListOpenWorkflowExecutionsResponse{
		Executions:    thriftWorkflowExecutionInfoArray(t.Executions),
		NextPageToken: t.NextPageToken,
	}
}

func thriftListTaskListPartitionsResponse(t *apiv1.ListTaskListPartitionsResponse) *shared.ListTaskListPartitionsResponse {
	if t == nil {
		return nil
	}
	return &shared.ListTaskListPartitionsResponse{
		ActivityTaskListPartitions: thriftTaskListPartitionMetadataArray(t.ActivityTaskListPartitions),
		DecisionTaskListPartitions: thriftTaskListPartitionMetadataArray(t.DecisionTaskListPartitions),
	}
}

func thriftListWorkflowExecutionsResponse(t *apiv1.ListWorkflowExecutionsResponse) *shared.ListWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &shared.ListWorkflowExecutionsResponse{
		Executions:    thriftWorkflowExecutionInfoArray(t.Executions),
		NextPageToken: t.NextPageToken,
	}
}


func thriftPollForActivityTaskResponse(t *apiv1.PollForActivityTaskResponse) *shared.PollForActivityTaskResponse {
	if t == nil {
		return nil
	}
	return &shared.PollForActivityTaskResponse{
		TaskToken:                       t.TaskToken,
		WorkflowExecution:               thriftWorkflowExecution(t.WorkflowExecution),
		ActivityId:                      &t.ActivityId,
		ActivityType:                    thriftActivityType(t.ActivityType),
		Input:                           thriftPayload(t.Input),
		ScheduledTimestamp:              timeToUnixNano(t.ScheduledTime),
		StartedTimestamp:                timeToUnixNano(t.StartedTime),
		ScheduleToCloseTimeoutSeconds:   durationToSeconds(t.ScheduleToCloseTimeout),
		StartToCloseTimeoutSeconds:      durationToSeconds(t.StartToCloseTimeout),
		HeartbeatTimeoutSeconds:         durationToSeconds(t.HeartbeatTimeout),
		Attempt:                         &t.Attempt,
		ScheduledTimestampOfThisAttempt: timeToUnixNano(t.ScheduledTimeOfThisAttempt),
		HeartbeatDetails:                thriftPayload(t.HeartbeatDetails),
		WorkflowType:                    thriftWorkflowType(t.WorkflowType),
		WorkflowDomain:                  &t.WorkflowDomain,
		Header:                          thriftHeader(t.Header),
	}
}


func thriftPollForDecisionTaskResponse(t *apiv1.PollForDecisionTaskResponse) *shared.PollForDecisionTaskResponse {
	if t == nil {
		return nil
	}
	return &shared.PollForDecisionTaskResponse{
		TaskToken:                 t.TaskToken,
		WorkflowExecution:         thriftWorkflowExecution(t.WorkflowExecution),
		WorkflowType:              thriftWorkflowType(t.WorkflowType),
		PreviousStartedEventId:    toInt64Value(t.PreviousStartedEventId),
		StartedEventId:            &t.StartedEventId,
		Attempt:                   &t.Attempt,
		BacklogCountHint:          &t.BacklogCountHint,
		History:                   thriftHistory(t.History),
		NextPageToken:             t.NextPageToken,
		Query:                     thriftWorkflowQuery(t.Query),
		WorkflowExecutionTaskList: thriftTaskList(t.WorkflowExecutionTaskList),
		ScheduledTimestamp:        timeToUnixNano(t.ScheduledTime),
		StartedTimestamp:          timeToUnixNano(t.StartedTime),
		Queries:                   thriftWorkflowQueryMap(t.Queries),
		NextEventId:               &t.NextEventId,
	}
}

func thriftQueryWorkflowResponse(t *apiv1.QueryWorkflowResponse) *shared.QueryWorkflowResponse {
	if t == nil {
		return nil
	}
	return &shared.QueryWorkflowResponse{
		QueryResult:   thriftPayload(t.QueryResult),
		QueryRejected: thriftQueryRejected(t.QueryRejected),
	}
}

func thriftRecordActivityTaskHeartbeatByIdResponse(t *apiv1.RecordActivityTaskHeartbeatByIDResponse) *shared.RecordActivityTaskHeartbeatResponse {
	if t == nil {
		return nil
	}
	return &shared.RecordActivityTaskHeartbeatResponse{
		CancelRequested: &t.CancelRequested,
	}
}

func thriftRecordActivityTaskHeartbeatResponse(t *apiv1.RecordActivityTaskHeartbeatResponse) *shared.RecordActivityTaskHeartbeatResponse {
	if t == nil {
		return nil
	}
	return &shared.RecordActivityTaskHeartbeatResponse{
		CancelRequested: &t.CancelRequested,
	}
}

func thriftResetWorkflowExecutionResponse(t *apiv1.ResetWorkflowExecutionResponse) *shared.ResetWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &shared.ResetWorkflowExecutionResponse{
		RunId: &t.RunId,
	}
}

func thriftRespondDecisionTaskCompletedResponse(t *apiv1.RespondDecisionTaskCompletedResponse) *shared.RespondDecisionTaskCompletedResponse {
	if t == nil {
		return nil
	}
	return &shared.RespondDecisionTaskCompletedResponse{
		DecisionTask:                thriftPollForDecisionTaskResponse(t.DecisionTask),
		ActivitiesToDispatchLocally: thriftActivityLocalDispatchInfoMap(t.ActivitiesToDispatchLocally),
	}
}

func thriftScanWorkflowExecutionsResponse(t *apiv1.ScanWorkflowExecutionsResponse) *shared.ListWorkflowExecutionsResponse {
	if t == nil {
		return nil
	}
	return &shared.ListWorkflowExecutionsResponse{
		Executions:    thriftWorkflowExecutionInfoArray(t.Executions),
		NextPageToken: t.NextPageToken,
	}
}

func thriftSignalWithStartWorkflowExecutionResponse(t *apiv1.SignalWithStartWorkflowExecutionResponse) *shared.StartWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &shared.StartWorkflowExecutionResponse{
		RunId: &t.RunId,
	}
}

func thriftStartWorkflowExecutionResponse(t *apiv1.StartWorkflowExecutionResponse) *shared.StartWorkflowExecutionResponse {
	if t == nil {
		return nil
	}
	return &shared.StartWorkflowExecutionResponse{
		RunId: &t.RunId,
	}
}

func thriftUpdateDomainResponse(t *apiv1.UpdateDomainResponse) *shared.UpdateDomainResponse {
	if t == nil || t.Domain == nil {
		return nil
	}
	return &shared.UpdateDomainResponse{
		DomainInfo: &shared.DomainInfo{
			Name:        &t.Domain.Name,
			Status:      thriftDomainStatus(t.Domain.Status),
			Description: &t.Domain.Description,
			OwnerEmail:  &t.Domain.OwnerEmail,
			Data:        t.Domain.Data,
			UUID:        &t.Domain.Id,
		},
		Configuration: &shared.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: durationToDays(t.Domain.WorkflowExecutionRetentionPeriod),
			EmitMetric:                             boolPtr(true),
			BadBinaries:                            thriftBadBinaries(t.Domain.BadBinaries),
			HistoryArchivalStatus:                  thriftArchivalStatus(t.Domain.HistoryArchivalStatus),
			HistoryArchivalURI:                     &t.Domain.HistoryArchivalUri,
			VisibilityArchivalStatus:               thriftArchivalStatus(t.Domain.VisibilityArchivalStatus),
			VisibilityArchivalURI:                  &t.Domain.VisibilityArchivalUri,
		},
		ReplicationConfiguration: &shared.DomainReplicationConfiguration{
			ActiveClusterName: &t.Domain.ActiveClusterName,
			Clusters:          thriftClusterReplicationConfigurationArray(t.Domain.Clusters),
		},
		FailoverVersion: &t.Domain.FailoverVersion,
		IsGlobalDomain:  &t.Domain.IsGlobalDomain,
	}
}


func thriftPayload(p *apiv1.Payload) []byte {
	if p == nil {
		return nil
	}
	if p.Data == nil {
		// FromPayload will not generate this case
		// however, Data field will be dropped by the encoding if it's empty
		// and receiver side will see nil for the Data field
		// since we already know p is not nil, Data field must be an empty byte array
		return []byte{}
	}
	return p.Data
}

func thriftFailureReason(failure *apiv1.Failure) *string {
	if failure == nil {
		return nil
	}
	return &failure.Reason
}

func thriftFailureDetails(failure *apiv1.Failure) []byte {
	if failure == nil {
		return nil
	}
	return failure.Details
}

func thriftDataBlob(t *apiv1.DataBlob) *shared.DataBlob {
	if t == nil {
		return nil
	}
	return &shared.DataBlob{
		EncodingType: thriftEncodingType(t.EncodingType),
		Data:         t.Data,
	}
}

func thriftWorkflowExecution(t *apiv1.WorkflowExecution) *shared.WorkflowExecution {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecution{
		WorkflowId: &t.WorkflowId,
		RunId:      &t.RunId,
	}
}

func thriftWorkflowId(t *apiv1.WorkflowExecution) *string {
	if t == nil {
		return nil
	}
	return &t.WorkflowId
}

func thriftRunId(t *apiv1.WorkflowExecution) *string {
	if t == nil {
		return nil
	}
	return &t.RunId
}

func thriftExternalWorkflowExecution(t *apiv1.ExternalExecutionInfo) *shared.WorkflowExecution {
	if t == nil {
		return nil
	}
	return thriftWorkflowExecution(t.WorkflowExecution)
}

func thriftWorkflowType(t *apiv1.WorkflowType) *shared.WorkflowType {
	if t == nil {
		return nil
	}
	return &shared.WorkflowType{
		Name: &t.Name,
	}
}

func thriftActivityType(t *apiv1.ActivityType) *shared.ActivityType {
	if t == nil {
		return nil
	}
	return &shared.ActivityType{
		Name: &t.Name,
	}
}

func thriftHeader(t *apiv1.Header) *shared.Header {
	if t == nil {
		return nil
	}
	return &shared.Header{
		Fields: thriftPayloadMap(t.Fields),
	}
}

func thriftMemo(t *apiv1.Memo) *shared.Memo {
	if t == nil {
		return nil
	}
	return &shared.Memo{
		Fields: thriftPayloadMap(t.Fields),
	}
}

func thriftSearchAttributes(t *apiv1.SearchAttributes) *shared.SearchAttributes {
	if t == nil {
		return nil
	}
	return &shared.SearchAttributes{
		IndexedFields: thriftPayloadMap(t.IndexedFields),
	}
}

func thriftRetryPolicy(t *apiv1.RetryPolicy) *shared.RetryPolicy {
	if t == nil {
		return nil
	}
	return &shared.RetryPolicy{
		InitialIntervalInSeconds:    durationToSeconds(t.InitialInterval),
		BackoffCoefficient:          &t.BackoffCoefficient,
		MaximumIntervalInSeconds:    durationToSeconds(t.MaximumInterval),
		MaximumAttempts:             &t.MaximumAttempts,
		NonRetriableErrorReasons:    t.NonRetryableErrorReasons,
		ExpirationIntervalInSeconds: durationToSeconds(t.ExpirationInterval),
	}
}

func thriftTaskList(t *apiv1.TaskList) *shared.TaskList {
	if t == nil {
		return nil
	}
	return &shared.TaskList{
		Name: &t.Name,
		Kind: thriftTaskListKind(t.Kind),
	}
}

func thriftBadBinaries(t *apiv1.BadBinaries) *shared.BadBinaries {
	if t == nil {
		return nil
	}
	return &shared.BadBinaries{
		Binaries: thriftBadBinaryInfoMap(t.Binaries),
	}
}

func thriftBadBinaryInfo(t *apiv1.BadBinaryInfo) *shared.BadBinaryInfo {
	if t == nil {
		return nil
	}
	return &shared.BadBinaryInfo{
		Reason:          &t.Reason,
		Operator:        &t.Operator,
		CreatedTimeNano: timeToUnixNano(t.CreatedTime),
	}
}

func thriftResetPoints(t *apiv1.ResetPoints) *shared.ResetPoints {
	if t == nil {
		return nil
	}
	return &shared.ResetPoints{
		Points: thriftResetPointInfoArray(t.Points),
	}
}

func thriftClusterReplicationConfiguration(t *apiv1.ClusterReplicationConfiguration) *shared.ClusterReplicationConfiguration {
	if t == nil {
		return nil
	}
	return &shared.ClusterReplicationConfiguration{
		ClusterName: &t.ClusterName,
	}
}

func thriftPollerInfo(t *apiv1.PollerInfo) *shared.PollerInfo {
	if t == nil {
		return nil
	}
	return &shared.PollerInfo{
		LastAccessTime: timeToUnixNano(t.LastAccessTime),
		Identity:       &t.Identity,
		RatePerSecond:  &t.RatePerSecond,
	}
}

func thriftResetPointInfo(t *apiv1.ResetPointInfo) *shared.ResetPointInfo {
	if t == nil {
		return nil
	}
	return &shared.ResetPointInfo{
		BinaryChecksum:           &t.BinaryChecksum,
		RunId:                    &t.RunId,
		FirstDecisionCompletedId: &t.FirstDecisionCompletedId,
		CreatedTimeNano:          timeToUnixNano(t.CreatedTime),
		ExpiringTimeNano:         timeToUnixNano(t.ExpiringTime),
		Resettable:               &t.Resettable,
	}
}

func thriftTaskListStatus(t *apiv1.TaskListStatus) *shared.TaskListStatus {
	if t == nil {
		return nil
	}
	return &shared.TaskListStatus{
		BacklogCountHint: &t.BacklogCountHint,
		ReadLevel:        &t.ReadLevel,
		AckLevel:         &t.AckLevel,
		RatePerSecond:    &t.RatePerSecond,
		TaskIDBlock:      thriftTaskIdBlock(t.TaskIdBlock),
	}
}

func thriftTaskIdBlock(t *apiv1.TaskIDBlock) *shared.TaskIDBlock {
	if t == nil {
		return nil
	}
	return &shared.TaskIDBlock{
		StartID: &t.StartId,
		EndID:   &t.EndId,
	}
}

func thriftWorkflowExecutionConfiguration(t *apiv1.WorkflowExecutionConfiguration) *shared.WorkflowExecutionConfiguration {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionConfiguration{
		TaskList:                            thriftTaskList(t.TaskList),
		ExecutionStartToCloseTimeoutSeconds: durationToSeconds(t.ExecutionStartToCloseTimeout),
		TaskStartToCloseTimeoutSeconds:      durationToSeconds(t.TaskStartToCloseTimeout),
	}
}

func thriftWorkflowExecutionInfo(t *apiv1.WorkflowExecutionInfo) *shared.WorkflowExecutionInfo {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionInfo{
		Execution:         thriftWorkflowExecution(t.WorkflowExecution),
		Type:              thriftWorkflowType(t.Type),
		StartTime:         timeToUnixNano(t.StartTime),
		CloseTime:         timeToUnixNano(t.CloseTime),
		CloseStatus:       thriftWorkflowExecutionCloseStatus(t.CloseStatus),
		HistoryLength:     &t.HistoryLength,
		ParentDomainId:    thriftParentDomainId(t.ParentExecutionInfo),
		ParentExecution:   thriftParentWorkflowExecution(t.ParentExecutionInfo),
		ExecutionTime:     timeToUnixNano(t.ExecutionTime),
		Memo:              thriftMemo(t.Memo),
		SearchAttributes:  thriftSearchAttributes(t.SearchAttributes),
		AutoResetPoints:   thriftResetPoints(t.AutoResetPoints),
		TaskList:          &t.TaskList,
		IsCron:            &t.IsCron,
	}
}

func thriftParentDomainId(pei *apiv1.ParentExecutionInfo) *string {
	if pei == nil {
		return nil
	}
	return &pei.DomainId
}

func thriftParentDomainName(pei *apiv1.ParentExecutionInfo) *string {
	if pei == nil {
		return nil
	}
	return &pei.DomainName
}

func thriftParentInitiatedId(pei *apiv1.ParentExecutionInfo) *int64 {
	if pei == nil {
		return nil
	}
	return &pei.InitiatedId
}

func thriftParentWorkflowExecution(pei *apiv1.ParentExecutionInfo) *shared.WorkflowExecution {
	if pei == nil {
		return nil
	}
	return thriftWorkflowExecution(pei.WorkflowExecution)
}

func thriftPendingActivityInfo(t *apiv1.PendingActivityInfo) *shared.PendingActivityInfo {
	if t == nil {
		return nil
	}
	return &shared.PendingActivityInfo{
		ActivityID:             &t.ActivityId,
		ActivityType:           thriftActivityType(t.ActivityType),
		State:                  thriftPendingActivityState(t.State),
		HeartbeatDetails:       thriftPayload(t.HeartbeatDetails),
		LastHeartbeatTimestamp: timeToUnixNano(t.LastHeartbeatTime),
		LastStartedTimestamp:   timeToUnixNano(t.LastStartedTime),
		Attempt:                &t.Attempt,
		MaximumAttempts:        &t.MaximumAttempts,
		ScheduledTimestamp:     timeToUnixNano(t.ScheduledTime),
		ExpirationTimestamp:    timeToUnixNano(t.ExpirationTime),
		LastFailureReason:      thriftFailureReason(t.LastFailure),
		LastFailureDetails:     thriftFailureDetails(t.LastFailure),
		LastWorkerIdentity:     &t.LastWorkerIdentity,
	}
}

func thriftPendingChildExecutionInfo(t *apiv1.PendingChildExecutionInfo) *shared.PendingChildExecutionInfo {
	if t == nil {
		return nil
	}
	return &shared.PendingChildExecutionInfo{
		WorkflowID:        thriftWorkflowId(t.WorkflowExecution),
		RunID:             thriftRunId(t.WorkflowExecution),
		WorkflowTypName:   &t.WorkflowTypeName,
		InitiatedID:       &t.InitiatedId,
		ParentClosePolicy: thriftParentClosePolicy(t.ParentClosePolicy),
	}
}

func thriftPendingDecisionInfo(t *apiv1.PendingDecisionInfo) *shared.PendingDecisionInfo {
	if t == nil {
		return nil
	}
	return &shared.PendingDecisionInfo{
		State:                      thriftPendingDecisionState(t.State),
		ScheduledTimestamp:         timeToUnixNano(t.ScheduledTime),
		StartedTimestamp:           timeToUnixNano(t.StartedTime),
		Attempt:                    common.Int64Ptr(int64(t.Attempt)),
		OriginalScheduledTimestamp: timeToUnixNano(t.OriginalScheduledTime),
	}
}

func thriftActivityLocalDispatchInfo(t *apiv1.ActivityLocalDispatchInfo) *shared.ActivityLocalDispatchInfo {
	if t == nil {
		return nil
	}
	return &shared.ActivityLocalDispatchInfo{
		ActivityId:                      &t.ActivityId,
		ScheduledTimestamp:              timeToUnixNano(t.ScheduledTime),
		StartedTimestamp:                timeToUnixNano(t.StartedTime),
		ScheduledTimestampOfThisAttempt: timeToUnixNano(t.ScheduledTimeOfThisAttempt),
		TaskToken:                       t.TaskToken,
	}
}

func thriftSupportedClientVersions(t *apiv1.SupportedClientVersions) *shared.SupportedClientVersions {
	if t == nil {
		return nil
	}
	return &shared.SupportedClientVersions{
		GoSdk:   &t.GoSdk,
		JavaSdk: &t.JavaSdk,
	}
}

func thriftWorkflowQuery(t *apiv1.WorkflowQuery) *shared.WorkflowQuery {
	if t == nil {
		return nil
	}
	return &shared.WorkflowQuery{
		QueryType: &t.QueryType,
		QueryArgs: thriftPayload(t.QueryArgs),
	}
}

func thriftDescribeDomainResponseDomain(t *apiv1.Domain) *shared.DescribeDomainResponse {
	if t == nil {
		return nil
	}
	return &shared.DescribeDomainResponse{
		DomainInfo: &shared.DomainInfo{
			Name:        &t.Name,
			Status:      thriftDomainStatus(t.Status),
			Description: &t.Description,
			OwnerEmail:  &t.OwnerEmail,
			Data:        t.Data,
			UUID:        &t.Id,
		},
		Configuration: &shared.DomainConfiguration{
			WorkflowExecutionRetentionPeriodInDays: durationToDays(t.WorkflowExecutionRetentionPeriod),
			EmitMetric:                             boolPtr(true),
			BadBinaries:                            thriftBadBinaries(t.BadBinaries),
			HistoryArchivalStatus:                  thriftArchivalStatus(t.HistoryArchivalStatus),
			HistoryArchivalURI:                     &t.HistoryArchivalUri,
			VisibilityArchivalStatus:               thriftArchivalStatus(t.VisibilityArchivalStatus),
			VisibilityArchivalURI:                  &t.VisibilityArchivalUri,
		},
		ReplicationConfiguration: &shared.DomainReplicationConfiguration{
			ActiveClusterName: &t.ActiveClusterName,
			Clusters:          thriftClusterReplicationConfigurationArray(t.Clusters),
		},
		FailoverVersion: &t.FailoverVersion,
		IsGlobalDomain:  &t.IsGlobalDomain,
	}
}

func thriftTaskListPartitionMetadata(t *apiv1.TaskListPartitionMetadata) *shared.TaskListPartitionMetadata {
	if t == nil {
		return nil
	}
	return &shared.TaskListPartitionMetadata{
		Key:           &t.Key,
		OwnerHostName: &t.OwnerHostName,
	}
}

func thriftQueryRejected(t *apiv1.QueryRejected) *shared.QueryRejected {
	if t == nil {
		return nil
	}
	return &shared.QueryRejected{
		CloseStatus: thriftWorkflowExecutionCloseStatus(t.CloseStatus),
	}
}

func thriftExternalInitiatedId(t *apiv1.ExternalExecutionInfo) *int64 {
	if t == nil {
		return nil
	}
	return &t.InitiatedId
}

func thriftHistory(t *apiv1.History) *shared.History {
	if t == nil {
		return nil
	}
	return &shared.History{
		Events: thriftHistoryEventArray(t.Events),
	}
}

func thriftHistoryEvent(e *apiv1.HistoryEvent) *shared.HistoryEvent {
	if e == nil {
		return nil
	}
	event := shared.HistoryEvent{
		EventId:   &e.EventId,
		Timestamp: timeToUnixNano(e.EventTime),
		Version:   &e.Version,
		TaskId:    &e.TaskId,
	}
	switch attr := e.Attributes.(type) {
	case *apiv1.HistoryEvent_WorkflowExecutionStartedEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionStarted.Ptr()
		event.WorkflowExecutionStartedEventAttributes = thriftWorkflowExecutionStartedEventAttributes(attr.WorkflowExecutionStartedEventAttributes)
	case *apiv1.HistoryEvent_WorkflowExecutionCompletedEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionCompleted.Ptr()
		event.WorkflowExecutionCompletedEventAttributes = thriftWorkflowExecutionCompletedEventAttributes(attr.WorkflowExecutionCompletedEventAttributes)
	case *apiv1.HistoryEvent_WorkflowExecutionFailedEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionFailed.Ptr()
		event.WorkflowExecutionFailedEventAttributes = thriftWorkflowExecutionFailedEventAttributes(attr.WorkflowExecutionFailedEventAttributes)
	case *apiv1.HistoryEvent_WorkflowExecutionTimedOutEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionTimedOut.Ptr()
		event.WorkflowExecutionTimedOutEventAttributes = thriftWorkflowExecutionTimedOutEventAttributes(attr.WorkflowExecutionTimedOutEventAttributes)
	case *apiv1.HistoryEvent_DecisionTaskScheduledEventAttributes:
		event.EventType = shared.EventTypeDecisionTaskScheduled.Ptr()
		event.DecisionTaskScheduledEventAttributes = thriftDecisionTaskScheduledEventAttributes(attr.DecisionTaskScheduledEventAttributes)
	case *apiv1.HistoryEvent_DecisionTaskStartedEventAttributes:
		event.EventType = shared.EventTypeDecisionTaskStarted.Ptr()
		event.DecisionTaskStartedEventAttributes = thriftDecisionTaskStartedEventAttributes(attr.DecisionTaskStartedEventAttributes)
	case *apiv1.HistoryEvent_DecisionTaskCompletedEventAttributes:
		event.EventType = shared.EventTypeDecisionTaskCompleted.Ptr()
		event.DecisionTaskCompletedEventAttributes = thriftDecisionTaskCompletedEventAttributes(attr.DecisionTaskCompletedEventAttributes)
	case *apiv1.HistoryEvent_DecisionTaskTimedOutEventAttributes:
		event.EventType = shared.EventTypeDecisionTaskTimedOut.Ptr()
		event.DecisionTaskTimedOutEventAttributes = thriftDecisionTaskTimedOutEventAttributes(attr.DecisionTaskTimedOutEventAttributes)
	case *apiv1.HistoryEvent_DecisionTaskFailedEventAttributes:
		event.EventType = shared.EventTypeDecisionTaskFailed.Ptr()
		event.DecisionTaskFailedEventAttributes = thriftDecisionTaskFailedEventAttributes(attr.DecisionTaskFailedEventAttributes)
	case *apiv1.HistoryEvent_ActivityTaskScheduledEventAttributes:
		event.EventType = shared.EventTypeActivityTaskScheduled.Ptr()
		event.ActivityTaskScheduledEventAttributes = thriftActivityTaskScheduledEventAttributes(attr.ActivityTaskScheduledEventAttributes)
	case *apiv1.HistoryEvent_ActivityTaskStartedEventAttributes:
		event.EventType = shared.EventTypeActivityTaskStarted.Ptr()
		event.ActivityTaskStartedEventAttributes = thriftActivityTaskStartedEventAttributes(attr.ActivityTaskStartedEventAttributes)
	case *apiv1.HistoryEvent_ActivityTaskCompletedEventAttributes:
		event.EventType = shared.EventTypeActivityTaskCompleted.Ptr()
		event.ActivityTaskCompletedEventAttributes = thriftActivityTaskCompletedEventAttributes(attr.ActivityTaskCompletedEventAttributes)
	case *apiv1.HistoryEvent_ActivityTaskFailedEventAttributes:
		event.EventType = shared.EventTypeActivityTaskFailed.Ptr()
		event.ActivityTaskFailedEventAttributes = thriftActivityTaskFailedEventAttributes(attr.ActivityTaskFailedEventAttributes)
	case *apiv1.HistoryEvent_ActivityTaskTimedOutEventAttributes:
		event.EventType = shared.EventTypeActivityTaskTimedOut.Ptr()
		event.ActivityTaskTimedOutEventAttributes = thriftActivityTaskTimedOutEventAttributes(attr.ActivityTaskTimedOutEventAttributes)
	case *apiv1.HistoryEvent_TimerStartedEventAttributes:
		event.EventType = shared.EventTypeTimerStarted.Ptr()
		event.TimerStartedEventAttributes = thriftTimerStartedEventAttributes(attr.TimerStartedEventAttributes)
	case *apiv1.HistoryEvent_TimerFiredEventAttributes:
		event.EventType = shared.EventTypeTimerFired.Ptr()
		event.TimerFiredEventAttributes = thriftTimerFiredEventAttributes(attr.TimerFiredEventAttributes)
	case *apiv1.HistoryEvent_ActivityTaskCancelRequestedEventAttributes:
		event.EventType = shared.EventTypeActivityTaskCancelRequested.Ptr()
		event.ActivityTaskCancelRequestedEventAttributes = thriftActivityTaskCancelRequestedEventAttributes(attr.ActivityTaskCancelRequestedEventAttributes)
	case *apiv1.HistoryEvent_RequestCancelActivityTaskFailedEventAttributes:
		event.EventType = shared.EventTypeRequestCancelActivityTaskFailed.Ptr()
		event.RequestCancelActivityTaskFailedEventAttributes = thriftRequestCancelActivityTaskFailedEventAttributes(attr.RequestCancelActivityTaskFailedEventAttributes)
	case *apiv1.HistoryEvent_ActivityTaskCanceledEventAttributes:
		event.EventType = shared.EventTypeActivityTaskCanceled.Ptr()
		event.ActivityTaskCanceledEventAttributes = thriftActivityTaskCanceledEventAttributes(attr.ActivityTaskCanceledEventAttributes)
	case *apiv1.HistoryEvent_TimerCanceledEventAttributes:
		event.EventType = shared.EventTypeTimerCanceled.Ptr()
		event.TimerCanceledEventAttributes = thriftTimerCanceledEventAttributes(attr.TimerCanceledEventAttributes)
	case *apiv1.HistoryEvent_CancelTimerFailedEventAttributes:
		event.EventType = shared.EventTypeCancelTimerFailed.Ptr()
		event.CancelTimerFailedEventAttributes = thriftCancelTimerFailedEventAttributes(attr.CancelTimerFailedEventAttributes)
	case *apiv1.HistoryEvent_MarkerRecordedEventAttributes:
		event.EventType = shared.EventTypeMarkerRecorded.Ptr()
		event.MarkerRecordedEventAttributes = thriftMarkerRecordedEventAttributes(attr.MarkerRecordedEventAttributes)
	case *apiv1.HistoryEvent_WorkflowExecutionSignaledEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionSignaled.Ptr()
		event.WorkflowExecutionSignaledEventAttributes = thriftWorkflowExecutionSignaledEventAttributes(attr.WorkflowExecutionSignaledEventAttributes)
	case *apiv1.HistoryEvent_WorkflowExecutionTerminatedEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionTerminated.Ptr()
		event.WorkflowExecutionTerminatedEventAttributes = thriftWorkflowExecutionTerminatedEventAttributes(attr.WorkflowExecutionTerminatedEventAttributes)
	case *apiv1.HistoryEvent_WorkflowExecutionCancelRequestedEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionCancelRequested.Ptr()
		event.WorkflowExecutionCancelRequestedEventAttributes = thriftWorkflowExecutionCancelRequestedEventAttributes(attr.WorkflowExecutionCancelRequestedEventAttributes)
	case *apiv1.HistoryEvent_WorkflowExecutionCanceledEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionCanceled.Ptr()
		event.WorkflowExecutionCanceledEventAttributes = thriftWorkflowExecutionCanceledEventAttributes(attr.WorkflowExecutionCanceledEventAttributes)
	case *apiv1.HistoryEvent_RequestCancelExternalWorkflowExecutionInitiatedEventAttributes:
		event.EventType = shared.EventTypeRequestCancelExternalWorkflowExecutionInitiated.Ptr()
		event.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes = thriftRequestCancelExternalWorkflowExecutionInitiatedEventAttributes(attr.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes)
	case *apiv1.HistoryEvent_RequestCancelExternalWorkflowExecutionFailedEventAttributes:
		event.EventType = shared.EventTypeRequestCancelExternalWorkflowExecutionFailed.Ptr()
		event.RequestCancelExternalWorkflowExecutionFailedEventAttributes = thriftRequestCancelExternalWorkflowExecutionFailedEventAttributes(attr.RequestCancelExternalWorkflowExecutionFailedEventAttributes)
	case *apiv1.HistoryEvent_ExternalWorkflowExecutionCancelRequestedEventAttributes:
		event.EventType = shared.EventTypeExternalWorkflowExecutionCancelRequested.Ptr()
		event.ExternalWorkflowExecutionCancelRequestedEventAttributes = thriftExternalWorkflowExecutionCancelRequestedEventAttributes(attr.ExternalWorkflowExecutionCancelRequestedEventAttributes)
	case *apiv1.HistoryEvent_WorkflowExecutionContinuedAsNewEventAttributes:
		event.EventType = shared.EventTypeWorkflowExecutionContinuedAsNew.Ptr()
		event.WorkflowExecutionContinuedAsNewEventAttributes = thriftWorkflowExecutionContinuedAsNewEventAttributes(attr.WorkflowExecutionContinuedAsNewEventAttributes)
	case *apiv1.HistoryEvent_StartChildWorkflowExecutionInitiatedEventAttributes:
		event.EventType = shared.EventTypeStartChildWorkflowExecutionInitiated.Ptr()
		event.StartChildWorkflowExecutionInitiatedEventAttributes = thriftStartChildWorkflowExecutionInitiatedEventAttributes(attr.StartChildWorkflowExecutionInitiatedEventAttributes)
	case *apiv1.HistoryEvent_StartChildWorkflowExecutionFailedEventAttributes:
		event.EventType = shared.EventTypeStartChildWorkflowExecutionFailed.Ptr()
		event.StartChildWorkflowExecutionFailedEventAttributes = thriftStartChildWorkflowExecutionFailedEventAttributes(attr.StartChildWorkflowExecutionFailedEventAttributes)
	case *apiv1.HistoryEvent_ChildWorkflowExecutionStartedEventAttributes:
		event.EventType = shared.EventTypeChildWorkflowExecutionStarted.Ptr()
		event.ChildWorkflowExecutionStartedEventAttributes = thriftChildWorkflowExecutionStartedEventAttributes(attr.ChildWorkflowExecutionStartedEventAttributes)
	case *apiv1.HistoryEvent_ChildWorkflowExecutionCompletedEventAttributes:
		event.EventType = shared.EventTypeChildWorkflowExecutionCompleted.Ptr()
		event.ChildWorkflowExecutionCompletedEventAttributes = thriftChildWorkflowExecutionCompletedEventAttributes(attr.ChildWorkflowExecutionCompletedEventAttributes)
	case *apiv1.HistoryEvent_ChildWorkflowExecutionFailedEventAttributes:
		event.EventType = shared.EventTypeChildWorkflowExecutionFailed.Ptr()
		event.ChildWorkflowExecutionFailedEventAttributes = thriftChildWorkflowExecutionFailedEventAttributes(attr.ChildWorkflowExecutionFailedEventAttributes)
	case *apiv1.HistoryEvent_ChildWorkflowExecutionCanceledEventAttributes:
		event.EventType = shared.EventTypeChildWorkflowExecutionCanceled.Ptr()
		event.ChildWorkflowExecutionCanceledEventAttributes = thriftChildWorkflowExecutionCanceledEventAttributes(attr.ChildWorkflowExecutionCanceledEventAttributes)
	case *apiv1.HistoryEvent_ChildWorkflowExecutionTimedOutEventAttributes:
		event.EventType = shared.EventTypeChildWorkflowExecutionTimedOut.Ptr()
		event.ChildWorkflowExecutionTimedOutEventAttributes = thriftChildWorkflowExecutionTimedOutEventAttributes(attr.ChildWorkflowExecutionTimedOutEventAttributes)
	case *apiv1.HistoryEvent_ChildWorkflowExecutionTerminatedEventAttributes:
		event.EventType = shared.EventTypeChildWorkflowExecutionTerminated.Ptr()
		event.ChildWorkflowExecutionTerminatedEventAttributes = thriftChildWorkflowExecutionTerminatedEventAttributes(attr.ChildWorkflowExecutionTerminatedEventAttributes)
	case *apiv1.HistoryEvent_SignalExternalWorkflowExecutionInitiatedEventAttributes:
		event.EventType = shared.EventTypeSignalExternalWorkflowExecutionInitiated.Ptr()
		event.SignalExternalWorkflowExecutionInitiatedEventAttributes = thriftSignalExternalWorkflowExecutionInitiatedEventAttributes(attr.SignalExternalWorkflowExecutionInitiatedEventAttributes)
	case *apiv1.HistoryEvent_SignalExternalWorkflowExecutionFailedEventAttributes:
		event.EventType = shared.EventTypeSignalExternalWorkflowExecutionFailed.Ptr()
		event.SignalExternalWorkflowExecutionFailedEventAttributes = thriftSignalExternalWorkflowExecutionFailedEventAttributes(attr.SignalExternalWorkflowExecutionFailedEventAttributes)
	case *apiv1.HistoryEvent_ExternalWorkflowExecutionSignaledEventAttributes:
		event.EventType = shared.EventTypeExternalWorkflowExecutionSignaled.Ptr()
		event.ExternalWorkflowExecutionSignaledEventAttributes = thriftExternalWorkflowExecutionSignaledEventAttributes(attr.ExternalWorkflowExecutionSignaledEventAttributes)
	case *apiv1.HistoryEvent_UpsertWorkflowSearchAttributesEventAttributes:
		event.EventType = shared.EventTypeUpsertWorkflowSearchAttributes.Ptr()
		event.UpsertWorkflowSearchAttributesEventAttributes = thriftUpsertWorkflowSearchAttributesEventAttributes(attr.UpsertWorkflowSearchAttributesEventAttributes)
	}
	return &event
}

func thriftActivityTaskCancelRequestedEventAttributes(t *apiv1.ActivityTaskCancelRequestedEventAttributes) *shared.ActivityTaskCancelRequestedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ActivityTaskCancelRequestedEventAttributes{
		ActivityId:                   &t.ActivityId,
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
	}
}

func thriftActivityTaskCanceledEventAttributes(t *apiv1.ActivityTaskCanceledEventAttributes) *shared.ActivityTaskCanceledEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ActivityTaskCanceledEventAttributes{
		Details:                      thriftPayload(t.Details),
		LatestCancelRequestedEventId: &t.LatestCancelRequestedEventId,
		ScheduledEventId:             &t.ScheduledEventId,
		StartedEventId:               &t.StartedEventId,
		Identity:                     &t.Identity,
	}
}

func thriftActivityTaskCompletedEventAttributes(t *apiv1.ActivityTaskCompletedEventAttributes) *shared.ActivityTaskCompletedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ActivityTaskCompletedEventAttributes{
		Result:           thriftPayload(t.Result),
		ScheduledEventId: &t.ScheduledEventId,
		StartedEventId:   &t.StartedEventId,
		Identity:         &t.Identity,
	}
}

func thriftActivityTaskFailedEventAttributes(t *apiv1.ActivityTaskFailedEventAttributes) *shared.ActivityTaskFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ActivityTaskFailedEventAttributes{
		Reason:           thriftFailureReason(t.Failure),
		Details:          thriftFailureDetails(t.Failure),
		ScheduledEventId: &t.ScheduledEventId,
		StartedEventId:   &t.StartedEventId,
		Identity:         &t.Identity,
	}
}

func thriftActivityTaskScheduledEventAttributes(t *apiv1.ActivityTaskScheduledEventAttributes) *shared.ActivityTaskScheduledEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ActivityTaskScheduledEventAttributes{
		ActivityId:                    &t.ActivityId,
		ActivityType:                  thriftActivityType(t.ActivityType),
		Domain:                        &t.Domain,
		TaskList:                      thriftTaskList(t.TaskList),
		Input:                         thriftPayload(t.Input),
		ScheduleToCloseTimeoutSeconds: durationToSeconds(t.ScheduleToCloseTimeout),
		ScheduleToStartTimeoutSeconds: durationToSeconds(t.ScheduleToStartTimeout),
		StartToCloseTimeoutSeconds:    durationToSeconds(t.StartToCloseTimeout),
		HeartbeatTimeoutSeconds:       durationToSeconds(t.HeartbeatTimeout),
		DecisionTaskCompletedEventId:  &t.DecisionTaskCompletedEventId,
		RetryPolicy:                   thriftRetryPolicy(t.RetryPolicy),
		Header:                        thriftHeader(t.Header),
	}
}

func thriftActivityTaskStartedEventAttributes(t *apiv1.ActivityTaskStartedEventAttributes) *shared.ActivityTaskStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ActivityTaskStartedEventAttributes{
		ScheduledEventId:   &t.ScheduledEventId,
		Identity:           &t.Identity,
		RequestId:          &t.RequestId,
		Attempt:            &t.Attempt,
		LastFailureReason:  thriftFailureReason(t.LastFailure),
		LastFailureDetails: thriftFailureDetails(t.LastFailure),
	}
}

func thriftActivityTaskTimedOutEventAttributes(t *apiv1.ActivityTaskTimedOutEventAttributes) *shared.ActivityTaskTimedOutEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ActivityTaskTimedOutEventAttributes{
		Details:            thriftPayload(t.Details),
		ScheduledEventId:   &t.ScheduledEventId,
		StartedEventId:     &t.StartedEventId,
		TimeoutType:        thriftTimeoutType(t.TimeoutType),
		LastFailureReason:  thriftFailureReason(t.LastFailure),
		LastFailureDetails: thriftFailureDetails(t.LastFailure),
	}
}

func thriftCancelTimerFailedEventAttributes(t *apiv1.CancelTimerFailedEventAttributes) *shared.CancelTimerFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.CancelTimerFailedEventAttributes{
		TimerId:                      &t.TimerId,
		Cause:                        &t.Cause,
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		Identity:                     &t.Identity,
	}
}

func thriftChildWorkflowExecutionCanceledEventAttributes(t *apiv1.ChildWorkflowExecutionCanceledEventAttributes) *shared.ChildWorkflowExecutionCanceledEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ChildWorkflowExecutionCanceledEventAttributes{
		Domain:            &t.Domain,
		WorkflowExecution: thriftWorkflowExecution(t.WorkflowExecution),
		WorkflowType:      thriftWorkflowType(t.WorkflowType),
		InitiatedEventId:  &t.InitiatedEventId,
		StartedEventId:    &t.StartedEventId,
		Details:           thriftPayload(t.Details),
	}
}

func thriftChildWorkflowExecutionCompletedEventAttributes(t *apiv1.ChildWorkflowExecutionCompletedEventAttributes) *shared.ChildWorkflowExecutionCompletedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ChildWorkflowExecutionCompletedEventAttributes{
		Domain:            &t.Domain,
		WorkflowExecution: thriftWorkflowExecution(t.WorkflowExecution),
		WorkflowType:      thriftWorkflowType(t.WorkflowType),
		InitiatedEventId:  &t.InitiatedEventId,
		StartedEventId:    &t.StartedEventId,
		Result:            thriftPayload(t.Result),
	}
}

func thriftChildWorkflowExecutionFailedEventAttributes(t *apiv1.ChildWorkflowExecutionFailedEventAttributes) *shared.ChildWorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ChildWorkflowExecutionFailedEventAttributes{
		Domain:            &t.Domain,
		WorkflowExecution: thriftWorkflowExecution(t.WorkflowExecution),
		WorkflowType:      thriftWorkflowType(t.WorkflowType),
		InitiatedEventId:  &t.InitiatedEventId,
		StartedEventId:    &t.StartedEventId,
		Reason:            thriftFailureReason(t.Failure),
		Details:           thriftFailureDetails(t.Failure),
	}
}

func thriftChildWorkflowExecutionStartedEventAttributes(t *apiv1.ChildWorkflowExecutionStartedEventAttributes) *shared.ChildWorkflowExecutionStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ChildWorkflowExecutionStartedEventAttributes{
		Domain:            &t.Domain,
		WorkflowExecution: thriftWorkflowExecution(t.WorkflowExecution),
		WorkflowType:      thriftWorkflowType(t.WorkflowType),
		InitiatedEventId:  &t.InitiatedEventId,
		Header:            thriftHeader(t.Header),
	}
}

func thriftChildWorkflowExecutionTerminatedEventAttributes(t *apiv1.ChildWorkflowExecutionTerminatedEventAttributes) *shared.ChildWorkflowExecutionTerminatedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ChildWorkflowExecutionTerminatedEventAttributes{
		Domain:            &t.Domain,
		WorkflowExecution: thriftWorkflowExecution(t.WorkflowExecution),
		WorkflowType:      thriftWorkflowType(t.WorkflowType),
		InitiatedEventId:  &t.InitiatedEventId,
		StartedEventId:    &t.StartedEventId,
	}
}

func thriftChildWorkflowExecutionTimedOutEventAttributes(t *apiv1.ChildWorkflowExecutionTimedOutEventAttributes) *shared.ChildWorkflowExecutionTimedOutEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ChildWorkflowExecutionTimedOutEventAttributes{
		Domain:            &t.Domain,
		WorkflowExecution: thriftWorkflowExecution(t.WorkflowExecution),
		WorkflowType:      thriftWorkflowType(t.WorkflowType),
		InitiatedEventId:  &t.InitiatedEventId,
		StartedEventId:    &t.StartedEventId,
		TimeoutType:       thriftTimeoutType(t.TimeoutType),
	}
}

func thriftDecisionTaskFailedEventAttributes(t *apiv1.DecisionTaskFailedEventAttributes) *shared.DecisionTaskFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.DecisionTaskFailedEventAttributes{
		ScheduledEventId: &t.ScheduledEventId,
		StartedEventId:   &t.StartedEventId,
		Cause:            thriftDecisionTaskFailedCause(t.Cause),
		Reason:           thriftFailureReason(t.Failure),
		Details:          thriftFailureDetails(t.Failure),
		Identity:         &t.Identity,
		BaseRunId:        &t.BaseRunId,
		NewRunId:         &t.NewRunId,
		ForkEventVersion: &t.ForkEventVersion,
		BinaryChecksum:   &t.BinaryChecksum,
	}
}

func thriftDecisionTaskScheduledEventAttributes(t *apiv1.DecisionTaskScheduledEventAttributes) *shared.DecisionTaskScheduledEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.DecisionTaskScheduledEventAttributes{
		TaskList:                   thriftTaskList(t.TaskList),
		StartToCloseTimeoutSeconds: durationToSeconds(t.StartToCloseTimeout),
		Attempt:                    common.Int64Ptr(int64(t.Attempt)),
	}
}

func thriftDecisionTaskStartedEventAttributes(t *apiv1.DecisionTaskStartedEventAttributes) *shared.DecisionTaskStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.DecisionTaskStartedEventAttributes{
		ScheduledEventId: &t.ScheduledEventId,
		Identity:         &t.Identity,
		RequestId:        &t.RequestId,
	}
}

func thriftDecisionTaskCompletedEventAttributes(t *apiv1.DecisionTaskCompletedEventAttributes) *shared.DecisionTaskCompletedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.DecisionTaskCompletedEventAttributes{
		ScheduledEventId: &t.ScheduledEventId,
		StartedEventId:   &t.StartedEventId,
		Identity:         &t.Identity,
		BinaryChecksum:   &t.BinaryChecksum,
		ExecutionContext: t.ExecutionContext,
	}
}

func thriftDecisionTaskTimedOutEventAttributes(t *apiv1.DecisionTaskTimedOutEventAttributes) *shared.DecisionTaskTimedOutEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.DecisionTaskTimedOutEventAttributes{
		ScheduledEventId: &t.ScheduledEventId,
		StartedEventId:   &t.StartedEventId,
		TimeoutType:      thriftTimeoutType(t.TimeoutType),
		BaseRunId:        &t.BaseRunId,
		NewRunId:         &t.NewRunId,
		ForkEventVersion: &t.ForkEventVersion,
		Reason:           &t.Reason,
		Cause:            thriftDecisionTaskTimedOutCause(t.Cause),
	}
}

func thriftExternalWorkflowExecutionCancelRequestedEventAttributes(t *apiv1.ExternalWorkflowExecutionCancelRequestedEventAttributes) *shared.ExternalWorkflowExecutionCancelRequestedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ExternalWorkflowExecutionCancelRequestedEventAttributes{
		InitiatedEventId:  &t.InitiatedEventId,
		Domain:            &t.Domain,
		WorkflowExecution: thriftWorkflowExecution(t.WorkflowExecution),
	}
}

func thriftExternalWorkflowExecutionSignaledEventAttributes(t *apiv1.ExternalWorkflowExecutionSignaledEventAttributes) *shared.ExternalWorkflowExecutionSignaledEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.ExternalWorkflowExecutionSignaledEventAttributes{
		InitiatedEventId:  &t.InitiatedEventId,
		Domain:            &t.Domain,
		WorkflowExecution: thriftWorkflowExecution(t.WorkflowExecution),
		Control:           t.Control,
	}
}

func thriftMarkerRecordedEventAttributes(t *apiv1.MarkerRecordedEventAttributes) *shared.MarkerRecordedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.MarkerRecordedEventAttributes{
		MarkerName:                   &t.MarkerName,
		Details:                      thriftPayload(t.Details),
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		Header:                       thriftHeader(t.Header),
	}
}

func thriftRequestCancelActivityTaskFailedEventAttributes(t *apiv1.RequestCancelActivityTaskFailedEventAttributes) *shared.RequestCancelActivityTaskFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.RequestCancelActivityTaskFailedEventAttributes{
		ActivityId:                   &t.ActivityId,
		Cause:                        &t.Cause,
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
	}
}

func thriftRequestCancelExternalWorkflowExecutionFailedEventAttributes(t *apiv1.RequestCancelExternalWorkflowExecutionFailedEventAttributes) *shared.RequestCancelExternalWorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.RequestCancelExternalWorkflowExecutionFailedEventAttributes{
		Cause:                        thriftCancelExternalWorkflowExecutionFailedCause(t.Cause),
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		Domain:                       &t.Domain,
		WorkflowExecution:            thriftWorkflowExecution(t.WorkflowExecution),
		InitiatedEventId:             &t.InitiatedEventId,
		Control:                      t.Control,
	}
}

func thriftRequestCancelExternalWorkflowExecutionInitiatedEventAttributes(t *apiv1.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes) *shared.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.RequestCancelExternalWorkflowExecutionInitiatedEventAttributes{
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		Domain:                       &t.Domain,
		WorkflowExecution:            thriftWorkflowExecution(t.WorkflowExecution),
		Control:                      t.Control,
		ChildWorkflowOnly:            &t.ChildWorkflowOnly,
	}
}

func thriftSignalExternalWorkflowExecutionFailedEventAttributes(t *apiv1.SignalExternalWorkflowExecutionFailedEventAttributes) *shared.SignalExternalWorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.SignalExternalWorkflowExecutionFailedEventAttributes{
		Cause:                        thriftSignalExternalWorkflowExecutionFailedCause(t.Cause),
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		Domain:                       &t.Domain,
		WorkflowExecution:            thriftWorkflowExecution(t.WorkflowExecution),
		InitiatedEventId:             &t.InitiatedEventId,
		Control:                      t.Control,
	}
}

func thriftSignalExternalWorkflowExecutionInitiatedEventAttributes(t *apiv1.SignalExternalWorkflowExecutionInitiatedEventAttributes) *shared.SignalExternalWorkflowExecutionInitiatedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.SignalExternalWorkflowExecutionInitiatedEventAttributes{
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		Domain:                       &t.Domain,
		WorkflowExecution:            thriftWorkflowExecution(t.WorkflowExecution),
		SignalName:                   &t.SignalName,
		Input:                        thriftPayload(t.Input),
		Control:                      t.Control,
		ChildWorkflowOnly:            &t.ChildWorkflowOnly,
	}
}

func thriftStartChildWorkflowExecutionFailedEventAttributes(t *apiv1.StartChildWorkflowExecutionFailedEventAttributes) *shared.StartChildWorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.StartChildWorkflowExecutionFailedEventAttributes{
		Domain:                       &t.Domain,
		WorkflowId:                   &t.WorkflowId,
		WorkflowType:                 thriftWorkflowType(t.WorkflowType),
		Cause:                        thriftChildWorkflowExecutionFailedCause(t.Cause),
		Control:                      t.Control,
		InitiatedEventId:             &t.InitiatedEventId,
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
	}
}

func thriftStartChildWorkflowExecutionInitiatedEventAttributes(t *apiv1.StartChildWorkflowExecutionInitiatedEventAttributes) *shared.StartChildWorkflowExecutionInitiatedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.StartChildWorkflowExecutionInitiatedEventAttributes{
		Domain:                              &t.Domain,
		WorkflowId:                          &t.WorkflowId,
		WorkflowType:                        thriftWorkflowType(t.WorkflowType),
		TaskList:                            thriftTaskList(t.TaskList),
		Input:                               thriftPayload(t.Input),
		ExecutionStartToCloseTimeoutSeconds: durationToSeconds(t.ExecutionStartToCloseTimeout),
		TaskStartToCloseTimeoutSeconds:      durationToSeconds(t.TaskStartToCloseTimeout),
		ParentClosePolicy:                   thriftParentClosePolicy(t.ParentClosePolicy),
		Control:                             t.Control,
		DecisionTaskCompletedEventId:        &t.DecisionTaskCompletedEventId,
		WorkflowIdReusePolicy:               thriftWorkflowIdReusePolicy(t.WorkflowIdReusePolicy),
		RetryPolicy:                         thriftRetryPolicy(t.RetryPolicy),
		CronSchedule:                        &t.CronSchedule,
		Header:                              thriftHeader(t.Header),
		Memo:                                thriftMemo(t.Memo),
		SearchAttributes:                    thriftSearchAttributes(t.SearchAttributes),
		DelayStartSeconds:                   durationToSeconds(t.DelayStart),
	}
}

func thriftTimerCanceledEventAttributes(t *apiv1.TimerCanceledEventAttributes) *shared.TimerCanceledEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.TimerCanceledEventAttributes{
		TimerId:                      &t.TimerId,
		StartedEventId:               &t.StartedEventId,
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		Identity:                     &t.Identity,
	}
}

func thriftTimerFiredEventAttributes(t *apiv1.TimerFiredEventAttributes) *shared.TimerFiredEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.TimerFiredEventAttributes{
		TimerId:        &t.TimerId,
		StartedEventId: &t.StartedEventId,
	}
}

func thriftTimerStartedEventAttributes(t *apiv1.TimerStartedEventAttributes) *shared.TimerStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.TimerStartedEventAttributes{
		TimerId:                      &t.TimerId,
		StartToFireTimeoutSeconds:    int32To64(durationToSeconds(t.StartToFireTimeout)),
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
	}
}

func thriftUpsertWorkflowSearchAttributesEventAttributes(t *apiv1.UpsertWorkflowSearchAttributesEventAttributes) *shared.UpsertWorkflowSearchAttributesEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.UpsertWorkflowSearchAttributesEventAttributes{
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		SearchAttributes:             thriftSearchAttributes(t.SearchAttributes),
	}
}

func thriftWorkflowExecutionCancelRequestedEventAttributes(t *apiv1.WorkflowExecutionCancelRequestedEventAttributes) *shared.WorkflowExecutionCancelRequestedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionCancelRequestedEventAttributes{
		Cause:                     &t.Cause,
		ExternalInitiatedEventId:  thriftExternalInitiatedId(t.ExternalExecutionInfo),
		ExternalWorkflowExecution: thriftExternalWorkflowExecution(t.ExternalExecutionInfo),
		Identity:                  &t.Identity,
	}
}

func thriftWorkflowExecutionCanceledEventAttributes(t *apiv1.WorkflowExecutionCanceledEventAttributes) *shared.WorkflowExecutionCanceledEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionCanceledEventAttributes{
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
		Details:                      thriftPayload(t.Details),
	}
}

func thriftWorkflowExecutionCompletedEventAttributes(t *apiv1.WorkflowExecutionCompletedEventAttributes) *shared.WorkflowExecutionCompletedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionCompletedEventAttributes{
		Result:                       thriftPayload(t.Result),
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
	}
}

func thriftWorkflowExecutionContinuedAsNewEventAttributes(t *apiv1.WorkflowExecutionContinuedAsNewEventAttributes) *shared.WorkflowExecutionContinuedAsNewEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionContinuedAsNewEventAttributes{
		NewExecutionRunId:                   &t.NewExecutionRunId,
		WorkflowType:                        thriftWorkflowType(t.WorkflowType),
		TaskList:                            thriftTaskList(t.TaskList),
		Input:                               thriftPayload(t.Input),
		ExecutionStartToCloseTimeoutSeconds: durationToSeconds(t.ExecutionStartToCloseTimeout),
		TaskStartToCloseTimeoutSeconds:      durationToSeconds(t.TaskStartToCloseTimeout),
		DecisionTaskCompletedEventId:        &t.DecisionTaskCompletedEventId,
		BackoffStartIntervalInSeconds:       durationToSeconds(t.BackoffStartInterval),
		Initiator:                           thriftContinueAsNewInitiator(t.Initiator),
		FailureReason:                       thriftFailureReason(t.Failure),
		FailureDetails:                      thriftFailureDetails(t.Failure),
		LastCompletionResult:                thriftPayload(t.LastCompletionResult),
		Header:                              thriftHeader(t.Header),
		Memo:                                thriftMemo(t.Memo),
		SearchAttributes:                    thriftSearchAttributes(t.SearchAttributes),
	}
}

func thriftWorkflowExecutionFailedEventAttributes(t *apiv1.WorkflowExecutionFailedEventAttributes) *shared.WorkflowExecutionFailedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionFailedEventAttributes{
		Reason:                       thriftFailureReason(t.Failure),
		Details:                      thriftFailureDetails(t.Failure),
		DecisionTaskCompletedEventId: &t.DecisionTaskCompletedEventId,
	}
}

func thriftWorkflowExecutionSignaledEventAttributes(t *apiv1.WorkflowExecutionSignaledEventAttributes) *shared.WorkflowExecutionSignaledEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionSignaledEventAttributes{
		SignalName: &t.SignalName,
		Input:      thriftPayload(t.Input),
		Identity:   &t.Identity,
	}
}

func thriftWorkflowExecutionStartedEventAttributes(t *apiv1.WorkflowExecutionStartedEventAttributes) *shared.WorkflowExecutionStartedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionStartedEventAttributes{
		WorkflowType:                        thriftWorkflowType(t.WorkflowType),
		ParentWorkflowDomain:                thriftParentDomainName(t.ParentExecutionInfo),
		ParentWorkflowExecution:             thriftParentWorkflowExecution(t.ParentExecutionInfo),
		ParentInitiatedEventId:              thriftParentInitiatedId(t.ParentExecutionInfo),
		TaskList:                            thriftTaskList(t.TaskList),
		Input:                               thriftPayload(t.Input),
		ExecutionStartToCloseTimeoutSeconds: durationToSeconds(t.ExecutionStartToCloseTimeout),
		TaskStartToCloseTimeoutSeconds:      durationToSeconds(t.TaskStartToCloseTimeout),
		ContinuedExecutionRunId:             &t.ContinuedExecutionRunId,
		Initiator:                           thriftContinueAsNewInitiator(t.Initiator),
		ContinuedFailureReason:              thriftFailureReason(t.ContinuedFailure),
		ContinuedFailureDetails:             thriftFailureDetails(t.ContinuedFailure),
		LastCompletionResult:                thriftPayload(t.LastCompletionResult),
		OriginalExecutionRunId:              &t.OriginalExecutionRunId,
		Identity:                            &t.Identity,
		FirstExecutionRunId:                 &t.FirstExecutionRunId,
		RetryPolicy:                         thriftRetryPolicy(t.RetryPolicy),
		Attempt:                             &t.Attempt,
		ExpirationTimestamp:                 timeToUnixNano(t.ExpirationTime),
		CronSchedule:                        &t.CronSchedule,
		FirstDecisionTaskBackoffSeconds:     durationToSeconds(t.FirstDecisionTaskBackoff),
		Memo:                                thriftMemo(t.Memo),
		SearchAttributes:                    thriftSearchAttributes(t.SearchAttributes),
		PrevAutoResetPoints:                 thriftResetPoints(t.PrevAutoResetPoints),
		Header:                              thriftHeader(t.Header),
	}
}

func thriftWorkflowExecutionTerminatedEventAttributes(t *apiv1.WorkflowExecutionTerminatedEventAttributes) *shared.WorkflowExecutionTerminatedEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionTerminatedEventAttributes{
		Reason:   &t.Reason,
		Details:  thriftPayload(t.Details),
		Identity: &t.Identity,
	}
}

func thriftWorkflowExecutionTimedOutEventAttributes(t *apiv1.WorkflowExecutionTimedOutEventAttributes) *shared.WorkflowExecutionTimedOutEventAttributes {
	if t == nil {
		return nil
	}
	return &shared.WorkflowExecutionTimedOutEventAttributes{
		TimeoutType: thriftTimeoutType(t.TimeoutType),
	}
}

func thriftHistoryEventArray(t []*apiv1.HistoryEvent) []*shared.HistoryEvent {
	if t == nil {
		return nil
	}
	v := make([]*shared.HistoryEvent, len(t))
	for i := range t {
		v[i] = thriftHistoryEvent(t[i])
	}
	return v
}

func thriftPayloadMap(t map[string]*apiv1.Payload) map[string][]byte {
	if t == nil {
		return nil
	}
	v := make(map[string][]byte, len(t))
	for key := range t {
		v[key] = thriftPayload(t[key])
	}
	return v
}

func thriftBadBinaryInfoMap(t map[string]*apiv1.BadBinaryInfo) map[string]*shared.BadBinaryInfo {
	if t == nil {
		return nil
	}
	v := make(map[string]*shared.BadBinaryInfo, len(t))
	for key := range t {
		v[key] = thriftBadBinaryInfo(t[key])
	}
	return v
}

func thriftClusterReplicationConfigurationArray(t []*apiv1.ClusterReplicationConfiguration) []*shared.ClusterReplicationConfiguration {
	if t == nil {
		return nil
	}
	v := make([]*shared.ClusterReplicationConfiguration, len(t))
	for i := range t {
		v[i] = thriftClusterReplicationConfiguration(t[i])
	}
	return v
}

func thriftPollerInfoArray(t []*apiv1.PollerInfo) []*shared.PollerInfo {
	if t == nil {
		return nil
	}
	v := make([]*shared.PollerInfo, len(t))
	for i := range t {
		v[i] = thriftPollerInfo(t[i])
	}
	return v
}

func thriftResetPointInfoArray(t []*apiv1.ResetPointInfo) []*shared.ResetPointInfo {
	if t == nil {
		return nil
	}
	v := make([]*shared.ResetPointInfo, len(t))
	for i := range t {
		v[i] = thriftResetPointInfo(t[i])
	}
	return v
}

func thriftPendingActivityInfoArray(t []*apiv1.PendingActivityInfo) []*shared.PendingActivityInfo {
	if t == nil {
		return nil
	}
	v := make([]*shared.PendingActivityInfo, len(t))
	for i := range t {
		v[i] = thriftPendingActivityInfo(t[i])
	}
	return v
}

func thriftPendingChildExecutionInfoArray(t []*apiv1.PendingChildExecutionInfo) []*shared.PendingChildExecutionInfo {
	if t == nil {
		return nil
	}
	v := make([]*shared.PendingChildExecutionInfo, len(t))
	for i := range t {
		v[i] = thriftPendingChildExecutionInfo(t[i])
	}
	return v
}

func thriftIndexedValueTypeMap(t map[string]apiv1.IndexedValueType) map[string]shared.IndexedValueType {
	if t == nil {
		return nil
	}
	v := make(map[string]shared.IndexedValueType, len(t))
	for key := range t {
		v[key] = thriftIndexedValueType(t[key])
	}
	return v
}

func thriftDataBlobArray(t []*apiv1.DataBlob) []*shared.DataBlob {
	if t == nil {
		return nil
	}
	v := make([]*shared.DataBlob, len(t))
	for i := range t {
		v[i] = thriftDataBlob(t[i])
	}
	return v
}

func thriftWorkflowExecutionInfoArray(t []*apiv1.WorkflowExecutionInfo) []*shared.WorkflowExecutionInfo {
	if t == nil {
		return nil
	}
	v := make([]*shared.WorkflowExecutionInfo, len(t))
	for i := range t {
		v[i] = thriftWorkflowExecutionInfo(t[i])
	}
	return v
}

func thriftDescribeDomainResponseArray(t []*apiv1.Domain) []*shared.DescribeDomainResponse {
	if t == nil {
		return nil
	}
	v := make([]*shared.DescribeDomainResponse, len(t))
	for i := range t {
		v[i] = thriftDescribeDomainResponseDomain(t[i])
	}
	return v
}

func thriftTaskListPartitionMetadataArray(t []*apiv1.TaskListPartitionMetadata) []*shared.TaskListPartitionMetadata {
	if t == nil {
		return nil
	}
	v := make([]*shared.TaskListPartitionMetadata, len(t))
	for i := range t {
		v[i] = thriftTaskListPartitionMetadata(t[i])
	}
	return v
}

func thriftWorkflowQueryMap(t map[string]*apiv1.WorkflowQuery) map[string]*shared.WorkflowQuery {
	if t == nil {
		return nil
	}
	v := make(map[string]*shared.WorkflowQuery, len(t))
	for key := range t {
		v[key] = thriftWorkflowQuery(t[key])
	}
	return v
}

func thriftActivityLocalDispatchInfoMap(t map[string]*apiv1.ActivityLocalDispatchInfo) map[string]*shared.ActivityLocalDispatchInfo {
	if t == nil {
		return nil
	}
	v := make(map[string]*shared.ActivityLocalDispatchInfo, len(t))
	for key := range t {
		v[key] = thriftActivityLocalDispatchInfo(t[key])
	}
	return v
}

func thriftDomainStatus(t apiv1.DomainStatus) *shared.DomainStatus {
	switch t {
	case apiv1.DomainStatus_DOMAIN_STATUS_INVALID:
		return nil
	case apiv1.DomainStatus_DOMAIN_STATUS_REGISTERED:
		return shared.DomainStatusRegistered.Ptr()
	case apiv1.DomainStatus_DOMAIN_STATUS_DEPRECATED:
		return shared.DomainStatusDeprecated.Ptr()
	case apiv1.DomainStatus_DOMAIN_STATUS_DELETED:
		return shared.DomainStatusDeleted.Ptr()
	}
	panic("unexpected enum value")
}

func thriftArchivalStatus(t apiv1.ArchivalStatus) *shared.ArchivalStatus {
	switch t {
	case apiv1.ArchivalStatus_ARCHIVAL_STATUS_INVALID:
		return nil
	case apiv1.ArchivalStatus_ARCHIVAL_STATUS_DISABLED:
		return shared.ArchivalStatusDisabled.Ptr()
	case apiv1.ArchivalStatus_ARCHIVAL_STATUS_ENABLED:
		return shared.ArchivalStatusEnabled.Ptr()
	}
	panic("unexpected enum value")
}

func thriftTaskListKind(t apiv1.TaskListKind) *shared.TaskListKind {
	switch t {
	case apiv1.TaskListKind_TASK_LIST_KIND_INVALID:
		return nil
	case apiv1.TaskListKind_TASK_LIST_KIND_NORMAL:
		return shared.TaskListKindNormal.Ptr()
	case apiv1.TaskListKind_TASK_LIST_KIND_STICKY:
		return shared.TaskListKindSticky.Ptr()
	}
	panic("unexpected enum value")
}

func thriftWorkflowExecutionCloseStatus(t apiv1.WorkflowExecutionCloseStatus) *shared.WorkflowExecutionCloseStatus {
	switch t {
	case apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_INVALID:
		return nil
	case apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_COMPLETED:
		return shared.WorkflowExecutionCloseStatusCompleted.Ptr()
	case apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_FAILED:
		return shared.WorkflowExecutionCloseStatusFailed.Ptr()
	case apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CANCELED:
		return shared.WorkflowExecutionCloseStatusCanceled.Ptr()
	case apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TERMINATED:
		return shared.WorkflowExecutionCloseStatusTerminated.Ptr()
	case apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_CONTINUED_AS_NEW:
		return shared.WorkflowExecutionCloseStatusContinuedAsNew.Ptr()
	case apiv1.WorkflowExecutionCloseStatus_WORKFLOW_EXECUTION_CLOSE_STATUS_TIMED_OUT:
		return shared.WorkflowExecutionCloseStatusTimedOut.Ptr()
	}
	panic("unexpected enum value")
}

func thriftPendingActivityState(t apiv1.PendingActivityState) *shared.PendingActivityState {
	switch t {
	case apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_INVALID:
		return nil
	case apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_SCHEDULED:
		return shared.PendingActivityStateScheduled.Ptr()
	case apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_STARTED:
		return shared.PendingActivityStateStarted.Ptr()
	case apiv1.PendingActivityState_PENDING_ACTIVITY_STATE_CANCEL_REQUESTED:
		return shared.PendingActivityStateCancelRequested.Ptr()
	}
	panic("unexpected enum value")
}

func thriftParentClosePolicy(t apiv1.ParentClosePolicy) *shared.ParentClosePolicy {
	switch t {
	case apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_INVALID:
		return nil
	case apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_ABANDON:
		return shared.ParentClosePolicyAbandon.Ptr()
	case apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_REQUEST_CANCEL:
		return shared.ParentClosePolicyRequestCancel.Ptr()
	case apiv1.ParentClosePolicy_PARENT_CLOSE_POLICY_TERMINATE:
		return shared.ParentClosePolicyTerminate.Ptr()
	}
	panic("unexpected enum value")
}

func thriftPendingDecisionState(t apiv1.PendingDecisionState) *shared.PendingDecisionState {
	switch t {
	case apiv1.PendingDecisionState_PENDING_DECISION_STATE_INVALID:
		return nil
	case apiv1.PendingDecisionState_PENDING_DECISION_STATE_SCHEDULED:
		return shared.PendingDecisionStateScheduled.Ptr()
	case apiv1.PendingDecisionState_PENDING_DECISION_STATE_STARTED:
		return shared.PendingDecisionStateStarted.Ptr()
	}
	panic("unexpected enum value")
}

func thriftIndexedValueType(t apiv1.IndexedValueType) shared.IndexedValueType {
	switch t {
	case apiv1.IndexedValueType_INDEXED_VALUE_TYPE_INVALID:
		panic("received IndexedValueType_INDEXED_VALUE_TYPE_INVALID")
	case apiv1.IndexedValueType_INDEXED_VALUE_TYPE_STRING:
		return shared.IndexedValueTypeString
	case apiv1.IndexedValueType_INDEXED_VALUE_TYPE_KEYWORD:
		return shared.IndexedValueTypeKeyword
	case apiv1.IndexedValueType_INDEXED_VALUE_TYPE_INT:
		return shared.IndexedValueTypeInt
	case apiv1.IndexedValueType_INDEXED_VALUE_TYPE_DOUBLE:
		return shared.IndexedValueTypeDouble
	case apiv1.IndexedValueType_INDEXED_VALUE_TYPE_BOOL:
		return shared.IndexedValueTypeBool
	case apiv1.IndexedValueType_INDEXED_VALUE_TYPE_DATETIME:
		return shared.IndexedValueTypeDatetime
	}
	panic("unexpected enum value")
}

func thriftEncodingType(t apiv1.EncodingType) *shared.EncodingType {
	switch t {
	case apiv1.EncodingType_ENCODING_TYPE_INVALID:
		return nil
	case apiv1.EncodingType_ENCODING_TYPE_THRIFTRW:
		return shared.EncodingTypeThriftRW.Ptr()
	case apiv1.EncodingType_ENCODING_TYPE_JSON:
		return shared.EncodingTypeJSON.Ptr()
	case apiv1.EncodingType_ENCODING_TYPE_PROTO3:
		panic("not supported yet")
	}
	panic("unexpected enum value")
}

func thriftTimeoutType(t apiv1.TimeoutType) *shared.TimeoutType {
	switch t {
	case apiv1.TimeoutType_TIMEOUT_TYPE_INVALID:
		return nil
	case apiv1.TimeoutType_TIMEOUT_TYPE_START_TO_CLOSE:
		return shared.TimeoutTypeStartToClose.Ptr()
	case apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_START:
		return shared.TimeoutTypeScheduleToStart.Ptr()
	case apiv1.TimeoutType_TIMEOUT_TYPE_SCHEDULE_TO_CLOSE:
		return shared.TimeoutTypeScheduleToClose.Ptr()
	case apiv1.TimeoutType_TIMEOUT_TYPE_HEARTBEAT:
		return shared.TimeoutTypeHeartbeat.Ptr()
	}
	panic("unexpected enum value")
}


func thriftDecisionTaskFailedCause(t apiv1.DecisionTaskFailedCause) *shared.DecisionTaskFailedCause {
	switch t {
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_INVALID:
		return nil
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_UNHANDLED_DECISION:
		return shared.DecisionTaskFailedCauseUnhandledDecision.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SCHEDULE_ACTIVITY_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadScheduleActivityAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_REQUEST_CANCEL_ACTIVITY_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadRequestCancelActivityAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_START_TIMER_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadStartTimerAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CANCEL_TIMER_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadCancelTimerAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_RECORD_MARKER_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadRecordMarkerAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_COMPLETE_WORKFLOW_EXECUTION_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadCompleteWorkflowExecutionAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_FAIL_WORKFLOW_EXECUTION_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadFailWorkflowExecutionAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CANCEL_WORKFLOW_EXECUTION_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadCancelWorkflowExecutionAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_REQUEST_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadRequestCancelExternalWorkflowExecutionAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_CONTINUE_AS_NEW_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadContinueAsNewAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_START_TIMER_DUPLICATE_ID:
		return shared.DecisionTaskFailedCauseStartTimerDuplicateID.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_RESET_STICKY_TASK_LIST:
		return shared.DecisionTaskFailedCauseResetStickyTasklist.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_WORKFLOW_WORKER_UNHANDLED_FAILURE:
		return shared.DecisionTaskFailedCauseWorkflowWorkerUnhandledFailure.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SIGNAL_WORKFLOW_EXECUTION_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadSignalWorkflowExecutionAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_START_CHILD_EXECUTION_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadStartChildExecutionAttributes.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_FORCE_CLOSE_DECISION:
		return shared.DecisionTaskFailedCauseForceCloseDecision.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_FAILOVER_CLOSE_DECISION:
		return shared.DecisionTaskFailedCauseFailoverCloseDecision.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SIGNAL_INPUT_SIZE:
		return shared.DecisionTaskFailedCauseBadSignalInputSize.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_RESET_WORKFLOW:
		return shared.DecisionTaskFailedCauseResetWorkflow.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_BINARY:
		return shared.DecisionTaskFailedCauseBadBinary.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_SCHEDULE_ACTIVITY_DUPLICATE_ID:
		return shared.DecisionTaskFailedCauseScheduleActivityDuplicateID.Ptr()
	case apiv1.DecisionTaskFailedCause_DECISION_TASK_FAILED_CAUSE_BAD_SEARCH_ATTRIBUTES:
		return shared.DecisionTaskFailedCauseBadSearchAttributes.Ptr()
	}
	panic("unexpected enum value")
}

func thriftDecisionTaskTimedOutCause(t apiv1.DecisionTaskTimedOutCause) *shared.DecisionTaskTimedOutCause {
	switch t {
	case apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_INVALID:
		return nil
	case apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_TIMEOUT:
		return shared.DecisionTaskTimedOutCauseTimeout.Ptr()
	case apiv1.DecisionTaskTimedOutCause_DECISION_TASK_TIMED_OUT_CAUSE_RESET:
		return shared.DecisionTaskTimedOutCauseReset.Ptr()
	}
	panic("unexpected enum value")
}

func thriftCancelExternalWorkflowExecutionFailedCause(t apiv1.CancelExternalWorkflowExecutionFailedCause) *shared.CancelExternalWorkflowExecutionFailedCause {
	switch t {
	case apiv1.CancelExternalWorkflowExecutionFailedCause_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID:
		return nil
	case apiv1.CancelExternalWorkflowExecutionFailedCause_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_UNKNOWN_EXTERNAL_WORKFLOW_EXECUTION:
		return shared.CancelExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution.Ptr()
	}
	panic("unexpected enum value")
}

func thriftSignalExternalWorkflowExecutionFailedCause(t apiv1.SignalExternalWorkflowExecutionFailedCause) *shared.SignalExternalWorkflowExecutionFailedCause {
	switch t {
	case apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID:
		return nil
	case apiv1.SignalExternalWorkflowExecutionFailedCause_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED_CAUSE_UNKNOWN_EXTERNAL_WORKFLOW_EXECUTION:
		return shared.SignalExternalWorkflowExecutionFailedCauseUnknownExternalWorkflowExecution.Ptr()
	}
	panic("unexpected enum value")
}

func thriftChildWorkflowExecutionFailedCause(t apiv1.ChildWorkflowExecutionFailedCause) *shared.ChildWorkflowExecutionFailedCause {
	switch t {
	case apiv1.ChildWorkflowExecutionFailedCause_CHILD_WORKFLOW_EXECUTION_FAILED_CAUSE_INVALID:
		return nil
	case apiv1.ChildWorkflowExecutionFailedCause_CHILD_WORKFLOW_EXECUTION_FAILED_CAUSE_WORKFLOW_ALREADY_RUNNING:
		return shared.ChildWorkflowExecutionFailedCauseWorkflowAlreadyRunning.Ptr()
	}
	panic("unexpected enum value")
}

func thriftWorkflowIdReusePolicy(t apiv1.WorkflowIdReusePolicy) *shared.WorkflowIdReusePolicy {
	switch t {
	case apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_INVALID:
		return nil
	case apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY:
		return shared.WorkflowIdReusePolicyAllowDuplicateFailedOnly.Ptr()
	case apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE:
		return shared.WorkflowIdReusePolicyAllowDuplicate.Ptr()
	case apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE:
		return shared.WorkflowIdReusePolicyRejectDuplicate.Ptr()
	case apiv1.WorkflowIdReusePolicy_WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING:
		return shared.WorkflowIdReusePolicyTerminateIfRunning.Ptr()
	}
	panic("unexpected enum value")
}

func thriftContinueAsNewInitiator(t apiv1.ContinueAsNewInitiator) *shared.ContinueAsNewInitiator {
	switch t {
	case apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_INVALID:
		return nil
	case apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_DECIDER:
		return shared.ContinueAsNewInitiatorDecider.Ptr()
	case apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_RETRY_POLICY:
		return shared.ContinueAsNewInitiatorRetryPolicy.Ptr()
	case apiv1.ContinueAsNewInitiator_CONTINUE_AS_NEW_INITIATOR_CRON_SCHEDULE:
		return shared.ContinueAsNewInitiatorCronSchedule.Ptr()
	}
	panic("unexpected enum value")
}