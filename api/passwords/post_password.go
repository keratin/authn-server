package passwords

import (
	"net/http"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/api/sessionz"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/services"
)

func postPassword(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var accountID int
		if r.FormValue("token") != "" {
			accountID, err = services.PasswordResetter(
				app.AccountStore,
				app.Reporter,
				app.Config,
				r.FormValue("token"),
				r.FormValue("password"),
			)
		} else {
			accountID = sessionz.GetAccountID(r)
			if accountID == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			err = services.PasswordChanger(
				app.AccountStore,
				app.Reporter,
				app.Config,
				accountID,
				r.FormValue("currentPassword"),
				r.FormValue("password"),
			)
		}

		if err != nil {
			if fe, ok := err.(services.FieldErrors); ok {
				api.WriteErrors(w, fe)
				return
			}

			panic(err)
		}

		sessionToken, identityToken, err := services.SessionCreator(
			app.AccountStore, app.RefreshTokenStore, app.KeyStore, app.Actives, app.Config, app.Reporter,
			accountID, route.MatchedDomain(r), sessionz.GetRefreshToken(r),
		)
		if err != nil {
			panic(err)
		}

		// Return the signed session in a cookie
		sessionz.Set(app.Config, w, sessionToken)

		// Return the signed identity token in the body
		api.WriteData(w, http.StatusCreated, map[string]string{
			"id_token": identityToken,
		})
	}
}
