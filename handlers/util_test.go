package handlers_test

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/handlers"
	"github.com/keratin/authn-server/services"
	"github.com/keratin/authn-server/tokens/sessions"
	"github.com/stretchr/testify/assert"
)

type HandlerFuncable func(w http.ResponseWriter, r *http.Request)
type ReqModder func(req *http.Request) *http.Request

func post(path string, h HandlerFuncable, params map[string]string, befores ...ReqModder) *httptest.ResponseRecorder {
	buffer := make([]string, 0)
	for k, v := range params {
		buffer = append(buffer, strings.Join([]string{k, v}, "="))
	}
	paramsStr := strings.Join(buffer, "&")

	res := httptest.NewRecorder()
	req := httptest.NewRequest("POST", path, strings.NewReader(paramsStr))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	for _, before := range befores {
		req = before(req)
	}

	http.HandlerFunc(h).ServeHTTP(res, req)
	return res
}

func get(path string, h HandlerFuncable, befores ...ReqModder) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	for _, before := range befores {
		req = before(req)
	}

	http.HandlerFunc(h).ServeHTTP(res, req)
	return res
}

func testApp() handlers.App {
	accountStore := mock.NewAccountStore()

	authnUrl, err := url.Parse("https://authn.example.com")
	if err != nil {
		panic(err)
	}

	weakKey, err := rsa.GenerateKey(rand.Reader, 512)
	if err != nil {
		panic(err)
	}

	cfg := config.Config{
		BcryptCost:         4,
		SessionSigningKey:  []byte("TODO"),
		IdentitySigningKey: weakKey,
		AuthNURL:           authnUrl,
		SessionCookieName:  "authn",
	}

	tokenStore := mock.NewRefreshTokenStore()

	return handlers.App{
		AccountStore:      accountStore,
		RefreshTokenStore: tokenStore,
		Config:            &cfg,
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

func assertCode(t *testing.T, rr *httptest.ResponseRecorder, expected int) {
	assert.Equal(t, expected, rr.Code)
}

func assertBody(t *testing.T, rr *httptest.ResponseRecorder, expected string) {
	assert.Equal(t, expected, rr.Body.String())
}

func assertErrors(t *testing.T, rr *httptest.ResponseRecorder, expected []services.Error) {
	assert.Equal(t, []string{"application/json"}, rr.HeaderMap["Content-Type"])

	j, err := json.Marshal(handlers.ServiceErrors{Errors: expected})
	if err != nil {
		panic(err)
	}

	assertBody(t, rr, string(j))
}

func assertSession(t *testing.T, rr *httptest.ResponseRecorder) {
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

func assertIdTokenResponse(t *testing.T, rr *httptest.ResponseRecorder, cfg *config.Config) {
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

// extracts the value from inside a successful result envelope. must be provided
// with `inner`, an empty struct that describes the expected (desired) shape of
// what is inside the envelope.
func extractResult(response *httptest.ResponseRecorder, inner interface{}) error {
	return json.Unmarshal([]byte(response.Body.String()), &handlers.ServiceData{inner})
}

func createSession(tokenStore data.RefreshTokenStore, cfg *config.Config, account_id int) *http.Cookie {
	sessionToken, err := sessions.New(tokenStore, cfg, account_id)
	if err != nil {
		panic(err)
	}

	sessionString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, sessionToken).SignedString(cfg.SessionSigningKey)
	if err != nil {
		panic(err)
	}

	return &http.Cookie{
		Name:  cfg.SessionCookieName,
		Value: sessionString,
	}
}
