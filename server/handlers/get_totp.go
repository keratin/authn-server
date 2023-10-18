package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/sessions"
)

// GetTOTP begins the TOTP onboarding process
func GetTOTP(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check for valid session with live token
		accountID := sessions.GetAccountID(r)
		if accountID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		totpKey, err := services.TOTPCreator(app.AccountStore, app.TOTPCache, accountID, route.MatchedDomain(r))
		if err != nil {
			panic(err)
		}

		WriteData(w, http.StatusOK, map[string]string{
			"secret": totpKey.Secret(),
			"url":    totpKey.URL(),
		})
	}
}
