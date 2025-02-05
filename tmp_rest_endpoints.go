package rest

import (
	"net/http"
	"os"

	"github.com/alexliesenfeld/health"
	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/nicklaw5/go-respond"
	"go.uber.org/zap"
	{{range .imports}}
	"{{.}}"
	{{- end}}
)

type (
	serverResponse map[string]interface{}

	// IndexEndpoint represents the index endpoint.
	IndexEndpoint struct{ log *logging.Logger }

	// DocsEndpoint represents the docs endpoint where the API documentation is served.
	DocsEndpoint struct{ log *logging.Logger }

	// StatusEndpoint represents the status endpoint.
	StatusEndpoint struct{ healthChecker health.Checker }

	// ExampleEndpoint represents the example endpoint.
	ExampleEndpoint struct {
		log    *logging.Logger
		client *httpclient.ExampleClient
	}
)

const (
	docsPath = "./docs"
)

// RegisterAPIEndpoints registers all REST API endpoints.
func RegisterAPIEndpoints(server *APIServer, endpoints []Endpoint) {
	server = server.withMetrics("{{.serviceName}}")

	server.log.Debug("registering all REST API endpoints")

	server.router.Use(
		chiMiddleware.StripSlashes,
		chiMiddleware.Recoverer,
	)

	server.router.Group(func(r chi.Router) {
		r.Use(
			serverAuthUserMiddleware(server.log),
			serverTracingMiddleware(server.log),
			serverMetricsMiddleware(server.collector, server.log),
			chiMiddleware.RequestLogger(newAPIRequestLogger(server.log.Logger)),
		)

		for _, endpoint := range endpoints {
			switch endpoint.(type) {
			case *ExampleEndpoint:
				r.Get("/example-endpoint", endpoint.Handler())
			}
		}
	})

	registerPublicEndpoints(server, endpoints)
}

func registerPublicEndpoints(server *APIServer, endpoints []Endpoint) {
	server.router.Group(func(r chi.Router) {
		for _, endpoint := range endpoints {
			switch endpoint.(type) {
			case *IndexEndpoint:
				r.Get("/", endpoint.Handler())
			case *DocsEndpoint:
				r.Handle("/docs/*", endpoint.Handler())
			case *StatusEndpoint:
				r.Get("/status", endpoint.Handler())
			}
		}
		r.Handle("/metrics", server.instrumentation.Endpoint())
	})
}

// Handler returns the handler function for the index endpoint.
func (handler *IndexEndpoint) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler.log.Debug("index endpoint hit")
		respond.NewResponse(w).Ok(serverResponse{"{{.serviceName}}": "online"})
	}
}

// Handler returns the handler function for the docs endpoint.
func (handler *DocsEndpoint) Handler() http.HandlerFunc {
	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		handler.log.Error("failed to fetch api docs", zap.Error(err))

		return func(w http.ResponseWriter, r *http.Request) {
			respond.NewResponse(w).InternalServerError(map[string]string{"error": err.Error()})
		}
	}

	fs := http.FileServer(http.Dir(docsPath))
	return http.StripPrefix("/docs", fs).ServeHTTP
}

// Handler returns the handler function for the status endpoint.
func (handler *StatusEndpoint) Handler() http.HandlerFunc {
	return health.NewHandler(handler.healthChecker)
}

// Handler returns the handler function for the example endpoint.
func (handler *ExampleEndpoint) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler.log.Debug("example endpoint hit")

		// call to external service should happen via the domain, like using a use case or domain service
		if err := handler.client.ExternalRequest(r.Context()); err != nil {
			handler.log.Error("failed to make external request", zap.Error(err))
			respond.NewResponse(w).InternalServerError(serverResponse{"error": err.Error()})
			return
		}

		respond.NewResponse(w).Ok(serverResponse{"message": "example endpoint hit"})
	}
}

