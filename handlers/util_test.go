package handlers_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/keratin/authn/config"
	"github.com/keratin/authn/data/sqlite3"
	"github.com/keratin/authn/handlers"
	"github.com/keratin/authn/services"
)

func App() handlers.App {
	db, err := sqlite3.TempDB()
	if err != nil {
		panic(err)
	}
	store := sqlite3.AccountStore{db}

	cfg := config.Config{
		BcryptCost: 4,
	}

	return handlers.App{Db: *db, AccountStore: store, Config: cfg}
}

func AssertCode(t *testing.T, rr *httptest.ResponseRecorder, expected int) {
	status := rr.Code
	if status != expected {
		t.Errorf("HTTP status:\n  expected: %v\n  actual:   %v", expected, status)
	}
}

func AssertBody(t *testing.T, rr *httptest.ResponseRecorder, expected string) {
	if rr.Body.String() != expected {
		t.Errorf("HTTP body:\n  expected: %v\n  actual:   %v", expected, rr.Body.String())
	}
}

func AssertErrors(t *testing.T, rr *httptest.ResponseRecorder, expected []services.Error) {
	j, err := json.Marshal(handlers.ServiceErrors{Errors: expected})
	if err != nil {
		panic(err)
	}

	AssertBody(t, rr, string(j))
}

func AssertResult(t *testing.T, rr *httptest.ResponseRecorder, expected interface{}) {
	j, err := json.Marshal(handlers.ServiceData{expected})
	if err != nil {
		panic(err)
	}

	AssertBody(t, rr, string(j))
}
