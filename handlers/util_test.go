package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/keratin/authn/config"
	dataRedis "github.com/keratin/authn/data/redis"
	"github.com/keratin/authn/data/sqlite3"
	"github.com/keratin/authn/handlers"
	"github.com/keratin/authn/services"
)

func testApp() handlers.App {
	db, err := sqlite3.TempDB()
	if err != nil {
		panic(err)
	}
	store := sqlite3.AccountStore{db}

	cfg := config.Config{
		BcryptCost:        4,
		SessionSigningKey: []byte("TODO"),
	}

	opts, err := redis.ParseURL("redis://127.0.0.1:6379/12")
	if err != nil {
		panic(err)
	}
	redis := redis.NewClient(opts)

	tokenStore := dataRedis.RefreshTokenStore{Client: redis, TTL: time.Minute}

	return handlers.App{
		Db:                *db,
		Redis:             redis,
		AccountStore:      &store,
		RefreshTokenStore: &tokenStore,
		Config:            cfg,
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
	status := rr.Code
	if status != expected {
		t.Errorf("HTTP status:\n  expected: %v\n  actual:   %v", expected, status)
	}
}

func assertBody(t *testing.T, rr *httptest.ResponseRecorder, expected string) {
	if rr.Body.String() != expected {
		t.Errorf("HTTP body:\n  expected: %v\n  actual:   %v", expected, rr.Body.String())
	}
}

func assertErrors(t *testing.T, rr *httptest.ResponseRecorder, expected []services.Error) {
	j, err := json.Marshal(handlers.ServiceErrors{Errors: expected})
	if err != nil {
		panic(err)
	}

	assertBody(t, rr, string(j))
}

func assertResult(t *testing.T, rr *httptest.ResponseRecorder, expected interface{}) {
	j, err := json.Marshal(handlers.ServiceData{expected})
	if err != nil {
		panic(err)
	}

	assertBody(t, rr, string(j))
}

func assertSession(t *testing.T, rr *httptest.ResponseRecorder) {
	session, err := readSetCookieValue("authn", rr)
	if err != nil {
		t.Error(err)
	}

	segments := strings.Split(session, ".")
	if len(segments) != 3 {
		t.Error("expected JWT with three segments, got: %v", session)
	}

	_, err = jwt.Parse(session, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}
		return []byte("TODO"), err
	})
	if err != nil {
		t.Error(err)
	}
}
