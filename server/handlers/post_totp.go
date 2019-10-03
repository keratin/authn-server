package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/server/sessions"
)

//PostTOTP finishes the TOTP onboarding process
func PostTOTP(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check for valid session with live token
		accountID := sessions.GetAccountID(r)
		if accountID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := services.TOTPSetter(app.AccountStore, app.TOTPCache, app.Config, accountID, r.FormValue("code")); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			if fe, ok := err.(services.FieldErrors); ok {
				WriteErrors(w, fe)
				return
			}

			panic(err)
		}

		w.WriteHeader(http.StatusOK)
	}
}
