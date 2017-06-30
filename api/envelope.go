package api

import (
	"encoding/json"
	"net/http"

	"github.com/keratin/authn-server/services"
)

type ServiceData struct {
	Result interface{} `json:"result"`
}

type ServiceErrors struct {
	Errors []services.Error `json:"errors"`
}

func WriteData(w http.ResponseWriter, httpCode int, d interface{}) {
	WriteJson(w, httpCode, ServiceData{Result: d})
}

func WriteErrors(w http.ResponseWriter, e []services.Error) {
	WriteJson(w, http.StatusUnprocessableEntity, ServiceErrors{Errors: e})
}

func WriteJson(w http.ResponseWriter, httpCode int, d interface{}) {
	j, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	w.Write(j)
}
