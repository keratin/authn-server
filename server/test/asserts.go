package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	jwt "github.com/go-jose/go-jose/v3/jwt"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/app/tokens/identities"
	"github.com/keratin/authn-server/app/tokens/sessions"
	"github.com/keratin/authn-server/server/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssertData(t *testing.T, res *http.Response, expected interface{}) {
	t.Helper()
	assert.Equal(t, []string{"application/json"}, res.Header["Content-Type"])

	j, err := json.Marshal(handlers.ServiceData{Result: expected})
	require.NoError(t, err)
	assert.Equal(t, string(j), string(ReadBody(res)))
}

func AssertErrors(t *testing.T, res *http.Response, expected services.FieldErrors) {
	t.Helper()
	assert.Equal(t, []string{"application/json"}, res.Header["Content-Type"])

	j, err := json.Marshal(handlers.ServiceErrors{Errors: expected})
	require.NoError(t, err)
	assert.Equal(t, string(j), string(ReadBody(res)))
}

func AssertSession(t *testing.T, cfg *app.Config, cookies []*http.Cookie, expectedAMR ...string) {
	t.Helper()
	session := ReadCookie(cookies, cfg.SessionCookieName)
	require.NotEmpty(t, session)

	claims, err := sessions.Parse(session.Value, cfg)
	assert.NoError(t, err)

	if len(expectedAMR) > 0 {
		assert.True(t, reflect.DeepEqual(claims.AuthMethodReference, expectedAMR), fmt.Sprintf("expected %v got %v", expectedAMR, claims.AuthMethodReference))
	}
}

func AssertIDTokenResponse(t *testing.T, res *http.Response, keyStore data.KeyStore, cfg *app.Config, expectedAMR ...string) {
	t.Helper()

	// check that the response contains the expected json
	assert.Equal(t, []string{"application/json"}, res.Header["Content-Type"])
	responseData := struct {
		IDToken string `json:"id_token"`
	}{}
	err := ExtractResult(res, &responseData)
	assert.NoError(t, err)

	tok, err := jwt.ParseSigned(responseData.IDToken)
	assert.NoError(t, err)

	claims := identities.Claims{}
	err = tok.Claims(keyStore.Key().Public(), &claims)
	if assert.NoError(t, err) {
		// check that the JWT contains nice things
		assert.Equal(t, cfg.AuthNURL.String(), claims.Issuer)
		if len(expectedAMR) > 0 {
			assert.True(t, reflect.DeepEqual(claims.AuthMethodReference, expectedAMR), fmt.Sprintf("expected %v got %v", expectedAMR, claims.AuthMethodReference))
		}
	}
}

func AssertRedirect(t *testing.T, res *http.Response, location string) bool {
	t.Helper()

	assert.Equal(t, http.StatusSeeOther, res.StatusCode)
	loc, err := res.Location()
	require.NoError(t, err)
	return assert.Equal(t, location, loc.String())
}
