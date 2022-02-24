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
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			WriteNotFound(w, "account")
			return
		}

		account, err := services.AccountGetter(app.AccountStore, id)
		if err != nil {
			if _, ok := err.(services.FieldErrors); ok {
				WriteNotFound(w, "account")
				return
			}

			panic(err)
		}

		WriteData(w, http.StatusOK, map[string]interface{}{
			"id":       account.ID,
			"username": account.Username,
			"locked":   account.Locked,
			"last_login_at": func() string {
				if account.LastLoginAt != nil {
					return account.LastLoginAt.String()
				} else {
					return ""
				}
			},
			"password_changed_at": account.PasswordChangedAt.String(),
			"deleted":             account.DeletedAt != nil,
		})
	}
}
