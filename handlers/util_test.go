package handlers_test

import (
	"database/sql"
	"net/http/httptest"
	"testing"

	"github.com/keratin/authn/handlers"

	_ "github.com/mattn/go-sqlite3"
)

func App() handlers.App {
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	return handlers.App{Db: *db}
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
