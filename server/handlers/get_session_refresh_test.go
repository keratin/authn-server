package handlers_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/app/data/redis"
	"github.com/keratin/authn-server/app/data/sqlite3"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/ops"
	"github.com/keratin/authn-server/server/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkGetSessionRefresh(b *testing.B) {
	verify := func(res *http.Response, err error) {
		if err != nil {
			panic(err)
		}
		if res.StatusCode != http.StatusCreated {
			panic(res.StatusCode)
		}
	}

	b.Run("mock    store", func(b *testing.B) {
		testApp := test.App()
		server := test.Server(testApp)
		defer server.Close()
		testApp.RefreshTokenStore = mock.NewRefreshTokenStore()
		client := route.NewClient(server.URL).
			Referred(&testApp.Config.ApplicationDomains[0]).
			WithCookie(test.CreateSession(testApp.RefreshTokenStore, testApp.Config, 12345))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			verify(client.Get("/session/refresh"))
		}
	})

	sqliteDB, err := sqlite3.NewDB("benchmarks")
	if err != nil {
		panic(err)
	}
	sqlite3.MigrateDB(sqliteDB)

	b.Run("sqlite3 store", func(b *testing.B) {
		testApp := test.App()
		server := test.Server(testApp)
		defer server.Close()
		testApp.RefreshTokenStore = &sqlite3.RefreshTokenStore{sqliteDB, time.Hour}
		client := route.NewClient(server.URL).
			Referred(&testApp.Config.ApplicationDomains[0]).
			WithCookie(test.CreateSession(testApp.RefreshTokenStore, testApp.Config, 12345))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			verify(client.Get("/session/refresh"))
		}
	})

	redisDB, err := redis.TestDB()
	if err != nil {
		panic(err)
	}

	b.Run("redis   store", func(b *testing.B) {
		testApp := test.App()
		server := test.Server(testApp)
		defer server.Close()
		testApp.RefreshTokenStore = &redis.RefreshTokenStore{redisDB, time.Hour}
		client := route.NewClient(server.URL).
			Referred(&testApp.Config.ApplicationDomains[0]).
			WithCookie(test.CreateSession(testApp.RefreshTokenStore, testApp.Config, 12345))

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			verify(client.Get("/session/refresh"))
		}
	})
}

func TestGetSessionRefreshSuccess(t *testing.T) {
	testApp := test.App()
	server := test.Server(testApp)
	defer server.Close()

	accountID := 82594
	existingSession := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, accountID)

	client := route.NewClient(server.URL).Referred(&testApp.Config.ApplicationDomains[0]).WithCookie(existingSession)
	res, err := client.Get("/session/refresh")
	require.NoError(t, err)

	if assert.Equal(t, http.StatusCreated, res.StatusCode) {
		test.AssertIDTokenResponse(t, res, testApp.KeyStore, testApp.Config)
	}
}

func TestGetSessionRefreshFailure(t *testing.T) {
	testApp := &app.App{
		Config: &app.Config{
			AuthNURL:           &url.URL{Scheme: "https", Path: "www.example.com"},
			SessionCookieName:  "authn-test",
			SessionSigningKey:  []byte("good"),
			ApplicationDomains: []route.Domain{{Hostname: "test.com"}},
		},
		RefreshTokenStore: mock.NewRefreshTokenStore(),
		Reporter:          &ops.LogReporter{logrus.New()},
		Logger:            logrus.New(),
	}
	server := test.Server(testApp)
	defer server.Close()

	testCases := []struct {
		signingKey []byte
		liveToken  bool
	}{
		// cookie with the wrong signature
		{[]byte("wrong"), true},
		// cookie with a revoked refresh token
		{testApp.Config.SessionSigningKey, false},
	}

	for idx, tc := range testCases {
		tcCfg := &app.Config{
			AuthNURL:           testApp.Config.AuthNURL,
			SessionCookieName:  testApp.Config.SessionCookieName,
			SessionSigningKey:  tc.signingKey,
			ApplicationDomains: []route.Domain{{Hostname: "test.com"}},
		}
		existingSession := test.CreateSession(testApp.RefreshTokenStore, tcCfg, idx+100)
		if !tc.liveToken {
			test.RevokeSession(testApp.RefreshTokenStore, testApp.Config, existingSession)
		}

		client := route.NewClient(server.URL).Referred(&testApp.Config.ApplicationDomains[0]).WithCookie(existingSession)
		res, err := client.Get("/session/refresh")
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	}
}
