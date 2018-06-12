package sessions

import (
	"net/http"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/models"
	"github.com/pkg/errors"
)

func getSessionRefresh(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check for valid session with live token
		accountID := api.GetSessionAccountID(r)
		if accountID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// refresh the refresh token
		session := api.GetSession(r)
		err := app.RefreshTokenStore.Touch(models.RefreshToken(session.Subject), accountID)
		if err != nil {
			panic(errors.Wrap(err, "Touch"))
		}

		// generate the requested identity token
		identityToken, err := api.IdentityForSession(app.KeyStore, app.Actives, app.Config, session, accountID, route.MatchedDomain(r))
		if err != nil {
			panic(errors.Wrap(err, "IdentityForSession"))
		}

		api.WriteData(w, http.StatusCreated, map[string]string{
			"id_token": identityToken,
		})
	}
}
