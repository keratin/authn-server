package services_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/models"
	"github.com/keratin/authn-server/app/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordlessTokenSender(t *testing.T) {
	remoteApp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()

		if !ok || u != "user" || p != "pass" {
			w.WriteHeader(http.StatusUnauthorized)
		} else if r.URL.Path == "/passwordless" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	serverURL, err := url.Parse(remoteApp.URL)
	require.NoError(t, err)

	authNURL := &url.URL{Scheme: "https", Host: "authn.example.com"}
	passwordlessURL := &url.URL{Scheme: "http", Host: serverURL.Host, Path: "/passwordless", User: url.UserPassword("user", "pass")}

	invoke := func(account *models.Account) error {
		cfg := &app.Config{
			AuthNURL:                    authNURL,
			AppPasswordlessTokenURL:     passwordlessURL,
			PasswordlessTokenSigningKey: []byte("passwordless"),
			PasswordlessTokenTTL:        time.Minute,
		}
		return services.PasswordlessTokenSender(cfg, account, logrus.New())
	}

	t.Run("posting to remote app", func(t *testing.T) {
		err := invoke(&models.Account{
			ID: 1234,
		})
		assert.NoError(t, err)
	})

	t.Run("with locked account", func(t *testing.T) {
		err := invoke(&models.Account{
			ID:     1234,
			Locked: true,
		})
		assert.NoError(t, err)
	})

	t.Run("with no account", func(t *testing.T) {
		err := invoke(nil)
		assert.NoError(t, err)
	})
}
