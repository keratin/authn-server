package test

import (
	"encoding/json"
	"net/http"
	"testing"

	jwt "gopkg.in/square/go-jose.v2/jwt"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/services"
	"github.com/keratin/authn-server/tokens/identities"
	"github.com/keratin/authn-server/tokens/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssertData(t *testing.T, res *http.Response, expected interface{}) {
	assert.Equal(t, []string{"application/json"}, res.Header["Content-Type"])

	j, err := json.Marshal(api.ServiceData{expected})
	require.NoError(t, err)
	assert.Equal(t, string(j), string(ReadBody(res)))
}

func AssertErrors(t *testing.T, res *http.Response, expected services.FieldErrors) {
	assert.Equal(t, []string{"application/json"}, res.Header["Content-Type"])

	j, err := json.Marshal(api.ServiceErrors{Errors: expected})
	require.NoError(t, err)
	assert.Equal(t, string(j), string(ReadBody(res)))
}

func AssertSession(t *testing.T, cfg *config.Config, cookies []*http.Cookie) {
	var session string
	for _, cookie := range cookies {
		if cookie.Name == cfg.SessionCookieName {
			session = cookie.Value
			break
		}
	}
	require.NotEmpty(t, session)

	_, err := sessions.Parse(session, cfg)
	assert.NoError(t, err)
}

func AssertIdTokenResponse(t *testing.T, res *http.Response, keyStore data.KeyStore, cfg *config.Config) {
	// check that the response contains the expected json
	assert.Equal(t, []string{"application/json"}, res.Header["Content-Type"])
	responseData := struct {
		IdToken string `json:"id_token"`
	}{}
	err := extractResult(res, &responseData)
	assert.NoError(t, err)

	tok, err := jwt.ParseSigned(responseData.IdToken)
	assert.NoError(t, err)

	claims := identities.Claims{}
	err = tok.Claims(keyStore.Key().Public(), &claims)
	if assert.NoError(t, err) {
		// check that the JWT contains nice things
		assert.Equal(t, cfg.AuthNURL.String(), claims.Issuer)
	}
}

// extracts the value from inside a successful result envelope. must be provided
// with `inner`, an empty struct that describes the expected (desired) shape of
// what is inside the envelope.
func extractResult(res *http.Response, inner interface{}) error {
	return json.Unmarshal(ReadBody(res), &api.ServiceData{inner})
}
