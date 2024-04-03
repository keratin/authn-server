package oauth_test

import (
	"encoding/hex"
	"net/http"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/keratin/authn-server/lib/oauth"
	"github.com/keratin/authn-server/lib/oauth/apple"
	"github.com/keratin/authn-server/lib/oauth/apple/test"
	"github.com/stretchr/testify/assert"
)

func TestAppleProvider(t *testing.T) {
	keyString := hex.EncodeToString([]byte(test.SampleKey))

	teamID, keyID, clientID := "teamID", "keyID", "clientID"
	p, err := oauth.NewAppleProvider(&oauth.Credentials{
		ID:     clientID,
		Secret: keyString,
		Additional: map[string]string{
			"teamID":        teamID,
			"keyID":         keyID,
			"expirySeconds": "3600",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, http.MethodPost, p.ReturnMethod())
	assert.Equal(t, 1, len(p.AuthCodeOptions()))

	got, err := p.Secret()
	assert.NoError(t, err)
	token, err := jwt.ParseSigned(got)
	assert.NoError(t, err)
	assert.NotNil(t, token)

	claims := map[string]interface{}{}

	// verification is tested directly in apple/privkey_test.go
	err = token.UnsafeClaimsWithoutVerification(&claims)
	assert.NoError(t, err)

	assert.Equal(t, teamID, claims["iss"])
	assert.Equal(t, clientID, claims["sub"])
	assert.Equal(t, apple.BaseURL, claims["aud"])
	assert.Less(t, float64(time.Now().Unix()+3600-1), claims["exp"].(float64))
	assert.Less(t, float64(time.Now().Unix()-1), claims["iat"].(float64))
}
