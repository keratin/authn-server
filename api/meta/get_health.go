package meta

import (
	"net/http"

	"github.com/keratin/authn-server/api"
)

type health struct {
	Http  bool `json:"http"`
	Db    bool `json:"db"`
	Redis bool `json:"redis"`
}

func getHealth(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := health{
			Http:  true,
			Redis: app.RedisCheck(),
			Db:    app.DbCheck(),
		}

		api.WriteJson(w, http.StatusOK, h)
	}
}
