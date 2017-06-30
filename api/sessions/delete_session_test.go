package sessions_test

import (
	"net/http"
	"testing"

	apiSessions "github.com/keratin/authn-server/api/sessions"
	"github.com/keratin/authn-server/api/test"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/tokens/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteSessionSuccess(t *testing.T) {
	app := test.App()
	account_id := 514628

	session := test.CreateSession(app.RefreshTokenStore, app.Config, account_id)

	// token exists
	claims, err := sessions.Parse(session.Value, app.Config)
	require.NoError(t, err)
	id, err := app.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	res := test.Delete("/session", apiSessions.DeleteSession(app), test.WithSession(session))

	// request always succeeds
	test.AssertCode(t, res, http.StatusOK)

	// token no longer exists
	id, err = app.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestDeleteSessionFailure(t *testing.T) {
	app := test.App()

	bad_config := &config.Config{
		AuthNURL:          app.Config.AuthNURL,
		SessionCookieName: app.Config.SessionCookieName,
		SessionSigningKey: []byte("wrong"),
	}
	session := test.CreateSession(app.RefreshTokenStore, bad_config, 123)

	res := test.Delete("/session", apiSessions.DeleteSession(app), test.WithSession(session))
	test.AssertCode(t, res, http.StatusOK)
}

func TestDeleteSessionWithoutSession(t *testing.T) {
	app := test.App()

	res := test.Delete("/session", apiSessions.DeleteSession(app))
	test.AssertCode(t, res, http.StatusOK)
}
