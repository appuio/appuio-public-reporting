package thanos

import "net/http"

// NoPartialResponseRoundTripper adds a new RoundTripper to the chain that sets the partial_response query parameter to false.
type NoPartialResponseRoundTripper struct {
	http.RoundTripper
}

// RoundTrip implements the RoundTripper interface.
func (t *NoPartialResponseRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	q := req.URL.Query()
	q.Set("partial_response", "false")
	req.URL.RawQuery = q.Encode()
	return t.RoundTripper.RoundTrip(req)
}
