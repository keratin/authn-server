package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/handlers"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/tokens/sessions"
)

func TestSessionSecurity(t *testing.T) {
	app := &handlers.App{
		Config: &config.Config{
			AuthNURL:          &url.URL{Scheme: "https", Path: "www.example.com"},
			SessionCookieName: "authn-test",
			SessionSigningKey: []byte("good"),
		},
		RefreshTokenStore: mock.NewRefreshTokenStore(),
	}

	expectedSuccessHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		account_id := r.Context().Value(handlers.AccountIDKey).(int)
		w.Write([]byte(strconv.Itoa(account_id)))
	})
	expectedFailureHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("handler should not be called"))
		w.WriteHeader(http.StatusInternalServerError)
	})

	testTable := []struct {
		cookieName string
		signingKey []byte
		liveToken  bool
		success    bool
	}{
		// cookie with correct name and current refresh token
		{app.Config.SessionCookieName, app.Config.SessionSigningKey, true, true},
		// cookie with the wrong name
		{"wrong", app.Config.SessionSigningKey, true, false},
		// cookie with the wrong signature
		{app.Config.SessionCookieName, []byte("wrong"), true, false},
		// cookie with a revoked refresh token
		{app.Config.SessionCookieName, app.Config.SessionSigningKey, false, false},
	}

	for idx, tt := range testTable {
		tt_cfg := &config.Config{
			AuthNURL:          app.Config.AuthNURL,
			SessionCookieName: tt.cookieName,
			SessionSigningKey: tt.signingKey,
		}
		existingSession := createSession(app.RefreshTokenStore, tt_cfg, idx+100)
		if !tt.liveToken {
			revokeSession(app.RefreshTokenStore, app.Config, existingSession)
		}

		adapter := handlers.SessionSecurity(app)
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", strings.NewReader(""))
		req.AddCookie(existingSession)
		if tt.success {
			adapter(expectedSuccessHandler).ServeHTTP(res, req)
			assertCode(t, res, http.StatusOK)
			assertBody(t, res, strconv.Itoa(idx+100))
		} else {
			adapter(expectedFailureHandler).ServeHTTP(res, req)
			assertCode(t, res, http.StatusUnauthorized)
		}
	}
}

func revokeSession(store data.RefreshTokenStore, cfg *config.Config, session *http.Cookie) {
	claims, err := sessions.Parse(session.Value, cfg)
	if err != nil {
		panic(err)
	}
	err = store.Revoke(models.RefreshToken(claims.Subject))
	if err != nil {
		panic(err)
	}
}
