package services_test

import (
	"net/url"
	"testing"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/ops"
	"github.com/keratin/authn-server/services"
	"github.com/keratin/authn-server/tokens/passwordless"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordlessTokenVerifier(t *testing.T) {
	accountStore := mock.NewAccountStore()
	cfg := &config.Config{
		AuthNURL:                    &url.URL{Scheme: "http", Host: "authn.example.com"},
		BcryptCost:                  4,
		PasswordMinComplexity:       1,
		PasswordlessTokenSigningKey: []byte("reset-a-reno"),
	}

	newToken := func(id int) string {
		claims, err := passwordless.New(cfg, id)
		require.NoError(t, err)
		token, err := claims.Sign(cfg.PasswordlessTokenSigningKey)
		require.NoError(t, err)
		return token
	}

	invoke := func(token string) error {
		_, err := services.PasswordlessTokenVerifier(accountStore, &ops.LogReporter{}, cfg, token)
		return err
	}

	t.Run("when token is invalid", func(t *testing.T) {
		token := "not.valid.jwt"

		err := invoke(token)
		assert.Equal(t, services.FieldErrors{{"token", services.ErrInvalidOrExpired}}, err)
	})

	t.Run("on an archived account", func(t *testing.T) {
		archived, err := accountStore.Create("archived@keratin.tech", []byte("old"))
		require.NoError(t, err)
		_, err = accountStore.Archive(archived.ID)
		require.NoError(t, err)

		token := newToken(archived.ID)

		err = invoke(token)
		assert.Equal(t, services.FieldErrors{{"account", "LOCKED"}}, err)
	})

	t.Run("on a locked account", func(t *testing.T) {
		locked, err := accountStore.Create("locked@keratin.tech", []byte("old"))
		require.NoError(t, err)
		_, err = accountStore.Lock(locked.ID)
		require.NoError(t, err)

		token := newToken(locked.ID)

		err = invoke(token)
		assert.Equal(t, services.FieldErrors{{"account", "LOCKED"}}, err)
	})

	t.Run("logging in account after require a passwordless token", func(t *testing.T) {
		account, err := accountStore.Create("account@keratin.tech", []byte("old"))
		require.NoError(t, err)

		token := newToken(account.ID)

		err = invoke(token)
		require.NoError(t, err)

		err = services.LastLoginUpdater(accountStore, account.ID)
		require.NoError(t, err)

		err = invoke(token)
		assert.Equal(t, services.FieldErrors{{"token", services.ErrInvalidOrExpired}}, err)
	})

	t.Run("with an unknown account", func(t *testing.T) {
		token := newToken(0)
		err := invoke(token)
		assert.Equal(t, services.FieldErrors{{"account", "NOT_FOUND"}}, err)
	})
}
