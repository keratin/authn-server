package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/keratin/authn/services"
)

type ServiceData struct {
	Result interface{} `json:"result"`
}

func writeData(w http.ResponseWriter, httpCode int, d interface{}) {
	writeJson(w, httpCode, ServiceData{Result: d})
}

type ServiceErrors struct {
	Errors []services.Error `json:"errors"`
}

func writeErrors(w http.ResponseWriter, e []services.Error) {
	writeJson(w, http.StatusUnprocessableEntity, ServiceErrors{Errors: e})
}

func writeJson(w http.ResponseWriter, httpCode int, d interface{}) {
	j, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	w.Write(j)
}
