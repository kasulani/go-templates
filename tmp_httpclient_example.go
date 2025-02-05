package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	{{range .imports}}
	"{{.}}"
	{{- end}}
)

type (
	// ExampleClient is a http client
	ExampleClient struct {
		*http.Client

		log    *logging.Logger
		config *config
	}

	examplePayload struct {
		EventType string                 `json:"eventType"`
		Payload   map[string]interface{} `json:"payload"`
	}
)

const (
	defaultTimeout  = 30 * time.Second
)

// NewExampleClient creates a new ExampleClient.
func NewExampleClient(instrumentation *telemetry.Instrumentation) *ExampleClient {
	name := "example.client"
	logger := logging.NewLogger()
	cfg := newConfig()

	return &ExampleClient{
		Client: newHTTPClient(
			useOTELRoundTripper(),
			useMetricsRoundTripper(name, instrumentation.Registry()),
			useLoggerRoundTripper(name, logger),
		),
		config: cfg,
		log:    logger,
	}
}

// ExternalRequest sends a request to an external service.
func (client *ExampleClient) ExternalRequest(ctx context.Context) error {
	payload := examplePayload{
		Payload: map[string]interface{}{
			"key": "value",
		},
	}

	token := "Bearer " + client.config.ExampleAPIKey

	headers := map[string]string{
		"Content-Type":  jsonContentType,
		"Authorization": token,
	}

	client.log.Debug(
		"make http request to create message",
		zap.Any("payload", payload),
		zap.Any("headers", headers),
	)

	req := &httpRequest{
		host:    client.config.ExampleHost + "/api/v1",
		payload: payload,
		headers: headers,
		method:  http.MethodPost,
	}

	client.log.Debug("sending http request to external service")

	response, err := req.make(ctx, client.Client)
	if err != nil {
		client.log.Error(
			"http request to create message failed",
			zap.Error(err),
			zap.String("responseBody", string(response.Body)),
			zap.Int("responseStatusCode", response.StatusCode),
		)
		return err
	}

	client.log.Debug(
		"http response from external service",
		zap.String("responseBody", string(response.Body)),
		zap.Int("responseStatusCode", response.StatusCode),
	)

	if response.StatusCode >= http.StatusIMUsed {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	client.log.Debug("success")

	return nil
}
