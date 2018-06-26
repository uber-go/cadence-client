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

package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
	"sync"
)

func Test_Counter(t *testing.T) {
	isReplay := true
	scope, closer, reporter := NewMetricsScope(&isReplay)
	scope.Counter("test-name").Inc(1)
	closer.Close()
	require.Equal(t, 0, len(reporter.Counts()))

	isReplay = false
	scope, closer, reporter = NewMetricsScope(&isReplay)
	scope.Counter("test-name").Inc(3)
	closer.Close()
	require.Equal(t, 1, len(reporter.Counts()))
	require.Equal(t, int64(3), reporter.Counts()[0].Value())
}

func Test_Gauge(t *testing.T) {
	isReplay := true
	scope, closer, reporter := NewMetricsScope(&isReplay)
	scope.Gauge("test-name").Update(1)
	closer.Close()
	require.Equal(t, 0, len(reporter.Gauges()))

	isReplay = false
	scope, closer, reporter = NewMetricsScope(&isReplay)
	scope.Gauge("test-name").Update(3)
	closer.Close()
	require.Equal(t, 1, len(reporter.Gauges()))
	require.Equal(t, float64(3), reporter.Gauges()[0].Value())
}

func Test_Timer(t *testing.T) {
	isReplay := true
	scope, closer, reporter := NewMetricsScope(&isReplay)
	scope.Timer("test-name").Record(time.Second)
	sw := scope.Timer("test-stopwatch").Start()
	sw.Stop()
	closer.Close()
	require.Equal(t, 0, len(reporter.Timers()))

	isReplay = false
	scope, closer, reporter = NewMetricsScope(&isReplay)
	scope.Timer("test-name").Record(time.Second)
	sw = scope.Timer("test-stopwatch").Start()
	sw.Stop()
	closer.Close()
	require.Equal(t, 2, len(reporter.Timers()))
	require.Equal(t, time.Second, reporter.Timers()[0].Value())
}

func Test_Histogram(t *testing.T) {
	isReplay := true
	scope, closer, reporter := NewMetricsScope(&isReplay)
	valueBuckets := tally.MustMakeLinearValueBuckets(0, 10, 10)
	scope.Histogram("test-hist-1", valueBuckets).RecordValue(5)
	scope.Histogram("test-hist-2", valueBuckets).RecordValue(15)
	closer.Close()
	require.Equal(t, 0, len(reporter.HistogramValueSamples()))
	scope, closer, reporter = NewMetricsScope(&isReplay)
	durationBuckets := tally.MustMakeLinearDurationBuckets(0, time.Hour, 10)
	scope.Histogram("test-hist-1", durationBuckets).RecordDuration(time.Minute)
	scope.Histogram("test-hist-2", durationBuckets).RecordDuration(time.Minute * 61)
	sw := scope.Histogram("test-hist-3", durationBuckets).Start()
	sw.Stop()
	closer.Close()
	require.Equal(t, 0, len(reporter.HistogramDurationSamples()))

	isReplay = false
	scope, closer, reporter = NewMetricsScope(&isReplay)
	valueBuckets = tally.MustMakeLinearValueBuckets(0, 10, 10)
	scope.Histogram("test-hist-1", valueBuckets).RecordValue(5)
	scope.Histogram("test-hist-2", valueBuckets).RecordValue(15)
	closer.Close()
	require.Equal(t, 2, len(reporter.HistogramValueSamples()))

	scope, closer, reporter = NewMetricsScope(&isReplay)
	durationBuckets = tally.MustMakeLinearDurationBuckets(0, time.Hour, 10)
	scope.Histogram("test-hist-1", durationBuckets).RecordDuration(time.Minute)
	scope.Histogram("test-hist-2", durationBuckets).RecordDuration(time.Minute * 61)
	sw = scope.Histogram("test-hist-3", durationBuckets).Start()
	sw.Stop()
	closer.Close()
	require.Equal(t, 3, len(reporter.HistogramDurationSamples()))
}

func Test_ScopeCoverage(t *testing.T) {
	isReplay := false
	scope, closer, reporter := NewMetricsScope(&isReplay)
	caps := scope.Capabilities()
	require.Equal(t, true, caps.Reporting())
	require.Equal(t, true, caps.Tagging())
	subScope := scope.SubScope("test")
	taggedScope := subScope.Tagged(make(map[string]string))
	taggedScope.Counter("test-counter").Inc(1)
	closer.Close()
	require.Equal(t, 1, len(reporter.Counts()))
}

func Test_TaggedScope(t *testing.T) {
	taggedScope, closer, reporter := NewTaggedMetricsScope()
	scope := taggedScope.GetTaggedScope("tag1", "val1")
	scope.Counter("test-name").Inc(3)
	closer.Close()
	require.Equal(t, 1, len(reporter.Counts()))
	require.Equal(t, int64(3), reporter.Counts()[0].Value())

	m := &sync.Map{}
	taggedScope, closer, reporter = NewTaggedMetricsScope()
	taggedScope.Map = m
	scope = taggedScope.GetTaggedScope("tag2", "val1")
	scope.Counter("test-name").Inc(2)
	taggedScope, closer2, reporter2 := NewTaggedMetricsScope()
	taggedScope.Map = m
	scope = taggedScope.GetTaggedScope("tag2", "val1")
	scope.Counter("test-name").Inc(1)
	closer2.Close()
	require.Equal(t, 0, len(reporter2.Counts()))
	closer.Close()
	require.Equal(t, 1, len(reporter.Counts()))
	require.Equal(t, int64(3), reporter.Counts()[0].Value())
}