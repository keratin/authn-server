package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/lib/parse"
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

		accountID, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			WriteNotFound(w, "account")
			return
		}

		var payload struct {
			OauthProviders []string `json:"oauth_providers"`
		}

		if err := parse.Payload(r, &payload); err != nil {
			WriteErrors(w, err)
			return
		}

		err = services.AccountOauthEnder(app.AccountStore, accountID, payload.OauthProviders)
		if err != nil {
			app.Logger.WithError(err).Error("AccountOauthEnder")

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
