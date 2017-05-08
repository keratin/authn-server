package handlers

import (
	"encoding/json"
	"net/http"
)

type ServiceError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ServiceData struct {
	Result interface{} `json:"result"`
}

func writeData(w http.ResponseWriter, d interface{}) {
	writeJson(w, ServiceData{Result: d})
}

type ServiceErrors struct {
	Errors []ServiceError `json:"errors"`
}

func writeErrors(w http.ResponseWriter, e []ServiceError) {
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
