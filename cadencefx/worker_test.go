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
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"go.uber.org/cadence/cadencefx/mocks"
	"go.uber.org/cadence/worker"
	"go.uber.org/config"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

type WorkerTestSuite struct {
	suite.Suite
}

func TestWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerTestSuite))
}

func (s *WorkerTestSuite) TestNewWorkersEmptyConfig() {
	result, err := NewWorkers(newWorkerParams(``))
	s.NoError(err)
	s.Equal(0, len(result))
}

//
// This test verifies providing HostSpecificWorkerConfig will cause
// a host specific worker to be created, independent of config yaml
//
func (s *WorkerTestSuite) TestNewWorkersEmptyConfigWithPrivateTaskList() {
	params := newWorkerParams(``)

	wConfig := HostSpecificWorkerConfig{Domain: hostSpecificDomainName,
		TaskList: hostSpecificTaskListName,
		Options:  options{DisableWorkflowWorker: true}}
	params.HostSpecificWorkerConfig = &wConfig

	result, err := NewWorkers(params)
	s.NoError(err)
	s.Equal(1, len(result), "expected a host specific worker to be created.")
}

func (s *WorkerTestSuite) TestNewWorkersOneWorker() {
	result, err := NewWorkers(newWorkerParams(`
cadence:
  workers:
  - domain: foo
    task_list: bar`))

	s.NoError(err)
	s.Equal(1, len(result), "expected 1 worker to be created")
}

func (s *WorkerTestSuite) TestNewWorkersMultipleWorkers() {
	result, err := NewWorkers(newWorkerParams(`
cadence:
  workers:
  - domain: foo
    task_list: bar
  - domain: foo
    task_list: bar2`))

	s.NoError(err)
	s.Equal(2, len(result), "expected 2 workers to be created")
}

func (s *WorkerTestSuite) TestNewWorkersNoDomain() {
	result, err := NewWorkers(newWorkerParams(`
cadence:
  workers:
  - task_list: bar`))

	s.Error(err)
	s.True(strings.Contains(err.Error(), ErrConfigDomainNotSetMsg))
	s.Equal(0, len(result), "expected 0 worker to be created")
}

func (s *WorkerTestSuite) TestNewWorkersNoTaskList() {
	result, err := NewWorkers(newWorkerParams(`
cadence:
  workers:
  - domain: foo`))

	s.Error(err)
	s.True(strings.Contains(err.Error(), ErrConfigTaskListNotSetMsg))
	s.Equal(0, len(result), "expected 0 worker to be created")
}

func (s *WorkerTestSuite) TestNewWorkersWithActivityContext() {
	params := newWorkerParams(`
cadence:
  workers:
  - domain: foo
    task_list: bar`)

	result, err := NewWorkers(params)
	s.NoError(err)
	s.Equal(1, len(result), "expected 1 worker to be created")
}

//
// This test verifies that when HostSpecificWorkerConfig is set in WorkerParams,
// 1 additional worker is created as the host specific worker.
//
func (s *WorkerTestSuite) TestNewWorkersWithHostSpecificWorker() {
	params := newWorkerParams(`
cadence:
  workers:
  - domain: foo
    task_list: bar
  - domain: foo
    task_list: bar2`)

	wConfig := HostSpecificWorkerConfig{Domain: hostSpecificDomainName,
		TaskList: hostSpecificTaskListName,
		Options:  options{DisableWorkflowWorker: true}}
	params.HostSpecificWorkerConfig = &wConfig

	result, err := NewWorkers(params)
	s.NoError(err)
	s.Equal(3, len(result), "expected 3 = 2 + 1 workers to be created.")
}

func (s *WorkerTestSuite) TestRegisterWorkersSuccess() {
	lf := fxtest.NewLifecycle(s.T())
	worker1 := &testWorker{}
	worker2 := &testWorker{}

	RegisterWorkers(
		RegisterParams{
			Lifecycle: lf,
			Logger:    zap.NewNop(),
		},
		[]worker.Worker{worker1, worker2},
	)
	lf.RequireStart().RequireStop()

	s.True(worker1.StartCalled)
	s.True(worker1.StopCalled)
	s.True(worker2.StartCalled)
	s.True(worker2.StopCalled)
}

func (s *WorkerTestSuite) TestRegisterWorkersStartFails() {
	lf := fxtest.NewLifecycle(s.T())
	worker0 := &testWorker{FailStart: true}

	RegisterWorkers(
		RegisterParams{
			Lifecycle: lf,
			Logger:    zap.NewNop(),
		},
		[]worker.Worker{worker0},
	)

	err := lf.Start(context.Background())
	s.Equal("worker start failed", err.Error())
	s.True(worker0.StartCalled)
}

type testWorker struct {
	FailStart bool

	StartCalled bool
	StopCalled  bool
}

func (w *testWorker) Start() error {
	w.StartCalled = true
	if w.FailStart {
		return errors.New("worker start failed")
	}
	return nil
}

func (w *testWorker) Stop() {
	w.StopCalled = true
}

func (w *testWorker) Run() error {
	return nil
}

func newWorkerParams(cfg string) WorkerParams {
	return WorkerParams{
		ServiceClient: new(mocks.Interface),
		Scope:         _testScope,
		Logger:        zap.NewNop(),
		Config: func() config.Provider {
			prov, _ := config.NewYAMLProviderFromBytes([]byte(cfg))
			return prov
		}(),
	}
}
