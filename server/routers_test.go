package server_test

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server"
	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCORS(t *testing.T) {
	app := test.App()
	domain := app.Config.ApplicationDomains[0]
	server := httptest.NewServer(server.Router(app))
	defer server.Close()

	client := route.NewClient(server.URL)
	res, err := client.Preflight(&domain, "PATCH", "/path")
	require.NoError(t, err)

	scheme := "http"
	if domain.Port == "443" {
		scheme = "https"
	}
	origin := fmt.Sprintf("%s://%s", scheme, domain.String())

	fmt.Println(res.Header)

	assert.Equal(t, "true", res.Header.Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "PATCH", res.Header.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, origin, res.Header.Get("Access-Control-Allow-Origin"))
}
