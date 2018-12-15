package oauth_test

import (
	"net/url"
	"testing"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/tokens/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthToken(t *testing.T) {
	cfg := &app.Config{
		AuthNURL:        &url.URL{Scheme: "https", Host: "authn.example.com"},
		OAuthSigningKey: []byte("key-a-reno"),
	}

	nonce := "rand123"

	t.Run("creating signing and parsing", func(t *testing.T) {
		token, err := oauth.New(cfg, nonce, "https://app.example.com/return")
		require.NoError(t, err)
		assert.Equal(t, "oauth", token.Scope)
		assert.Equal(t, nonce, token.RequestForgeryProtection)
		assert.Equal(t, "https://app.example.com/return", token.Destination)
		assert.Equal(t, "https://authn.example.com", token.Issuer)
		assert.True(t, token.Audience.Contains("https://authn.example.com"))
		assert.NotEmpty(t, token.IssuedAt)

		tokenStr, err := token.Sign(cfg.OAuthSigningKey)
		require.NoError(t, err)

		_, err = oauth.Parse(tokenStr, cfg, nonce)
		require.NoError(t, err)
	})

	t.Run("parsing with an unknown nonce", func(t *testing.T) {
		token, err := oauth.New(cfg, nonce, "https://app.example.com/return")
		require.NoError(t, err)

		tokenStr, err := token.Sign(cfg.OAuthSigningKey)
		require.NoError(t, err)

		_, err = oauth.Parse(tokenStr, cfg, "wrong")
		assert.Error(t, err)
	})

	t.Run("parsing with a different key", func(t *testing.T) {
		oldCfg := app.Config{
			AuthNURL:        cfg.AuthNURL,
			OAuthSigningKey: []byte("old-a-reno"),
		}
		token, err := oauth.New(cfg, nonce, "https://app.example.com/return")
		require.NoError(t, err)
		tokenStr, err := token.Sign(oldCfg.OAuthSigningKey)
		require.NoError(t, err)
		_, err = oauth.Parse(tokenStr, cfg, nonce)
		assert.Error(t, err)
	})

	t.Run("parsing with an unknown issuer and audience", func(t *testing.T) {
		oldCfg := app.Config{
			AuthNURL:        &url.URL{Scheme: "https", Host: "unknown.com"},
			OAuthSigningKey: cfg.OAuthSigningKey,
		}
		token, err := oauth.New(&oldCfg, nonce, "https://app.example.com/return")
		require.NoError(t, err)
		tokenStr, err := token.Sign(cfg.OAuthSigningKey)
		require.NoError(t, err)
		_, err = oauth.Parse(tokenStr, cfg, nonce)
		assert.Error(t, err)
	})
}
