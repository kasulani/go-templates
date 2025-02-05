package httpclient

import "net/http"

type roundTripper interface {
	use(client *http.Client)
}
