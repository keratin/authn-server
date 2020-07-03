package cors

import (
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/keratin/authn-server/app"
)

func Middleware(app *app.App) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return handlers.CORS(
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}),
			handlers.AllowCredentials(),
			handlers.AllowedOrigins([]string{}), // see: https://github.com/gorilla/handlers/issues/117
			handlers.AllowedOriginValidator(OriginValidator(app.Config.ApplicationDomains)),
			handlers.AllowedHeaders([]string{"content-type"}),
		)(h)
	}
}
