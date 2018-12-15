package passwordless_test

import (
	"net/url"
	"testing"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/tokens/passwordless"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordlessToken(t *testing.T) {
	cfg := &app.Config{
		AuthNURL:                    &url.URL{Scheme: "https", Host: "authn.example.com"},
		PasswordlessTokenSigningKey: []byte("key-a-reno"),
		PasswordlessTokenTTL:        3600,
	}

	accountID := 52167

	t.Run("creating signing and parsing", func(t *testing.T) {
		token, err := passwordless.New(cfg, accountID)
		require.NoError(t, err)
		assert.Equal(t, "passwordless", token.Scope)
		assert.Equal(t, "https://authn.example.com", token.Issuer)
		assert.Equal(t, "52167", token.Subject)
		assert.True(t, token.Audience.Contains("https://authn.example.com"))
		assert.NotEmpty(t, token.Expiry)
		assert.NotEmpty(t, token.IssuedAt)

		tokenStr, err := token.Sign(cfg.PasswordlessTokenSigningKey)
		require.NoError(t, err)

		_, err = passwordless.Parse(tokenStr, cfg)
		require.NoError(t, err)
	})

	t.Run("parsing with a different key", func(t *testing.T) {
		oldCfg := app.Config{
			AuthNURL:                    cfg.AuthNURL,
			PasswordlessTokenSigningKey: []byte("old-a-reno"),
			PasswordlessTokenTTL:        cfg.PasswordlessTokenTTL,
		}
		token, err := passwordless.New(&oldCfg, accountID)
		require.NoError(t, err)
		tokenStr, err := token.Sign(oldCfg.PasswordlessTokenSigningKey)
		require.NoError(t, err)
		_, err = passwordless.Parse(tokenStr, cfg)
		assert.Error(t, err)
	})

	t.Run("parsing with an unknown issuer and audience", func(t *testing.T) {
		oldCfg := app.Config{
			AuthNURL:                    &url.URL{Scheme: "https", Host: "unknown.com"},
			PasswordlessTokenSigningKey: cfg.PasswordlessTokenSigningKey,
			PasswordlessTokenTTL:        cfg.PasswordlessTokenTTL,
		}
		token, err := passwordless.New(&oldCfg, accountID)
		require.NoError(t, err)
		tokenStr, err := token.Sign(oldCfg.PasswordlessTokenSigningKey)
		require.NoError(t, err)
		_, err = passwordless.Parse(tokenStr, cfg)
		assert.Error(t, err)
	})

}
