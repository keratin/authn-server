package sessions_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/keratin/authn-server/api"
	apiSessions "github.com/keratin/authn-server/api/sessions"
	"github.com/keratin/authn-server/api/test"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/data/redis"
	"github.com/keratin/authn-server/data/sqlite3"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/ops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkGetSessionRefresh(b *testing.B) {
	b.Run("mock    store", func(b *testing.B) {
		app := test.App()
		server := test.Server(app, apiSessions.Routes(app))
		defer server.Close()
		app.RefreshTokenStore = mock.NewRefreshTokenStore()
		client := route.NewClient(server.URL).
			Referred(&app.Config.ApplicationDomains[0]).
			WithCookie(test.CreateSession(app.RefreshTokenStore, app.Config, 12345))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			client.Get("/session/refresh")
		}
	})

	sqliteDB, err := sqlite3.NewDB("benchmarks")
	if err != nil {
		panic(err)
	}
	sqlite3.MigrateDB(sqliteDB)

	b.Run("sqlite3 store", func(b *testing.B) {
		app := test.App()
		server := test.Server(app, apiSessions.Routes(app))
		defer server.Close()
		app.RefreshTokenStore = &sqlite3.RefreshTokenStore{sqliteDB, time.Hour}
		client := route.NewClient(server.URL).
			Referred(&app.Config.ApplicationDomains[0]).
			WithCookie(test.CreateSession(app.RefreshTokenStore, app.Config, 12345))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			client.Get("/session/refresh")
		}
	})

	redisDB, err := redis.TestDB()
	if err != nil {
		panic(err)
	}

	b.Run("redis   store", func(b *testing.B) {
		app := test.App()
		server := test.Server(app, apiSessions.Routes(app))
		defer server.Close()
		app.RefreshTokenStore = &redis.RefreshTokenStore{redisDB, time.Hour}
		client := route.NewClient(server.URL).
			Referred(&app.Config.ApplicationDomains[0]).
			WithCookie(test.CreateSession(app.RefreshTokenStore, app.Config, 12345))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			client.Get("/session/refresh")
		}
	})
}

func TestGetSessionRefreshSuccess(t *testing.T) {
	app := test.App()
	server := test.Server(app, apiSessions.Routes(app))
	defer server.Close()

	accountID := 82594
	existingSession := test.CreateSession(app.RefreshTokenStore, app.Config, accountID)

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0]).WithCookie(existingSession)
	res, err := client.Get("/session/refresh")
	require.NoError(t, err)

	if assert.Equal(t, http.StatusCreated, res.StatusCode) {
		test.AssertIDTokenResponse(t, res, app.KeyStore, app.Config)
	}
}

func TestGetSessionRefreshFailure(t *testing.T) {
	app := &api.App{
		Config: &config.Config{
			AuthNURL:           &url.URL{Scheme: "https", Path: "www.example.com"},
			SessionCookieName:  "authn-test",
			SessionSigningKey:  []byte("good"),
			ApplicationDomains: []route.Domain{{Hostname: "test.com"}},
		},
		RefreshTokenStore: mock.NewRefreshTokenStore(),
		Reporter:          &ops.LogReporter{},
	}
	server := test.Server(app, apiSessions.Routes(app))
	defer server.Close()

	testCases := []struct {
		signingKey []byte
		liveToken  bool
	}{
		// cookie with the wrong signature
		{[]byte("wrong"), true},
		// cookie with a revoked refresh token
		{app.Config.SessionSigningKey, false},
	}

	for idx, tc := range testCases {
		tcCfg := &config.Config{
			AuthNURL:           app.Config.AuthNURL,
			SessionCookieName:  app.Config.SessionCookieName,
			SessionSigningKey:  tc.signingKey,
			ApplicationDomains: []route.Domain{{Hostname: "test.com"}},
		}
		existingSession := test.CreateSession(app.RefreshTokenStore, tcCfg, idx+100)
		if !tc.liveToken {
			test.RevokeSession(app.RefreshTokenStore, app.Config, existingSession)
		}

		client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0]).WithCookie(existingSession)
		res, err := client.Get("/session/refresh")
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	}
}
