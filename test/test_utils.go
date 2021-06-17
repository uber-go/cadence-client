// Copyright (c) 2017-2020 Uber Technologies Inc.
// Portions of the Software are attributed to Copyright (c) 2020 Temporal Technologies Inc.
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

package test

import (
	"context"
	"fmt"
	"os"
	"strings"

	apiv1 "go.uber.org/cadence/v2/.gen/proto/api/v1"
	"go.uber.org/cadence/v2/internal/api"
	"go.uber.org/cadence/v2/workflow"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/grpc"
)

type (
	// Config contains the integration test configuration
	Config struct {
		ServiceAddr       string
		ServiceName       string
		IsStickyOff       bool
		Debug             bool
	}

	// context.WithValue need this type instead of basic type string to avoid lint error
	contextKey string
)

func newConfig() Config {
	cfg := Config{
		ServiceName:       "cadence-frontend",
		ServiceAddr:       "127.0.0.1:7833",
		IsStickyOff:       true,
	}
	if name := getEnvServiceName(); name != "" {
		cfg.ServiceName = name
	}
	if addr := getEnvServiceAddr(); addr != "" {
		cfg.ServiceAddr = addr
	}
	if so := getEnvStickyOff(); so != "" {
		cfg.IsStickyOff = so == "true"
	}
	if debug := getDebug(); debug != "" {
		cfg.Debug = debug == "true"
	}
	return cfg
}

func getEnvServiceName() string {
	return strings.TrimSpace(os.Getenv("SERVICE_NAME"))
}

func getEnvServiceAddr() string {
	return strings.TrimSpace(os.Getenv("SERVICE_ADDR"))
}

func getEnvStickyOff() string {
	return strings.ToLower(strings.TrimSpace(os.Getenv("STICKY_OFF")))
}

func getDebug() string {
	return strings.ToLower(strings.TrimSpace(os.Getenv("DEBUG")))
}

type cadenceAPI struct {
	apiv1.DomainAPIYARPCClient
	apiv1.WorkflowAPIYARPCClient
	apiv1.VisibilityAPIYARPCClient
	apiv1.WorkerAPIYARPCClient
}

type rpcClient struct {
	api.Interface
	dispatcher *yarpc.Dispatcher
}

func (c *rpcClient) Close() {
	c.dispatcher.Stop()
}

// newRPCClient builds and returns a new rpc client that is able to
// make calls to the localhost cadence-server container
func newRPCClient(
	serviceName string, serviceAddr string) (*rpcClient, error) {
	transport := grpc.NewTransport()
	outbound := transport.NewSingleOutbound(serviceAddr)
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: "integration-test",
		Outbounds: yarpc.Outbounds{
			serviceName: {
				Unary: outbound,
			},
		},
	})
	if err := dispatcher.Start(); err != nil {
		return nil, err
	}
	clientConfig := dispatcher.ClientConfig(serviceName)
	service := cadenceAPI{
		DomainAPIYARPCClient:     apiv1.NewDomainAPIYARPCClient(clientConfig),
		VisibilityAPIYARPCClient: apiv1.NewVisibilityAPIYARPCClient(clientConfig),
		WorkflowAPIYARPCClient:   apiv1.NewWorkflowAPIYARPCClient(clientConfig),
		WorkerAPIYARPCClient:     apiv1.NewWorkerAPIYARPCClient(clientConfig),
	}
	return &rpcClient{Interface: service, dispatcher: dispatcher}, nil
}

// stringMapPropagator propagates the list of keys across a workflow,
// interpreting the payloads as strings.
// BORROWED FROM 'internal' PACKAGE TESTS.
type stringMapPropagator struct {
	keys map[string]struct{}
}

// NewStringMapPropagator returns a context propagator that propagates a set of
// string key-value pairs across a workflow
func NewStringMapPropagator(keys []string) workflow.ContextPropagator {
	keyMap := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		keyMap[key] = struct{}{}
	}
	return &stringMapPropagator{keyMap}
}

// Inject injects values from context into headers for propagation
func (s *stringMapPropagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	for key := range s.keys {
		value, ok := ctx.Value(contextKey(key)).(string)
		if !ok {
			return fmt.Errorf("unable to extract key from context %v", key)
		}
		writer.Set(key, []byte(value))
	}
	return nil
}

// InjectFromWorkflow injects values from context into headers for propagation
func (s *stringMapPropagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	for key := range s.keys {
		value, ok := ctx.Value(contextKey(key)).(string)
		if !ok {
			return fmt.Errorf("unable to extract key from context %v", key)
		}
		writer.Set(key, []byte(value))
	}
	return nil
}

// Extract extracts values from headers and puts them into context
func (s *stringMapPropagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	if err := reader.ForEachKey(func(key string, value []byte) error {
		if _, ok := s.keys[key]; ok {
			ctx = context.WithValue(ctx, contextKey(key), string(value))
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return ctx, nil
}

// ExtractToWorkflow extracts values from headers and puts them into context
func (s *stringMapPropagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	if err := reader.ForEachKey(func(key string, value []byte) error {
		if _, ok := s.keys[key]; ok {
			ctx = workflow.WithValue(ctx, contextKey(key), string(value))
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return ctx, nil
}
