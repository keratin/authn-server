package services_test

import (
	"net/url"
	"testing"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/app/tokens/passwordless"
	"github.com/keratin/authn-server/ops"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordlessTokenVerifier(t *testing.T) {
	accountStore := mock.NewAccountStore()
	cfg := &app.Config{
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
		_, err := services.PasswordlessTokenVerifier(accountStore, &ops.LogReporter{logrus.New()}, cfg, token)
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

	t.Run("when account has logged in again", func(t *testing.T) {
		account, err := accountStore.Create("account@keratin.tech", []byte("old"))
		require.NoError(t, err)

		token := newToken(account.ID)

		_, err = accountStore.SetLastLogin(account.ID)
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
