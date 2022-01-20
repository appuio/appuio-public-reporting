package thanos

import (
	"net/http"

	promapi "github.com/prometheus/client_golang/api"
)

type ThanosRoundTripper struct {
	http.RoundTripper
}

func (t *ThanosRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt := t.RoundTripper
	if rt == nil {
		rt = promapi.DefaultRoundTripper
	}

	q := req.URL.Query()
	q.Set("dedup", "true")
	req.URL.RawQuery = q.Encode()

	return rt.RoundTrip(req)
}
