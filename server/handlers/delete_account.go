package handlers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
)

func DeleteAccount(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			WriteNotFound(w, "account")
			return
		}

		err = services.AccountArchiver(app.AccountStore, app.RefreshTokenStore, id)
		if err != nil {
			if _, ok := err.(services.FieldErrors); ok {
				WriteNotFound(w, "account")
				return
			}

			panic(err)
		}

		w.WriteHeader(http.StatusOK)
	}
}
