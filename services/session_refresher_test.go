package services_test

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/ops"
	"github.com/keratin/authn-server/services"
	"github.com/keratin/authn-server/tokens/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func TestSessionRefresher(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 512)
	require.NoError(t, err)
	keyStore := mock.NewKeyStore(rsaKey)
	cfg := &config.Config{
		AuthNURL: &url.URL{Scheme: "http", Host: "authn.example.com"},
	}
	refreshStore := mock.NewRefreshTokenStore()
	reporter := &ops.LogReporter{}

	accountID := 0
	audience := &route.Domain{"authn.example.com", "8080"}
	session, err := sessions.New(refreshStore, cfg, accountID, audience.String())
	require.NoError(t, err)

	t.Run("tracks actives while generating token", func(t *testing.T) {
		activesStore := mock.NewActives()

		identityToken, err := services.SessionRefresher(
			refreshStore, keyStore, activesStore, cfg, reporter,
			session, accountID, audience,
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, identityToken)

		report, err := activesStore.ActivesByDay()
		require.NoError(t, err)
		assert.Len(t, report, 1)
	})

	t.Run("ignores actives when not configured", func(t *testing.T) {
		identityToken, err := services.SessionRefresher(
			refreshStore, keyStore, nil, cfg, reporter,
			session, accountID, audience,
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, identityToken)
	})
}
