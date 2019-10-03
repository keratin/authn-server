package handlers_test

import (
	"net/http"
	"testing"

	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteTOTPSuccess(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	account, _ := app.AccountStore.Create("account@keratin.tech", []byte("password"))
	existingSession := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0]).WithCookie(existingSession)
	res, err := client.Delete("/totp")
	require.NoError(t, err)
	body := test.ReadBody(res)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, []byte{}, body)
}

func TestDeleteTOTPUnauthenticated(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0])
	res, err := client.Delete("/totp")

	require.NoError(t, err)
	body := test.ReadBody(res)

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	assert.Equal(t, []byte{}, body)
}
