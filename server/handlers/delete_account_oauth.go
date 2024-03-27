package handlers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/lib/parse"
)

func DeleteAccountOauth(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			WriteNotFound(w, "account")
			return
		}

		var payload struct {
			OauthProviders []string `json:"oauth_providers"`
		}

		if err := parse.Payload(r, &payload); err != nil {
			WriteErrors(w, err)
			return
		}

		result, err := services.AccountOauthEnder(app.AccountStore, accountID, payload.OauthProviders)
		if err != nil {
			app.Logger.WithError(err).Error("AccountOauthEnder")

			if _, ok := err.(services.FieldErrors); ok {
				WriteNotFound(w, "account")
				return
			}

			WriteErrors(w, err)
			return
		}

		WriteData(w, http.StatusOK, map[string]interface{}{
			"require_password_reset": result.RequirePasswordReset,
		})
	}
}
