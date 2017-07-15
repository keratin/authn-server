package sessions_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/keratin/authn-server/api"
	apiSessions "github.com/keratin/authn-server/api/sessions"
	"github.com/keratin/authn-server/api/test"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/data/mock"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/tokens/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSessionRefreshSuccess(t *testing.T) {
	app := test.App()
	server := test.Server(app, apiSessions.Routes(app))
	defer server.Close()

	accountId := 82594
	existingSession := test.CreateSession(app.RefreshTokenStore, app.Config, accountId)

	client := test.NewClient(server).Referred(app.Config).WithSession(existingSession)
	res, err := client.Get("/session/refresh")
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	test.AssertIdTokenResponse(t, res, app.Config)
}

func TestGetSessionRefreshFailure(t *testing.T) {
	app := &api.App{
		Config: &config.Config{
			AuthNURL:           &url.URL{Scheme: "https", Path: "www.example.com"},
			SessionCookieName:  "authn-test",
			SessionSigningKey:  []byte("good"),
			ApplicationDomains: []string{"test.com"},
		},
		RefreshTokenStore: mock.NewRefreshTokenStore(),
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
			AuthNURL:          app.Config.AuthNURL,
			SessionCookieName: app.Config.SessionCookieName,
			SessionSigningKey: tc.signingKey,
		}
		existingSession := test.CreateSession(app.RefreshTokenStore, tcCfg, idx+100)
		if !tc.liveToken {
			revokeSession(app.RefreshTokenStore, app.Config, existingSession)
		}

		client := test.NewClient(server).Referred(app.Config).WithSession(existingSession)
		res, err := client.Get("/session/refresh")
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
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
