package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
)

func GetPasswordScore(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		password := r.URL.Query().Get("password")
		if password == "" {
			WriteErrors(w, services.FieldErrors{services.FieldError{
				Field:   "password",
				Message: services.ErrMissing,
			}})
			return
		}

		score := services.CalcPasswordScore(password)

		WriteData(w, http.StatusOK, map[string]interface{}{
			"score":         score,
			"requiredScore": app.Config.PasswordMinComplexity,
		})
	}
}
