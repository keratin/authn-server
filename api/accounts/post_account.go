package accounts

import (
	"net/http"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/services"
)

func postAccount(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create the account
		account, err := services.AccountCreator(
			app.AccountStore,
			app.Config,
			r.FormValue("username"),
			r.FormValue("password"),
		)
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

		sessionToken, identityToken, err := api.NewSession(app.RefreshTokenStore, app.KeyStore, app.Actives, app.Config, account.Id, api.MatchedDomain(r))
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
