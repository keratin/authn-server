package handlers_test

import (
	"net/http"
	"testing"

	"github.com/keratin/authn-server/server/test"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/app/models"
	"github.com/keratin/authn-server/app/tokens/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteSessionSuccess(t *testing.T) {
	testApp := test.App()
	server := test.Server(testApp)
	defer server.Close()

	accountID := 514628
	session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, accountID)

	// token exists
	claims, err := sessions.Parse(session.Value, testApp.Config)
	require.NoError(t, err)
	id, err := testApp.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	client := route.NewClient(server.URL).Referred(&testApp.Config.ApplicationDomains[0]).WithCookie(session)
	res, err := client.Delete("/session")
	require.NoError(t, err)

	// request always succeeds
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// token no longer exists
	id, err = testApp.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestDeleteSessionFailure(t *testing.T) {
	testApp := test.App()
	server := test.Server(testApp)
	defer server.Close()

	badCfg := &app.Config{
		AuthNURL:           testApp.Config.AuthNURL,
		SessionCookieName:  testApp.Config.SessionCookieName,
		SessionSigningKey:  []byte("wrong"),
		ApplicationDomains: testApp.Config.ApplicationDomains,
	}
	session := test.CreateSession(testApp.RefreshTokenStore, badCfg, 123)

	client := route.NewClient(server.URL).Referred(&testApp.Config.ApplicationDomains[0]).WithCookie(session)
	res, err := client.Delete("/session")
	require.NoError(t, err)

	// request still succeeds
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestDeleteSessionWithoutSession(t *testing.T) {
	testApp := test.App()
	server := test.Server(testApp)
	defer server.Close()

	client := route.NewClient(server.URL).Referred(&testApp.Config.ApplicationDomains[0])
	res, err := client.Delete("/session")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
}
