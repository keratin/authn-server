package resets_test

import (
	"net/url"
	"testing"
	"time"

	jwt "gopkg.in/square/go-jose.v2/jwt"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/tokens/resets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordResetToken(t *testing.T) {
	cfg := &app.Config{
		AuthNURL:        &url.URL{Scheme: "https", Host: "authn.example.com"},
		ResetSigningKey: []byte("key-a-reno"),
		RefreshTokenTTL: 3600,
	}

	then := time.Now().Add(time.Duration(-1) * time.Second).Truncate(time.Second) // 1 second ago
	accountID := 52167

	t.Run("creating signing and parsing", func(t *testing.T) {
		token, err := resets.New(cfg, accountID, then)
		require.NoError(t, err)
		assert.Equal(t, "reset", token.Scope)
		assert.Equal(t, then, token.Lock.Time())
		assert.Equal(t, "https://authn.example.com", token.Issuer)
		assert.Equal(t, "52167", token.Subject)
		assert.True(t, token.Audience.Contains("https://authn.example.com"))
		assert.NotEmpty(t, token.Expiry)
		assert.NotEmpty(t, token.IssuedAt)

		tokenStr, err := token.Sign(cfg.ResetSigningKey)
		require.NoError(t, err)

		_, err = resets.Parse(tokenStr, cfg)
		require.NoError(t, err)
	})

	t.Run("parsing with a different key", func(t *testing.T) {
		oldCfg := app.Config{
			AuthNURL:        cfg.AuthNURL,
			ResetSigningKey: []byte("old-a-reno"),
			RefreshTokenTTL: cfg.RefreshTokenTTL,
		}
		token, err := resets.New(&oldCfg, accountID, then)
		require.NoError(t, err)
		tokenStr, err := token.Sign(oldCfg.ResetSigningKey)
		require.NoError(t, err)
		_, err = resets.Parse(tokenStr, cfg)
		assert.Error(t, err)
	})

	t.Run("parsing with an unknown issuer and audience", func(t *testing.T) {
		oldCfg := app.Config{
			AuthNURL:        &url.URL{Scheme: "https", Host: "unknown.com"},
			ResetSigningKey: cfg.ResetSigningKey,
			RefreshTokenTTL: cfg.RefreshTokenTTL,
		}
		token, err := resets.New(&oldCfg, accountID, then)
		require.NoError(t, err)
		tokenStr, err := token.Sign(oldCfg.ResetSigningKey)
		require.NoError(t, err)
		_, err = resets.Parse(tokenStr, cfg)
		assert.Error(t, err)
	})

	t.Run("checking lock expiration", func(t *testing.T) {
		claims := resets.Claims{Lock: jwt.NewNumericDate(then)}
		assert.False(t, claims.LockExpired(then))
		assert.False(t, claims.LockExpired(then.Add(time.Microsecond)))
		assert.True(t, claims.LockExpired(then.Add(time.Second)))
	})
}
