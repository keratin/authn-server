package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
)

func DeleteAccountOauth(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		isNotFoundErr := func(err error) (string, bool) {
			var fieldErr services.FieldErrors

			if errors.As(err, &fieldErr) {
				for _, err := range fieldErr {
					if err.Message == services.ErrNotFound {
						return err.Field, true
					}
				}
			}

			return "", false
		}

		provider := mux.Vars(r)["name"]
		accountID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			WriteNotFound(w, "account")
			return
		}

		err = services.IdentityRemover(app.AccountStore, accountID, []string{provider})
		if err != nil {
			app.Logger.WithError(err).Error("IdentityRemover")

			if resource, ok := isNotFoundErr(err); ok {
				WriteNotFound(w, resource)
				return
			}

			WriteErrors(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
