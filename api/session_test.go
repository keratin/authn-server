package api_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/api/test"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/ops"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession(t *testing.T) {
	app := &api.App{
		Config: &config.Config{
			SessionCookieName:  "authn-test",
			SessionSigningKey:  []byte("drinkme"),
			AuthNURL:           &url.URL{Scheme: "http", Host: "authn.example.com"},
			ApplicationDomains: []route.Domain{{Hostname: "example.com"}},
		},
		RefreshTokenStore: mock.NewRefreshTokenStore(),
		Reporter:          &ops.LogReporter{},
	}

	t.Run("valid session", func(t *testing.T) {
		accountID := 60090
		session := test.CreateSession(app.RefreshTokenStore, app.Config, accountID)

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, api.GetSession(r))
			assert.Equal(t, accountID, api.GetSessionAccountID(r))

			w.WriteHeader(http.StatusOK)
		}
		server := httptest.NewServer(api.Session(app)(http.HandlerFunc(handler)))
		defer server.Close()

		client := test.NewClient(server).WithSession(session)
		res, err := client.Get("/")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("invalid session", func(t *testing.T) {
		oldConfig := &config.Config{
			SessionCookieName:  app.Config.SessionCookieName,
			SessionSigningKey:  []byte("previouskey"),
			AuthNURL:           app.Config.AuthNURL,
			ApplicationDomains: app.Config.ApplicationDomains,
		}
		accountID := 52444
		session := test.CreateSession(app.RefreshTokenStore, oldConfig, accountID)

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Empty(t, api.GetSession(r))
			assert.Empty(t, api.GetSessionAccountID(r))

			w.WriteHeader(http.StatusOK)
		}
		server := httptest.NewServer(api.Session(app)(http.HandlerFunc(handler)))
		defer server.Close()

		client := test.NewClient(server).WithSession(session)
		res, err := client.Get("/")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("revoked session", func(t *testing.T) {
		accountID := 10001
		session := test.CreateSession(app.RefreshTokenStore, app.Config, accountID)
		test.RevokeSession(app.RefreshTokenStore, app.Config, session)

		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.NotEmpty(t, api.GetSession(r))
			assert.Empty(t, api.GetSessionAccountID(r))

			w.WriteHeader(http.StatusOK)
		}
		server := httptest.NewServer(api.Session(app)(http.HandlerFunc(handler)))
		defer server.Close()

		client := test.NewClient(server).WithSession(session)
		res, err := client.Get("/")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("missing session", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) {
			assert.Empty(t, api.GetSession(r))
			assert.Empty(t, api.GetSessionAccountID(r))

			w.WriteHeader(http.StatusOK)
		}
		server := httptest.NewServer(api.Session(app)(http.HandlerFunc(handler)))
		defer server.Close()

		client := test.NewClient(server)
		res, err := client.Get("/")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}
