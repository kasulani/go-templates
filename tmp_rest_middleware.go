package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
    {{range .imports}}
	"{{.}}"
	{{- end}}
)

func serverTracingMiddleware(log *logging.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			rw := &responseWriter{ResponseWriter: w}

			ctx, span := otel.Tracer("api-server").Start(r.Context(), r.URL.Path)
			defer span.End()

			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				id := uuid.Must(uuid.NewV4())
				requestID = id.String()
				r.Header.Add("X-Request-ID", requestID)
			}

			corID := uuid.Must(uuid.NewV4())

			ctx = context.WithValue(ctx, apiRequestID, requestID)
			ctx = context.WithValue(ctx, logging.CorrelationID, corID)
			ctx = context.WithValue(ctx, apiTraceID, span.SpanContext().TraceID())
			ctx = context.WithValue(ctx, apiSpanID, span.SpanContext().SpanID())

			span.SetName(r.URL.Path)
			span.SetAttributes(
				attribute.String("request_id", requestID),
				attribute.String("correlation_id", corID.String()),
			)

			next.ServeHTTP(rw, r.WithContext(ctx))

			if rw.Code() >= http.StatusBadRequest {
				err := errors.New(fmt.Sprintf("unsuccessful api request: %d", rw.Code()))

				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
		}

		log.Debug("use server tracing middleware")
		return http.HandlerFunc(fn)
	}
}

func serverMetricsMiddleware(
	collector *telemetry.MetricsCollector,
	log *logging.Logger,
) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			rw := &responseWriter{ResponseWriter: w}
			collector.RecordTotalOpsMetric()

			startTime := time.Now()

			next.ServeHTTP(rw, r)

			collector.RecordHTTPLatencyMetricWithLabels(
				startTime,
				telemetry.NewHTTPMetricLabels(r.Method, r.URL.Path, rw.Code()),
			)

			collector.RecordHTTPMetric(
				telemetry.NewHTTPMetricLabels(r.Method, r.URL.Path, rw.Code()),
			)
		}

		log.Debug("use server metrics middleware")
		return http.HandlerFunc(fn)
	}
}

func serverAuthUserMiddleware(log *logging.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			rw := &responseWriter{ResponseWriter: w}

			next.ServeHTTP(rw, r.WithContext(ctx))
		}

		log.Debug("use server auth middleware")
		return http.HandlerFunc(fn)
	}
}
