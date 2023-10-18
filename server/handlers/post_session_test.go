package handlers_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPostSessionSuccess(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	b, _ := bcrypt.GenerateFromPassword([]byte("bar"), 4)
	_, err := app.AccountStore.Create("foo", b)
	require.NoError(t, err)

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0])
	res, err := client.PostForm("/session", url.Values{
		"username": []string{"foo"},
		"password": []string{"bar"},
	})
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	test.AssertSession(t, app.Config, res.Cookies())
	test.AssertIDTokenResponse(t, res, app.KeyStore, app.Config)
}

func TestPostSessionSuccessWithSession(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	b, _ := bcrypt.GenerateFromPassword([]byte("bar"), 4)
	_, _ = app.AccountStore.Create("foo", b)

	accountID := 8642
	session := test.CreateSession(app.RefreshTokenStore, app.Config, accountID)

	// before
	refreshTokens, err := app.RefreshTokenStore.FindAll(accountID)
	require.NoError(t, err)
	refreshToken := refreshTokens[0]

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0]).WithCookie(session)
	_, err = client.PostForm("/session", url.Values{
		"username": []string{"foo"},
		"password": []string{"bar"},
	})
	require.NoError(t, err)

	// after
	id, err := app.RefreshTokenStore.Find(refreshToken)
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestPostSessionFailure(t *testing.T) {
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	var testCases = []struct {
		username string
		password string
		errors   services.FieldErrors
	}{
		{"", "", services.FieldErrors{{Field: "credentials", Message: "FAILED"}}},
	}

	for _, tc := range testCases {
		client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0])
		res, err := client.PostForm("/session", url.Values{
			"username": []string{tc.username},
			"password": []string{tc.password},
		})
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		test.AssertErrors(t, res, tc.errors)
	}
}

func TestPostSessionSuccessWithTOTP(t *testing.T) {
	// nolint: gosec
	totpSecret := "JKK5AG4NDAWSZSR4ZFKZBWZ7OJGLB2JM"
	totpSecretEnc := []byte("cli6azfL5i7PAnh8U/w3Zbglsm3XcdaGODy+Ga5QqT02c9hotDAR1Y28--3UihzsJhw/+EU3R6--qUw9L8DwN5XPVfOStshKzA==")

	app := test.App()
	server := test.Server(app)
	defer server.Close()

	b, _ := bcrypt.GenerateFromPassword([]byte("bar"), 4)
	account, _ := app.AccountStore.Create("foo", b)

	ok, err := app.AccountStore.SetTOTPSecret(account.ID, totpSecretEnc)
	assert.True(t, ok)
	require.NoError(t, err)

	code, err := totp.GenerateCode(totpSecret, time.Now())
	require.NoError(t, err)

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0])
	res, err := client.PostForm("/session", url.Values{
		"username": []string{"foo"},
		"password": []string{"bar"},
		"otp":      []string{code},
	})
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, res.StatusCode)
	test.AssertSession(t, app.Config, res.Cookies())
	test.AssertIDTokenResponse(t, res, app.KeyStore, app.Config)
}

func TestPostSessionSuccessWithSessionAndTOTP(t *testing.T) {
	// nolint: gosec
	totpSecret := "JKK5AG4NDAWSZSR4ZFKZBWZ7OJGLB2JM"
	totpSecretEnc := []byte("cli6azfL5i7PAnh8U/w3Zbglsm3XcdaGODy+Ga5QqT02c9hotDAR1Y28--3UihzsJhw/+EU3R6--qUw9L8DwN5XPVfOStshKzA==")

	app := test.App()
	server := test.Server(app)
	defer server.Close()

	b, _ := bcrypt.GenerateFromPassword([]byte("bar"), 4)
	account, _ := app.AccountStore.Create("foo", b)

	accountID := 8642
	session := test.CreateSession(app.RefreshTokenStore, app.Config, accountID)

	//Generate TOTP code
	ok, err := app.AccountStore.SetTOTPSecret(account.ID, totpSecretEnc)
	assert.True(t, ok)
	require.NoError(t, err)

	code, err := totp.GenerateCode(totpSecret, time.Now())
	require.NoError(t, err)

	// before
	refreshTokens, err := app.RefreshTokenStore.FindAll(accountID)
	require.NoError(t, err)
	refreshToken := refreshTokens[0]

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0]).WithCookie(session)
	_, err = client.PostForm("/session", url.Values{
		"username": []string{"foo"},
		"password": []string{"bar"},
		"otp":      []string{code},
	})
	require.NoError(t, err)

	// after
	id, err := app.RefreshTokenStore.Find(refreshToken)
	require.NoError(t, err)
	assert.Empty(t, id)
}

func TestPostSessionFailureWithTOTP(t *testing.T) {
	totpSecretEnc := []byte("cli6azfL5i7PAnh8U/w3Zbglsm3XcdaGODy+Ga5QqT02c9hotDAR1Y28--3UihzsJhw/+EU3R6--qUw9L8DwN5XPVfOStshKzA==")

	app := test.App()
	server := test.Server(app)
	defer server.Close()

	b, _ := bcrypt.GenerateFromPassword([]byte("bar"), 4)
	account, _ := app.AccountStore.Create("foo", b)

	ok, err := app.AccountStore.SetTOTPSecret(account.ID, totpSecretEnc)
	assert.True(t, ok)
	require.NoError(t, err)

	var testCases = []struct {
		username string
		password string
		totpCode string
		errors   services.FieldErrors
	}{
		{"foo", "bar", "12345", services.FieldErrors{{Field: "totp", Message: "INVALID_OR_EXPIRED"}}},
	}

	for _, tc := range testCases {
		client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0])
		res, err := client.PostForm("/session", url.Values{
			"username": []string{tc.username},
			"password": []string{tc.password},
			"otp":      []string{tc.totpCode},
		})
		require.NoError(t, err)

		assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
		test.AssertErrors(t, res, tc.errors)
	}
}
