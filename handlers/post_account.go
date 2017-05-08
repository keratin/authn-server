package handlers

import (
	"net/http"
)

type response struct {
	IdToken string `json:"id_token"`
}

func (app App) PostAccount(w http.ResponseWriter, req *http.Request) {
	service := AccountCreator{
		Username: req.FormValue("username"),
		Password: req.FormValue("password"),
	}

	_, errors := service.perform()
	if errors != nil {
		writeErrors(w, errors)
		return
	}

	token := "j.w.t"

	w.WriteHeader(http.StatusCreated)
	writeData(w, response{token})
}

type AccountCreator struct {
	Username string
	Password string
}

func (s *AccountCreator) perform() (bool, []ServiceError) {
	if s.Username != "" && s.Password != "" {
		return true, nil
	} else {
		errors := make([]ServiceError, 0, 1)
		errors = append(errors, ServiceError{Field: "foo", Message: "bar"})

		return false, errors
	}
}
