package accounts

import (
	"net/http"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/services"
)

func getAccountsAvailable(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		account, err := app.AccountStore.FindByUsername(req.FormValue("username"))
		if err != nil {
			panic(err)
		}

		if account == nil {
			api.WriteErrors(w, []services.Error{{"username", services.ErrTaken}})
		} else {
			api.WriteData(w, http.StatusOK, true)
		}
	}
}
