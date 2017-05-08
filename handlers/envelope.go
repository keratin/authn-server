package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/keratin/authn/services"
)

type ServiceData struct {
	Result interface{} `json:"result"`
}

func writeData(w http.ResponseWriter, d interface{}) {
	writeJson(w, ServiceData{Result: d})
}

type ServiceErrors struct {
	Errors []services.Error `json:"errors"`
}

func writeErrors(w http.ResponseWriter, e []services.Error) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	writeJson(w, ServiceErrors{Errors: e})
}

func writeJson(w http.ResponseWriter, d interface{}) {
	j, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}
