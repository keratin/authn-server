package sessions_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data/mock"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/ops"
	"github.com/keratin/authn-server/server/sessions"
	"github.com/keratin/authn-server/server/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession(t *testing.T) {
	testApp := &app.App{
		Config: &app.Config{
			SessionCookieName:  "authn-test",
			SessionSigningKey:  []byte("drinkme"),
			AuthNURL:           &url.URL{Scheme: "http", Host: "authn.example.com"},
			ApplicationDomains: []route.Domain{{Hostname: "example.com"}},
		},
		RefreshTokenStore: mock.NewRefreshTokenStore(),
		Reporter:          &ops.LogReporter{logrus.New()},
	}

	t.Run("valid session", func(t *testing.T) {
		accountID := 60090
		session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, accountID)

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, sessions.Get(r))
			assert.Equal(t, accountID, sessions.GetAccountID(r))

			w.WriteHeader(http.StatusOK)
		}
		server := httptest.NewServer(sessions.Middleware(testApp)(http.HandlerFunc(handler)))
		defer server.Close()

		client := route.NewClient(server.URL).WithCookie(session)
		res, err := client.Get("/")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("invalid session", func(t *testing.T) {
		oldConfig := &app.Config{
			SessionCookieName:  testApp.Config.SessionCookieName,
			SessionSigningKey:  []byte("previouskey"),
			AuthNURL:           testApp.Config.AuthNURL,
			ApplicationDomains: testApp.Config.ApplicationDomains,
		}
		accountID := 52444
		session := test.CreateSession(testApp.RefreshTokenStore, oldConfig, accountID)

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Empty(t, sessions.Get(r))
			assert.Empty(t, sessions.GetAccountID(r))

			w.WriteHeader(http.StatusOK)
		}
		server := httptest.NewServer(sessions.Middleware(testApp)(http.HandlerFunc(handler)))
		defer server.Close()

		client := route.NewClient(server.URL).WithCookie(session)
		res, err := client.Get("/")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("revoked session", func(t *testing.T) {
		accountID := 10001
		session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, accountID)
		test.RevokeSession(testApp.RefreshTokenStore, testApp.Config, session)

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, sessions.Get(r))
			assert.Empty(t, sessions.GetAccountID(r))

			w.WriteHeader(http.StatusOK)
		}
		server := httptest.NewServer(sessions.Middleware(testApp)(http.HandlerFunc(handler)))
		defer server.Close()

		client := route.NewClient(server.URL).WithCookie(session)
		res, err := client.Get("/")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("missing session", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Empty(t, sessions.Get(r))
			assert.Empty(t, sessions.GetAccountID(r))

			w.WriteHeader(http.StatusOK)
		}
		server := httptest.NewServer(sessions.Middleware(testApp)(http.HandlerFunc(handler)))
		defer server.Close()

		client := route.NewClient(server.URL)
		res, err := client.Get("/")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}
