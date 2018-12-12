package accounts

import (
	"net/http"

	"github.com/keratin/authn-server/api/util"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/services"
)

func GetAccountsAvailable(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account, err := app.AccountStore.FindByUsername(r.FormValue("username"))
		if err != nil {
			panic(err)
		}

		if account == nil {
			util.WriteData(w, http.StatusOK, true)
		} else {
			util.WriteErrors(w, services.FieldErrors{{"username", services.ErrTaken}})
		}
	}
}
