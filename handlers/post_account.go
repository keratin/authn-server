package handlers

import (
	"encoding/json"
	"net/http"
)

type response struct {
	IdToken string `json:"id_token"`
}

func (app App) PostAccount(w http.ResponseWriter, req *http.Request) {
	if req.ContentLength > 0 {
		token := "j.w.t"

		w.WriteHeader(http.StatusCreated)
		writeData(w, response{token})
	} else {
		e := make([]ServiceError, 0, 1)
		e = append(e, ServiceError{Field: "foo", Message: "bar"})

		writeErrors(w, e)
	}
}

type ServiceError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ServiceErrors struct {
	Errors []ServiceError `json:"errors"`
}

type ServiceData struct {
	Result interface{} `json:"result"`
}

func writeData(w http.ResponseWriter, d interface{}) {
	writeJson(w, ServiceData{Result: d})
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
