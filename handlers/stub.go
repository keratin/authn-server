package handlers

import (
	"net/http"
)

func (app App) Stub(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("not implemented"))
}
