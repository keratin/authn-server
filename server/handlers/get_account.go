package handlers

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
)

func GetAccount(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var paramID *int
		var paramUsername *string

		idOrUsername := mux.Vars(r)["id"]
		if idOrUsername == "" {
			return
		}

		id, err := strconv.Atoi(idOrUsername)
		if err != nil {
			paramUsername = &idOrUsername
		} else {
			paramID = &id
		}

		account, err := services.AccountGetter(app.AccountStore, services.AccountGetterParams{
			AccountID: paramID,
			Username:  paramUsername,
		})
		if err != nil {
			if _, ok := err.(services.FieldErrors); ok {
				WriteNotFound(w, "account")
				return
			}

			panic(err)
		}

		WriteData(w, http.StatusOK, account)
	}
}
