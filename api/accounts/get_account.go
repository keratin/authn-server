package accounts

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/api/util"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/services"
)

func getAccount(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			util.WriteNotFound(w, "account")
			return
		}

		account, err := services.AccountGetter(app.AccountStore, id)
		if err != nil {
			if _, ok := err.(services.FieldErrors); ok {
				util.WriteNotFound(w, "account")
				return
			}

			panic(err)
		}

		util.WriteData(w, http.StatusOK, map[string]interface{}{
			"id":       account.ID,
			"username": account.Username,
			"locked":   account.Locked,
			"deleted":  account.DeletedAt != nil,
		})
	}
}
