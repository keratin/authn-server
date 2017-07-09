package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssertData(t *testing.T, res *http.Response, expected interface{}) {
	assert.Equal(t, []string{"application/json"}, res.Header["Content-Type"])

	j, err := json.Marshal(api.ServiceData{expected})
	require.NoError(t, err)
	assert.Equal(t, string(j), string(ReadBody(res)))
}

func AssertErrors(t *testing.T, res *http.Response, expected []services.Error) {
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

	segments := strings.Split(session, ".")
	assert.Len(t, segments, 3)

	_, err := jwt.Parse(session, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}
		return []byte("TODO"), nil
	})
	assert.NoError(t, err)
}

func AssertIdTokenResponse(t *testing.T, res *http.Response, cfg *config.Config) {
	// check that the response contains the expected json
	assert.Equal(t, []string{"application/json"}, res.Header["Content-Type"])
	responseData := struct {
		IdToken string `json:"id_token"`
	}{}
	err := extractResult(res, &responseData)
	assert.NoError(t, err)

	// check that the IdToken is JWT-ish
	identityToken, err := jwt.Parse(responseData.IdToken, func(tkn *jwt.Token) (interface{}, error) {
		return cfg.IdentitySigningKey.Public(), nil
	})
	if assert.NoError(t, err) {
		// check that the JWT contains nice things
		assert.Equal(t, cfg.AuthNURL.String(), identityToken.Claims.(jwt.MapClaims)["iss"])
	}
}

// extracts the value from inside a successful result envelope. must be provided
// with `inner`, an empty struct that describes the expected (desired) shape of
// what is inside the envelope.
func extractResult(res *http.Response, inner interface{}) error {
	return json.Unmarshal(ReadBody(res), &api.ServiceData{inner})
}
