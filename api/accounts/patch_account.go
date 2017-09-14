package accounts

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/services"
)

func patchAccount(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}

		err = services.AccountUpdater(app.AccountStore, app.Config, id, r.FormValue("username"))
		if err != nil {
			if fe, ok := err.(services.FieldErrors); ok {
				api.WriteJSON(w, http.StatusNotFound, api.ServiceErrors{fe})
				return
			}

			panic(err)
		}

		w.WriteHeader(http.StatusOK)
	}
}
