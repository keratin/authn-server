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

		account, err := services.AccountImporter(
			app.AccountStore,
			app.Config,
			r.FormValue("username"),
			r.FormValue("password"),
			locked,
		)
		if err != nil {
			if fe, ok := err.(services.FieldErrors); ok {
				api.WriteErrors(w, fe)
				return
			}

			panic(err)
		}

		api.WriteData(w, http.StatusCreated, map[string]int{
			"id": account.ID,
		})
	}
}
