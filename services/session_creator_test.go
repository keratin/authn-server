package services_test

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/ops"
	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func TestSessionCreator(t *testing.T) {
	cfg := &config.Config{
		AuthNURL: &url.URL{Scheme: "http", Host: "authn.example.com"},
	}
	rsaKey, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err)
	keyStore := mock.NewKeyStore(rsaKey)
	refreshStore := mock.NewRefreshTokenStore()
	accountStore := mock.NewAccountStore()
	reporter := &ops.LogReporter{}

	audience := &route.Domain{"authn.example.com", "8080"}
	account, err := accountStore.Create("existing", []byte("secret"))
	require.NoError(t, err)

	t.Run("tracks last login while generating tokens", func(t *testing.T) {
		identityToken, refreshToken, err := services.SessionCreator(
			accountStore, refreshStore, keyStore, nil, cfg, reporter,
			account.ID, audience, nil,
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, identityToken)
		assert.NotEmpty(t, refreshToken)

		found, err := accountStore.Find(account.ID)
		require.NoError(t, err)
		assert.NotEqual(t, nil, found.LastLoginAt)
	})

	t.Run("tracks actives", func(t *testing.T) {
		activesStore := mock.NewActives()
		_, _, err := services.SessionCreator(
			accountStore, refreshStore, keyStore, activesStore, cfg, reporter,
			account.ID, audience, nil,
		)

		report, err := activesStore.ActivesByDay()
		require.NoError(t, err)
		assert.Len(t, report, 1)
	})

	t.Run("ends existing session", func(t *testing.T) {
		token, err := refreshStore.Create(account.ID)
		require.NoError(t, err)

		_, _, err = services.SessionCreator(
			accountStore, refreshStore, keyStore, nil, cfg, reporter,
			account.ID, audience, &token,
		)
		assert.NoError(t, err)

		foundID, err := refreshStore.Find(token)
		assert.Empty(t, foundID)
		assert.NoError(t, err)
	})
}
