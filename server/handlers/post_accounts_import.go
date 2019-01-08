package handlers

import (
	"net/http"
	"regexp"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
)

func PostAccountsImport(app *app.App) http.HandlerFunc {
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
				WriteErrors(w, fe)
				return
			}

			panic(err)
		}

		WriteData(w, http.StatusCreated, map[string]int{
			"id": account.ID,
		})
	}
}
