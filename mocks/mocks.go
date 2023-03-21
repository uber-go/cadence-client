package mocks

// Type ''go generate'' to rebuild needed mock.
//go:generate go install -v github.com/vektra/mockery/v2@v2.23.0
//go:generate mockery --dir=../client --name=Client
//go:generate mockery --dir=../client --name=DomainClient
//go:generate mockery --dir=../internal --name=HistoryEventIterator
//go:generate mockery --dir=../internal --name=Value
//go:generate mockery --dir=../internal --name=WorkflowRun
//go:generate mockery --dir=../worker --name=Registry
//go:generate mockery --dir=../worker --name=Worker
