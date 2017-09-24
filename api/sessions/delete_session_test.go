package sessions_test

import (
	"net/http"
	"testing"

	apiSessions "github.com/keratin/authn-server/api/sessions"
	"github.com/keratin/authn-server/api/test"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/tokens/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteSessionSuccess(t *testing.T) {
	app := test.App()
	server := test.Server(app, apiSessions.Routes(app))
	defer server.Close()

	accountID := 514628
	session := test.CreateSession(app.RefreshTokenStore, app.Config, accountID)

	// token exists
	claims, err := sessions.Parse(session.Value, app.Config)
	require.NoError(t, err)
	id, err := app.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0]).WithCookie(session)
	res, err := client.Delete("/session")
	require.NoError(t, err)

	// request always succeeds
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// token no longer exists
	id, err = app.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestDeleteSessionFailure(t *testing.T) {
	app := test.App()
	server := test.Server(app, apiSessions.Routes(app))
	defer server.Close()

	badCfg := &config.Config{
		AuthNURL:           app.Config.AuthNURL,
		SessionCookieName:  app.Config.SessionCookieName,
		SessionSigningKey:  []byte("wrong"),
		ApplicationDomains: app.Config.ApplicationDomains,
	}
	session := test.CreateSession(app.RefreshTokenStore, badCfg, 123)

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0]).WithCookie(session)
	res, err := client.Delete("/session")
	require.NoError(t, err)

	// request still succeeds
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestDeleteSessionWithoutSession(t *testing.T) {
	app := test.App()
	server := test.Server(app, apiSessions.Routes(app))
	defer server.Close()

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0])
	res, err := client.Delete("/session")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
}
