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
	"errors"

	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/client"
	"go.uber.org/config"
	"go.uber.org/yarpc"
)

const (
	_svcName                 = "cadence"
	_yarpcConfigKey          = "yarpc"
	_yarpcOutboundsConfigKey = "outbounds"
)

// NewServiceClient instantiates a new Cadence service client based on provided
// configuration.
func NewServiceClient(
	configProvider config.Provider,
	yarpcProvider yarpc.ClientConfig,
) (workflowserviceclient.Interface, error) {
	// Note: Cannot read directly into yarpc.Config here, it does not map to config directly.
	// Read into outbounds instead
	var yOutbounds yarpc.Outbounds
	err := configProvider.Get(_yarpcConfigKey).Get(_yarpcOutboundsConfigKey).Populate(&yOutbounds)
	if err != nil {
		return nil, err
	}
	if len(yOutbounds) == 0 {
		return nil, errors.New("provide 'yarpc: outbounds' config")
	}
	if _, ok := yOutbounds[_svcName]; !ok {
		return nil, errors.New("provide 'yarpc: outbounds: cadence' config")
	}
	return workflowserviceclient.New(yarpcProvider.ClientConfig(_svcName)), nil
}

// NewCadenceClient instantiates a new Cadence service client that can be used
// to interact with workflows and activities.
func NewCadenceClient(service workflowserviceclient.Interface,
	configProvider config.Provider,
	scope tally.Scope) (client.Client, error) {
	var domain string
	clientConfig, clientConfigErr := loadClientConfig(configProvider)
	if clientConfigErr != nil {
		// If there is an error loading the client config, we try to fallback to worker config.
		workerConfig, workerConfigErr := loadWorkerConfig(configProvider)
		if workerConfigErr != nil || workerConfig.Workers == nil || len(workerConfig.Workers) == 0 {
			// Return the original client config err if both configs are incomplete.
			return nil, clientConfigErr
		}
		// Test that all workers share the same domain
		domain = workerConfig.Workers[0].Domain
		for _, d := range workerConfig.Workers {
			if domain != d.Domain {
				return nil, clientConfigErr
			}
		}
	} else {
		domain = clientConfig.Client.Domain
	}
	return client.NewClient(service, domain, &client.Options{
		MetricsScope: scope,
	}), nil
}

// NewDomainClient instantiates a new Cadence service client that can be used
// to manipulate domains.
func NewDomainClient(service workflowserviceclient.Interface, scope tally.Scope) client.DomainClient {
	return client.NewDomainClient(
		service, &client.Options{
			MetricsScope: scope,
		})
}
