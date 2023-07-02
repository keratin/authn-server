package handlers

import (
	"bytes"
	"net/http"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/server/views"
)

func GetRoot(app *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var buf bytes.Buffer
		views.Root(&buf)

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	}
}
