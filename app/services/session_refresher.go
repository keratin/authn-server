package services

import (
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/app/models"
	"github.com/keratin/authn-server/ops"
	"github.com/keratin/authn-server/app/tokens/identities"
	"github.com/keratin/authn-server/app/tokens/sessions"
	"github.com/pkg/errors"
)

func SessionRefresher(
	refreshTokenStore data.RefreshTokenStore, keyStore data.KeyStore, actives data.Actives, cfg *app.Config, reporter ops.ErrorReporter,
	session *sessions.Claims, accountID int, audience *route.Domain,
) (string, error) {
	// track actives
	if actives != nil {
		err := actives.Track(accountID)
		if err != nil {
			reporter.ReportError(errors.Wrap(err, "Track"))
		}
	}

	// extend refresh token expiration
	err := refreshTokenStore.Touch(models.RefreshToken(session.Subject), accountID)
	if err != nil {
		return "", errors.Wrap(err, "Touch")
	}

	// create new identity token
	identityToken, err := identities.New(cfg, session, accountID, audience.String()).Sign(keyStore.Key())
	if err != nil {
		return "", errors.Wrap(err, "New")
	}

	return identityToken, nil
}
