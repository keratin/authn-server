package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keratin/authn/handlers"
	"github.com/keratin/authn/services"
)

func TestRefererSecurity(t *testing.T) {
	testTable := []struct {
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

	for _, tt := range testTable {
		adapter := handlers.RefererSecurity([]string{tt.domain})
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", strings.NewReader(""))
		req.Header.Add("Referer", tt.referer)
		if tt.success {
			adapter(expectedSuccessHandler).ServeHTTP(res, req)
			if res.Body.String() != "success" {
				t.Errorf("expected success with %v", tt)
			}
		} else {
			adapter(expectedFailureHandler).ServeHTTP(res, req)
			assertErrors(t, res, []services.Error{{"referer", "is not a trusted host"}})
		}
	}
}
