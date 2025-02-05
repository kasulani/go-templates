package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/spf13/cast"
)

type (
	httpRequest struct {
		host        string
		method      string
		payload     interface{}
		headers     map[string]string
		queryParams map[string]string
	}

	httpResponse struct {
		Body       []byte
		StatusCode int
	}

	config struct {
		Environment   string `envconfig:"ENVIRONMENT" required:"true"`
		ExampleHost   string `envconfig:"EXAMPLE_HOST" default:"http://mock-server:8080"`
		ExampleAPIKey string `envconfig:"EXAMPLE_API_KEY"`
	}
)

const (
	urlEncodedContentType = "application/x-www-form-urlencoded"
	jsonContentType       = "application/json"
	contentTypeHeader     = "Content-Type"
)

func (r *httpRequest) make(ctx context.Context, client *http.Client) (*httpResponse, error) {
	var req *http.Request
	var err error

	switch r.headers[contentTypeHeader] {
	case urlEncodedContentType:
		req, err = r.prepareFormURLEncodedRequest(ctx)
	case jsonContentType:
		req, err = r.prepareJSONEncodedRequest(ctx)
	default:
		return &httpResponse{nil, 0}, errors.New("invalid content type")
	}

	if err != nil {
		return &httpResponse{nil, 0}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return &httpResponse{nil, 0}, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &httpResponse{nil, resp.StatusCode}, err
	}

	return &httpResponse{body, resp.StatusCode}, nil
}

func (r *httpRequest) prepareJSONEncodedRequest(ctx context.Context) (*http.Request, error) {
	var body []byte
	var err error

	if r.method == http.MethodPost || r.method == http.MethodPatch {
		body, err = json.Marshal(r.payload)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(
		ctx,
		r.method,
		r.host,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}

	for k, v := range r.headers {
		req.Header.Add(k, v)
	}

	query := req.URL.Query()
	for k, v := range r.queryParams {
		query.Add(k, v)
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}

func (r *httpRequest) prepareFormURLEncodedRequest(ctx context.Context) (*http.Request, error) {
	var err error

	req, err := http.NewRequestWithContext(
		ctx,
		r.method,
		r.host,
		strings.NewReader(cast.ToString(r.payload)),
	)
	if err != nil {
		return nil, err
	}

	for k, v := range r.headers {
		req.Header.Add(k, v)
	}

	query := req.URL.Query()
	for k, v := range r.queryParams {
		query.Add(k, v)
	}
	req.URL.RawQuery = query.Encode()

	return req, nil
}
