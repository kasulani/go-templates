package httpclient

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"time"

	promClient "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	{{range .imports}}
	"{{.}}"
	{{- end}}
)

type (
	metricsRoundTripper struct {
		next      http.RoundTripper
		collector *telemetry.MetricsCollector
	}

	loggingRoundTripper struct {
		next   http.RoundTripper
		logger *logging.Logger
	}

	roundTripperFunc func(client *http.Client)
)

func useOTELRoundTripper() roundTripper {
	return roundTripperFunc(func(client *http.Client) {
		client.Transport = otelhttp.NewTransport(client.Transport)
	})
}

func useMetricsRoundTripper(name string, registry *promClient.Registry) roundTripper {
	return roundTripperFunc(func(client *http.Client) {
		collector := telemetry.NewMetricsCollector(
			name,
			telemetry.TotalOperations(),
			telemetry.TotalHTTPOperationsWithLabels(),
			telemetry.HTTPLatencyWithLabels(),
		)

		client.Transport = &metricsRoundTripper{
			client.Transport,
			collector,
		}

		registry.MustRegister(
			collector.Counter(),
			collector.CounterVec(),
			collector.LatencyVec(),
		)
	})
}

func useLoggerRoundTripper(name string, logger *logging.Logger) roundTripper {
	logger.Logger = logger.Logger.With(zap.String("client", name))

	return roundTripperFunc(func(client *http.Client) {
		client.Transport = &loggingRoundTripper{
			client.Transport,
			logger,
		}
	})
}

func (fn roundTripperFunc) use(client *http.Client) {
	fn(client)
}

func (recorder *metricsRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	recorder.collector.RecordTotalOpsMetric()

	startTime := time.Now()
	response, err := recorder.next.RoundTrip(request)

	recorder.collector.RecordHTTPLatencyMetricWithLabels(
		startTime,
		telemetry.NewHTTPMetricLabels(request.Method, request.URL.Host, response.StatusCode),
	)

	if err != nil {
		recorder.collector.RecordErrorMetric()
		return response, err
	}

	recorder.collector.RecordHTTPMetric(
		telemetry.NewHTTPMetricLabels(request.Method, request.URL.Host, response.StatusCode),
	)

	return response, err
}

func (tripper *loggingRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	var requestBody string
	if request.Method == http.MethodPost {
		reqBody, err := ioutil.ReadAll(request.Body)
		if err != nil {
			tripper.logger.Error("failed to read request body", zap.Error(err))
			return nil, err
		}

		requestBody = string(reqBody)

		_ = request.Body.Close()
		request.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
	}

	tripper.logger.Debug(
		"outgoing http request",
		zap.String("method", request.Method),
		zap.String("raw_query", request.URL.RawQuery),
		zap.String("path", request.URL.Path),
		zap.String("host", request.URL.Host),
		zap.String("request_body", logging.SanitizeSecrets(requestBody)),
	)

	response, err := tripper.next.RoundTrip(request)
	if err != nil {
		tripper.logger.Error("request failed with error", zap.Error(err))
		return response, err
	}

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		tripper.logger.Error("failed to read response", zap.Error(err))
		return response, err
	}

	_ = response.Body.Close()
	response.Body = ioutil.NopCloser(bytes.NewBuffer(respBody))

	tripper.logger.Debug(
		"response to outgoing http request",
		zap.String("response_body", logging.SanitizeSecrets(string(respBody))),
		zap.Int("status_code", response.StatusCode),
		zap.String("method", request.Method),
		zap.String("raw_query", request.URL.RawQuery),
		zap.String("path", request.URL.Path),
		zap.String("host", request.URL.Host),
		zap.String("request_body", logging.SanitizeSecrets(requestBody)),
	)

	return response, err
}
