package passwords

import (
	"net/http"

	"github.com/keratin/authn-server/api"
)

func getPasswordReset(app *api.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
