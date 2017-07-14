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

	t.Run("sets new password", func(t *testing.T) {
		expired, err := accountStore.Create("expired@keratin.tech", []byte("old"))
		require.NoError(t, err)
		err = accountStore.RequireNewPassword(expired.Id)
		require.NoError(t, err)

		claims, err := password_resets.New(cfg, expired.Id, *expired.PasswordChangedAt)
		require.NoError(t, err)
		token, err := claims.Sign(cfg.ResetSigningKey)
		require.NoError(t, err)

		err = services.PasswordResetter(accountStore, cfg, token, "0a0b0c0d0e0f")
		assert.NoError(t, err)
		account, err := accountStore.Find(expired.Id)
		require.NoError(t, err)
		assert.NotEqual(t, expired.Password, account.Password)
		assert.False(t, account.RequireNewPassword)
	})

	t.Run("when token is invalid", func(t *testing.T) {
		token := "not.valid.jwt"

		err := services.PasswordResetter(accountStore, cfg, token, "0a0b0c0d0e0f")
		assert.Equal(t, services.FieldErrors{{"token", services.ErrInvalidOrExpired}}, err)
	})

	account, err := accountStore.Create("existing@keratin.tech", []byte("old"))
	require.NoError(t, err)

	changed, err := accountStore.Create("changed@keratin.tech", []byte("old"))
	require.NoError(t, err)

	locked, err := accountStore.Create("locked@keratin.tech", []byte("old"))
	require.NoError(t, err)
	err = accountStore.Lock(locked.Id)
	require.NoError(t, err)

	archived, err := accountStore.Create("archived@keratin.tech", []byte("old"))
	require.NoError(t, err)
	err = accountStore.Archive(archived.Id)
	require.NoError(t, err)

	failureCases := []struct {
		account_id int
		lock       time.Time
		password   string
		errors     services.FieldErrors
	}{
		{changed.Id, changed.PasswordChangedAt.Add(time.Duration(-1) * time.Hour), "0a0b0c0d0e0f", services.FieldErrors{{"token", "INVALID_OR_EXPIRED"}}},
		{archived.Id, *archived.PasswordChangedAt, "0a0b0c0d0e0f", services.FieldErrors{{"account", "LOCKED"}}},
		{locked.Id, *locked.PasswordChangedAt, "0a0b0c0d0e0f", services.FieldErrors{{"account", "LOCKED"}}},
		{account.Id, *account.PasswordChangedAt, "abc", services.FieldErrors{{"password", "INSECURE"}}},
		{account.Id, *account.PasswordChangedAt, "", services.FieldErrors{{"password", "MISSING"}}},
		{0, time.Now(), "0a0b0c0d0e0f", services.FieldErrors{{"account", "NOT_FOUND"}}},
	}

	for _, tc := range failureCases {
		t.Run(tc.errors.Error(), func(t *testing.T) {
			claims, err := password_resets.New(cfg, tc.account_id, tc.lock)
			require.NoError(t, err)
			token, err := claims.Sign(cfg.ResetSigningKey)
			require.NoError(t, err)

			err = services.PasswordResetter(accountStore, cfg, token, tc.password)
			assert.Equal(t, tc.errors, err)
		})
	}
}
