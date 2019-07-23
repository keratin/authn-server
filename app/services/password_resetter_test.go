package services_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/app/tokens/resets"
	"github.com/keratin/authn-server/ops"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordResetter(t *testing.T) {
	accountStore := mock.NewAccountStore()
	cfg := &app.Config{
		AuthNURL:              &url.URL{Scheme: "http", Host: "authn.example.com"},
		BcryptCost:            4,
		PasswordMinComplexity: 1,
		ResetSigningKey:       []byte("reset-a-reno"),
	}

	newToken := func(id int, lock time.Time) string {
		claims, err := resets.New(cfg, id, lock)
		require.NoError(t, err)
		token, err := claims.Sign(cfg.ResetSigningKey)
		require.NoError(t, err)
		return token
	}

	invoke := func(token string, password string) error {
		_, err := services.PasswordResetter(accountStore, &ops.LogReporter{logrus.New()}, cfg, token, password)
		return err
	}

	account, err := accountStore.Create("existing@keratin.tech", []byte("old"))
	require.NoError(t, err)

	t.Run("sets new password", func(t *testing.T) {
		expired, err := accountStore.Create("expired@keratin.tech", []byte("old"))
		require.NoError(t, err)
		_, err = accountStore.RequireNewPassword(expired.ID)
		require.NoError(t, err)

		err = invoke(newToken(expired.ID, expired.PasswordChangedAt), "0a0b0c0d0e0f")
		assert.NoError(t, err)

		account, err := accountStore.Find(expired.ID)
		require.NoError(t, err)
		assert.NotEqual(t, expired.Password, account.Password)
		assert.False(t, account.RequireNewPassword)
	})

	t.Run("when token is invalid", func(t *testing.T) {
		token := "not.valid.jwt"

		err := invoke(token, "0a0b0c0d0e0f")
		assert.Equal(t, services.FieldErrors{{"token", services.ErrInvalidOrExpired}}, err)
	})

	t.Run("after a password change", func(t *testing.T) {
		previousPasswordChange := account.PasswordChangedAt.Add(time.Duration(-1) * time.Hour)
		token := newToken(account.ID, previousPasswordChange)

		err := invoke(token, "0a0b0c0d0e0f")
		assert.Equal(t, services.FieldErrors{{"token", "INVALID_OR_EXPIRED"}}, err)
	})

	t.Run("on an archived account", func(t *testing.T) {
		archived, err := accountStore.Create("archived@keratin.tech", []byte("old"))
		require.NoError(t, err)
		_, err = accountStore.Archive(archived.ID)
		require.NoError(t, err)

		token := newToken(archived.ID, archived.PasswordChangedAt)

		err = invoke(token, "0a0b0c0d0e0f")
		assert.Equal(t, services.FieldErrors{{"account", "LOCKED"}}, err)
	})

	t.Run("on a locked account", func(t *testing.T) {
		locked, err := accountStore.Create("locked@keratin.tech", []byte("old"))
		require.NoError(t, err)
		_, err = accountStore.Lock(locked.ID)
		require.NoError(t, err)

		token := newToken(locked.ID, locked.PasswordChangedAt)

		err = invoke(token, "0a0b0c0d0e0f")
		assert.Equal(t, services.FieldErrors{{"account", "LOCKED"}}, err)
	})

	t.Run("with a missing password", func(t *testing.T) {
		token := newToken(account.ID, account.PasswordChangedAt)
		err := invoke(token, "")
		assert.Equal(t, services.FieldErrors{{"password", "MISSING"}}, err)
	})

	t.Run("with an insecure password", func(t *testing.T) {
		token := newToken(account.ID, account.PasswordChangedAt)
		err := invoke(token, "abc")
		assert.Equal(t, services.FieldErrors{{"password", "INSECURE"}}, err)
	})

	t.Run("with an unknown account", func(t *testing.T) {
		token := newToken(0, time.Now())
		err := invoke(token, "0a0b0c0d0e0f")
		assert.Equal(t, services.FieldErrors{{"account", "NOT_FOUND"}}, err)
	})
}
