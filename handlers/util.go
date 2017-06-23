package handlers

import (
	"net/http"

	"github.com/keratin/authn/config"
	"github.com/keratin/authn/data"
	"github.com/keratin/authn/models"
	"github.com/keratin/authn/tokens/identities"
	"github.com/keratin/authn/tokens/sessions"
)

func establishSession(refreshTokenStore data.RefreshTokenStore, cfg *config.Config, account_id int) (sessionToken string, identityToken string, err error) {
	session, err := sessions.New(refreshTokenStore, cfg, account_id)
	if err != nil {
		return
	}

	identity, err := identities.New(refreshTokenStore, cfg, session)
	if err != nil {
		return
	}

	sessionToken, err = session.Sign(cfg.SessionSigningKey)
	if err != nil {
		return
	}

	identityToken, err = identity.Sign(cfg.IdentitySigningKey)
	if err != nil {
		return
	}

	return
}

func revokeSession(refreshTokenStore data.RefreshTokenStore, cfg *config.Config, req *http.Request) (err error) {
	cookie, err := req.Cookie(cfg.SessionCookieName)
	if err != nil {
		return
	}

	oldSession, err := sessions.Parse(cookie.Value, cfg)
	if err != nil {
		return
	}

	return refreshTokenStore.Revoke(models.RefreshToken(oldSession.Subject))
}
