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
	"testing"

	"go.uber.org/cadence/cadencefx/mocks"
	"go.uber.org/cadence/cadencefx/mocks/transport"

	"github.com/stretchr/testify/suite"
	"github.com/uber-go/tally"
	"go.uber.org/config"
)

const (
	hostSpecificTaskListName = "TestTaskList"
	hostSpecificDomainName   = "TestDomain"
)

type taggedScope struct {
	tally.Scope
}

func (ts *taggedScope) Capabilities() tally.Capabilities {
	return &capabilities{tagging: true}
}

type capabilities struct {
	reporting bool
	tagging   bool
}

func (c *capabilities) Reporting() bool {
	return c.reporting
}

func (c *capabilities) Tagging() bool {
	return c.tagging
}

var _testScope = &taggedScope{tally.NoopScope}

type ClientTestSuite struct {
	suite.Suite
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (s *ClientTestSuite) TestNewServiceClientEmptyConfig() {
	prov, err := config.NewYAMLProviderFromBytes([]byte(``))
	s.NoError(err)

	cc := new(mocks.ClientConfig)
	client, err := NewServiceClient(prov, cc)
	s.Nil(client)
	s.Equal("provide 'yarpc: outbounds' config", err.Error())
}

func (s *ClientTestSuite) TestNewServiceClientNoOutbounds() {
	prov, err := config.NewYAMLProviderFromBytes([]byte(`
    yarpc:
      outbounds:`))
	s.NoError(err)

	cc := new(mocks.ClientConfig)
	client, err := NewServiceClient(prov, cc)
	s.Nil(client)
	s.Equal("provide 'yarpc: outbounds' config", err.Error())
}

func (s *ClientTestSuite) TestNewServiceClientDuplicateCadenceFrontendConfig() {
	prov, err := config.NewYAMLProviderFromBytes([]byte(`
  yarpc:
    outbounds:
      xyz:`))
	s.NoError(err)

	cc := new(mocks.ClientConfig)
	client, err := NewServiceClient(prov, cc)
	s.Nil(client)
	s.Contains(err.Error(), "provide 'yarpc: outbounds: cadence'")
}

func (s *ClientTestSuite) TestNewServiceClientHostAndPort() {
	prov, err := config.NewYAMLProviderFromBytes([]byte(`
  yarpc:
    inbounds:
      tchannel:
        address: ""
    outbounds:
      cadence:
        service: cadence-frontend
        tchannel:
          peer: "127.0.0.1:1234"`))
	s.NoError(err)
	s.NotNil(prov)

	cc := new(mocks.ClientConfig)
	ccTransport := new(mockstransport.ClientConfig)
	cc.On("ClientConfig", "cadence").Return(ccTransport)
	client, err := NewServiceClient(prov, cc)
	s.NotNil(client)
	s.NoError(err)
}

func (s *ClientTestSuite) TestNewCadenceClientClientConfig() {
	service := new(mocks.Interface)
	prov, err := config.NewYAMLProviderFromBytes([]byte(`cadence:
    client:
      domain: domain1`))
	s.NoError(err)
	s.NotNil(prov)
	client, err := NewCadenceClient(service, prov, _testScope)
	s.NotNil(client)
	s.NoError(err)
}

func (s *ClientTestSuite) TestNewCadenceClientWorkerConfig() {
	service := new(mocks.Interface)
	prov, err := config.NewYAMLProviderFromBytes([]byte(`
cadence:
  workers:
  - domain: foo
    task_list: bar`))
	s.NoError(err)
	s.NotNil(prov)
	client, err := NewCadenceClient(service, prov, _testScope)
	s.NotNil(client)
	s.NoError(err)
}

// Ensures that we don't use a domain unless it's universally shared by all workers or explicitly
// set via the `client` config.
func (s *ClientTestSuite) TestNewCadenceClientWorkerDifferentDomainsConfig() {
	service := new(mocks.Interface)
	prov, err := config.NewYAMLProviderFromBytes([]byte(`
cadence:
  workers:
  - domain: foo
    task_list: bar
  - domain: notFoo
    task_list: bar`))
	s.NoError(err)
	s.NotNil(prov)
	client, err := NewCadenceClient(service, prov, _testScope)
	s.Nil(client)
	s.Contains(err.Error(), "Domain not set")
}

func (s *ClientTestSuite) TestNewCadenceClientBothConfigsButDifferentDomain() {
	service := new(mocks.Interface)
	prov, err := config.NewYAMLProviderFromBytes([]byte(`
cadence:
  client:
    domain: notFoo
  workers:
  - domain: foo
    task_list: bar`))
	s.NoError(err)
	s.NotNil(prov)
	client, err := NewCadenceClient(service, prov, _testScope)
	s.NotNil(client)
	s.NoError(err)
}

func (s *ClientTestSuite) TestNewCadenceClientEmptyDomain() {
	service := new(mocks.Interface)
	prov, err := config.NewYAMLProviderFromBytes([]byte(`cadence:
    client:
      domain: `))
	s.NoError(err)

	client, err := NewCadenceClient(service, prov, _testScope)
	s.Nil(client)
	s.Contains(err.Error(), "Domain not set")
}

func (s *ClientTestSuite) TestNewCadenceClientNoDomain() {
	service := new(mocks.Interface)
	prov, err := config.NewYAMLProviderFromBytes([]byte(``))
	s.NoError(err)

	client, err := NewCadenceClient(service, prov, _testScope)
	s.Nil(client)
	s.Contains(err.Error(), "Domain not set")
}

func (s *ClientTestSuite) TestNewDomainClient() {
	service := new(mocks.Interface)

	client := NewDomainClient(service, _testScope)
	s.NotNil(client)
}
