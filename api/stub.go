package api

import (
	"net/http"
)

func Stub(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not implemented"))
	}
}
