package handlers

import (
	"fmt"
	"net/http"
)

func (app App) Stub(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(fmt.Sprintf("[%v] not implemented", app.Name)))
}
