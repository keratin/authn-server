package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/lib/parse"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/sessions"
)

func PostPassword(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var credentials struct {
			Token           string
			Password        string
			CurrentPassword string
			OTP             string
		}
		if err := parse.Payload(r, &credentials); err != nil {
			WriteErrors(w, err)
			return
		}

		var err error
		var accountID int
		if credentials.Token != "" {
			accountID, err = services.PasswordResetter(
				app.AccountStore,
				app.Reporter,
				app.Config,
				credentials.Token,
				credentials.Password,
				credentials.OTP,
			)
		} else {
			accountID = sessions.GetAccountID(r)
			if accountID == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			err = services.PasswordChanger(
				app.AccountStore,
				app.Reporter,
				app.Config,
				accountID,
				credentials.CurrentPassword,
				credentials.Password,
			)
		}

		if err != nil {
			if fe, ok := err.(services.FieldErrors); ok {
				WriteErrors(w, fe)
				return
			}

			panic(err)
		}

		if app.Config.PasswordChangeLogout {
			err = services.SessionBatchEnder(app.RefreshTokenStore, accountID)
			if err != nil {
				panic(err)
			}
		}

		sessionToken, identityToken, err := services.SessionCreator(
			app.AccountStore, app.RefreshTokenStore, app.KeyStore, app.Actives, app.Config, app.Reporter,
			accountID, route.MatchedDomain(r), sessions.GetRefreshToken(r),
		)
		if err != nil {
			panic(err)
		}

		// Return the signed session in a cookie
		sessions.Set(app.Config, w, sessionToken)

		// Return the signed identity token in the body
		WriteData(w, http.StatusCreated, map[string]string{
			"id_token": identityToken,
		})
	}
}
