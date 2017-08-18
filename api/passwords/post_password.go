package passwords

import (
	"net/http"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/services"
)

func postPassword(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		var accountId int
		if r.FormValue("token") != "" {
			accountId, err = services.PasswordResetter(
				app.AccountStore,
				app.Config,
				r.FormValue("token"),
				r.FormValue("password"),
			)
		} else {
			accountId = api.GetSessionAccountId(r)
			if accountId == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			err = services.PasswordChanger(
				app.AccountStore,
				app.Config,
				accountId,
				r.FormValue("password"),
			)
		}

		if err != nil {
			if fe, ok := err.(services.FieldErrors); ok {
				api.WriteErrors(w, fe)
				return
			} else {
				panic(err)
			}
		}

		err = api.RevokeSession(app.RefreshTokenStore, app.Config, r)
		if err != nil {
			// TODO: alert but continue
		}

		sessionToken, identityToken, err := api.NewSession(app.RefreshTokenStore, app.KeyStore, app.Config, accountId, r.Host)
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
