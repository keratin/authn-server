package handlers

import (
	"net/http"
)

type health struct {
	Http  bool `json:"http"`
	Db    bool `json:"db"`
	Redis bool `json:"redis"`
}

func (app App) Health(w http.ResponseWriter, req *http.Request) {
	h := health{
		Http:  true,
		Redis: app.Redis.Ping().Err() == nil,
		Db:    app.Db.Ping() == nil,
	}

	writeJson(w, http.StatusOK, h)
}
