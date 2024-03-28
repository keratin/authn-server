package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/server/sessions"
)

func GetOauthInfo(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID := sessions.GetAccountID(r)
		if accountID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		oAccounts, err := services.AccountOauthGetter(app.AccountStore, accountID)
		if err != nil {
			WriteErrors(w, err)
			return
		}

		WriteData(w, http.StatusOK, oAccounts)
	}
}
