package accounts

import (
	"net/http"
	"regexp"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/services"
)

func postAccountsImport(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		locked, err := regexp.MatchString("^(?i:t|true|yes)$", r.FormValue("locked"))
		if err != nil {
			panic(err)
		}

		account, errors := services.AccountImporter(
			app.AccountStore,
			app.Config,
			r.FormValue("username"),
			r.FormValue("password"),
			locked,
		)
		if errors != nil {
			api.WriteErrors(w, errors)
		} else {
			api.WriteData(w, http.StatusCreated, map[string]int{"id": account.Id})
		}
	}
}
