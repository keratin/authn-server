package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keratin/authn/data/sqlite3"
)

func TestHealth(t *testing.T) {
	db, err := sqlite3.TempDB()
	if err != nil {
		panic(err)
	}

	app := testApp()
	app.Db = *db

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	handler := http.HandlerFunc(app.Health)
	handler.ServeHTTP(res, req)

	assertCode(t, res, http.StatusOK)
	assertBody(t, res, `{"http":true,"db":true,"redis":true}`)
}
