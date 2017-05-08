package handlers

import (
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
