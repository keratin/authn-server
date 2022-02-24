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

		var formattedLastLogin string
		if account.LastLoginAt == nil {
			formattedLastLogin = ""
		} else {
			formattedLastLogin = account.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
		}

		var formattedPasswordChangedAt string
		if account.PasswordChangedAt.IsZero() {
			formattedPasswordChangedAt = ""
		} else {
			formattedPasswordChangedAt = account.PasswordChangedAt.Format("2006-01-02T15:04:05Z07:00")
		}

		WriteData(w, http.StatusOK, map[string]interface{}{
			"id":                  account.ID,
			"username":            account.Username,
			"last_login_at":       formattedLastLogin,
			"password_changed_at": formattedPasswordChangedAt,
			"locked":              account.Locked,
			"deleted":             account.DeletedAt != nil,
		})
	}
}
