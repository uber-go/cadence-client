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

import "go.uber.org/cadence/internal"

type (
	// HeaderReader is an interface to read information from cadence headers
	HeaderReader = internal.HeaderReader

	// HeaderWriter is an interface to write information to cadence headers
	HeaderWriter = internal.HeaderWriter

	// ContextPropagator determines what information from context to pass along.
	//
	// The information passed is called Headers - a sequence of string to []byte
	// tuples of serialized data that should follow workflow and activity execution
	// around.
	//
	// Inject* methods are used on the way from the process to persistence in
	// Cadence - thus they use HeaderWriter-s to write the metadata. Extract*
	// methods are used on the way from persisted state in Cadence to execution
	// - thus they use HeaderReader-s to read the metadata and fill it in the
	// returned context. Returning error from Extract* methods prevents the
	// successful workflow run.
	//
	// The whole sequence of execution is:
	//
	// Process initiating the workflow -> Inject -> Cadence -> Go Workflow Worker
	// -> ExtractToWorkflow -> Start executing a workflow -> InjectFromWorkflow ->
	// Cadence -> Go Activity Worker -> Extract -> Execute Activity
	ContextPropagator = internal.ContextPropagator
)
