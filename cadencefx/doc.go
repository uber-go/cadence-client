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

// Package cadencefx configures workers and clients for a Cadence environment.
// Outside of use with Fx, the individual types and methods can be manually
// invoked if desired.
//
// Dependencies
//
// You will need a YARPC dispatcher, as well as a logging via Zap and metrics
// via Tally
//
// Usage
//
// To use this module in your fx application just pass the module variable to
// the fx.App constructor.
//
//  func main() {
//    fx.New(
//      // other fx modules
//      ...
//      cadencefx.Module,
//      ...
//    ).Run()
//  }
//
// Configuration
//
// The configuration for the module may be specified in YAML. The module will
// read the 'cadence' configuration section and the 'cadence' outbound under the
// 'yarpc' section.
//
// The 'cadence' configuration section accepts the following top level attributes:
// workers.
//
//    workers:
//      ...
//
// Cadence service endpoint configuration
//
// The 'cadence' outbound under the 'yarpc' section configures which Cadence
// service the clients and workers will be connected to. You can use any of the
// configuration options supported by YARPC outbounds for configuring the Cadence
// endpoint. You can specify the endpoint as 'host:port' or use one of TChannel's
// autodetection mechanisms, i.e. UNS or Muttley.
//
// Additionally, the you must specify a service name override for this outbound.
// When connecting to the Cadence 'staging' environment the service name override
// must be set to 'cadence-frontend-staging'. For all other environments, i.e
// development, local, or production, the service name override must be 'cadence-frontend',
// as shown in the samples below.
//
// To connect via 'host:port' to service running locally:
//  yarpc:
//    outbounds:
//      cadence:
//        service: cadence-frontend
//        tchannel:
//          peer: "127.0.0.1:1234"
//
// To connect via Muttley:
//  yarpc:
//    outbounds:
//      cadence:
//        service: cadence-frontend
//        tchannel:
//          with: muttley
//
// To connect via UNS:
//  yarpc:
//    outbounds:
//      cadence:
//        service: cadence-frontend
//        tchannel:
//          round-robin:
//            uns:
//              path: uns://dca1/dca1-prod01/us1/cadence-frontend:tchannel
//
// Worker configuration
//
// The 'workers' attribute configures one or more worker instances that poll the
// the Cadence service for 'tasks' to execute. This attribute is an array of
// workerConfig structures specifying the 'domain' and 'taskList' to poll.
//
// To configure a single worker:
//  workers:
//  - domain: sample-domain
//    taskList: sample-taskList
//
// To configure multiple workers polling different tasksLists in the same
// domain:
//  workers:
//  - domain: sample-domain
//    taskList: sample-taskList-one
//  - domain: sample-domain
//    taskList: sample-taskList-two
//
// To configure multiple workers polling taskLists in different domains:
//  workers:
//  - domain: sample-domain-one
//    taskList: sample-taskList
//  - domain: sample-domain-two
//    taskList: sample-taskList
//
// Provide context with outbound clients for activities
//
// Often activities will require that they have access to types that are
// created at runtime (often themselves constructed by Fx). To extensibly
// add your requirements to the context received by all activity invocations,
// the ContextInjectorResult and the related ContextInjector should be used:
//
//	fx.Provide(func(client myservice.Interface) ContextInjectorResult {
//		return ContextInjectorResult{
//			ContextInjector: func(ctx context.Context) context.Context {
//				// a separate method should be used by activities to retrieve
//				// the client from the context
//				return context.WithValue(ctx, clientKey, client)
//			},
//		}
//	})
package cadencefx
