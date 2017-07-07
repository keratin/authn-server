package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/api/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicAuthSecurity(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("success"))
	})

	adapter := api.BasicAuthSecurity("user", "pass", "authn-server tests")
	server := httptest.NewServer(adapter(nextHandler))
	defer server.Close()

	testCases := []struct {
		username string
		password string
		success  bool
	}{
		{"user", "pass", true},
		{"user", "unknown", false},
		{"unknown", "pass", false},
		{"USER", "PASS", false},
	}

	for _, tc := range testCases {
		req, err := http.NewRequest("GET", server.URL, nil)
		require.NoError(t, err)
		req.SetBasicAuth(tc.username, tc.password)

		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		if tc.success {
			assert.Equal(t, string(test.ReadBody(res)), "success")
		} else {
			assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
		}
	}
}
