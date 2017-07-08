package accounts

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/services"
)

func patchAccountExpirePassword(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			panic(err)
		}

		errors := services.PasswordExpirer(app.AccountStore, app.RefreshTokenStore, id)

		if errors == nil {
			w.WriteHeader(http.StatusOK)
		} else {
			api.WriteJson(w, http.StatusNotFound, api.ServiceErrors{errors})
		}
	}
}
