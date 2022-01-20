package thanos_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/appuio/appuio-cloud-reporting/pkg/thanos"
	"github.com/stretchr/testify/require"
)

func TestThanosRoundTripperAddsDedupURLParameter(t *testing.T) {
	subject := thanos.ThanosRoundTripper{&testRT{T: t}}
	subject.RoundTrip(httptest.NewRequest("GET", "/query", nil))
	require.True(t, subject.RoundTripper.(*testRT).called)
}

type testRT struct {
	*testing.T
	called bool
}

func (t *testRT) RoundTrip(req *http.Request) (*http.Response, error) {
	t.called = true
	require.Equal(t, req.URL.Query().Get("dedup"), "true")
	return nil, nil
}
