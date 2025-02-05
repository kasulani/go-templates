package rest

import (
	"net/http"
)

// Endpoint defines a rest handler method.
type Endpoint interface {
	Handler() http.HandlerFunc
}
