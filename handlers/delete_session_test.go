package handlers_test

import (
	"net/http"
	"testing"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/handlers"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/tokens/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteSessionSuccess(t *testing.T) {
	app := testApp()
	account_id := 514628

	session := createSession(app.RefreshTokenStore, app.Config, account_id)

	// token exists
	claims, err := sessions.Parse(session.Value, app.Config)
	require.NoError(t, err)
	id, err := app.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	res := delete("/session", handlers.DeleteSession(app), withSession(session))

	// request always succeeds
	assertCode(t, res, http.StatusOK)

	// token no longer exists
	id, err = app.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestDeleteSessionFailure(t *testing.T) {
	app := testApp()

	bad_config := &config.Config{
		AuthNURL:          app.Config.AuthNURL,
		SessionCookieName: app.Config.SessionCookieName,
		SessionSigningKey: []byte("wrong"),
	}
	session := createSession(app.RefreshTokenStore, bad_config, 123)

	res := delete("/session", handlers.DeleteSession(app), withSession(session))
	assertCode(t, res, http.StatusOK)

	res = delete("/session", handlers.DeleteSession(app))
	assertCode(t, res, http.StatusOK)
}

func TestDeleteSessionWithoutSession(t *testing.T) {
	app := testApp()

	res := delete("/session", handlers.DeleteSession(app))
	assertCode(t, res, http.StatusOK)
}
