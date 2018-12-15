package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/app"
)

func GetStats(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		daily, err := app.Actives.ActivesByDay()
		if err != nil {
			panic(err)
		}

		weekly, err := app.Actives.ActivesByWeek()
		if err != nil {
			panic(err)
		}

		monthly, err := app.Actives.ActivesByMonth()
		if err != nil {
			panic(err)
		}

		actives := struct {
			Daily   map[string]int `json:"daily"`
			Weekly  map[string]int `json:"weekly"`
			Monthly map[string]int `json:"monthly"`
		}{
			Daily:   daily,
			Weekly:  weekly,
			Monthly: monthly,
		}

		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"actives": actives,
		})
	}
}
