package sessions

import (
	"net/http"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/services"
)

func postSessionToken(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var accountID int

		accountID, err = services.PasswordlessTokenVerifier(
			app.AccountStore,
			app.Reporter,
			app.Config,
			r.FormValue("token"),
		)

		if err != nil {
			if fe, ok := err.(services.FieldErrors); ok {
				api.WriteErrors(w, fe)
				return
			}

			panic(err)
		}

		sessionToken, identityToken, err := services.SessionCreator(
			app.AccountStore, app.RefreshTokenStore, app.KeyStore, app.Actives, app.Config, app.Reporter,
			accountID, route.MatchedDomain(r), api.GetRefreshToken(r),
		)
		if err != nil {
			panic(err)
		}

		// Return the signed session in a cookie
		api.SetSession(app.Config, w, sessionToken)

		// Return the signed identity token in the body
		api.WriteData(w, http.StatusCreated, map[string]string{
			"id_token": identityToken,
		})
	}
}
