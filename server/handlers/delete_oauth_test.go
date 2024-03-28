package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	oauthlib "github.com/keratin/authn-server/lib/oauth"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
)

func TestDeleteOauthAccount(t *testing.T) {
	providerServer := httptest.NewServer(test.ProviderApp())
	defer providerServer.Close()

	app := test.App()
	app.OauthProviders["test"] = *oauthlib.NewTestProvider(providerServer)

	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0]).WithCookie(&http.Cookie{
		Name:  app.Config.OAuthCookieName,
		Value: "",
	})

	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	t.Run("unanthorized", func(t *testing.T) {
		res, err := client.Delete("/oauth/test")
		require.NoError(t, err)

		require.Equal(t, http.StatusUnauthorized, res.StatusCode)
		require.Equal(t, []byte{}, test.ReadBody(res))
	})

	t.Run("delete social account", func(t *testing.T) {
		account, err := app.AccountStore.Create("deleted@keratin.tech", []byte("password"))
		require.NoError(t, err)

		time.Sleep(5 * time.Second)

		err = app.AccountStore.AddOauthAccount(account.ID, "test", "DELETEDID", "email", "TOKEN")
		require.NoError(t, err)

		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		res, err := client.WithCookie(session).Delete("/oauth/test")
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Equal(t, []byte{}, test.ReadBody(res))
	})

	t.Run("return unprocessable entity when user requires a new password", func(t *testing.T) {
		expected := "{\"errors\":[{\"field\":\"password\",\"message\":\"NEW_PASSWORD_REQUIRED\"}]}"
		account, err := app.AccountStore.Create("deleted-unprocessable-entity@keratin.tech", []byte("password"))
		require.NoError(t, err)

		err = app.AccountStore.AddOauthAccount(account.ID, "test", "DELETEDID4", "email", "TOKEN")
		require.NoError(t, err)

		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		payload := map[string]interface{}{"oauth_providers": []string{"test"}}
		res, err := client.WithCookie(session).DeleteJSON("/oauth/test", payload)
		require.NoError(t, err)

		require.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		require.Equal(t, []byte(expected), test.ReadBody(res))
	})
}
