package api

import (
	"net/http"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/tokens/identities"
	"github.com/keratin/authn-server/tokens/sessions"
	"github.com/pkg/errors"
)

func NewSession(refreshTokenStore data.RefreshTokenStore, keyStore data.KeyStore, actives data.Actives, cfg *config.Config, accountID int, authorizedAudience *route.Domain) (string, string, error) {
	session, err := sessions.New(refreshTokenStore, cfg, accountID, authorizedAudience.String())
	if err != nil {
		return "", "", errors.Wrap(err, "New")
	}

	sessionToken, err := session.Sign(cfg.SessionSigningKey)
	if err != nil {
		return "", "", errors.Wrap(err, "Sign")
	}

	identityToken, err := IdentityForSession(keyStore, actives, cfg, session, accountID)
	if err != nil {
		return "", "", errors.Wrap(err, "IdentityForSession")
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

func IdentityForSession(keyStore data.KeyStore, actives data.Actives, cfg *config.Config, session *sessions.Claims, accountID int) (string, error) {
	if actives != nil {
		actives.Track(accountID)
	}
	identity := identities.New(cfg, session, accountID)
	identityToken, err := identity.Sign(keyStore.Key())
	if err != nil {
		return "", errors.Wrap(err, "Track")
	}

	return identityToken, nil
}
