package handlers

import (
	"net/http"
)

func Stub(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("not implemented"))
	}
}
