package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/app/services"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/server/sessions"
)

func DeleteTOTP(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID := sessions.GetAccountID(r)
		if accountID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := services.TOTPDeleter(app.AccountStore, accountID); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
