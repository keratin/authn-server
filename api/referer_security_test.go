package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/api/test"
	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefererSecurity(t *testing.T) {
	testCases := []struct {
		domain  string
		referer string
		success bool
	}{
		{"example.com", "http://example.com", true},
		{"www.example.com", "http://www.example.com", true},
		{"example.com:8080", "http://example.com:8080", true},
		{"www.example.com", "http://example.com", false},
		{"example.com", "http://example.com:8080", false},
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("success"))
	})

	for _, tc := range testCases {
		adapter := api.RefererSecurity([]string{tc.domain})

		server := httptest.NewServer(adapter(nextHandler))
		defer server.Close()

		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)
		req.Header.Add("Referer", tc.referer)
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if tc.success {
			assert.Equal(t, string(test.ReadBody(res)), "success")
		} else {
			test.AssertErrors(t, res, services.FieldErrors{{"referer", "is not a trusted host"}})
		}
	}
}
