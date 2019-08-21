package route_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keratin/authn-server/lib/route"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const successBody = "success"

func TestOriginSecurity(t *testing.T) {
	testCases := []struct {
		domain      string
		goodOrigins []string
		badOrigins  []string
	}{
		{"example.com", []string{"http://example.com", "http://example.com:3000"}, nil},
		{"www.example.com", []string{"http://www.example.com"}, []string{"http://example.com"}},
		{"example.com:3000", []string{"http://example.com:3000"}, []string{"http://example.com:8080"}},
		{"example.com:80", []string{"http://example.com"}, []string{"https://example.com", "http://example.com:3000"}},
		{"example.com:443", []string{"https://example.com"}, []string{"http://example.com", "https://example.com:3000"}},
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(successBody))
	})

	for _, tc := range testCases {
		t.Run(tc.domain, func(t *testing.T) {
			adapter := route.OriginSecurity([]route.Domain{route.ParseDomain(tc.domain)}, logrus.New())

			server := httptest.NewServer(adapter(nextHandler))
			defer server.Close()
			// Undefined - referer and origin missing
			assertOriginDenied(t, server, "", "")

			for _, goodOrigin := range tc.goodOrigins {
				// Cross-site case
				assertOriginAccepted(t, server, goodOrigin, goodOrigin)
				assertOriginAccepted(t, server, goodOrigin, "")
				// Same-origin case
				assertOriginAccepted(t, server, "", goodOrigin)

				for _, badOrigin := range tc.badOrigins {
					// Shouldn't happen, but origin takes precedent and referer is ignored
					assertOriginAccepted(t, server, goodOrigin, badOrigin)
					// Origin takes precedent so fails
					assertOriginDenied(t, server, badOrigin, goodOrigin)
				}
			}
			for _, badOrigin := range tc.badOrigins {
				assertOriginDenied(t, server, badOrigin, "")
				assertOriginDenied(t, server, badOrigin, badOrigin)
			}
		})
	}
}

func assertOriginAccepted(t *testing.T, server *httptest.Server, origin, referer string) {
	res := originDo(t, server, origin, referer)
	assert.Equal(t, successBody, readBody(t, res))
	assert.Equal(t, res.StatusCode, http.StatusOK)
}

func assertOriginDenied(t *testing.T, server *httptest.Server, origin, referer string) {
	res := originDo(t, server, origin, referer)
	assert.Equal(t, "Origin is not a trusted host.", readBody(t, res))
	assert.Equal(t, res.StatusCode, http.StatusForbidden)
}

func originDo(t *testing.T, server *httptest.Server, origin, referer string) *http.Response {
	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)
	req.Header.Add("Origin", origin)
	req.Header.Add("Referer", referer)
	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return res
}

func readBody(t *testing.T, res *http.Response) string {
	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)
	err = res.Body.Close()
	require.NoError(t, err)
	return string(body)
}
