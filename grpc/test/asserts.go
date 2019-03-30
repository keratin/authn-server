package test

import (
	"testing"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/app/tokens/identities"
	"github.com/keratin/authn-server/app/tokens/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"gopkg.in/square/go-jose.v2/jwt"
)

func AssertSession(t *testing.T, cfg *app.Config, md metadata.MD) {
	session := ReadMetadata(md, cfg.SessionCookieName)
	require.NotEmpty(t, session)

	_, err := sessions.Parse(session, cfg)
	assert.NoError(t, err)
}

func AssertIDTokenResponse(t *testing.T, idToken string, keyStore data.KeyStore, cfg *app.Config) {

	tok, err := jwt.ParseSigned(idToken)
	assert.NoError(t, err)

	claims := identities.Claims{}
	err = tok.Claims(keyStore.Key().Public(), &claims)
	if assert.NoError(t, err) {
		// check that the JWT contains nice things
		assert.Equal(t, cfg.AuthNURL.String(), claims.Issuer)
	}
}

func ReadMetadata(md metadata.MD, name string) string {
	for k, v := range md {
		if k == name {
			if len(v) > 0 {
				return v[0]
			}
			return ""
		}
	}
	return ""
}
