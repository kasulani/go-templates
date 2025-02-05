package telemetry

import (
	"log"
	"strings"

	"github.com/kelseyhightower/envconfig"
	promClient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

const (
	serviceName            = "{{.serviceName}}"
	developmentEnvironment = "development"
	stagingEnvironment     = "staging"
	productionEnvironment  = "production"
)

// NewMetricsCollector returns an instance of MetricsCollector.
func NewMetricsCollector(name string, metrics ...CollectorMetric) *MetricsCollector {
	if name == "" {
		log.Fatal("metrics collector has no name")
	}

	name = strings.Replace(name, ".", "_", strings.Count(name, "."))
	collector := &MetricsCollector{tag: name}

	if len(metrics) == 0 {
		metrics = defaultCollectorMetrics
	}

	for _, opt := range metrics {
		opt.Apply(collector, name)
	}

	return collector
}

func newMetricsExporter(registry *promClient.Registry) *prometheus.Exporter {
	exporter, err := prometheus.New(
		prometheus.WithRegisterer(registry),
		prometheus.WithoutScopeInfo(),
		prometheus.WithoutUnits(),
	)
	if err != nil {
		log.Fatalf("failed to create prometheus exporter: %q", err)
	}

	customizedResource, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(serviceName)),
	)

	defaultHistogramBoundaries := []float64{
		1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200,
		250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000,
	}
	bytesHistogramBoundaries := []float64{
		1024, 2048, 4096, 16384, 65536, 262144, 1048576,
		4194304, 16777216, 67108864, 268435456, 1073741824, 4294967296,
	}

	otel.SetMeterProvider(
		metric.NewMeterProvider(
			metric.WithReader(exporter),
			metric.WithView(metric.NewView(
				metric.Instrument{Name: "http.server.duration*"},
				metric.Stream{
					Aggregation: metric.AggregationExplicitBucketHistogram{
						Boundaries: defaultHistogramBoundaries,
					},
					AttributeFilter: httpMetricsAttributeFilter(),
				},
			)),
			metric.WithView(metric.NewView(
				metric.Instrument{Name: "http.server.response_count*"},
				metric.Stream{},
			)),
			metric.WithView(metric.NewView(
				metric.Instrument{Name: "http.server*"},
				metric.Stream{
					AttributeFilter: httpMetricsAttributeFilter(),
				},
			)),
			metric.WithView(metric.NewView(
				metric.Instrument{Name: "*bytes", Kind: metric.InstrumentKindHistogram},
				metric.Stream{
					Aggregation: metric.AggregationExplicitBucketHistogram{
						Boundaries: bytesHistogramBoundaries,
					},
				},
			)),
			metric.WithView(metric.NewView(
				metric.Instrument{Name: "*", Kind: metric.InstrumentKindHistogram},
				metric.Stream{
					Aggregation: metric.AggregationExplicitBucketHistogram{
						Boundaries: defaultHistogramBoundaries,
					},
				},
			)),
			metric.WithView(metric.NewView(
				metric.Instrument{Name: "*", Kind: metric.InstrumentKindCounter},
				metric.Stream{
					Unit: "",
				},
			)),
			metric.WithResource(customizedResource),
		),
	)

	return exporter
}

func httpMetricsAttributeFilter() func(kv attribute.KeyValue) bool {
	return attribute.NewAllowKeysFilter(
		"http.flavor",
		"http.method",
		"http.scheme",
	)
}

// NewMetricsRegistry returns an instance of promClient.Registry.
func NewMetricsRegistry() *promClient.Registry {
	registry := promClient.NewRegistry()

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return registry
}

func newConfig() *config {
	cfg := new(config)
	err := envconfig.Process("", cfg)
	if err != nil {
		log.Fatalf("failed to load configuration: %q", err)
	}

	return cfg
}

// NewInstrumentation returns an instance of Instrumentation.
func NewInstrumentation(registry *promClient.Registry) *Instrumentation {
	config := newConfig()

	traceResource, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)

	var (
		traceExporter sdkTrace.SpanExporter
		traceProvider *sdkTrace.TracerProvider
	)

	switch config.Environment {
	case developmentEnvironment, stagingEnvironment, productionEnvironment:
		traceExporter, _ = jaeger.New(
			jaeger.WithAgentEndpoint(
				jaeger.WithAgentHost(config.Jager.Host),
				jaeger.WithAgentPort(config.Jager.Port),
			),
		)
		traceProvider = sdkTrace.NewTracerProvider(
			sdkTrace.WithBatcher(traceExporter),
			sdkTrace.WithResource(traceResource),
			sdkTrace.WithSampler(sdkTrace.TraceIDRatioBased(config.Jager.SampleRate)),
		)
	default:
		log.Fatalf("intrustmentation: unknown environment %s", config.Environment)
	}

	otel.SetTracerProvider(traceProvider)

	return &Instrumentation{
		serviceName:   serviceName,
		exporter:      newMetricsExporter(registry),
		registry:      registry,
		traceProvider: traceProvider,
	}
}

// NewHTTPMetricLabels returns an instance of HTTPMetricLabels.
func NewHTTPMetricLabels(method, url string, code int) *HTTPMetricLabels {
	return &HTTPMetricLabels{
		method: method,
		url:    url,
		status: code,
	}
}

/*

type ExampleCustomMetricsCollector struct { counterVec *promClient.CounterVec }

func NewExampleCustomMetricsCollector() *ExampleCustomMetricsCollector {
	collector := &ExampleCustomMetricsCollector{
		counterVec: promauto.NewCounterVec(
			promClient.CounterOpts{
				Name: "some_business_metric",
				Help: "Some business metric",
			},
			[]string{"tag1", "tag2"},
		),
	}

	return collector
}*/
