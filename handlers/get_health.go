package handlers

import (
	"net/http"
)

type health struct {
	Http  bool `json:"http"`
	Db    bool `json:"db"`
	Redis bool `json:"redis"`
}

func Health(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		h := health{
			Http:  true,
			Redis: app.RedisCheck(),
			Db:    app.DbCheck(),
		}

		writeJson(w, http.StatusOK, h)
	}
}
