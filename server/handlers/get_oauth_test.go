package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/keratin/authn-server/server/test"
	oauthlib "github.com/keratin/authn-server/lib/oauth"
	"github.com/keratin/authn-server/lib/route"
	"github.com/stretchr/testify/require"
)

func TestGetOauth(t *testing.T) {
	// start a fake oauth provider
	providerServer := httptest.NewServer(test.ProviderApp())
	defer providerServer.Close()

	// configure a client for the fake oauth provider
	providerClient := oauthlib.NewTestProvider(providerServer)

	// configure and start the authn test server
	app := test.App()
	app.OauthProviders["test"] = *providerClient
	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0])

	t.Run("when provider is configured", func(t *testing.T) {
		res, err := client.Get("/oauth/test?redirect_uri=http://test.com/finish")
		require.NoError(t, err)
		assert.Equal(t, http.StatusSeeOther, res.StatusCode)
		assert.NotNil(t, test.ReadCookie(res.Cookies(), app.Config.OAuthCookieName))

		location, err := res.Location()
		require.NoError(t, err)
		require.NotEmpty(t, location.Query().Get("state"))
		require.Equal(t, "https://authn.example.com/oauth/test/return", location.Query().Get("redirect_uri"))
	})

	t.Run("unknown provider", func(t *testing.T) {
		res, err := client.Get("/oauth/unknown")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("unknown redirect domain", func(t *testing.T) {
		res, err := client.Get("/oauth/test?redirect_uri=http://evil.com")
		require.NoError(t, err)
		test.AssertRedirect(t, res, "http://test.com")
	})
}
