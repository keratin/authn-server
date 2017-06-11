package handlers

import (
	"net/http"

	"github.com/keratin/authn/services"
)

type request struct {
	Username string
	Password string
}

type response struct {
	IdToken string `json:"id_token"`
}

func (app App) PostAccount(w http.ResponseWriter, req *http.Request) {
	account, errors := services.AccountCreator(
		app.AccountStore,
		&app.Config,
		req.FormValue("username"),
		req.FormValue("password"),
	)
	if errors != nil {
		writeErrors(w, errors)
		return
	}

	_, err := app.RefreshTokenStore.Create(account.Id)
	if err != nil {
		panic(err)
	}
	accessToken := "j.w.t"

	w.WriteHeader(http.StatusCreated)
	writeData(w, response{accessToken})
}
