package api

import (
	"net/http"

	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/tokens/identities"
	"github.com/keratin/authn-server/tokens/sessions"
)

func NewSession(refreshTokenStore data.RefreshTokenStore, keyStore data.KeyStore, actives data.Actives, cfg *config.Config, accountId int, authorizedAudience *config.Domain) (string, string, error) {
	session, err := sessions.New(refreshTokenStore, cfg, accountId, authorizedAudience.String())
	if err != nil {
		return "", "", err
	}

	sessionToken, err := session.Sign(cfg.SessionSigningKey)
	if err != nil {
		return "", "", err
	}

	identityToken, err := IdentityForSession(keyStore, actives, cfg, session, accountId)
	if err != nil {
		return "", "", err
	}

	return sessionToken, identityToken, err
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

func IdentityForSession(keyStore data.KeyStore, actives data.Actives, cfg *config.Config, session *sessions.Claims, accountId int) (string, error) {
	actives.Track(accountId)
	identity := identities.New(cfg, session, accountId)
	identityToken, err := identity.Sign(keyStore.Key())
	if err != nil {
		return "", err
	}

	return identityToken, nil
}
