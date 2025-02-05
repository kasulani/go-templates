package telemetry

type (
	// CollectorMetric defines the Apply method.
	CollectorMetric interface {
		Apply(collector *MetricsCollector, name string)
	}
)
