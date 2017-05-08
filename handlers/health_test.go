package handlers_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keratin/authn/handlers"

	_ "github.com/mattn/go-sqlite3"
)

func TestHealth(t *testing.T) {
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	app := handlers.App{Db: *db}

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	handler := http.HandlerFunc(app.Health)
	handler.ServeHTTP(res, req)

	AssertCode(t, res, http.StatusOK)
	AssertBody(t, res, `{"http":true,"db":true}`)
}
