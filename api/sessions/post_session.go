package sessions

import (
	"net/http"

	"github.com/keratin/authn-server/api/util"
	"github.com/keratin/authn-server/api/sessionz"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/services"
)

func PostSession(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check the password
		account, err := services.CredentialsVerifier(
			app.AccountStore,
			app.Config,
			r.FormValue("username"),
			r.FormValue("password"),
		)
		if err != nil {
			if fe, ok := err.(services.FieldErrors); ok {
				util.WriteErrors(w, fe)
				return
			}

			panic(err)
		}

		sessionToken, identityToken, err := services.SessionCreator(
			app.AccountStore, app.RefreshTokenStore, app.KeyStore, app.Actives, app.Config, app.Reporter,
			account.ID, route.MatchedDomain(r), sessionz.GetRefreshToken(r),
		)
		if err != nil {
			panic(err)
		}

		// Return the signed session in a cookie
		sessionz.Set(app.Config, w, sessionToken)

		// Return the signed identity token in the body
		util.WriteData(w, http.StatusCreated, map[string]string{
			"id_token": identityToken,
		})
	}
}
