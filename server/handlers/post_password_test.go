package handlers_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/keratin/authn-server/app/models"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/app/tokens/resets"
	"github.com/keratin/authn-server/app/tokens/sessions"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
	"github.com/pkg/errors"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostPassword(t *testing.T) {
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
		assert.NotEqual(t, found.Password, account.Password)
	}

	factory := func(username string, password string) (*models.Account, error) {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), app.Config.BcryptCost)
		if err != nil {
			return nil, errors.Wrap(err, "bcrypt")
		}

		return app.AccountStore.Create(username, hash)
	}

	t.Run("valid reset token", func(t *testing.T) {
		// given an account
		account, err := factory("valid.token@authn.tech", "oldpwd")
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
		test.AssertErrors(t, res, services.FieldErrors{{Field: "token", Message: "INVALID_OR_EXPIRED"}})
	})

	t.Run("valid session", func(t *testing.T) {
		// given an account
		account, err := factory("valid.session@authn.tech", "oldpwd")
		require.NoError(t, err)

		// given a session
		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		// invoking the endpoint
		res, err := client.WithCookie(session).PostForm("/password", url.Values{
			"currentPassword": []string{"oldpwd"},
			"password":        []string{"0a0b0c0d0"},
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
		account, err := factory("bad.password@authn.tech", "oldpwd")
		require.NoError(t, err)

		// given a session
		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		// invoking the endpoint
		res, err := client.WithCookie(session).PostForm("/password", url.Values{
			"currentPassword": []string{"oldpwd"},
			"password":        []string{"a"},
		})
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		test.AssertErrors(t, res, services.FieldErrors{{Field: "password", Message: "INSECURE"}})
	})

	t.Run("valid session and bad currentPassword", func(t *testing.T) {
		// given an account
		account, err := factory("bad.currentPassword@authn.tech", "oldpwd")
		require.NoError(t, err)

		// given a session
		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		// invoking the endpoint
		res, err := client.WithCookie(session).PostForm("/password", url.Values{
			"currentPassword": []string{"wrong"},
			"password":        []string{"0a0b0c0d0"},
		})
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		test.AssertErrors(t, res, services.FieldErrors{{Field: "credentials", Message: "FAILED"}})
	})

	t.Run("invalid session", func(t *testing.T) {
		session := &http.Cookie{
			Name:  app.Config.SessionCookieName,
			Value: "invalid",
		}

		res, err := client.WithCookie(session).PostForm("/password", url.Values{
			"currentPassword": []string{"oldpwd"},
			"password":        []string{"0a0b0c0d0"},
		})
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("token AND session", func(t *testing.T) {
		// given an account
		tokenAccount, err := factory("token@authn.tech", "oldpwd")
		require.NoError(t, err)
		// with a reset token
		token, err := resets.New(app.Config, tokenAccount.ID, tokenAccount.PasswordChangedAt)
		require.NoError(t, err)
		tokenStr, err := token.Sign(app.Config.ResetSigningKey)
		require.NoError(t, err)

		// given another account
		sessionAccount, err := factory("session@authn.tech", "oldpwd")
		require.NoError(t, err)
		// with a session
		session := test.CreateSession(app.RefreshTokenStore, app.Config, sessionAccount.ID)

		// invoking the endpoint
		res, err := client.WithCookie(session).PostForm("/password", url.Values{
			"token":           []string{tokenStr},
			"currentPassword": []string{"oldpwd"},
			"password":        []string{"0a0b0c0d0"},
		})
		require.NoError(t, err)

		// works
		assertSuccess(t, res, tokenAccount)
	})

	t.Run("multiple sessions with PasswordChangeLogout", func(t *testing.T) {
		app.Config.PasswordChangeLogout = true
		defer func() { app.Config.PasswordChangeLogout = false }()

		// given an account
		account, err := factory("PasswordChangeLogout@authn.tech", "oldpwd")
		require.NoError(t, err)

		// given a session
		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)
		otherSession := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		// invoking the endpoint
		res, err := client.WithCookie(session).PostForm("/password", url.Values{
			"currentPassword": []string{"oldpwd"},
			"password":        []string{"0a0b0c0d0"},
		})
		require.NoError(t, err)

		// works
		assertSuccess(t, res, account)

		// invalidates other session
		claims, err := sessions.Parse(otherSession.Value, app.Config)
		require.NoError(t, err)
		id, err := app.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
		require.NoError(t, err)
		assert.Empty(t, id)
	})

	t.Run("multiple sessions without PasswordChangeLogout", func(t *testing.T) {
		app.Config.PasswordChangeLogout = false

		// given an account
		account, err := factory("NoPasswordChangeLogout@authn.tech", "oldpwd")
		require.NoError(t, err)

		// given a session
		session := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)
		otherSession := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

		// invoking the endpoint
		res, err := client.WithCookie(session).PostForm("/password", url.Values{
			"currentPassword": []string{"oldpwd"},
			"password":        []string{"0a0b0c0d0"},
		})
		require.NoError(t, err)

		// works
		assertSuccess(t, res, account)

		// preserves other session
		claims, err := sessions.Parse(otherSession.Value, app.Config)
		require.NoError(t, err)
		id, err := app.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
		require.NoError(t, err)
		assert.Equal(t, account.ID, id)
	})
}

func TestPostPasswordWithTOTP(t *testing.T) {
	// nolint: gosec
	totpSecret := "JKK5AG4NDAWSZSR4ZFKZBWZ7OJGLB2JM"
	totpSecretEnc := []byte("cli6azfL5i7PAnh8U/w3Zbglsm3XcdaGODy+Ga5QqT02c9hotDAR1Y28--3UihzsJhw/+EU3R6--qUw9L8DwN5XPVfOStshKzA==")

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
		assert.NotEqual(t, found.Password, account.Password)
	}

	factory := func(username string, password string) (*models.Account, error) {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), app.Config.BcryptCost)
		if err != nil {
			return nil, errors.Wrap(err, "bcrypt")
		}

		return app.AccountStore.Create(username, hash)
	}

	t.Run("valid totp code", func(t *testing.T) {
		// given an account
		account, err := factory("valid@authn.tech", "oldpwd")
		require.NoError(t, err)
		_, err = app.AccountStore.SetTOTPSecret(account.ID, totpSecretEnc)
		require.NoError(t, err)

		// given a reset token
		token, err := resets.New(app.Config, account.ID, account.PasswordChangedAt)
		require.NoError(t, err)
		tokenStr, err := token.Sign(app.Config.ResetSigningKey)
		require.NoError(t, err)

		// given a totp code
		code, err := totp.GenerateCode(totpSecret, time.Now())
		require.NoError(t, err)

		// invoking the endpoint
		res, err := client.PostForm("/password", url.Values{
			"token":    []string{tokenStr},
			"password": []string{"0a0b0c0d0"},
			"totp":     []string{code},
		})
		require.NoError(t, err)

		// works
		assertSuccess(t, res, account)
	})

	t.Run("invalid totp code", func(t *testing.T) {
		// given an account
		account, err := factory("invaild@authn.tech", "oldpwd")
		require.NoError(t, err)
		_, err = app.AccountStore.SetTOTPSecret(account.ID, totpSecretEnc)
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
			"totp":     []string{"12345"},
		})
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		test.AssertErrors(t, res, services.FieldErrors{{Field: "totp", Message: "INVALID_OR_EXPIRED"}})
	})
}
