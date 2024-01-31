package apple_test

import (
	"crypto/ecdsa"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/google/uuid"
	"github.com/keratin/authn-server/lib/oauth/apple"
	"github.com/keratin/authn-server/lib/oauth/apple/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAppleSecret(t *testing.T) {
	keyID := "kid"
	clientID := "clientID"
	teamID := "teamID"
	expiresInSeconds := int64(3600)

	signingKey, err := apple.ParsePrivateKey([]byte(test.SampleKey), uuid.NewString())
	assert.NoError(t, err)

	jwk, ok := signingKey.Key.(jose.JSONWebKey)
	require.True(t, ok)
	pk, ok := jwk.Key.(*ecdsa.PrivateKey)
	require.True(t, ok)

	got, err := apple.GenerateSecret(*signingKey, keyID, clientID, teamID, expiresInSeconds)
	assert.NoError(t, err)
	assert.NotNil(t, got)

	token, err := jwt.ParseSigned(got)
	assert.NoError(t, err)
	assert.NotNil(t, token)

	claims := map[string]interface{}{}

	publicKey := jose.JSONWebKey{Key: &pk.PublicKey, Algorithm: string(jose.ES256), Use: "sig"}

	err = token.Claims(publicKey, &claims)
	assert.NoError(t, err)

	assert.Equal(t, teamID, claims["iss"])
	assert.Equal(t, clientID, claims["sub"])
	assert.Equal(t, apple.BaseURL, claims["aud"])
	assert.Less(t, float64(time.Now().Unix()+expiresInSeconds-1), claims["exp"].(float64))
	assert.Less(t, float64(time.Now().Unix()-1), claims["iat"].(float64))
}

func TestParseAppleKey(t *testing.T) {
	t.Run("invalid key", func(t *testing.T) {
		_, err := apple.ParsePrivateKey([]byte("invalid"), uuid.NewString())
		assert.Error(t, err)
	})
	t.Run("valid", func(t *testing.T) {
		keyID := uuid.NewString()

		got, err := apple.ParsePrivateKey([]byte(test.SampleKey), keyID)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		assert.Equal(t, jose.ES256, got.Algorithm)
	})
}
