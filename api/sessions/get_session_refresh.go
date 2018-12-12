package sessions

import (
	"net/http"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/api/sessionz"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/services"
	"github.com/pkg/errors"
)

func getSessionRefresh(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check for valid session with live token
		accountID := sessionz.GetAccountID(r)
		if accountID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		identityToken, err := services.SessionRefresher(
			app.RefreshTokenStore, app.KeyStore, app.Actives, app.Config, app.Reporter,
			sessionz.Get(r), accountID, route.MatchedDomain(r),
		)
		if err != nil {
			panic(errors.Wrap(err, "IdentityForSession"))
		}

		api.WriteData(w, http.StatusCreated, map[string]string{
			"id_token": identityToken,
		})
	}
}
