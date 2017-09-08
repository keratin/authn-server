package meta

import (
	"net/http"

	"github.com/keratin/authn-server/api"
)

func getStats(app *api.App) http.HandlerFunc {
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

		api.WriteJSON(w, http.StatusOK, struct {
			Daily   map[string]int `json:"daily"`
			Weekly  map[string]int `json:"weekly"`
			Monthly map[string]int `json:"monthly"`
		}{
			Daily:   daily,
			Weekly:  weekly,
			Monthly: monthly,
		})
	}
}
