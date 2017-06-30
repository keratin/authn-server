package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
)

func AssertCode(t *testing.T, rr *httptest.ResponseRecorder, expected int) {
	assert.Equal(t, expected, rr.Code)
}

func AssertBody(t *testing.T, rr *httptest.ResponseRecorder, expected string) {
	assert.Equal(t, expected, rr.Body.String())
}

func AssertErrors(t *testing.T, rr *httptest.ResponseRecorder, expected []services.Error) {
	assert.Equal(t, []string{"application/json"}, rr.HeaderMap["Content-Type"])

	j, err := json.Marshal(api.ServiceErrors{Errors: expected})
	if err != nil {
		panic(err)
	}

	AssertBody(t, rr, string(j))
}

func AssertSession(t *testing.T, rr *httptest.ResponseRecorder) {
	session, err := readSetCookieValue("authn", rr)
	assert.NoError(t, err)

	segments := strings.Split(session, ".")
	assert.Len(t, segments, 3)

	_, err = jwt.Parse(session, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}
		return []byte("TODO"), err
	})
	assert.NoError(t, err)
}

func AssertIdTokenResponse(t *testing.T, rr *httptest.ResponseRecorder, cfg *config.Config) {
	// check that the response contains the expected json
	assert.Equal(t, []string{"application/json"}, rr.HeaderMap["Content-Type"])
	responseData := struct {
		IdToken string `json:"id_token"`
	}{}
	err := extractResult(rr, &responseData)
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

// apparently you can't fully restore a Cookie from the Set-Cookie header without
// in-depth parsing hijinx like in net/http/cookie.go's readSetCookies.
//
// you can't even partially restore a Cookie without going through a new Request:
// http://jonnyreeves.co.uk/2016/testing-setting-http-cookies-in-go/
func readSetCookieValue(name string, recorder *httptest.ResponseRecorder) (string, error) {
	request := http.Request{
		Header: http.Header{"Cookie": recorder.HeaderMap["Set-Cookie"]},
	}
	cookie, err := request.Cookie(name)
	if err != nil {
		return "", err
	} else {
		return cookie.Value, nil
	}
}

// extracts the value from inside a successful result envelope. must be provided
// with `inner`, an empty struct that describes the expected (desired) shape of
// what is inside the envelope.
func extractResult(response *httptest.ResponseRecorder, inner interface{}) error {
	return json.Unmarshal([]byte(response.Body.String()), &api.ServiceData{inner})
}
