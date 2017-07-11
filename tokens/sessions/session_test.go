package sessions_test

import (
	"net/url"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/tokens/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAndParseAndSign(t *testing.T) {
	store := mock.NewRefreshTokenStore()
	cfg := config.Config{
		AuthNURL:          &url.URL{Scheme: "http", Host: "authn.example.com"},
		SessionSigningKey: []byte("key-a-reno"),
	}

	token, err := sessions.New(store, &cfg, 658908)
	require.NoError(t, err)
	assert.Equal(t, "http://authn.example.com", token.Issuer)
	assert.Equal(t, "http://authn.example.com", token.Audience)
	assert.Equal(t, "RefreshToken:658908", token.Subject)
	assert.Equal(t, "", token.Azp)
	assert.NotEmpty(t, token.IssuedAt)

	sessionString, err := token.Sign(cfg.SessionSigningKey)
	require.NoError(t, err)

	claims, err := sessions.Parse(sessionString, &cfg)
	require.NoError(t, err)
	assert.Equal(t, "http://authn.example.com", claims.Issuer)
	assert.Equal(t, "http://authn.example.com", claims.Audience)
	assert.Equal(t, "RefreshToken:658908", claims.Subject)
	assert.Equal(t, "", claims.Azp)
	assert.NotEmpty(t, claims.IssuedAt)
}

func TestParseInvalidSessionJWT(t *testing.T) {
	store := mock.NewRefreshTokenStore()
	authn := url.URL{Scheme: "http", Host: "authn.example.com"}
	app := url.URL{Scheme: "http", Host: "app.example.com"}
	key := []byte("current key")
	oldKey := []byte("old key")

	invalids := []string{}
	var token *sessions.Claims
	var tokenStr string
	var cfg config.Config
	var err error

	// This invalid JWT was signed with an old key.
	cfg = config.Config{AuthNURL: &authn, SessionSigningKey: oldKey}
	token, err = sessions.New(store, &cfg, 1)
	require.NoError(t, err)
	tokenStr, err = token.Sign(cfg.SessionSigningKey)
	require.NoError(t, err)
	invalids = append(invalids, tokenStr)

	// This invalid JWT was signed for a different audience.
	cfg = config.Config{AuthNURL: &authn, SessionSigningKey: key}
	token, err = sessions.New(store, &cfg, 2)
	require.NoError(t, err)
	token.Audience = app.String()
	tokenStr, err = token.Sign(cfg.SessionSigningKey)
	require.NoError(t, err)
	invalids = append(invalids, tokenStr)

	// This invalid JWT was signed with "none" alg
	cfg = config.Config{AuthNURL: &authn}
	token, err = sessions.New(store, &cfg, 3)
	require.NoError(t, err)
	tokenStr, err = jwt.NewWithClaims(jwt.SigningMethodNone, token).SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)
	invalids = append(invalids, tokenStr)

	cfg = config.Config{AuthNURL: &authn, SessionSigningKey: key}
	for _, invalid := range invalids {
		_, err := sessions.Parse(invalid, &cfg)
		assert.Error(t, err)
	}
}
