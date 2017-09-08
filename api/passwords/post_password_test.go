package passwords_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/keratin/authn-server/api/passwords"
	"github.com/keratin/authn-server/api/test"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/services"
	"github.com/keratin/authn-server/tokens/resets"
	"github.com/keratin/authn-server/tokens/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostPassword(t *testing.T) {
	app := test.App()
	server := test.Server(app, passwords.Routes(app))
	defer server.Close()

	client := test.NewClient(server).Referred(app.Config)

	assertSuccess := func(t *testing.T, res *http.Response, account *models.Account) {
		assert.Equal(t, http.StatusCreated, res.StatusCode)
		test.AssertSession(t, app.Config, res.Cookies())
		test.AssertIDTokenResponse(t, res, app.KeyStore, app.Config)
		found, err := app.AccountStore.Find(account.ID)
		require.NoError(t, err)
		assert.NotEqual(t, found.Password, account.Password)
	}

	t.Run("valid reset token", func(t *testing.T) {
		// given an account
		account, err := app.AccountStore.Create("valid.token@authn.tech", []byte("oldpwd"))
		require.NoError(t, err)

		// given a reset token
		token, err := resets.New(app.Config, account.ID, account.PasswordChangedAt)
		require.NoError(t, err)
		tokenStr, err := token.Sign(app.Config.ResetSigningKey)
		require.NoError(t, err)

		// invoking the endpoint
		res, err := client.PostForm("/password", url.Values{
			"token":    []string{tokenStr},
			"password": []string{"0a0b0c0d0"},
		})
		require.NoError(t, err)

		// works
		assertSuccess(t, res, account)
	})

	t.Run("invalid reset token", func(t *testing.T) {
		// invoking the endpoint
		res, err := client.PostForm("/password", url.Values{
			"token":    []string{"invalid"},
			"password": []string{"0a0b0c0d0"},
		})
		require.NoError(t, err)

		// does not work
		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		test.AssertErrors(t, res, services.FieldErrors{{"token", "INVALID_OR_EXPIRED"}})
	})

	t.Run("valid session", func(t *testing.T) {
		// given an account
		account, err := app.AccountStore.Create("valid.session@authn.tech", []byte("oldpwd"))
		require.NoError(t, err)

		// given a session
		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		// invoking the endpoint
		res, err := client.WithSession(session).PostForm("/password", url.Values{
			"password": []string{"0a0b0c0d0"},
		})
		require.NoError(t, err)

		// works
		assertSuccess(t, res, account)

		// invalidates old session
		claims, err := sessions.Parse(session.Value, app.Config)
		require.NoError(t, err)
		id, err := app.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
		require.NoError(t, err)
		assert.Empty(t, id)
	})

	t.Run("valid session and bad password", func(t *testing.T) {
		// given an account
		account, err := app.AccountStore.Create("bad.password@authn.tech", []byte("oldpwd"))
		require.NoError(t, err)

		// given a session
		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		// invoking the endpoint
		res, err := client.WithSession(session).PostForm("/password", url.Values{
			"password": []string{"a"},
		})
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		test.AssertErrors(t, res, services.FieldErrors{{"password", "INSECURE"}})
	})

	t.Run("invalid session", func(t *testing.T) {
		session := &http.Cookie{
			Name:  app.Config.SessionCookieName,
			Value: "invalid",
		}

		res, err := client.WithSession(session).PostForm("/password", url.Values{
			"password": []string{"0a0b0c0d0"},
		})
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("token AND session", func(t *testing.T) {
		// given an account
		tokenAccount, err := app.AccountStore.Create("token@authn.tech", []byte("oldpwd"))
		require.NoError(t, err)
		// with a reset token
		token, err := resets.New(app.Config, tokenAccount.ID, tokenAccount.PasswordChangedAt)
		require.NoError(t, err)
		tokenStr, err := token.Sign(app.Config.ResetSigningKey)
		require.NoError(t, err)

		// given another account
		sessionAccount, err := app.AccountStore.Create("session@authn.tech", []byte("oldpwd"))
		require.NoError(t, err)
		// with a session
		session := test.CreateSession(app.RefreshTokenStore, app.Config, sessionAccount.ID)

		// invoking the endpoint
		res, err := client.WithSession(session).PostForm("/password", url.Values{
			"token":    []string{tokenStr},
			"password": []string{"0a0b0c0d0"},
		})
		require.NoError(t, err)

		// works
		assertSuccess(t, res, tokenAccount)
	})
}
