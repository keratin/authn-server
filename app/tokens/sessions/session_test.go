package sessions_test

import (
	"net/url"
	"testing"

	jwt "gopkg.in/square/go-jose.v2/jwt"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/tokens/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAndParseAndSign(t *testing.T) {
	store := mock.NewRefreshTokenStore()
	cfg := app.Config{
		AuthNURL:          &url.URL{Scheme: "http", Host: "authn.example.com"},
		SessionSigningKey: []byte("key-a-reno"),
	}

	token, err := sessions.New(store, &cfg, 658908, "example.com")
	require.NoError(t, err)
	assert.Equal(t, "refresh", token.Scope)
	assert.Equal(t, "http://authn.example.com", token.Issuer)
	assert.True(t, token.Audience.Contains("http://authn.example.com"))
	assert.NotEmpty(t, token.Subject)
	assert.Equal(t, "example.com", token.Azp)
	assert.NotEmpty(t, token.IssuedAt)

	sessionString, err := token.Sign(cfg.SessionSigningKey)
	require.NoError(t, err)

	claims, err := sessions.Parse(sessionString, &cfg)
	require.NoError(t, err)
	assert.Equal(t, "http://authn.example.com", claims.Issuer)
	assert.True(t, token.Audience.Contains("http://authn.example.com"))
	assert.NotEmpty(t, token.Subject)
	assert.Equal(t, "example.com", claims.Azp)
	assert.NotEmpty(t, claims.IssuedAt)
}

func TestParseInvalidSessionJWT(t *testing.T) {
	store := mock.NewRefreshTokenStore()
	authn := url.URL{Scheme: "http", Host: "authn.example.com"}
	mainApp := url.URL{Scheme: "http", Host: "app.example.com"}
	key := []byte("current key")
	cfg := app.Config{AuthNURL: &authn, SessionSigningKey: key}

	t.Run("old key", func(t *testing.T) {
		token, err := sessions.New(store, &app.Config{AuthNURL: &authn}, 1, mainApp.Host)
		require.NoError(t, err)
		tokenStr, err := token.Sign([]byte("old key"))
		require.NoError(t, err)

		_, err = sessions.Parse(tokenStr, &cfg)
		assert.Error(t, err)
	})

	t.Run("different audience", func(t *testing.T) {
		token, err := sessions.New(store, &app.Config{AuthNURL: &authn}, 2, mainApp.Host)
		require.NoError(t, err)
		token.Audience = jwt.Audience{mainApp.String()}
		tokenStr, err := token.Sign(key)
		require.NoError(t, err)

		_, err = sessions.Parse(tokenStr, &cfg)
		assert.Error(t, err)
	})
}
