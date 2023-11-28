package handlers_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostTOTPCreateSuccess(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	account, _ := app.AccountStore.Create("account@keratin.tech", []byte("password"))
	existingSession := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0]).WithCookie(existingSession)
	res, err := client.PostForm("/totp/new", url.Values{})
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, []string{"application/json"}, res.Header["Content-Type"])
	responseData := struct {
		Secret string `json:"secret"`
		Url    string `json:"url"`
	}{}
	err = test.ExtractResult(res, &responseData)
	require.NoError(t, err)

	assert.NotEqual(t, responseData.Secret, "")
	assert.Contains(t, responseData.Url, "otpauth://totp/test.com:account@keratin.tech")
	assert.Contains(t, responseData.Url, responseData.Secret)
}

func TestPostTOTPCreateUnauthenticated(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0])
	res, err := client.PostForm("/totp/new", url.Values{})

	require.NoError(t, err)
	body := test.ReadBody(res)

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	assert.Equal(t, []byte{}, body)
}
