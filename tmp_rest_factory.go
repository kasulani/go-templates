package rest

import (
	"log"
	"net/http"

	"github.com/alexliesenfeld/health"
	"github.com/go-chi/chi"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"

    {{range .imports}}
	"{{.}}"
	{{- end}}
)

type (
	config struct {}
)

func newConfig() *config {
	cfg := new(config)
	err := envconfig.Process("", cfg)
	if err != nil {
		log.Fatalf("failed to load configuration: %q", err)
	}

	return cfg
}

func newAPIRequestLogger(logger *zap.Logger) apiRequestLogger {
	return apiRequestLogger{
		Log: logger,
	}
}

// NewAPIServer returns a new instance of APIServer.
func NewAPIServer(instrumentation *telemetry.Instrumentation) *APIServer {
	return &APIServer{
		Server:          &http.Server{ReadHeaderTimeout: timeout},
		router:          chi.NewMux(),
		config:          newConfig(),
		log:             logging.NewLogger(),
		instrumentation: instrumentation,
	}
}

// NewIndexEndpoint returns a new instance of indexEndpoint.
func NewIndexEndpoint() *IndexEndpoint {
	return &IndexEndpoint{log: logging.NewLogger()}
}

// NewDocsEndpoint returns a new instance of docsEndpoint.
func NewDocsEndpoint() *DocsEndpoint {
	return &DocsEndpoint{log: logging.NewLogger()}
}

// NewStatusEndpoint returns a new instance of statusEndpoint.
func NewStatusEndpoint(healthChecker health.Checker) *StatusEndpoint {
	return &StatusEndpoint{healthChecker: healthChecker}
}

// NewExampleEndpoint returns a new instance of ExampleEndpoint.
func NewExampleEndpoint(client *httpclient.ExampleClient) *ExampleEndpoint {
	return &ExampleEndpoint{
		log:    logging.NewLogger(),
		client: client,
	}
}
