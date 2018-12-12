package accounts

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/api/util"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/services"
)

func PatchAccountLock(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			util.WriteNotFound(w, "account")
			return
		}

		err = services.AccountLocker(app.AccountStore, app.RefreshTokenStore, id)
		if err != nil {
			if _, ok := err.(services.FieldErrors); ok {
				util.WriteNotFound(w, "account")
				return
			}

			panic(err)
		}

		w.WriteHeader(http.StatusOK)
	}
}
