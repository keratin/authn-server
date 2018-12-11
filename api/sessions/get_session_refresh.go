package sessions

import (
	"github.com/keratin/authn-server/services"
	"net/http"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/lib/route"
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

		identityToken, err := services.SessionRefresher(
			app.RefreshTokenStore, app.KeyStore, app.Actives, app.Config,
			api.GetSession(r), accountID, route.MatchedDomain(r),
		)
		if err != nil {
			panic(errors.Wrap(err, "IdentityForSession"))
		}

		api.WriteData(w, http.StatusCreated, map[string]string{
			"id_token": identityToken,
		})
	}
}
