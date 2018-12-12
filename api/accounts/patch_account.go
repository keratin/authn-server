package accounts

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/api/util"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/services"
)

func patchAccount(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			util.WriteNotFound(w, "account")
			return
		}

		err = services.AccountUpdater(app.AccountStore, app.Config, id, r.FormValue("username"))
		if err != nil {
			if fe, ok := err.(services.FieldErrors); ok {
				if fe[0].Message == services.ErrNotFound {
					util.WriteNotFound(w, "account")
				} else {
					util.WriteErrors(w, fe)
				}
				return
			}

			panic(err)
		}

		w.WriteHeader(http.StatusOK)
	}
}
