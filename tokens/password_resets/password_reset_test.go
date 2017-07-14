package password_resets_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/tokens/password_resets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordResetToken(t *testing.T) {
	cfg := config.Config{
		AuthNURL:        &url.URL{Scheme: "https", Host: "authn.example.com"},
		ResetSigningKey: []byte("key-a-reno"),
		RefreshTokenTTL: 3600,
	}

	then := time.Now().Add(time.Duration(-1) * time.Second).Round(time.Second) // 1 second ago
	timestamp := then.Unix()
	account_id := 52167

	t.Run("creating signing and parsing", func(t *testing.T) {
		token, err := password_resets.New(&cfg, account_id, then)
		require.NoError(t, err)
		assert.Equal(t, "reset", token.Scope)
		assert.Equal(t, timestamp, token.Lock)
		assert.Equal(t, "https://authn.example.com", token.Issuer)
		assert.Equal(t, "52167", token.Subject)
		assert.Equal(t, "https://authn.example.com", token.Audience)
		assert.NotEmpty(t, token.ExpiresAt)
		assert.NotEmpty(t, token.IssuedAt)

		tokenStr, err := token.Sign(cfg.ResetSigningKey)
		require.NoError(t, err)

		_, err = password_resets.Parse(tokenStr, &cfg)
		require.NoError(t, err)
	})

	t.Run("parsing with a different key", func(t *testing.T) {
		oldCfg := config.Config{
			AuthNURL:        cfg.AuthNURL,
			ResetSigningKey: []byte("old-a-reno"),
			RefreshTokenTTL: cfg.RefreshTokenTTL,
		}
		token, err := password_resets.New(&oldCfg, account_id, then)
		require.NoError(t, err)
		tokenStr, err := token.Sign(oldCfg.ResetSigningKey)
		require.NoError(t, err)
		_, err = password_resets.Parse(tokenStr, &cfg)
		assert.Error(t, err)
	})

	t.Run("parsing with an unknown issuer and audience", func(t *testing.T) {
		oldCfg := config.Config{
			AuthNURL:        &url.URL{Scheme: "https", Host: "unknown.com"},
			ResetSigningKey: cfg.ResetSigningKey,
			RefreshTokenTTL: cfg.RefreshTokenTTL,
		}
		token, err := password_resets.New(&oldCfg, account_id, then)
		require.NoError(t, err)
		tokenStr, err := token.Sign(oldCfg.ResetSigningKey)
		require.NoError(t, err)
		_, err = password_resets.Parse(tokenStr, &cfg)
		assert.Error(t, err)
	})

	t.Run("checking lock expiration", func(t *testing.T) {
		claims := password_resets.Claims{Lock: timestamp}
		assert.False(t, claims.LockExpired(&then))
		later := then.Add(time.Hour)
		assert.True(t, claims.LockExpired(&later))
	})
}
