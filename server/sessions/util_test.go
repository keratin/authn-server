package sessions

import (
	"github.com/keratin/authn-server/app"
	"github.com/test-go/testify/assert"
	"net/http/httptest"
	"testing"
)

// ensures that Domain for session cookie is set when app.cfg.SessionDomain exists
func TestSettingCookieDomain(t *testing.T) {
	// given
	cfg := &app.Config{
		SessionCookieName: "aCookie",
		SessionDomain:     "example.com",
	}
	w := httptest.NewRecorder()

	// when
	Set(cfg, w, "aCookieValue")

	// then
	assert.Equal(t, "aCookie=aCookieValue; Domain=example.com; HttpOnly", w.Header().Get("Set-Cookie"))
}

// ensures that Domain for session cookie is NOT set when app.cfg.SessionDomain is empty
func TestThatDomainIsNotSetWhenNotPresent(t *testing.T) {
	// given
	cfg := &app.Config{
		SessionCookieName: "aCookie",
	}
	w := httptest.NewRecorder()

	// when
	Set(cfg, w, "aCookieValue")

	// then
	assert.Equal(t, "aCookie=aCookieValue; HttpOnly", w.Header().Get("Set-Cookie"))
}