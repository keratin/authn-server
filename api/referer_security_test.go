package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/api/test"
	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
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

	expectedSuccessHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("success"))
	})
	expectedFailureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("failure"))
	})

	for _, tc := range testCases {
		adapter := api.RefererSecurity([]string{tc.domain})
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", strings.NewReader(""))
		req.Header.Add("Referer", tc.referer)
		if tc.success {
			adapter(expectedSuccessHandler).ServeHTTP(res, req)
			assert.Equal(t, res.Body.String(), "success")
		} else {
			adapter(expectedFailureHandler).ServeHTTP(res, req)
			test.AssertErrors(t, res, []services.Error{{"referer", "is not a trusted host"}})
		}
	}
}
