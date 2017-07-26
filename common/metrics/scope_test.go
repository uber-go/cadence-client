package metrics

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
)

func Test_Counter(t *testing.T) {
	isReplay := true
	scope, closer, reporter := newMetricsScope(&isReplay)
	scope.Counter("test-name").Inc(1)
	closer.Close()
	require.Equal(t, 0, len(reporter.counts))

	isReplay = false
	scope, closer, reporter = newMetricsScope(&isReplay)
	scope.Counter("test-name").Inc(3)
	closer.Close()
	require.Equal(t, 1, len(reporter.counts))
	require.Equal(t, int64(3), reporter.counts[0].value)
}

func Test_Gauge(t *testing.T) {
	isReplay := true
	scope, closer, reporter := newMetricsScope(&isReplay)
	scope.Gauge("test-name").Update(1)
	closer.Close()
	require.Equal(t, 0, len(reporter.gauges))

	isReplay = false
	scope, closer, reporter = newMetricsScope(&isReplay)
	scope.Gauge("test-name").Update(3)
	closer.Close()
	require.Equal(t, 1, len(reporter.gauges))
	require.Equal(t, float64(3), reporter.gauges[0].value)
}

func Test_Timer(t *testing.T) {
	isReplay := true
	scope, closer, reporter := newMetricsScope(&isReplay)
	scope.Timer("test-name").Record(time.Second)
	sw := scope.Timer("test-stopwatch").Start()
	sw.Stop()
	closer.Close()
	require.Equal(t, 0, len(reporter.timers))

	isReplay = false
	scope, closer, reporter = newMetricsScope(&isReplay)
	scope.Timer("test-name").Record(time.Second)
	sw = scope.Timer("test-stopwatch").Start()
	sw.Stop()
	closer.Close()
	require.Equal(t, 2, len(reporter.timers))
	require.Equal(t, time.Second, reporter.timers[0].value)
}

func Test_Histogram(t *testing.T) {
	isReplay := true
	scope, closer, reporter := newMetricsScope(&isReplay)
	valueBuckets := tally.MustMakeLinearValueBuckets(0, 10, 10)
	scope.Histogram("test-hist-1", valueBuckets).RecordValue(5)
	scope.Histogram("test-hist-2", valueBuckets).RecordValue(15)
	closer.Close()
	require.Equal(t, 0, len(reporter.histogramValueSamples))
	require.Equal(t, 0, len(reporter.histogramDurationSamples))

	isReplay = false
	scope, closer, reporter = newMetricsScope(&isReplay)
	valueBuckets = tally.MustMakeLinearValueBuckets(0, 10, 10)
	scope.Histogram("test-hist-1", valueBuckets).RecordValue(5)
	scope.Histogram("test-hist-2", valueBuckets).RecordValue(15)
	closer.Close()
	require.Equal(t, 2, len(reporter.histogramValueSamples))

	scope, closer, reporter = newMetricsScope(&isReplay)
	durationBuckets := tally.MustMakeLinearDurationBuckets(0, time.Hour, 10)
	scope.Histogram("test-hist-1", durationBuckets).RecordDuration(time.Minute)
	scope.Histogram("test-hist-2", durationBuckets).RecordDuration(time.Minute * 61)
	sw := scope.Histogram("test-hist-3", durationBuckets).Start()
	sw.Stop()
	closer.Close()
	require.Equal(t, 3, len(reporter.histogramDurationSamples))
}

func newMetricsScope(isReplay *bool) (tally.Scope, io.Closer, *capturingStatsReporter) {
	reporter := &capturingStatsReporter{}
	opts := tally.ScopeOptions{Reporter: reporter}
	scope, closer := tally.NewRootScope(opts, time.Second)
	return WrapScope(isReplay, scope, &realClock{}), closer, reporter
}

type realClock struct {
}

func (c *realClock) Now() time.Time {
	return time.Now()
}

// capturingStatsReporter is a reporter used by tests to capture the metric so we can verify our tests.
type capturingStatsReporter struct {
	counts                   []capturedCount
	gauges                   []capturedGauge
	timers                   []capturedTimer
	histogramValueSamples    []capturedHistogramValueSamples
	histogramDurationSamples []capturedHistogramDurationSamples
	capabilities             int
	flush                    int
}

type capturedCount struct {
	name  string
	tags  map[string]string
	value int64
}

type capturedGauge struct {
	name  string
	tags  map[string]string
	value float64
}

type capturedTimer struct {
	name  string
	tags  map[string]string
	value time.Duration
}

type capturedHistogramValueSamples struct {
	name             string
	tags             map[string]string
	bucketLowerBound float64
	bucketUpperBound float64
	samples          int64
}

type capturedHistogramDurationSamples struct {
	name             string
	tags             map[string]string
	bucketLowerBound time.Duration
	bucketUpperBound time.Duration
	samples          int64
}

func (r *capturingStatsReporter) ReportCounter(
	name string,
	tags map[string]string,
	value int64,
) {
	r.counts = append(r.counts, capturedCount{name, tags, value})
}

func (r *capturingStatsReporter) ReportGauge(
	name string,
	tags map[string]string,
	value float64,
) {
	r.gauges = append(r.gauges, capturedGauge{name, tags, value})
}

func (r *capturingStatsReporter) ReportTimer(
	name string,
	tags map[string]string,
	value time.Duration,
) {
	r.timers = append(r.timers, capturedTimer{name, tags, value})
}

func (r *capturingStatsReporter) ReportHistogramValueSamples(
	name string,
	tags map[string]string,
	buckets tally.Buckets,
	bucketLowerBound,
	bucketUpperBound float64,
	samples int64,
) {
	elem := capturedHistogramValueSamples{name, tags,
		bucketLowerBound, bucketUpperBound, samples}
	r.histogramValueSamples = append(r.histogramValueSamples, elem)
}

func (r *capturingStatsReporter) ReportHistogramDurationSamples(
	name string,
	tags map[string]string,
	buckets tally.Buckets,
	bucketLowerBound,
	bucketUpperBound time.Duration,
	samples int64,
) {
	elem := capturedHistogramDurationSamples{name, tags,
		bucketLowerBound, bucketUpperBound, samples}
	r.histogramDurationSamples = append(r.histogramDurationSamples, elem)
}

func (r *capturingStatsReporter) Capabilities() tally.Capabilities {
	r.capabilities++
	return r
}

func (r *capturingStatsReporter) Reporting() bool {
	return true
}

func (r *capturingStatsReporter) Tagging() bool {
	return true
}

func (r *capturingStatsReporter) Flush() {
	r.flush++
}
