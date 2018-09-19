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
	"time"

	"go.uber.org/config"
)

const (
	// ConfigurationKey is the portion of the service configuration that this
	// package reads.
	ConfigurationKey = "cadence"

	// ErrConfigDomainNotSetMsg helpful message specifying the domain is not set
	ErrConfigDomainNotSetMsg = "Domain not set. "

	// ErrConfigTaskListNotSetMsg helpful message specifying the tasklist is not set
	ErrConfigTaskListNotSetMsg = "TaskList not set. "

	// ErrClientConfigNotSetMsg helpful message specifying the correct config
	ErrClientConfigNotSetMsg = `
	Client config should be similar to below:
	client:
  	domain: domain1
	`

	// ErrWorkerConfigNotSetMsg helpful message specifying the correct config
	ErrWorkerConfigNotSetMsg = `
	Worker config should be similar to below:
	workers:
  - domain: domain1
		task_list: tasklist1
	- domain: domain2
		task_list: tasklist2
	`
)

type (
	clientConfig struct {
		Domain string `yaml:"domain"`
	}

	workerConfig struct {
		Domain   string `yaml:"domain"`
		TaskList string `yaml:"task_list"`
		Options  options
	}

	options struct {
		TaskListActivitiesPerSecond  float64
		WorkerActivitiesPerSecond    float64
		AutoHeartBeat                bool
		Identity                     string
		EnableLoggingInReplay        bool
		DisableWorkflowWorker        bool
		DisableActivityWorker        bool
		DisableStickyExecution       bool
		StickyScheduleToStartTimeout time.Duration
	}

	moduleConfig struct {
		Workers []workerConfig
		Client  clientConfig
	}
)

func loadClientConfig(provider config.Provider) (moduleConfig, error) {
	var config moduleConfig
	err := provider.Get(ConfigurationKey).Populate(&config)
	if err != nil {
		return moduleConfig{}, err
	}
	if len(config.Client.Domain) == 0 || config.Client.Domain == "" {
		return moduleConfig{}, errors.New(ErrConfigDomainNotSetMsg + ErrClientConfigNotSetMsg)
	}
	return config, nil
}

func loadWorkerConfig(provider config.Provider) (moduleConfig, error) {
	var config moduleConfig

	err := provider.Get(ConfigurationKey).Populate(&config)
	if err != nil {
		return moduleConfig{}, err
	}

	for _, config := range config.Workers {
		if err := validateWorkerConfig(config); err != nil {
			return moduleConfig{}, err
		}
	}

	return config, nil
}

func validateWorkerConfig(config workerConfig) error {
	if len(config.Domain) == 0 {
		return errors.New(ErrConfigDomainNotSetMsg + ErrWorkerConfigNotSetMsg)
	}
	if len(config.TaskList) == 0 {
		return errors.New(ErrConfigTaskListNotSetMsg + ErrWorkerConfigNotSetMsg)
	}
	return nil
}
