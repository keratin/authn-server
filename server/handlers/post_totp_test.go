package handlers_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostTOTPSuccess(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	// nolint: gosec
	totpSecret := "JKK5AG4NDAWSZSR4ZFKZBWZ7OJGLB2JM"
	account, _ := app.AccountStore.Create("account@keratin.tech", []byte("password"))
	existingSession := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)
	err := app.TOTPCache.CacheTOTPSecret(account.ID, []byte(totpSecret))
	require.NoError(t, err)

	code, err := totp.GenerateCode(totpSecret, time.Now())
	fmt.Println(code)
	require.NoError(t, err)

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0]).WithCookie(existingSession)
	res, err := client.PostForm("/totp", url.Values{
		"code": []string{code},
	})
	require.NoError(t, err)
	body := test.ReadBody(res)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "", string(body))
}

func TestPostTOTPFailure(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	// nolint: gosec
	totpSecret := "JKK5AG4NDAWSZSR4ZFKZBWZ7OJGLB2JM"
	account, _ := app.AccountStore.Create("account@keratin.tech", []byte("password"))
	existingSession := test.CreateSession(app.RefreshTokenStore, app.Config, account.ID)
	err := app.TOTPCache.CacheTOTPSecret(account.ID, []byte(totpSecret))
	require.NoError(t, err)

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0]).WithCookie(existingSession)
	res, err := client.PostForm("/totp", url.Values{
		"code": []string{"12345"}, //Invalid code
	})
	require.NoError(t, err)
	body := test.ReadBody(res)

	assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
	assert.Equal(t, "{\"errors\":[{\"field\":\"totp\",\"message\":\"INVALID_OR_EXPIRED\"}]}", string(body))
}
