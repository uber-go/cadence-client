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

package cadencefx

import (
	"go.uber.org/fx"
)

// Module includes everything in ClientModule
// Additionally, it provides constructor for:
// 1. Cadence workers
//
// On start, it will spin up cadence workers based
// on the specification in config/[base,development,prod].yaml
//
// Use Module if the FX app is a cadence worker.
//
// Sample yaml for configuring a worker:
// cadencde:
//   workers:
//	 - domain: domain1
//	   task_list: tasklist1
//     options:
//        disableworkflowworker : true
//   - domain: domain2
//     task_list: tasklist2
//
var Module = fx.Options(
	ClientModule,
	fx.Provide(NewWorkers),
	fx.Invoke(RegisterWorkers),
)

// ClientModule provides constructors for:
// 1. Cadence thrift client
// 2. Cadence client.Client
// 3. Cadence domain client
//
// It invokes the version reporter but does not start any workers.
//
// ClientModule should be used for Fx app's that just want
// the various clients but do not need/want any cadence workers.
//
// The clients are initialized using config values in config/[base/development].yaml
//
// Sample yaml for configuring a worker:
// cadence:
//    client:
//	  - domain: samples-domain
//
// In addition, yarpc needs to know about cadence thrift client:
// yarpc:
//   outbounds:
//     cadence:
//       service: cadence-frontend
//       tchannel:
//         peer: "127.0.0.1:7933"
//
var ClientModule = fx.Options(
	fx.Provide(NewServiceClient),
	fx.Provide(NewCadenceClient),
	fx.Provide(NewDomainClient),
)
