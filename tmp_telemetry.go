package telemetry

import (
	"fmt"
	"net/http"
	"time"

	promClient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cast"
	"go.opentelemetry.io/otel/exporters/prometheus"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
)

type (
	// MetricsCollector defines the metrics that will be collected.
	MetricsCollector struct {
		tag        string
		counter    promClient.Counter
		counterVec *promClient.CounterVec
		latencyVec *promClient.HistogramVec
	}

	// HTTPMetricLabels defines the fields in an HTTP metric that is collected.
	HTTPMetricLabels struct {
		method string
		url    string
		status int
	}

	collectorMetricFunc func(collector *MetricsCollector, name string)

	// Instrumentation is used to instrument the application.
	Instrumentation struct {
		serviceName   string
		exporter      *prometheus.Exporter
		registry      *promClient.Registry
		traceProvider *sdkTrace.TracerProvider
	}

	config struct {
		Environment string `envconfig:"ENVIRONMENT" required:"true"`
		Jager       *jaegerConfig
	}

	jaegerConfig struct {
		Host       string  `envconfig:"JAEGER_AGENT_HOST" required:"true"`
		Port       string  `envconfig:"JAEGER_AGENT_PORT" required:"true"`
		SampleRate float64 `envconfig:"JAEGER_SAMPLE_RATE" required:"true"`
	}
)

var histogramBuckets = []float64{
	1, 2, 5, 10, 25, 50, 100, 250, 500, 1000, 2500,
	5000, 10000, 15000, 20000, 25000, 50000, 100000,
}

func (instrumentation *Instrumentation) ServiceName() string {
	return instrumentation.serviceName
}

func (instrumentation *Instrumentation) TraceProvider() *sdkTrace.TracerProvider {
	return instrumentation.traceProvider
}

func (instrumentation *Instrumentation) Registry() *promClient.Registry {
	return instrumentation.registry
}

func (collector *MetricsCollector) Counter() promClient.Counter {
	return collector.counter
}

func (collector *MetricsCollector) CounterVec() *promClient.CounterVec {
	return collector.counterVec
}

func (collector *MetricsCollector) LatencyVec() *promClient.HistogramVec {
	return collector.latencyVec
}

func (fn collectorMetricFunc) Apply(collector *MetricsCollector, name string) {
	fn(collector, name)
}

var defaultCollectorMetrics = []CollectorMetric{
	TotalOperations(),
	totalOperationsWithLabels(),
	Latency(),
}

func TotalOperations() CollectorMetric {
	return collectorMetricFunc(func(collector *MetricsCollector, name string) {
		collector.counter = promauto.NewCounter(promClient.CounterOpts{
			Name: fmt.Sprintf("%s_processed_ops_total", name),
			Help: "The total number of processed operations",
		})
	})
}

func totalOperationsWithLabels() CollectorMetric {
	return collectorMetricFunc(func(collector *MetricsCollector, name string) {
		collector.counterVec = promauto.NewCounterVec(
			promClient.CounterOpts{
				Name: fmt.Sprintf("%s_processed_ops_count", name),
				Help: "The total number of processed operations grouped by labels",
			},
			[]string{"status", "tag"},
		)
	})
}

func Latency() CollectorMetric {
	return collectorMetricFunc(func(collector *MetricsCollector, name string) {
		collector.latencyVec = promauto.NewHistogramVec(
			promClient.HistogramOpts{
				Name:    fmt.Sprintf("%s_processed_ops_latency", name),
				Help:    "The total latency in milliseconds of processed operations grouped in buckets",
				Buckets: histogramBuckets,
			},
			[]string{"tag"},
		)
	})
}

func LatencyWithLabels() CollectorMetric {
	return collectorMetricFunc(func(collector *MetricsCollector, name string) {
		collector.latencyVec = promauto.NewHistogramVec(
			promClient.HistogramOpts{
				Name:    fmt.Sprintf("%s_processed_ops_latency", name),
				Help:    "The total latency in milliseconds of processed operations grouped in buckets with labels",
				Buckets: histogramBuckets,
			},
			[]string{"method", "tag"},
		)
	})
}

func HTTPLatencyWithLabels() CollectorMetric {
	return collectorMetricFunc(func(collector *MetricsCollector, name string) {
		collector.latencyVec = promauto.NewHistogramVec(
			promClient.HistogramOpts{
				Name:    fmt.Sprintf("%s_processed_ops_http_latency", name),
				Help:    "The total HTTP latency in milliseconds of processed operations grouped in buckets with labels",
				Buckets: histogramBuckets,
			},
			[]string{"method", "url", "status", "tag"},
		)
	})
}

func TotalHTTPOperationsWithLabels() CollectorMetric {
	return collectorMetricFunc(func(collector *MetricsCollector, name string) {
		collector.counterVec = promauto.NewCounterVec(
			promClient.CounterOpts{
				Name: fmt.Sprintf("%s_processed_ops_count", name),
				Help: "The total number of http operations grouped by labels",
			},
			[]string{"method", "url", "status", "tag"},
		)
	})
}

func (collector *MetricsCollector) RecordLatencyMetric(startTime time.Time) {
	collector.latencyVec.With(
		promClient.Labels{
			"tag": collector.tag,
		},
	).Observe(float64(time.Since(startTime).Milliseconds()))
}

func (collector *MetricsCollector) RecordLatencyMetricWithLabels(startTime time.Time, method string) {
	collector.latencyVec.With(
		promClient.Labels{
			"tag":    collector.tag,
			"method": method,
		},
	).Observe(float64(time.Since(startTime).Milliseconds()))
}

func (collector *MetricsCollector) RecordHTTPLatencyMetricWithLabels(startTime time.Time, params *HTTPMetricLabels) {
	collector.latencyVec.With(
		promClient.Labels{
			"tag":    collector.tag,
			"status": cast.ToString(params.status),
			"method": params.method,
			"url":    params.url,
		},
	).Observe(float64(time.Since(startTime).Milliseconds()))
}

func (collector *MetricsCollector) RecordErrorMetric() {
	collector.counterVec.With(
		promClient.Labels{
			"tag":    collector.tag,
			"status": "error",
		},
	).Inc()
}

func (collector *MetricsCollector) RecordSuccessMetric() {
	collector.counterVec.With(
		promClient.Labels{
			"tag":    collector.tag,
			"status": "success",
		},
	).Inc()
}

func (collector *MetricsCollector) RecordHTTPMetric(params *HTTPMetricLabels) {
	collector.counterVec.With(
		promClient.Labels{
			"tag":    collector.tag,
			"status": cast.ToString(params.status),
			"method": params.method,
			"url":    params.url,
		},
	).Inc()
}

func (collector *MetricsCollector) RecordTotalOpsMetric() {
	collector.counter.Inc()
}

func (instrumentation *Instrumentation) Endpoint() http.Handler {
	registry := instrumentation.Registry()
	handlerOptions := promhttp.HandlerOpts{Registry: registry}

	return promhttp.InstrumentMetricHandler(
		registry,
		promhttp.HandlerFor(registry, handlerOptions),
	)
}
