package httpclient

import (
	"log"
	"net/http"

	"github.com/kelseyhightower/envconfig"
)

func newConfig() *config {
	cfg := new(config)
	err := envconfig.Process("", cfg)
	if err != nil {
		log.Fatalf("failed to load configuration: %q", err)
	}

	return cfg
}

func newHTTPClient(transports ...roundTripper) *http.Client {
	client := &http.Client{Transport: http.DefaultTransport, Timeout: defaultTimeout}

	for _, transport := range transports {
		transport.use(client)
	}

	return client
}
