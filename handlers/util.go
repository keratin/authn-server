package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/tokens/identities"
	"github.com/keratin/authn-server/tokens/sessions"
)

// A specialized handler that looks like any other middleware adapter but is known to serve a
// particular purpose.
type SecurityHandler func(http.Handler) http.Handler

func establishSession(refreshTokenStore data.RefreshTokenStore, cfg *config.Config, account_id int) (string, string, error) {
	session, err := sessions.New(refreshTokenStore, cfg, account_id)
	if err != nil {
		return "", "", err
	}

	sessionToken, err := session.Sign(cfg.SessionSigningKey)
	if err != nil {
		return "", "", err
	}

	identityToken, err := identityForSession(cfg, session, account_id)
	if err != nil {
		return "", "", err
	}

	return sessionToken, identityToken, err
}

func revokeSession(refreshTokenStore data.RefreshTokenStore, cfg *config.Config, req *http.Request) (err error) {
	oldSession, err := currentSession(cfg, req)
	if err != nil {
		return err
	}
	if oldSession != nil {
		return refreshTokenStore.Revoke(models.RefreshToken(oldSession.Subject))
	}
	return nil
}

func setSession(cfg *config.Config, w http.ResponseWriter, val string) {
	cookie := &http.Cookie{
		Name:     cfg.SessionCookieName,
		Value:    val,
		Path:     cfg.MountedPath,
		Secure:   cfg.ForceSSL,
		HttpOnly: true,
	}
	if val == "" {
		cookie.MaxAge = -1
	}
	http.SetCookie(w, cookie)
}

func currentSession(cfg *config.Config, req *http.Request) (*sessions.Claims, error) {
	cookie, err := req.Cookie(cfg.SessionCookieName)
	if err == http.ErrNoCookie {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return sessions.Parse(cookie.Value, cfg)
}

func identityForSession(cfg *config.Config, session *sessions.Claims, account_id int) (string, error) {
	identity := identities.New(cfg, session, account_id)
	identityToken, err := identity.Sign(cfg.IdentitySigningKey)
	if err != nil {
		return "", err
	}

	return identityToken, nil
}
