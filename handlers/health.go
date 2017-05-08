package handlers

import (
	"encoding/json"
	"net/http"
)

type health struct {
	Http bool `json:"http"`
}

func (app App) Health(w http.ResponseWriter, req *http.Request) {
	h := health{true}

	j, err := json.Marshal(h)
	if err != nil {
		panic("TODO: gorilla RecoveryHandler")
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
