package services_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/services"
	"github.com/keratin/authn-server/tokens/password_resets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordResetter(t *testing.T) {
	accountStore := mock.NewAccountStore()
	cfg := &config.Config{
		AuthNURL:              &url.URL{Scheme: "http", Host: "authn.example.com"},
		BcryptCost:            4,
		PasswordMinComplexity: 1,
		ResetSigningKey:       []byte("reset-a-reno"),
	}

	newToken := func(id int, lock time.Time) string {
		claims, err := password_resets.New(cfg, id, lock)
		require.NoError(t, err)
		token, err := claims.Sign(cfg.ResetSigningKey)
		require.NoError(t, err)
		return token
	}

	invoke := func(token string, password string) error {
		_, err := services.PasswordResetter(accountStore, cfg, token, password)
		return err
	}

	account, err := accountStore.Create("existing@keratin.tech", []byte("old"))
	require.NoError(t, err)

	t.Run("sets new password", func(t *testing.T) {
		expired, err := accountStore.Create("expired@keratin.tech", []byte("old"))
		require.NoError(t, err)
		err = accountStore.RequireNewPassword(expired.Id)
		require.NoError(t, err)

		err = invoke(newToken(expired.Id, expired.PasswordChangedAt), "0a0b0c0d0e0f")
		assert.NoError(t, err)

		account, err := accountStore.Find(expired.Id)
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
		token := newToken(account.Id, previousPasswordChange)

		err := invoke(token, "0a0b0c0d0e0f")
		assert.Equal(t, services.FieldErrors{{"token", "INVALID_OR_EXPIRED"}}, err)
	})

	t.Run("on an archived account", func(t *testing.T) {
		archived, err := accountStore.Create("archived@keratin.tech", []byte("old"))
		require.NoError(t, err)
		err = accountStore.Archive(archived.Id)
		require.NoError(t, err)

		token := newToken(archived.Id, archived.PasswordChangedAt)

		err = invoke(token, "0a0b0c0d0e0f")
		assert.Equal(t, services.FieldErrors{{"account", "LOCKED"}}, err)
	})

	t.Run("on a locked account", func(t *testing.T) {
		locked, err := accountStore.Create("locked@keratin.tech", []byte("old"))
		require.NoError(t, err)
		err = accountStore.Lock(locked.Id)
		require.NoError(t, err)

		token := newToken(locked.Id, locked.PasswordChangedAt)

		err = invoke(token, "0a0b0c0d0e0f")
		assert.Equal(t, services.FieldErrors{{"account", "LOCKED"}}, err)
	})

	t.Run("with a missing password", func(t *testing.T) {
		token := newToken(account.Id, account.PasswordChangedAt)
		err := invoke(token, "")
		assert.Equal(t, services.FieldErrors{{"password", "MISSING"}}, err)
	})

	t.Run("with an insecure password", func(t *testing.T) {
		token := newToken(account.Id, account.PasswordChangedAt)
		err := invoke(token, "abc")
		assert.Equal(t, services.FieldErrors{{"password", "INSECURE"}}, err)
	})

	t.Run("with an unknown account", func(t *testing.T) {
		token := newToken(0, time.Now())
		err := invoke(token, "0a0b0c0d0e0f")
		assert.Equal(t, services.FieldErrors{{"account", "NOT_FOUND"}}, err)
	})
}
