package rest

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
    {{range .imports}}
	"{{.}}"
	{{- end}}
)

type (
	// APIServer is REST API server.
	APIServer struct {
		*http.Server

		router          *chi.Mux
		config          *config
		log             *logging.Logger
		instrumentation *telemetry.Instrumentation
		collector       *telemetry.MetricsCollector
	}

	apiRequestLogger struct {
		Log *zap.Logger
	}

	responseWriter struct {
		http.ResponseWriter

		code int
		size int
	}
)

const (
	apiRequestID     = logging.ContextKey("apiRequestID")
	apiTraceID       = logging.ContextKey("apiTraceID")
	apiSpanID        = logging.ContextKey("apiSpanID")

	timeout = time.Second * 10
)

func (server *APIServer) haltOnSignal(ctx context.Context, signal <-chan os.Signal) {
	<-signal

	timeOutctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	server.log.Info("shutting down REST API server")

	if err := server.instrumentation.TraceProvider().Shutdown(timeOutctx); err != nil {
		server.log.Fatal("REST API server shutdown did not complete successfully", zap.Error(err))
	}

	if err := server.Shutdown(timeOutctx); err != nil {
		server.log.Fatal("REST API server shutdown did not complete successfully", zap.Error(err))
	}
}

func (server *APIServer) withMetrics(name string) *APIServer {
	server.collector = telemetry.NewMetricsCollector(
		name,
		telemetry.TotalOperations(),
		telemetry.TotalHTTPOperationsWithLabels(),
		telemetry.HTTPLatencyWithLabels(),
	)

	server.instrumentation.Registry().MustRegister(
		server.collector.Counter(),
		server.collector.CounterVec(),
		server.collector.LatencyVec(),
	)

	return server
}

// BootstrapAPIServer starts up the API server.
func BootstrapAPIServer(ctx context.Context, server *APIServer, address string) {
	server.Addr = address

	server.Handler = otelhttp.NewHandler(server.router, "")

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go server.haltOnSignal(ctx, stop)

	serverErr := make(chan error, 1)
	go func() {
		server.log.Debug(fmt.Sprintf("starting REST API server on port %s", server.Addr))
		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-stop:
		server.log.Info("graceful shutdown of REST API server")
		os.Exit(0)
	case err := <-serverErr:
		server.log.Error("error occurred while listening to http requests", zap.Error(err))
		os.Exit(1)
	}
}

// WriteHeader implements http.ResponseWriter interface.
func (r *responseWriter) WriteHeader(code int) {
	if r.Code() == 0 {
		r.code = code
		r.ResponseWriter.WriteHeader(code)
	}
}

// Write implements http.ResponseWriter interface.
func (r *responseWriter) Write(body []byte) (int, error) {
	if r.Code() == 0 {
		r.WriteHeader(http.StatusOK)
	}

	var err error
	r.size, err = r.ResponseWriter.Write(body)

	return r.size, err
}

// Flush implements http.ResponseWriter interface.
func (r *responseWriter) Flush() {
	if fl, ok := r.ResponseWriter.(http.Flusher); ok {
		if r.Code() == 0 {
			r.WriteHeader(http.StatusOK)
		}

		fl.Flush()
	}
}

// Hijack implements http.ResponseWriter interface.
func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the hijacker interface is not supported")
	}

	return hj.Hijack()
}

// Code implements http.ResponseWriter interface.
func (r *responseWriter) Code() int {
	return r.code
}

// Size implements http.ResponseWriter interface.
func (r *responseWriter) Size() int {
	return r.size
}

// NewLogEntry returns a new LogEntry.
func (logger apiRequestLogger) NewLogEntry(request *http.Request) middleware.LogEntry {
	ctx := request.Context()

	return apiRequestLogger{
		Log: logger.Log.With(
			zap.String("http_method", request.Method),
			zap.String("http_url", request.URL.Path),
			zap.String("user_agent", request.UserAgent()),
			zap.Any("request_id", ctx.Value(apiRequestID)),
			zap.Any("trace_id", ctx.Value(apiTraceID).(trace.TraceID)),
			zap.Any("span_id", ctx.Value(apiSpanID).(trace.SpanID)),
		),
	}
}

// Write will add additional fields to the apiRequestLogger.
func (logger apiRequestLogger) Write(status, _ int, _ http.Header, elapsed time.Duration, _ interface{}) {
	l := logger.Log.With(
		zap.Int("http_status", status),
		zap.Float64("response_time_ms", float64(elapsed.Nanoseconds())/1000000.0),
	)

	switch {
	case status >= 200 && status < 300:
		l.Debug("Successful request")
	case status < 400:
		l.Info("Redirected request")
	case status < 500:
		l.Info("Invalid client request")
	case status >= 500:
		l.Error("Internal service error")
	}
}

// Panic will make our apiRequestLogger panic.
func (logger apiRequestLogger) Panic(v interface{}, stack []byte) {
	logger.Log.Panic(fmt.Sprintf("%+v", v), zap.String("stack", string(stack)))
}
