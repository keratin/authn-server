package services_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
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
		err := services.WebhookSender(successURL, &url.Values{}, noRetry, nil)
		assert.NoError(t, err)
	})

	t.Run("without auth", func(t *testing.T) {
		err := services.WebhookSender(unauthedURL, &url.Values{}, noRetry, nil)
		if assert.Error(t, err) {
			assert.Equal(t, "PostForm: Status Code: 401", err.Error())
		}
	})

	t.Run("without configured url", func(t *testing.T) {
		err := services.WebhookSender(nil, &url.Values{}, noRetry, nil)
		if assert.Error(t, err) {
			assert.Equal(t, "URL unconfigured", err.Error())
		}
	})

	t.Run("with remote app failure", func(t *testing.T) {
		err := services.WebhookSender(failureURL, &url.Values{}, noRetry, nil)
		if assert.Error(t, err) {
			assert.Equal(t, "PostForm: Status Code: 500", err.Error())
		}
	})
}

func TestWebhookSenderSignature(t *testing.T) {
	key := []byte(uuid.NewString())

	verifier := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h := r.Header.Get("X-Authn-Webhook-Signature"); h != "" {
			hm := hmac.New(sha256.New, key)
			_, err := hm.Write([]byte(r.Form.Encode()))
			require.NoError(t, err)
			exp := hex.EncodeToString(hm.Sum(nil))
			if exp == h {
				// request is verified
				return
			}
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	verifierURL, err := url.Parse(verifier.URL)
	require.NoError(t, err)
	requireSigURL := &url.URL{Scheme: "http", Host: verifierURL.Host, Path: "/mustsign"}
	t.Run("without signing key", func(t *testing.T) {
		err := services.WebhookSender(requireSigURL, &url.Values{}, noRetry, nil)
		if assert.Error(t, err) {
			assert.Equal(t, "PostForm: Status Code: 401", err.Error())
		}
	})

	t.Run("with signing key", func(t *testing.T) {
		err := services.WebhookSender(requireSigURL, &url.Values{}, noRetry, key)
		assert.NoError(t, err)
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

	err = services.WebhookSender(serverURL, &url.Values{}, fastRetry, nil)
	assert.NoError(t, err)
}
