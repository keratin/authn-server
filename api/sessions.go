package api

import (
	"net/http"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/tokens/identities"
	"github.com/keratin/authn-server/tokens/sessions"
	"github.com/keratin/authn-server/services"
	"github.com/pkg/errors"
)

func NewSession(store data.AccountStore, refreshTokenStore data.RefreshTokenStore, keyStore data.KeyStore, actives data.Actives, cfg *config.Config, accountID int, authorizedAudience *route.Domain) (string, string, error) {
	session, err := sessions.New(refreshTokenStore, cfg, accountID, authorizedAudience.String())
	if err != nil {
		return "", "", errors.Wrap(err, "New")
	}

	sessionToken, err := session.Sign(cfg.SessionSigningKey)
	if err != nil {
		return "", "", errors.Wrap(err, "Sign")
	}

	identityToken, err := IdentityForSession(keyStore, actives, cfg, session, accountID, authorizedAudience)
	if err != nil {
		return "", "", errors.Wrap(err, "IdentityForSession")
	}

	err = services.LastLoginUpdater(store, accountID)
	if err != nil {
		return "", "", errors.Wrap(err, "UpdateLastLoginAt")
	}

	return sessionToken, identityToken, nil
}

func RevokeSession(refreshTokenStore data.RefreshTokenStore, cfg *config.Config, r *http.Request) (err error) {
	oldSession := GetSession(r)
	if oldSession == nil {
		return nil
	}
	return refreshTokenStore.Revoke(models.RefreshToken(oldSession.Subject))
}

func SetSession(cfg *config.Config, w http.ResponseWriter, val string) {
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

func IdentityForSession(keyStore data.KeyStore, actives data.Actives, cfg *config.Config, session *sessions.Claims, accountID int, audience *route.Domain) (string, error) {
	if actives != nil {
		actives.Track(accountID)
	}
	identityToken, err := identities.New(cfg, session, accountID, audience.String()).Sign(keyStore.Key())
	if err != nil {
		return "", errors.Wrap(err, "New")
	}

	return identityToken, nil
}
