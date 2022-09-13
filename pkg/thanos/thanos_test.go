package thanos

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoPartialResponseRoundTripper(t *testing.T) {
	rt := NoPartialResponseRoundTripper{roundTripFunc(func(r *http.Request) (*http.Response, error) {
		require.Regexp(t, `partial_response=false`, r.URL.RawQuery)
		return nil, errors.New("not implemented")
	})}

	_, _ = rt.RoundTrip(httptest.NewRequest("GET", "https://thanos.io", nil))
	_, _ = rt.RoundTrip(httptest.NewRequest("GET", "https://thanos.io?testly=blub", nil))
}

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (s roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return s(r)
}
