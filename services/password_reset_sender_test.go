package services_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordResetSender(t *testing.T) {
	remoteApp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()

		if !ok || u != "user" || p != "pass" {
			w.WriteHeader(http.StatusUnauthorized)
		} else if r.URL.Path == "/success" {
			w.WriteHeader(http.StatusOK)
		} else if r.URL.Path == "/failure" {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	serverURL, err := url.Parse(remoteApp.URL)
	require.NoError(t, err)

	authNURL := &url.URL{Scheme: "https", Host: "authn.example.com"}
	successURL := &url.URL{Scheme: "http", Host: serverURL.Host, Path: "/success", User: url.UserPassword("user", "pass")}
	failureURL := &url.URL{Scheme: "http", Host: serverURL.Host, Path: "/failure", User: url.UserPassword("user", "pass")}

	invoke := func(resetURL *url.URL, account *models.Account) error {
		cfg := &config.Config{
			AuthNURL:            authNURL,
			AppPasswordResetURL: resetURL,
			ResetSigningKey:     []byte("resets"),
			ResetTokenTTL:       time.Minute,
		}
		return services.PasswordResetSender(cfg, account)
	}

	t.Run("posting to remote app", func(t *testing.T) {
		err := invoke(successURL, &models.Account{
			Id:                1234,
			PasswordChangedAt: time.Now(),
		})
		assert.NoError(t, err)
	})

	t.Run("without configured url", func(t *testing.T) {
		err := invoke(nil, &models.Account{
			Id:                1234,
			PasswordChangedAt: time.Now(),
		})
		assert.Equal(t, "AppPasswordResetURL unconfigured", err.Error())
	})

	t.Run("with locked account", func(t *testing.T) {
		err := invoke(successURL, &models.Account{
			Id:                1234,
			PasswordChangedAt: time.Now(),
			Locked:            true,
		})
		assert.Equal(t, services.FieldErrors{{"account", "LOCKED"}}, err)
	})

	t.Run("with remote app failure", func(t *testing.T) {
		err := invoke(failureURL, &models.Account{
			Id:                1234,
			PasswordChangedAt: time.Now(),
		})
		assert.Equal(t, "Status Code: 500", err.Error())
	})

	t.Run("with no account", func(t *testing.T) {
		err := invoke(successURL, nil)
		assert.Equal(t, services.FieldErrors{{"account", "MISSING"}}, err)
	})
}
