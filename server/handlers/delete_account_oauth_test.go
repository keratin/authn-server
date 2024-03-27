package handlers_test

import (
	"net/http"
	"testing"

	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/require"
)

func TestDeleteAccountOauth(t *testing.T) {
	app := test.App()

	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Authenticated(app.Config.AuthUsername, app.Config.AuthPassword)

	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	t.Run("delete social account", func(t *testing.T) {
		expected := "{\"result\":{\"require_password_reset\":true}}"
		account, err := app.AccountStore.Create("deleted-social-account@keratin.tech", []byte("password"))
		require.NoError(t, err)

		err = app.AccountStore.AddOauthAccount(account.ID, "test", "DELETEDID", "TOKEN")
		require.NoError(t, err)

		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		payload := map[string]interface{}{"oauth_providers": []string{"test"}}
		res, err := client.WithCookie(session).DeleteJSON("/accounts/1/oauth", payload)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Equal(t, []byte(expected), test.ReadBody(res))
	})
}
