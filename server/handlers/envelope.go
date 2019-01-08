package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/keratin/authn-server/app/services"
)

type ServiceData struct {
	Result interface{} `json:"result"`
}

type ServiceErrors struct {
	Errors services.FieldErrors `json:"errors"`
}

func WriteData(w http.ResponseWriter, httpCode int, d interface{}) {
	WriteJSON(w, httpCode, ServiceData{Result: d})
}

func WriteErrors(w http.ResponseWriter, e services.FieldErrors) {
	WriteJSON(w, http.StatusUnprocessableEntity, ServiceErrors{Errors: e})
}

func WriteNotFound(w http.ResponseWriter, resource string) {
	WriteJSON(w, http.StatusNotFound, ServiceErrors{Errors: services.FieldErrors{{resource, services.ErrNotFound}}})
}

func WriteJSON(w http.ResponseWriter, httpCode int, d interface{}) {
	j, err := json.Marshal(d)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	w.Write(j)
}
