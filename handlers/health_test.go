package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-redis/redis"
	"github.com/keratin/authn/data/sqlite3"
)

func TestHealth(t *testing.T) {
	db, err := sqlite3.TempDB()
	if err != nil {
		panic(err)
	}

	opts, err := redis.ParseURL("redis://127.0.0.1:6379/12")
	if err != nil {
		panic(err)
	}
	redis := redis.NewClient(opts)

	app := testApp()
	app.Db = *db
	app.Redis = redis

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)

	handler := http.HandlerFunc(app.Health)
	handler.ServeHTTP(res, req)

	assertCode(t, res, http.StatusOK)
	assertBody(t, res, `{"http":true,"db":true,"redis":true}`)
}
