package handlers_test

import (
	"net/http"
	"net/url"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/keratin/authn-server/server/test"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/app/models"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/app/tokens/passwordless"
	"github.com/keratin/authn-server/app/tokens/sessions"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostSessionToken(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0])

	assertSuccess := func(t *testing.T, res *http.Response, account *models.Account) {
		assert.Equal(t, http.StatusCreated, res.StatusCode)
		test.AssertSession(t, app.Config, res.Cookies())
		test.AssertIDTokenResponse(t, res, app.KeyStore, app.Config)
		found, err := app.AccountStore.Find(account.ID)
		require.NoError(t, err)
		assert.Equal(t, found.Password, account.Password)
	}

	factory := func(username string, password string) (*models.Account, error) {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), app.Config.BcryptCost)
		if err != nil {
			return nil, errors.Wrap(err, "bcrypt")
		}

		return app.AccountStore.Create(username, hash)
	}

	t.Run("valid passwordless token", func(t *testing.T) {
		// given an account
		account, err := factory("valid.token@authn.tech", "oldpwd")
		require.NoError(t, err)

		// given a passwordless token
		token, err := passwordless.New(app.Config, account.ID)
		require.NoError(t, err)
		tokenStr, err := token.Sign(app.Config.PasswordlessTokenSigningKey)
		require.NoError(t, err)

		// invoking the endpoint
		res, err := client.PostForm("/session/token", url.Values{
			"token": []string{tokenStr},
		})
		require.NoError(t, err)

		// works
		assertSuccess(t, res, account)
	})

	t.Run("invalid passwordless token", func(t *testing.T) {
		// invoking the endpoint
		res, err := client.PostForm("/session/token", url.Values{
			"token": []string{"invalid"},
		})
		require.NoError(t, err)

		// does not work
		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		test.AssertErrors(t, res, services.FieldErrors{{"token", "INVALID_OR_EXPIRED"}})
	})

	t.Run("valid session", func(t *testing.T) {
		// given an account
		account, err := factory("valid.session@authn.tech", "oldpwd")
		require.NoError(t, err)

		// given a session
		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		// given a passwordless token
		token, err := passwordless.New(app.Config, account.ID)
		require.NoError(t, err)
		tokenStr, err := token.Sign(app.Config.PasswordlessTokenSigningKey)
		require.NoError(t, err)

		// invoking the endpoint
		res, err := client.WithCookie(session).PostForm("/session/token", url.Values{
			"token": []string{tokenStr},
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

}
