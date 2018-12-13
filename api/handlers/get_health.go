package handlers

import (
	"net/http"

	"github.com/keratin/authn-server/app"
)

type health struct {
	HTTP  bool `json:"http"`
	Db    bool `json:"db"`
	Redis bool `json:"redis"`
}

func GetHealth(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := health{
			HTTP:  true,
			Redis: app.RedisCheck(),
			Db:    app.DbCheck(),
		}

		WriteJSON(w, http.StatusOK, h)
	}
}
