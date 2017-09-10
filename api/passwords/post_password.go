package passwords

import (
	"net/http"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/services"
)

func postPassword(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var accountID int
		if r.FormValue("token") != "" {
			accountID, err = services.PasswordResetter(
				app.AccountStore,
				app.Config,
				r.FormValue("token"),
				r.FormValue("password"),
			)
		} else {
			accountID = api.GetSessionAccountID(r)
			if accountID == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			err = services.PasswordChanger(
				app.AccountStore,
				app.Config,
				accountID,
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

		err = api.RevokeSession(app.RefreshTokenStore, app.Config, r)
		if err != nil {
			app.Reporter.ReportError(err)
		}

		sessionToken, identityToken, err := api.NewSession(app.RefreshTokenStore, app.KeyStore, app.Actives, app.Config, accountID, api.MatchedDomain(r))
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
