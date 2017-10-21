package route_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keratin/authn-server/lib/route"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOriginSecurity(t *testing.T) {
	readBody := func(res *http.Response) []byte {
		body, err := ioutil.ReadAll(res.Body)
		require.NoError(t, err)
		res.Body.Close()
		return body
	}

	testCases := []struct {
		domain  string
		origin  string
		success bool
	}{
		{"example.com", "http://example.com", true},
		{"example.com", "http://example.com:3000", true},
		{"www.example.com", "http://www.example.com", true},
		{"www.example.com", "http://example.com", false},
		{"example.com:3000", "http://example.com:3000", true},
		{"example.com:3000", "http://example.com:8080", false},
		{"example.com:80", "http://example.com", true},
		{"example.com:80", "https://example.com", false},
		{"example.com:80", "http://example.com:3000", false},
		{"example.com:443", "https://example.com", true},
		{"example.com:443", "http://example.com", false},
		{"example.com:443", "https://example.com:3000", false},
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("success"))
	})

	for _, tc := range testCases {
		t.Run(tc.domain, func(t *testing.T) {
			adapter := route.OriginSecurity([]route.Domain{route.ParseDomain(tc.domain)})

			server := httptest.NewServer(adapter(nextHandler))
			defer server.Close()

			req, err := http.NewRequest("GET", server.URL, nil)
			require.NoError(t, err)
			req.Header.Add("Origin", tc.origin)
			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if tc.success {
				assert.Equal(t, string(readBody(res)), "success")
			} else {
				assert.Equal(t, http.StatusForbidden, res.StatusCode)
			}
		})
	}
}
