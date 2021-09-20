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

package client

import (
	"time"

	cadence "go.uber.org/cadence/v2"
)

type (
	// DomainReplicationConfig configures domain replication. Use GlobalDomain or LocalDomain to create it.
	DomainReplicationConfig interface {
		domainReplicationConfig()
	}

	// DomainRegisterOption allows passing optional parameters when registering a domain
	DomainRegisterOption interface {
		domainRegisterOption()
	}

	// DomainUpdateOption allows passing optional parameters when updating the domain
	DomainUpdateOption interface {
		domainUpdateOption()
	}

	// DomainListOption allows passing optional parameters when listing domains
	DomainListOption interface {
		domainListOption()
	}

	// DomainDescribeOption allows passing optional parameters when describing domain
	DomainDescribeOption interface {
		domainDescribeOption()
	}

	// DomainFailoverOption allows passing optional parameters when making a domain failover
	DomainFailoverOption interface {
		domainFailoverOption()
	}

	// DomainDeprecateOption allows passing optional parameters when deprecating a domain
	DomainDeprecateOption interface {
		domainDeprecateOption()
	}

	// BadBinaryListOption allows passing optional parameters when listing bad binaries
	BadBinaryListOption interface {
		badBinaryListOption()
	}

	// BadBinaryAddOption allows passing optional parameters when adding bad binaries
	BadBinaryAddOption interface {
		badBinaryAddOption()
	}

	// BadBinaryDeleteOption allows passing optional parameters when deleting bad binaries
	BadBinaryDeleteOption interface {
		badBinaryDeleteOption()
	}

	// DomainOption can be used when registering or updating a domain
	DomainOption interface {
		DomainRegisterOption
		DomainUpdateOption
	}
)

// GlobalDomain can be used when registering a new domain to create its replication configuration.
// - clusters specify in which clusters domain will be registered
// - activeCluster specify an inital active cluster for the domain. This can later be changed with domain failover operation.
func GlobalDomain(activeCluster string, clusters ...string) DomainReplicationConfig {
	return &globalDomain{activeCluster, clusters}
}

// LocalDomain can be used when registering a new domain to create its replication configuration.
// LocalDomain will register domain only in one specified cluster and will not setup any replication.
func LocalDomain(cluster string) DomainReplicationConfig {
	return &localDomain{cluster}
}

// WithDomainDescription will set description of the domain
func WithDomainDescription(description string) DomainOption {
	return &domainDescription{description}
}

// WithDomainOwnerEmail will set owner email of the domain
func WithDomainOwnerEmail(email string) DomainOption {
	return &domainOwnerEmail{email}
}

// WithWorkflowRetentionPeriod will set workflow execution retention period of the domain
func WithWorkflowRetentionPeriod(period time.Duration) DomainOption {
	return &domainWorkflowRetention{period}
}

// WithDomainData will set arbitrary data map provided by the user for the domain
func WithDomainData(data map[string]string) DomainOption {
	return &domainData{data}
}

// WithHistoryArchival will set history archival parameters for the domain
func WithHistoryArchival(status cadence.ArchivalStatus, uri string) DomainOption {
	return &historyArchival{status, uri}
}

// WithVisibilityArchival will set visibility archival parameters for the domain
func WithVisibilityArchival(status cadence.ArchivalStatus, uri string) DomainOption {
	return &visibilityArchival{status, uri}
}

// WithGracefulFailover will use a graceful domain failover with provided timeout
func WithGracefulFailover(timeout time.Duration) DomainFailoverOption {
	return &gracefulFailover{timeout}
}

type (
	globalDomain struct {
		activeCluster string
		clusters      []string
	}
	localDomain struct {
		cluster string
	}
	domainDescription struct {
		description string
	}
	domainOwnerEmail struct {
		ownerEmail string
	}
	domainWorkflowRetention struct {
		period time.Duration
	}
	domainData struct {
		data map[string]string
	}
	historyArchival struct {
		status cadence.ArchivalStatus
		uri    string
	}
	visibilityArchival struct {
		status cadence.ArchivalStatus
		uri    string
	}
	gracefulFailover struct {
		timeout time.Duration
	}
)

func (globalDomain) domainReplicationConfig()
func (localDomain) domainReplicationConfig()

func (domainDescription) domainRegisterOption()
func (domainOwnerEmail) domainRegisterOption()
func (domainWorkflowRetention) domainRegisterOption()
func (domainData) domainRegisterOption()
func (historyArchival) domainRegisterOption()
func (visibilityArchival) domainRegisterOption()

func (domainDescription) domainUpdateOption()
func (domainOwnerEmail) domainUpdateOption()
func (domainWorkflowRetention) domainUpdateOption()
func (domainData) domainUpdateOption()
func (historyArchival) domainUpdateOption()
func (visibilityArchival) domainUpdateOption()

func (gracefulFailover) domainFailoverOption()
