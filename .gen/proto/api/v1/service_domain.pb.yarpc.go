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

// Code generated by protoc-gen-yarpc-go. DO NOT EDIT.
// source: uber/cadence/api/v1/service_domain.proto

package apiv1

import (
	"context"
	"io/ioutil"
	"reflect"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"go.uber.org/fx"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/api/x/restriction"
	"go.uber.org/yarpc/encoding/protobuf"
	"go.uber.org/yarpc/encoding/protobuf/reflection"
)

var _ = ioutil.NopCloser

// DomainAPIYARPCClient is the YARPC client-side interface for the DomainAPI service.
type DomainAPIYARPCClient interface {
	RegisterDomain(context.Context, *RegisterDomainRequest, ...yarpc.CallOption) (*RegisterDomainResponse, error)
	DescribeDomain(context.Context, *DescribeDomainRequest, ...yarpc.CallOption) (*DescribeDomainResponse, error)
	ListDomains(context.Context, *ListDomainsRequest, ...yarpc.CallOption) (*ListDomainsResponse, error)
	UpdateDomain(context.Context, *UpdateDomainRequest, ...yarpc.CallOption) (*UpdateDomainResponse, error)
	DeprecateDomain(context.Context, *DeprecateDomainRequest, ...yarpc.CallOption) (*DeprecateDomainResponse, error)
}

func newDomainAPIYARPCClient(clientConfig transport.ClientConfig, anyResolver jsonpb.AnyResolver, options ...protobuf.ClientOption) DomainAPIYARPCClient {
	return &_DomainAPIYARPCCaller{protobuf.NewStreamClient(
		protobuf.ClientParams{
			ServiceName:  "uber.cadence.api.v1.DomainAPI",
			ClientConfig: clientConfig,
			AnyResolver:  anyResolver,
			Options:      options,
		},
	)}
}

// NewDomainAPIYARPCClient builds a new YARPC client for the DomainAPI service.
func NewDomainAPIYARPCClient(clientConfig transport.ClientConfig, options ...protobuf.ClientOption) DomainAPIYARPCClient {
	return newDomainAPIYARPCClient(clientConfig, nil, options...)
}

// DomainAPIYARPCServer is the YARPC server-side interface for the DomainAPI service.
type DomainAPIYARPCServer interface {
	RegisterDomain(context.Context, *RegisterDomainRequest) (*RegisterDomainResponse, error)
	DescribeDomain(context.Context, *DescribeDomainRequest) (*DescribeDomainResponse, error)
	ListDomains(context.Context, *ListDomainsRequest) (*ListDomainsResponse, error)
	UpdateDomain(context.Context, *UpdateDomainRequest) (*UpdateDomainResponse, error)
	DeprecateDomain(context.Context, *DeprecateDomainRequest) (*DeprecateDomainResponse, error)
}

type buildDomainAPIYARPCProceduresParams struct {
	Server      DomainAPIYARPCServer
	AnyResolver jsonpb.AnyResolver
}

func buildDomainAPIYARPCProcedures(params buildDomainAPIYARPCProceduresParams) []transport.Procedure {
	handler := &_DomainAPIYARPCHandler{params.Server}
	return protobuf.BuildProcedures(
		protobuf.BuildProceduresParams{
			ServiceName: "uber.cadence.api.v1.DomainAPI",
			UnaryHandlerParams: []protobuf.BuildProceduresUnaryHandlerParams{
				{
					MethodName: "RegisterDomain",
					Handler: protobuf.NewUnaryHandler(
						protobuf.UnaryHandlerParams{
							Handle:      handler.RegisterDomain,
							NewRequest:  newDomainAPIServiceRegisterDomainYARPCRequest,
							AnyResolver: params.AnyResolver,
						},
					),
				},
				{
					MethodName: "DescribeDomain",
					Handler: protobuf.NewUnaryHandler(
						protobuf.UnaryHandlerParams{
							Handle:      handler.DescribeDomain,
							NewRequest:  newDomainAPIServiceDescribeDomainYARPCRequest,
							AnyResolver: params.AnyResolver,
						},
					),
				},
				{
					MethodName: "ListDomains",
					Handler: protobuf.NewUnaryHandler(
						protobuf.UnaryHandlerParams{
							Handle:      handler.ListDomains,
							NewRequest:  newDomainAPIServiceListDomainsYARPCRequest,
							AnyResolver: params.AnyResolver,
						},
					),
				},
				{
					MethodName: "UpdateDomain",
					Handler: protobuf.NewUnaryHandler(
						protobuf.UnaryHandlerParams{
							Handle:      handler.UpdateDomain,
							NewRequest:  newDomainAPIServiceUpdateDomainYARPCRequest,
							AnyResolver: params.AnyResolver,
						},
					),
				},
				{
					MethodName: "DeprecateDomain",
					Handler: protobuf.NewUnaryHandler(
						protobuf.UnaryHandlerParams{
							Handle:      handler.DeprecateDomain,
							NewRequest:  newDomainAPIServiceDeprecateDomainYARPCRequest,
							AnyResolver: params.AnyResolver,
						},
					),
				},
			},
			OnewayHandlerParams: []protobuf.BuildProceduresOnewayHandlerParams{},
			StreamHandlerParams: []protobuf.BuildProceduresStreamHandlerParams{},
		},
	)
}

// BuildDomainAPIYARPCProcedures prepares an implementation of the DomainAPI service for YARPC registration.
func BuildDomainAPIYARPCProcedures(server DomainAPIYARPCServer) []transport.Procedure {
	return buildDomainAPIYARPCProcedures(buildDomainAPIYARPCProceduresParams{Server: server})
}

// FxDomainAPIYARPCClientParams defines the input
// for NewFxDomainAPIYARPCClient. It provides the
// paramaters to get a DomainAPIYARPCClient in an
// Fx application.
type FxDomainAPIYARPCClientParams struct {
	fx.In

	Provider    yarpc.ClientConfig
	AnyResolver jsonpb.AnyResolver  `name:"yarpcfx" optional:"true"`
	Restriction restriction.Checker `optional:"true"`
}

// FxDomainAPIYARPCClientResult defines the output
// of NewFxDomainAPIYARPCClient. It provides a
// DomainAPIYARPCClient to an Fx application.
type FxDomainAPIYARPCClientResult struct {
	fx.Out

	Client DomainAPIYARPCClient

	// We are using an fx.Out struct here instead of just returning a client
	// so that we can add more values or add named versions of the client in
	// the future without breaking any existing code.
}

// NewFxDomainAPIYARPCClient provides a DomainAPIYARPCClient
// to an Fx application using the given name for routing.
//
//  fx.Provide(
//    apiv1.NewFxDomainAPIYARPCClient("service-name"),
//    ...
//  )
func NewFxDomainAPIYARPCClient(name string, options ...protobuf.ClientOption) interface{} {
	return func(params FxDomainAPIYARPCClientParams) FxDomainAPIYARPCClientResult {
		cc := params.Provider.ClientConfig(name)

		if params.Restriction != nil {
			if namer, ok := cc.GetUnaryOutbound().(transport.Namer); ok {
				if err := params.Restriction.Check(protobuf.Encoding, namer.TransportName()); err != nil {
					panic(err.Error())
				}
			}
		}

		return FxDomainAPIYARPCClientResult{
			Client: newDomainAPIYARPCClient(cc, params.AnyResolver, options...),
		}
	}
}

// FxDomainAPIYARPCProceduresParams defines the input
// for NewFxDomainAPIYARPCProcedures. It provides the
// paramaters to get DomainAPIYARPCServer procedures in an
// Fx application.
type FxDomainAPIYARPCProceduresParams struct {
	fx.In

	Server      DomainAPIYARPCServer
	AnyResolver jsonpb.AnyResolver `name:"yarpcfx" optional:"true"`
}

// FxDomainAPIYARPCProceduresResult defines the output
// of NewFxDomainAPIYARPCProcedures. It provides
// DomainAPIYARPCServer procedures to an Fx application.
//
// The procedures are provided to the "yarpcfx" value group.
// Dig 1.2 or newer must be used for this feature to work.
type FxDomainAPIYARPCProceduresResult struct {
	fx.Out

	Procedures     []transport.Procedure `group:"yarpcfx"`
	ReflectionMeta reflection.ServerMeta `group:"yarpcfx"`
}

// NewFxDomainAPIYARPCProcedures provides DomainAPIYARPCServer procedures to an Fx application.
// It expects a DomainAPIYARPCServer to be present in the container.
//
//  fx.Provide(
//    apiv1.NewFxDomainAPIYARPCProcedures(),
//    ...
//  )
func NewFxDomainAPIYARPCProcedures() interface{} {
	return func(params FxDomainAPIYARPCProceduresParams) FxDomainAPIYARPCProceduresResult {
		return FxDomainAPIYARPCProceduresResult{
			Procedures: buildDomainAPIYARPCProcedures(buildDomainAPIYARPCProceduresParams{
				Server:      params.Server,
				AnyResolver: params.AnyResolver,
			}),
			ReflectionMeta: reflection.ServerMeta{
				ServiceName:     "uber.cadence.api.v1.DomainAPI",
				FileDescriptors: yarpcFileDescriptorClosure2e37d15268893114,
			},
		}
	}
}

type _DomainAPIYARPCCaller struct {
	streamClient protobuf.StreamClient
}

func (c *_DomainAPIYARPCCaller) RegisterDomain(ctx context.Context, request *RegisterDomainRequest, options ...yarpc.CallOption) (*RegisterDomainResponse, error) {
	responseMessage, err := c.streamClient.Call(ctx, "RegisterDomain", request, newDomainAPIServiceRegisterDomainYARPCResponse, options...)
	if responseMessage == nil {
		return nil, err
	}
	response, ok := responseMessage.(*RegisterDomainResponse)
	if !ok {
		return nil, protobuf.CastError(emptyDomainAPIServiceRegisterDomainYARPCResponse, responseMessage)
	}
	return response, err
}

func (c *_DomainAPIYARPCCaller) DescribeDomain(ctx context.Context, request *DescribeDomainRequest, options ...yarpc.CallOption) (*DescribeDomainResponse, error) {
	responseMessage, err := c.streamClient.Call(ctx, "DescribeDomain", request, newDomainAPIServiceDescribeDomainYARPCResponse, options...)
	if responseMessage == nil {
		return nil, err
	}
	response, ok := responseMessage.(*DescribeDomainResponse)
	if !ok {
		return nil, protobuf.CastError(emptyDomainAPIServiceDescribeDomainYARPCResponse, responseMessage)
	}
	return response, err
}

func (c *_DomainAPIYARPCCaller) ListDomains(ctx context.Context, request *ListDomainsRequest, options ...yarpc.CallOption) (*ListDomainsResponse, error) {
	responseMessage, err := c.streamClient.Call(ctx, "ListDomains", request, newDomainAPIServiceListDomainsYARPCResponse, options...)
	if responseMessage == nil {
		return nil, err
	}
	response, ok := responseMessage.(*ListDomainsResponse)
	if !ok {
		return nil, protobuf.CastError(emptyDomainAPIServiceListDomainsYARPCResponse, responseMessage)
	}
	return response, err
}

func (c *_DomainAPIYARPCCaller) UpdateDomain(ctx context.Context, request *UpdateDomainRequest, options ...yarpc.CallOption) (*UpdateDomainResponse, error) {
	responseMessage, err := c.streamClient.Call(ctx, "UpdateDomain", request, newDomainAPIServiceUpdateDomainYARPCResponse, options...)
	if responseMessage == nil {
		return nil, err
	}
	response, ok := responseMessage.(*UpdateDomainResponse)
	if !ok {
		return nil, protobuf.CastError(emptyDomainAPIServiceUpdateDomainYARPCResponse, responseMessage)
	}
	return response, err
}

func (c *_DomainAPIYARPCCaller) DeprecateDomain(ctx context.Context, request *DeprecateDomainRequest, options ...yarpc.CallOption) (*DeprecateDomainResponse, error) {
	responseMessage, err := c.streamClient.Call(ctx, "DeprecateDomain", request, newDomainAPIServiceDeprecateDomainYARPCResponse, options...)
	if responseMessage == nil {
		return nil, err
	}
	response, ok := responseMessage.(*DeprecateDomainResponse)
	if !ok {
		return nil, protobuf.CastError(emptyDomainAPIServiceDeprecateDomainYARPCResponse, responseMessage)
	}
	return response, err
}

type _DomainAPIYARPCHandler struct {
	server DomainAPIYARPCServer
}

func (h *_DomainAPIYARPCHandler) RegisterDomain(ctx context.Context, requestMessage proto.Message) (proto.Message, error) {
	var request *RegisterDomainRequest
	var ok bool
	if requestMessage != nil {
		request, ok = requestMessage.(*RegisterDomainRequest)
		if !ok {
			return nil, protobuf.CastError(emptyDomainAPIServiceRegisterDomainYARPCRequest, requestMessage)
		}
	}
	response, err := h.server.RegisterDomain(ctx, request)
	if response == nil {
		return nil, err
	}
	return response, err
}

func (h *_DomainAPIYARPCHandler) DescribeDomain(ctx context.Context, requestMessage proto.Message) (proto.Message, error) {
	var request *DescribeDomainRequest
	var ok bool
	if requestMessage != nil {
		request, ok = requestMessage.(*DescribeDomainRequest)
		if !ok {
			return nil, protobuf.CastError(emptyDomainAPIServiceDescribeDomainYARPCRequest, requestMessage)
		}
	}
	response, err := h.server.DescribeDomain(ctx, request)
	if response == nil {
		return nil, err
	}
	return response, err
}

func (h *_DomainAPIYARPCHandler) ListDomains(ctx context.Context, requestMessage proto.Message) (proto.Message, error) {
	var request *ListDomainsRequest
	var ok bool
	if requestMessage != nil {
		request, ok = requestMessage.(*ListDomainsRequest)
		if !ok {
			return nil, protobuf.CastError(emptyDomainAPIServiceListDomainsYARPCRequest, requestMessage)
		}
	}
	response, err := h.server.ListDomains(ctx, request)
	if response == nil {
		return nil, err
	}
	return response, err
}

func (h *_DomainAPIYARPCHandler) UpdateDomain(ctx context.Context, requestMessage proto.Message) (proto.Message, error) {
	var request *UpdateDomainRequest
	var ok bool
	if requestMessage != nil {
		request, ok = requestMessage.(*UpdateDomainRequest)
		if !ok {
			return nil, protobuf.CastError(emptyDomainAPIServiceUpdateDomainYARPCRequest, requestMessage)
		}
	}
	response, err := h.server.UpdateDomain(ctx, request)
	if response == nil {
		return nil, err
	}
	return response, err
}

func (h *_DomainAPIYARPCHandler) DeprecateDomain(ctx context.Context, requestMessage proto.Message) (proto.Message, error) {
	var request *DeprecateDomainRequest
	var ok bool
	if requestMessage != nil {
		request, ok = requestMessage.(*DeprecateDomainRequest)
		if !ok {
			return nil, protobuf.CastError(emptyDomainAPIServiceDeprecateDomainYARPCRequest, requestMessage)
		}
	}
	response, err := h.server.DeprecateDomain(ctx, request)
	if response == nil {
		return nil, err
	}
	return response, err
}

func newDomainAPIServiceRegisterDomainYARPCRequest() proto.Message {
	return &RegisterDomainRequest{}
}

func newDomainAPIServiceRegisterDomainYARPCResponse() proto.Message {
	return &RegisterDomainResponse{}
}

func newDomainAPIServiceDescribeDomainYARPCRequest() proto.Message {
	return &DescribeDomainRequest{}
}

func newDomainAPIServiceDescribeDomainYARPCResponse() proto.Message {
	return &DescribeDomainResponse{}
}

func newDomainAPIServiceListDomainsYARPCRequest() proto.Message {
	return &ListDomainsRequest{}
}

func newDomainAPIServiceListDomainsYARPCResponse() proto.Message {
	return &ListDomainsResponse{}
}

func newDomainAPIServiceUpdateDomainYARPCRequest() proto.Message {
	return &UpdateDomainRequest{}
}

func newDomainAPIServiceUpdateDomainYARPCResponse() proto.Message {
	return &UpdateDomainResponse{}
}

func newDomainAPIServiceDeprecateDomainYARPCRequest() proto.Message {
	return &DeprecateDomainRequest{}
}

func newDomainAPIServiceDeprecateDomainYARPCResponse() proto.Message {
	return &DeprecateDomainResponse{}
}

var (
	emptyDomainAPIServiceRegisterDomainYARPCRequest   = &RegisterDomainRequest{}
	emptyDomainAPIServiceRegisterDomainYARPCResponse  = &RegisterDomainResponse{}
	emptyDomainAPIServiceDescribeDomainYARPCRequest   = &DescribeDomainRequest{}
	emptyDomainAPIServiceDescribeDomainYARPCResponse  = &DescribeDomainResponse{}
	emptyDomainAPIServiceListDomainsYARPCRequest      = &ListDomainsRequest{}
	emptyDomainAPIServiceListDomainsYARPCResponse     = &ListDomainsResponse{}
	emptyDomainAPIServiceUpdateDomainYARPCRequest     = &UpdateDomainRequest{}
	emptyDomainAPIServiceUpdateDomainYARPCResponse    = &UpdateDomainResponse{}
	emptyDomainAPIServiceDeprecateDomainYARPCRequest  = &DeprecateDomainRequest{}
	emptyDomainAPIServiceDeprecateDomainYARPCResponse = &DeprecateDomainResponse{}
)

var yarpcFileDescriptorClosure2e37d15268893114 = [][]byte{
	// uber/cadence/api/v1/service_domain.proto
	[]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x57, 0xdb, 0x6e, 0xdb, 0x46,
		0x13, 0x8e, 0x7c, 0x8a, 0x3c, 0xb4, 0x65, 0x7b, 0x7d, 0x62, 0x14, 0xe0, 0xff, 0x05, 0x05, 0x6d,
		0xd5, 0xb4, 0xa0, 0x6a, 0xa5, 0x27, 0x34, 0x57, 0xb6, 0xe5, 0xd4, 0x45, 0x9b, 0x40, 0xa0, 0xe3,
		0x02, 0x6d, 0x2f, 0xd8, 0x25, 0x39, 0x92, 0x17, 0xa2, 0x48, 0x76, 0x77, 0x29, 0x47, 0x79, 0x90,
		0x5e, 0xf6, 0xb9, 0xfa, 0x38, 0x05, 0x97, 0x4b, 0x47, 0x07, 0xca, 0x16, 0x12, 0xe7, 0x8e, 0x9a,
		0xfd, 0xe6, 0x9b, 0xd9, 0x99, 0xf9, 0x86, 0x22, 0x34, 0x12, 0x17, 0x79, 0xd3, 0xa3, 0x3e, 0x86,
		0x1e, 0x36, 0x69, 0xcc, 0x9a, 0xc3, 0xa3, 0xa6, 0x40, 0x3e, 0x64, 0x1e, 0x3a, 0x7e, 0x34, 0xa0,
		0x2c, 0xb4, 0x62, 0x1e, 0xc9, 0x88, 0xec, 0xa6, 0x48, 0x4b, 0x23, 0x2d, 0x1a, 0x33, 0x6b, 0x78,
		0x54, 0xfd, 0x5f, 0x2f, 0x8a, 0x7a, 0x01, 0x36, 0x15, 0xc4, 0x4d, 0xba, 0x4d, 0x3f, 0xe1, 0x54,
		0xb2, 0x48, 0x3b, 0x55, 0x6b, 0xd3, 0xe7, 0x5d, 0x86, 0x81, 0xef, 0x0c, 0xa8, 0xe8, 0xe7, 0x88,
		0xa2, 0x04, 0xc6, 0x03, 0xd7, 0xff, 0x5d, 0x83, 0x7d, 0x1b, 0x7b, 0x4c, 0x48, 0xe4, 0x6d, 0x75,
		0x60, 0xe3, 0x5f, 0x09, 0x0a, 0x49, 0x3e, 0x81, 0x8a, 0x40, 0x2f, 0xe1, 0x4c, 0x8e, 0x1c, 0x19,
		0xf5, 0x31, 0x34, 0x4b, 0xb5, 0x52, 0x63, 0xdd, 0xde, 0xcc, 0xad, 0xaf, 0x53, 0x23, 0x21, 0xb0,
		0x12, 0xd2, 0x01, 0x9a, 0x4b, 0xea, 0x50, 0x3d, 0x93, 0x1a, 0x18, 0x3e, 0x0a, 0x8f, 0xb3, 0x38,
		0xcd, 0xd6, 0x5c, 0x56, 0x47, 0xe3, 0x26, 0xf2, 0x7f, 0x30, 0xa2, 0xeb, 0x10, 0xb9, 0x83, 0x03,
		0xca, 0x02, 0x73, 0x45, 0x21, 0x40, 0x99, 0xce, 0x52, 0x0b, 0xb9, 0x82, 0x27, 0xd7, 0x11, 0xef,
		0x77, 0x83, 0xe8, 0xda, 0xc1, 0x37, 0xe8, 0x25, 0xa9, 0x9b, 0xc3, 0x51, 0x62, 0xa8, 0x9e, 0x62,
		0xe4, 0x2c, 0xf2, 0xcd, 0xd5, 0x5a, 0xa9, 0x61, 0xb4, 0x1e, 0x59, 0x59, 0x25, 0xac, 0xbc, 0x12,
		0x56, 0x5b, 0x57, 0xca, 0xae, 0xe5, 0x2c, 0x67, 0x39, 0x89, 0x9d, 0x73, 0x74, 0x14, 0x05, 0xe9,
		0x40, 0xd9, 0x0b, 0x92, 0xf4, 0xfe, 0xc2, 0x5c, 0xab, 0x2d, 0x37, 0x8c, 0xd6, 0xd7, 0x56, 0x41,
		0x37, 0xac, 0xd3, 0x0c, 0x64, 0x63, 0x1c, 0x30, 0x4f, 0x91, 0x9f, 0x46, 0x61, 0x97, 0xf5, 0xf2,
		0x48, 0x37, 0x2c, 0xc4, 0x82, 0x5d, 0xea, 0x49, 0x36, 0x44, 0x47, 0x9b, 0x1c, 0x55, 0xa1, 0x87,
		0xea, 0x92, 0x3b, 0xd9, 0x91, 0x66, 0x7b, 0x95, 0x96, 0xeb, 0x1c, 0x56, 0x7c, 0x2a, 0xa9, 0x59,
		0xbe, 0x25, 0x7a, 0x61, 0x8f, 0xac, 0x36, 0x95, 0xf4, 0x2c, 0x94, 0x7c, 0x64, 0x2b, 0x06, 0xd2,
		0x80, 0x6d, 0x26, 0x9c, 0x5e, 0x10, 0xb9, 0x34, 0xd0, 0x03, 0x66, 0xae, 0xd7, 0x4a, 0x8d, 0xb2,
		0x5d, 0x61, 0xe2, 0x47, 0x65, 0xce, 0x08, 0xc8, 0x1f, 0x70, 0x78, 0xc5, 0x84, 0x8c, 0xf8, 0xc8,
		0xa1, 0xdc, 0xbb, 0x62, 0x43, 0x1a, 0x38, 0x42, 0x52, 0x99, 0x08, 0x13, 0x6a, 0xa5, 0x46, 0xa5,
		0xf5, 0xa4, 0x30, 0x8d, 0x63, 0x8d, 0xbd, 0x50, 0x50, 0x7b, 0x5f, 0x73, 0x4c, 0x9a, 0xc9, 0x57,
		0xb0, 0x37, 0x43, 0x9e, 0x70, 0x66, 0x1a, 0xaa, 0x02, 0x64, 0xca, 0xe9, 0x92, 0x33, 0x42, 0xa1,
		0x3a, 0x64, 0x82, 0xb9, 0x2c, 0x48, 0xc7, 0x6d, 0x3a, 0xa3, 0x8d, 0xc5, 0x33, 0x32, 0xdf, 0xd1,
		0x4c, 0x25, 0xf5, 0x2d, 0x1c, 0x16, 0x85, 0x48, 0xf3, 0xda, 0x54, 0x79, 0xed, 0xcf, 0xba, 0x5e,
		0x72, 0x56, 0xfd, 0x0e, 0xd6, 0x6f, 0xca, 0x4c, 0xb6, 0x61, 0xb9, 0x8f, 0x23, 0xad, 0x84, 0xf4,
		0x91, 0xec, 0xc1, 0xea, 0x90, 0x06, 0x49, 0x2e, 0x80, 0xec, 0xc7, 0x0f, 0x4b, 0xdf, 0x97, 0xea,
		0x26, 0x1c, 0x4c, 0x77, 0x4d, 0xc4, 0x51, 0x28, 0xb0, 0xfe, 0x4f, 0x19, 0x76, 0x2f, 0x63, 0x9f,
		0x4a, 0xbc, 0x37, 0xc9, 0x3d, 0x07, 0x23, 0x51, 0x8c, 0x4a, 0xfe, 0xaa, 0x87, 0x46, 0xab, 0x3a,
		0xa3, 0x8b, 0x17, 0xe9, 0x86, 0x78, 0x49, 0x45, 0xdf, 0x86, 0x0c, 0x9e, 0x3e, 0x4f, 0xeb, 0xd5,
		0xb8, 0x53, 0xaf, 0x1b, 0x33, 0x7a, 0x7d, 0xa1, 0x67, 0x78, 0x53, 0xcd, 0x70, 0xab, 0xb0, 0x55,
		0x05, 0x57, 0x9e, 0x99, 0xe0, 0x05, 0x75, 0x5f, 0xf9, 0x70, 0xdd, 0x9f, 0xc2, 0x86, 0x4b, 0x7d,
		0xc7, 0x65, 0x21, 0xe5, 0x0c, 0x85, 0xb9, 0xa5, 0x28, 0x6b, 0x85, 0x99, 0x9f, 0x50, 0xff, 0x44,
		0xe3, 0x6c, 0xc3, 0x7d, 0xf7, 0xe3, 0x36, 0x19, 0x6d, 0x7f, 0x34, 0x19, 0xed, 0xbc, 0xa7, 0x8c,
		0xc8, 0x47, 0x96, 0xd1, 0xee, 0x2d, 0x32, 0x9a, 0xb7, 0x14, 0xf7, 0xe6, 0x2d, 0xc5, 0xf1, 0xb5,
		0xbc, 0x7f, 0x2f, 0x6b, 0xf9, 0x29, 0xec, 0xf8, 0x18, 0xa0, 0x44, 0xe7, 0xa6, 0xef, 0x23, 0xf3,
		0x40, 0xc5, 0xdf, 0xca, 0x0e, 0xf2, 0x36, 0x8f, 0x48, 0x1b, 0xb6, 0xbb, 0x94, 0x05, 0xd1, 0x10,
		0xb9, 0x23, 0xd9, 0x00, 0xa3, 0x44, 0x9a, 0x87, 0x77, 0xcd, 0xdc, 0x56, 0xee, 0xf2, 0x3a, 0xf3,
		0x78, 0xff, 0xd5, 0xf1, 0x33, 0xec, 0x4d, 0x8a, 0x25, 0x5b, 0x1c, 0xe4, 0x19, 0xac, 0xe9, 0xad,
		0x5e, 0x52, 0xc9, 0x3c, 0x2e, 0x2c, 0x89, 0x76, 0xd2, 0xd0, 0xfa, 0x05, 0x1c, 0xb4, 0x31, 0xe6,
		0xe8, 0xdd, 0xe3, 0xbe, 0xa9, 0x3f, 0x82, 0xc3, 0x19, 0x52, 0xbd, 0xdd, 0x5e, 0xc1, 0x7e, 0x5b,
		0xad, 0x0e, 0x77, 0x2a, 0xdc, 0x36, 0x2c, 0x31, 0x3f, 0x0b, 0x71, 0xfe, 0xc0, 0x5e, 0x62, 0x3e,
		0xd9, 0x1b, 0x67, 0x3e, 0x7f, 0x90, 0x71, 0x9f, 0x6c, 0xe6, 0xeb, 0xc8, 0x45, 0xc7, 0x1d, 0xd5,
		0x5f, 0xa6, 0xf9, 0x4f, 0xf2, 0x7d, 0x48, 0x39, 0x7e, 0x03, 0xf2, 0x0b, 0x13, 0x32, 0xb3, 0x8a,
		0x3c, 0xb7, 0xc7, 0xb0, 0x1e, 0xd3, 0x1e, 0x3a, 0x82, 0xbd, 0x45, 0xc5, 0xb6, 0x6a, 0x97, 0x53,
		0xc3, 0x05, 0x7b, 0x8b, 0xe4, 0x53, 0xd8, 0x0a, 0xf1, 0x8d, 0x74, 0x14, 0x22, 0x2b, 0x54, 0x9a,
		0xf1, 0x86, 0xbd, 0x99, 0x9a, 0x3b, 0xb4, 0x87, 0xaa, 0x50, 0x75, 0x09, 0xbb, 0x13, 0xd4, 0x3a,
		0xcd, 0x6f, 0xe0, 0x61, 0x16, 0x5b, 0x98, 0x25, 0x35, 0xc9, 0xb7, 0xe6, 0x99, 0x63, 0x17, 0x8d,
		0xda, 0xfa, 0x7b, 0x05, 0xd6, 0x33, 0xdf, 0xe3, 0xce, 0x4f, 0x84, 0x41, 0x65, 0xf2, 0xad, 0x43,
		0x9e, 0x2e, 0xfe, 0x87, 0xa2, 0xfa, 0xc5, 0x42, 0x58, 0x7d, 0x2f, 0x06, 0x95, 0xc9, 0xc6, 0xcc,
		0x09, 0x55, 0x38, 0x0d, 0x73, 0x42, 0xcd, 0xe9, 0xf4, 0x9f, 0x60, 0x8c, 0x55, 0x96, 0x7c, 0x56,
		0xe8, 0x3b, 0xdb, 0xd6, 0x6a, 0xe3, 0x6e, 0xa0, 0x8e, 0xe0, 0xc1, 0xc6, 0xb8, 0xe4, 0x48, 0x63,
		0xd1, 0x57, 0x58, 0xf5, 0xf3, 0x05, 0x90, 0x3a, 0x48, 0x00, 0x5b, 0x53, 0xaa, 0x21, 0xf3, 0xca,
		0x50, 0x24, 0xd8, 0xea, 0x97, 0x8b, 0x81, 0xb3, 0x68, 0x27, 0xbf, 0xc2, 0xa1, 0x17, 0x0d, 0x8a,
		0x5c, 0x4e, 0xca, 0xc7, 0x31, 0xeb, 0xa4, 0x0b, 0xac, 0x53, 0xfa, 0xbd, 0xd9, 0x63, 0xf2, 0x2a,
		0x71, 0x2d, 0x2f, 0x1a, 0x34, 0x27, 0xbe, 0x17, 0xac, 0x1e, 0x86, 0xd9, 0xc7, 0x85, 0xfe, 0x74,
		0x78, 0x4e, 0x63, 0x36, 0x3c, 0x72, 0xd7, 0x94, 0xed, 0xd9, 0x7f, 0x01, 0x00, 0x00, 0xff, 0xff,
		0xb6, 0x3b, 0xc8, 0x88, 0xdf, 0x0c, 0x00, 0x00,
	},
	// google/protobuf/duration.proto
	[]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x4b, 0xcf, 0xcf, 0x4f,
		0xcf, 0x49, 0xd5, 0x2f, 0x28, 0xca, 0x2f, 0xc9, 0x4f, 0x2a, 0x4d, 0xd3, 0x4f, 0x29, 0x2d, 0x4a,
		0x2c, 0xc9, 0xcc, 0xcf, 0xd3, 0x03, 0x8b, 0x08, 0xf1, 0x43, 0xe4, 0xf5, 0x60, 0xf2, 0x4a, 0x56,
		0x5c, 0x1c, 0x2e, 0x50, 0x25, 0x42, 0x12, 0x5c, 0xec, 0xc5, 0xa9, 0xc9, 0xf9, 0x79, 0x29, 0xc5,
		0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0xcc, 0x41, 0x30, 0xae, 0x90, 0x08, 0x17, 0x6b, 0x5e, 0x62, 0x5e,
		0x7e, 0xb1, 0x04, 0x93, 0x02, 0xa3, 0x06, 0x6b, 0x10, 0x84, 0xe3, 0xd4, 0xcc, 0xc8, 0x25, 0x9c,
		0x9c, 0x9f, 0xab, 0x87, 0x66, 0xa6, 0x13, 0x2f, 0xcc, 0xc4, 0x00, 0x90, 0x48, 0x00, 0x63, 0x94,
		0x21, 0x54, 0x45, 0x7a, 0x7e, 0x4e, 0x62, 0x5e, 0xba, 0x5e, 0x7e, 0x51, 0x3a, 0xc2, 0x81, 0x25,
		0x95, 0x05, 0xa9, 0xc5, 0xfa, 0xd9, 0x79, 0xf9, 0xe5, 0x79, 0x70, 0xc7, 0x16, 0x24, 0xfd, 0x60,
		0x64, 0x5c, 0xc4, 0xc4, 0xec, 0x1e, 0xe0, 0xb4, 0x8a, 0x49, 0xce, 0x1d, 0xa2, 0x39, 0x00, 0xaa,
		0x43, 0x2f, 0x3c, 0x35, 0x27, 0xc7, 0x1b, 0xa4, 0x3e, 0x04, 0xa4, 0x35, 0x89, 0x0d, 0x6c, 0x94,
		0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0xef, 0x8a, 0xb4, 0xc3, 0xfb, 0x00, 0x00, 0x00,
	},
	// google/protobuf/field_mask.proto
	[]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x48, 0xcf, 0xcf, 0x4f,
		0xcf, 0x49, 0xd5, 0x2f, 0x28, 0xca, 0x2f, 0xc9, 0x4f, 0x2a, 0x4d, 0xd3, 0x4f, 0xcb, 0x4c, 0xcd,
		0x49, 0x89, 0xcf, 0x4d, 0x2c, 0xce, 0xd6, 0x03, 0x8b, 0x09, 0xf1, 0x43, 0x54, 0xe8, 0xc1, 0x54,
		0x28, 0x29, 0x72, 0x71, 0xba, 0x81, 0x14, 0xf9, 0x26, 0x16, 0x67, 0x0b, 0x89, 0x70, 0xb1, 0x16,
		0x24, 0x96, 0x64, 0x14, 0x4b, 0x30, 0x2a, 0x30, 0x6b, 0x70, 0x06, 0x41, 0x38, 0x4e, 0xad, 0x8c,
		0x5c, 0xc2, 0xc9, 0xf9, 0xb9, 0x7a, 0x68, 0x5a, 0x9d, 0xf8, 0xe0, 0x1a, 0x03, 0x40, 0x42, 0x01,
		0x8c, 0x51, 0x46, 0x50, 0x25, 0xe9, 0xf9, 0x39, 0x89, 0x79, 0xe9, 0x7a, 0xf9, 0x45, 0xe9, 0x08,
		0xa7, 0x94, 0x54, 0x16, 0xa4, 0x16, 0xeb, 0x67, 0xe7, 0xe5, 0x97, 0xe7, 0x41, 0x9c, 0x05, 0x72,
		0x55, 0x41, 0xd2, 0x0f, 0x46, 0xc6, 0x45, 0x4c, 0xcc, 0xee, 0x01, 0x4e, 0xab, 0x98, 0xe4, 0xdc,
		0x21, 0xba, 0x03, 0xa0, 0x5a, 0xf4, 0xc2, 0x53, 0x73, 0x72, 0xbc, 0x41, 0x1a, 0x42, 0x40, 0x7a,
		0x93, 0xd8, 0xc0, 0x66, 0x19, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x94, 0x66, 0x94, 0x9d, 0xe6,
		0x00, 0x00, 0x00,
	},
	// uber/cadence/api/v1/domain.proto
	[]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x55, 0xdd, 0x72, 0xdb, 0x44,
		0x18, 0x45, 0x76, 0xea, 0x3a, 0x9f, 0x1c, 0xd7, 0x6c, 0x1b, 0xa2, 0x1a, 0x86, 0xa8, 0xe9, 0x30,
		0x63, 0x7a, 0x21, 0x13, 0xc3, 0x40, 0x0b, 0xc3, 0x85, 0x6d, 0x69, 0x8a, 0x99, 0x10, 0x3c, 0xb2,
		0x9b, 0x0b, 0xb8, 0xd0, 0xac, 0xa4, 0xb5, 0xbd, 0x53, 0x59, 0xab, 0x59, 0xfd, 0x84, 0xdc, 0x31,
		0x3c, 0x16, 0x0f, 0xc1, 0x33, 0x31, 0x5a, 0xad, 0x1c, 0xff, 0x68, 0x9a, 0xde, 0xed, 0x7e, 0x3f,
		0xe7, 0x3b, 0x3a, 0x3a, 0xbb, 0x0b, 0x7a, 0xea, 0x12, 0xde, 0xf7, 0xb0, 0x4f, 0x42, 0x8f, 0xf4,
		0x71, 0x44, 0xfb, 0xd9, 0x65, 0xdf, 0x67, 0x6b, 0x4c, 0x43, 0x23, 0xe2, 0x2c, 0x61, 0xe8, 0x69,
		0x5e, 0x61, 0xc8, 0x0a, 0x03, 0x47, 0xd4, 0xc8, 0x2e, 0xbb, 0x5f, 0x2e, 0x19, 0x5b, 0x06, 0xa4,
		0x2f, 0x4a, 0xdc, 0x74, 0xd1, 0xf7, 0x53, 0x8e, 0x13, 0xca, 0x64, 0x53, 0xf7, 0x7c, 0x3f, 0x9f,
		0xd0, 0x35, 0x89, 0x13, 0xbc, 0x8e, 0x8a, 0x82, 0x8b, 0xff, 0x1e, 0x43, 0xc3, 0x14, 0x63, 0x50,
		0x1b, 0x6a, 0xd4, 0xd7, 0x14, 0x5d, 0xe9, 0x1d, 0xdb, 0x35, 0xea, 0x23, 0x04, 0x47, 0x21, 0x5e,
		0x13, 0xad, 0x26, 0x22, 0x62, 0x8d, 0xde, 0x40, 0x23, 0x4e, 0x70, 0x92, 0xc6, 0x5a, 0x5d, 0x57,
		0x7a, 0xed, 0xc1, 0x0b, 0xa3, 0x82, 0x95, 0x51, 0x00, 0xce, 0x44, 0xa1, 0x2d, 0x1b, 0x90, 0x0e,
		0xaa, 0x4f, 0x62, 0x8f, 0xd3, 0x28, 0xe7, 0xa7, 0x1d, 0x09, 0xd4, 0xed, 0x10, 0x3a, 0x07, 0x95,
		0xdd, 0x86, 0x84, 0x3b, 0x64, 0x8d, 0x69, 0xa0, 0x3d, 0x12, 0x15, 0x20, 0x42, 0x56, 0x1e, 0x41,
		0x6f, 0xe0, 0xc8, 0xc7, 0x09, 0xd6, 0x1a, 0x7a, 0xbd, 0xa7, 0x0e, 0xbe, 0xfa, 0xc0, 0x6c, 0xc3,
		0xc4, 0x09, 0xb6, 0xc2, 0x84, 0xdf, 0xd9, 0xa2, 0x05, 0xad, 0xe0, 0xe5, 0x2d, 0xe3, 0xef, 0x17,
		0x01, 0xbb, 0x75, 0xc8, 0x5f, 0xc4, 0x4b, 0xf3, 0x89, 0x0e, 0x27, 0x09, 0x09, 0xc5, 0x2a, 0x22,
		0x9c, 0x32, 0x5f, 0x7b, 0xac, 0x2b, 0x3d, 0x75, 0xf0, 0xdc, 0x28, 0x64, 0x33, 0x4a, 0xd9, 0x0c,
		0x53, 0xca, 0x6a, 0xeb, 0x25, 0x8a, 0x55, 0x82, 0xd8, 0x25, 0xc6, 0x54, 0x40, 0xa0, 0x31, 0xb4,
		0x5c, 0xec, 0x3b, 0x2e, 0x0d, 0x31, 0xa7, 0x24, 0xd6, 0x9a, 0x02, 0x52, 0xaf, 0x24, 0x3b, 0xc2,
		0xfe, 0x48, 0xd6, 0xd9, 0xaa, 0x7b, 0xbf, 0x41, 0x7f, 0xc2, 0xd9, 0x8a, 0xc6, 0x09, 0xe3, 0x77,
		0x0e, 0xe6, 0xde, 0x8a, 0x66, 0x38, 0x70, 0xa4, 0xf0, 0xc7, 0x42, 0xf8, 0x97, 0x95, 0x78, 0x43,
		0x59, 0x2b, 0xa5, 0x3f, 0x95, 0x18, 0xbb, 0x61, 0xf4, 0x0d, 0x3c, 0x3b, 0x00, 0x4f, 0x39, 0xd5,
		0x40, 0x08, 0x8e, 0xf6, 0x9a, 0xde, 0x71, 0x8a, 0x30, 0x74, 0x33, 0x1a, 0x53, 0x97, 0x06, 0x34,
		0x39, 0x64, 0xa4, 0x7e, 0x3c, 0x23, 0xed, 0x1e, 0x66, 0x8f, 0xd4, 0xf7, 0x70, 0x56, 0x35, 0x22,
		0xe7, 0xd5, 0x12, 0xbc, 0x4e, 0x0f, 0x5b, 0x73, 0x6a, 0x06, 0x3c, 0xc5, 0x5e, 0x42, 0x33, 0xe2,
		0x78, 0x41, 0x1a, 0x27, 0x84, 0x3b, 0xc2, 0xb4, 0x27, 0xa2, 0xe7, 0xd3, 0x22, 0x35, 0x2e, 0x32,
		0xd7, 0xb9, 0x83, 0xa7, 0xd0, 0x94, 0x85, 0xb1, 0xd6, 0x16, 0x3e, 0xfa, 0xae, 0x92, 0xb8, 0xec,
		0xb1, 0x49, 0x14, 0x50, 0x4f, 0xfc, 0xfb, 0x31, 0x0b, 0x17, 0x74, 0x59, 0x1a, 0x61, 0x83, 0x82,
		0xbe, 0x86, 0xce, 0x02, 0xd3, 0x80, 0x65, 0x84, 0x3b, 0x19, 0xe1, 0x71, 0xee, 0xee, 0x27, 0xba,
		0xd2, 0xab, 0xdb, 0x4f, 0xca, 0xf8, 0x4d, 0x11, 0x46, 0x3d, 0xe8, 0xd0, 0xd8, 0x59, 0x06, 0xcc,
		0xc5, 0x81, 0x53, 0x9c, 0x6e, 0xad, 0xa3, 0x2b, 0xbd, 0xa6, 0xdd, 0xa6, 0xf1, 0x5b, 0x11, 0x2e,
		0xfc, 0xdb, 0xfd, 0x01, 0x8e, 0x37, 0x16, 0x46, 0x1d, 0xa8, 0xbf, 0x27, 0x77, 0xf2, 0x68, 0xe6,
		0x4b, 0xf4, 0x0c, 0x1e, 0x65, 0x38, 0x48, 0xcb, 0xc3, 0x59, 0x6c, 0x7e, 0xac, 0xbd, 0x56, 0x2e,
		0x4c, 0x38, 0x7f, 0x80, 0x3a, 0x7a, 0x01, 0xad, 0x1d, 0xad, 0x0a, 0x5c, 0xd5, 0xbb, 0x57, 0xe9,
		0xe2, 0x5f, 0x05, 0xd4, 0x2d, 0x73, 0xa2, 0x5f, 0xa1, 0xb9, 0x31, 0xb4, 0x22, 0x54, 0x33, 0x1e,
		0x32, 0xb4, 0x51, 0x2e, 0x8a, 0x63, 0xb8, 0xe9, 0xef, 0x3a, 0x70, 0xb2, 0x93, 0xaa, 0xf8, 0xbc,
		0xd7, 0xdb, 0x9f, 0xa7, 0x0e, 0x2e, 0x3e, 0x38, 0xeb, 0x6e, 0x12, 0x2e, 0xd8, 0xb6, 0x04, 0xff,
		0x28, 0x70, 0xb2, 0x93, 0x44, 0x9f, 0x41, 0x83, 0x13, 0x1c, 0xb3, 0x50, 0x0e, 0x91, 0x3b, 0xd4,
		0x85, 0x26, 0x8b, 0x08, 0xc7, 0x09, 0xe3, 0x52, 0xc9, 0xcd, 0x1e, 0xfd, 0x0c, 0x2d, 0x8f, 0x13,
		0x9c, 0x10, 0xdf, 0xc9, 0x2f, 0x4d, 0x71, 0xe1, 0xa9, 0x83, 0xee, 0xc1, 0xd5, 0x30, 0x2f, 0x6f,
		0x54, 0x5b, 0x95, 0xf5, 0x79, 0xe4, 0xd5, 0xdf, 0x0a, 0xb4, 0xb6, 0xef, 0x41, 0xf4, 0x1c, 0x4e,
		0xcd, 0xdf, 0x7f, 0x1b, 0x4e, 0xae, 0x9d, 0xd9, 0x7c, 0x38, 0x7f, 0x37, 0x73, 0x26, 0xd7, 0x37,
		0xc3, 0xab, 0x89, 0xd9, 0xf9, 0x04, 0x7d, 0x01, 0xda, 0x6e, 0xca, 0xb6, 0xde, 0x4e, 0x66, 0x73,
		0xcb, 0xb6, 0xcc, 0x8e, 0x72, 0x98, 0x35, 0xad, 0xa9, 0x6d, 0x8d, 0x87, 0x73, 0xcb, 0xec, 0xd4,
		0x0e, 0x61, 0x4d, 0xeb, 0xca, 0xca, 0x53, 0xf5, 0x57, 0x2b, 0x68, 0xef, 0x1d, 0xb2, 0xcf, 0xe1,
		0x6c, 0x68, 0x8f, 0x7f, 0x99, 0xdc, 0x0c, 0xaf, 0x2a, 0x59, 0xec, 0x27, 0xcd, 0xc9, 0x6c, 0x38,
		0xba, 0x12, 0x2c, 0x2a, 0x5a, 0xad, 0xeb, 0x22, 0x59, 0x1b, 0xdd, 0xc0, 0x99, 0xc7, 0xd6, 0x55,
		0x7f, 0x69, 0xd4, 0x1c, 0x46, 0x74, 0x9a, 0x6b, 0x35, 0x55, 0xfe, 0xe8, 0x2f, 0x69, 0xb2, 0x4a,
		0x5d, 0xc3, 0x63, 0xeb, 0xfe, 0xce, 0x7b, 0x67, 0x2c, 0x49, 0x58, 0xbc, 0x51, 0xf2, 0xe9, 0xfb,
		0x09, 0x47, 0x34, 0xbb, 0x74, 0x1b, 0x22, 0xf6, 0xed, 0xff, 0x01, 0x00, 0x00, 0xff, 0xff, 0xd4,
		0x52, 0xa7, 0xe0, 0x1e, 0x07, 0x00, 0x00,
	},
	// google/protobuf/timestamp.proto
	[]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x4f, 0xcf, 0xcf, 0x4f,
		0xcf, 0x49, 0xd5, 0x2f, 0x28, 0xca, 0x2f, 0xc9, 0x4f, 0x2a, 0x4d, 0xd3, 0x2f, 0xc9, 0xcc, 0x4d,
		0x2d, 0x2e, 0x49, 0xcc, 0x2d, 0xd0, 0x03, 0x0b, 0x09, 0xf1, 0x43, 0x14, 0xe8, 0xc1, 0x14, 0x28,
		0x59, 0x73, 0x71, 0x86, 0xc0, 0xd4, 0x08, 0x49, 0x70, 0xb1, 0x17, 0xa7, 0x26, 0xe7, 0xe7, 0xa5,
		0x14, 0x4b, 0x30, 0x2a, 0x30, 0x6a, 0x30, 0x07, 0xc1, 0xb8, 0x42, 0x22, 0x5c, 0xac, 0x79, 0x89,
		0x79, 0xf9, 0xc5, 0x12, 0x4c, 0x0a, 0x8c, 0x1a, 0xac, 0x41, 0x10, 0x8e, 0x53, 0x2b, 0x23, 0x97,
		0x70, 0x72, 0x7e, 0xae, 0x1e, 0x9a, 0xa1, 0x4e, 0x7c, 0x70, 0x23, 0x03, 0x40, 0x42, 0x01, 0x8c,
		0x51, 0x46, 0x50, 0x25, 0xe9, 0xf9, 0x39, 0x89, 0x79, 0xe9, 0x7a, 0xf9, 0x45, 0xe9, 0x48, 0x6e,
		0xac, 0x2c, 0x48, 0x2d, 0xd6, 0xcf, 0xce, 0xcb, 0x2f, 0xcf, 0x43, 0xb8, 0xb7, 0x20, 0xe9, 0x07,
		0x23, 0xe3, 0x22, 0x26, 0x66, 0xf7, 0x00, 0xa7, 0x55, 0x4c, 0x72, 0xee, 0x10, 0xdd, 0x01, 0x50,
		0x2d, 0x7a, 0xe1, 0xa9, 0x39, 0x39, 0xde, 0x20, 0x0d, 0x21, 0x20, 0xbd, 0x49, 0x6c, 0x60, 0xb3,
		0x8c, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0xae, 0x65, 0xce, 0x7d, 0xff, 0x00, 0x00, 0x00,
	},
}

func init() {
	yarpc.RegisterClientBuilder(
		func(clientConfig transport.ClientConfig, structField reflect.StructField) DomainAPIYARPCClient {
			return NewDomainAPIYARPCClient(clientConfig, protobuf.ClientBuilderOptions(clientConfig, structField)...)
		},
	)
}
