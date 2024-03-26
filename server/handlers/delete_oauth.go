package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/server/sessions"
)

func DeleteOauth(app *app.App, providerName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID := sessions.GetAccountID(r)
		if accountID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		result, err := services.AccountOauthEnder(app.AccountStore, accountID, providerName)
		if err != nil {
			app.Logger.WithError(err).Error("AccountOauthEnder")
			WriteErrors(w, err)
			return
		}

		WriteData(w, http.StatusOK, map[string]interface{}{
			"require_password_reset": result.RequirePasswordReset,
		})
	}
}
