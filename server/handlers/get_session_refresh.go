package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/server/sessions"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/app/services"
	"github.com/pkg/errors"
)

func GetSessionRefresh(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check for valid session with live token
		accountID := sessions.GetAccountID(r)
		if accountID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		identityToken, err := services.SessionRefresher(
			app.RefreshTokenStore, app.KeyStore, app.Actives, app.Config, app.Reporter,
			sessions.Get(r), accountID, route.MatchedDomain(r),
		)
		if err != nil {
			panic(errors.Wrap(err, "IdentityForSession"))
		}

		WriteData(w, http.StatusCreated, map[string]string{
			"id_token": identityToken,
		})
	}
}
