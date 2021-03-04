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

// Code generated by thriftrw v1.11.0. DO NOT EDIT.
// @generated

package shadower

import (
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/thriftrw/thriftreflect"
)

// ThriftModule represents the IDL file used to generate this package.
var ThriftModule = &thriftreflect.ThriftModule{
	Name:     "shadower",
	Package:  "go.uber.org/cadence/.gen/go/shadower",
	FilePath: "shadower.thrift",
	SHA1:     "ed3629d23f7839fcad723a1e0463a86900a2cd55",
	Includes: []*thriftreflect.ThriftModule{
		shared.ThriftModule,
	},
	Raw: rawIDL,
}

const rawIDL = "// Copyright (c) 2017-2021 Uber Technologies, Inc.\n//\n// Permission is hereby granted, free of charge, to any person obtaining a copy\n// of this software and associated documentation files (the \"Software\"), to deal\n// in the Software without restriction, including without limitation the rights\n// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell\n// copies of the Software, and to permit persons to whom the Software is\n// furnished to do so, subject to the following conditions:\n//\n// The above copyright notice and this permission notice shall be included in\n// all copies or substantial portions of the Software.\n//\n// THE SOFTWARE IS PROVIDED \"AS IS\", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR\n// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,\n// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE\n// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER\n// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,\n// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN\n// THE SOFTWARE.\n\nnamespace java com.uber.cadence.shadower\n\ninclude \"shared.thrift\"\n\nconst string ShadowerLocalDomainName = \"cadence-shadower\"\nconst string ShadowerTaskList = \"cadence-shadower-tl\"\n\nconst string ShadowWorkflowName = \"cadence-shadow-workflow\"\n\nconst string ScanWorkflowActivityName = \"scanWorkflowActivity\"\nconst string ReplayWorkflowActivityName = \"replayWorkflowActivity\"\n\nconst string ScanWorkflowIDSuffix = \"-shadow-workflow\"\n\nconst string ErrReasonDomainNotExists = \"domain not exists\"\nconst string ErrReasonInvalidQuery = \"invalid visibility query\"\nconst string ErrReasonWorkflowTypeNotRegistered = \"workflow type not registered\"\n\nenum ShadowMode {\n  Normal,\n  Continuous,\n}\n\nstruct ExitCondition {\n  10: optional i32 expirationIntervalInSeconds\n  20: optional i32 shadowCount\n}\n\nstruct ShadowWorkflowParams {\n  10: optional string domain \n  20: optional string taskList\n  30: optional string workflowQuery\n  40: optional binary nextPageToken\n  50: optional double samplingRate\n  60: optional ShadowMode shadowMode\n  70: optional ExitCondition exitCondition\n  80: optional i32 concurrency\n  90: optional ShadowWorkflowResult lastRunResult\n}\n\nstruct ShadowWorkflowResult {\n  10: optional i32 succeeded\n  20: optional i32 skipped\n  30: optional i32 failed\n}\n\nstruct ScanWorkflowActivityParams {\n  10: optional string domain\n  20: optional string workflowQuery\n  30: optional binary nextPageToken\n  40: optional i32 pageSize\n  50: optional double samplingRate\n}\n\nstruct ScanWorkflowActivityResult {\n  10: optional list<shared.WorkflowExecution> executions\n  20: optional binary nextPageToken\n}\n\nstruct ReplayWorkflowActivityParams {\n  10: optional string domain\n  20: optional list<shared.WorkflowExecution> executions\n}\n\nstruct ReplayWorkflowActivityResult {\n  10: optional i32 succeeded\n  20: optional i32 skipped\n  30: optional i32 failed\n}\n"
