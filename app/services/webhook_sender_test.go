package services_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/keratin/authn-server/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var noRetry = []time.Duration{}
var fastRetry = []time.Duration{time.Duration(1) * time.Nanosecond}

func TestWebhookSender(t *testing.T) {
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
	defer remoteApp.Close()
	serverURL, err := url.Parse(remoteApp.URL)
	require.NoError(t, err)

	unauthedURL := &url.URL{Scheme: "http", Host: serverURL.Host, Path: "/success"}
	successURL := &url.URL{Scheme: "http", Host: serverURL.Host, Path: "/success", User: url.UserPassword("user", "pass")}
	failureURL := &url.URL{Scheme: "http", Host: serverURL.Host, Path: "/failure", User: url.UserPassword("user", "pass")}

	t.Run("posting to remote app", func(t *testing.T) {
		err := services.WebhookSender(successURL, &url.Values{}, noRetry)
		assert.NoError(t, err)
	})

	t.Run("without auth", func(t *testing.T) {
		err := services.WebhookSender(unauthedURL, &url.Values{}, noRetry)
		if assert.Error(t, err) {
			assert.Equal(t, "PostForm: Status Code: 401", err.Error())
		}
	})

	t.Run("without configured url", func(t *testing.T) {
		err := services.WebhookSender(nil, &url.Values{}, noRetry)
		if assert.Error(t, err) {
			assert.Equal(t, "URL unconfigured", err.Error())
		}
	})

	t.Run("with remote app failure", func(t *testing.T) {
		err := services.WebhookSender(failureURL, &url.Values{}, noRetry)
		if assert.Error(t, err) {
			assert.Equal(t, "PostForm: Status Code: 500", err.Error())
		}
	})
}

func TestWebhookSenderRetries(t *testing.T) {
	var attempt int
	remoteApp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempt == 0 {
			attempt = attempt + 1
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	serverURL, err := url.Parse(remoteApp.URL)
	require.NoError(t, err)

	err = services.WebhookSender(serverURL, &url.Values{}, fastRetry)
	assert.NoError(t, err)
}
