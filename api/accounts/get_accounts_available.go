package accounts

import (
	"net/http"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/services"
)

func getAccountsAvailable(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, err := app.AccountStore.FindByUsername(r.FormValue("username"))
		if err != nil {
			panic(err)
		}

		if account == nil {
			api.WriteErrors(w, services.FieldErrors{{"username", services.ErrTaken}})
		} else {
			api.WriteData(w, http.StatusOK, true)
		}
	}
}
