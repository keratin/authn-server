package handlers_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/handlers"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/tokens/sessions"
)

func TestGetSessionRefreshSuccess(t *testing.T) {
	app := testApp()

	account_id := 82594
	existingSession := createSession(app.RefreshTokenStore, app.Config, account_id)

	res := get("/session/refresh", app.GetSessionRefresh, withSession(existingSession))

	assertCode(t, res, http.StatusCreated)
	assertIdTokenResponse(t, res, app.Config)
}

func TestGetSessionRefreshFailure(t *testing.T) {
	app := &handlers.App{
		Config: &config.Config{
			AuthNURL:          &url.URL{Scheme: "https", Path: "www.example.com"},
			SessionCookieName: "authn-test",
			SessionSigningKey: []byte("good"),
		},
		RefreshTokenStore: mock.NewRefreshTokenStore(),
	}

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
		tc_cfg := &config.Config{
			AuthNURL:          app.Config.AuthNURL,
			SessionCookieName: app.Config.SessionCookieName,
			SessionSigningKey: tc.signingKey,
		}
		existingSession := createSession(app.RefreshTokenStore, tc_cfg, idx+100)
		if !tc.liveToken {
			revokeSession(app.RefreshTokenStore, app.Config, existingSession)
		}

		res := get("/session/refresh", app.GetSessionRefresh, withSession(existingSession))

		assertCode(t, res, http.StatusUnauthorized)
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
