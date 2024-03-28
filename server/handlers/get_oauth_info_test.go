package handlers_test

import (
	"net/http"
	"testing"

	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/require"
)

func TestGetOauthInfo(t *testing.T) {
	app := test.App()

	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0]).WithCookie(&http.Cookie{
		Name:  app.Config.OAuthCookieName,
		Value: "",
	})

	t.Run("unanthorized", func(t *testing.T) {
		res, err := client.Get("/oauth/info")
		require.NoError(t, err)

		require.Equal(t, http.StatusUnauthorized, res.StatusCode)
		require.Equal(t, []byte{}, test.ReadBody(res))
	})

	t.Run("get oauth info", func(t *testing.T) {
		expected := "{\"result\":[{\"email\":\"email\",\"provider\":\"test\",\"provider_account_id\":\"ID\"}]}"
		account, err := app.AccountStore.Create("get-oauth-info@keratin.tech", []byte("password"))
		require.NoError(t, err)

		err = app.AccountStore.AddOauthAccount(account.ID, "test", "ID", "email", "TOKEN")
		require.NoError(t, err)

		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		res, err := client.WithCookie(session).Get("/oauth/info")
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Equal(t, []byte(expected), test.ReadBody(res))
	})
}
