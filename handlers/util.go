package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/tokens/identities"
	"github.com/keratin/authn-server/tokens/sessions"
)

func establishSession(refreshTokenStore data.RefreshTokenStore, cfg *config.Config, account_id int) (string, string, error) {
	session, err := sessions.New(refreshTokenStore, cfg, account_id)
	if err != nil {
		return "", "", err
	}

	sessionToken, err := session.Sign(cfg.SessionSigningKey)
	if err != nil {
		return "", "", err
	}

	identityToken, err := identityForSession(refreshTokenStore, cfg, session)
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

func identityForSession(store data.RefreshTokenStore, cfg *config.Config, session *sessions.Claims) (string, error) {
	identity, err := identities.New(store, cfg, session)
	if err != nil {
		return "", err
	}

	identityToken, err := identity.Sign(cfg.IdentitySigningKey)
	if err != nil {
		return "", err
	}

	return identityToken, nil
}
