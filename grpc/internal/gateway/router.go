package gateway

import (
	"net/http"
	"os"

	gorilla "github.com/gorilla/handlers"
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/ops"
)

func WrapRouter(r http.Handler, app *api.App) http.Handler {
	stack := gorilla.CombinedLoggingHandler(os.Stdout, r)

	stack = api.Session(app)(stack)

	stack = gorilla.CORS(
		gorilla.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}),
		gorilla.AllowCredentials(),
		gorilla.AllowedOrigins([]string{}), // see: https://github.com/gorilla/handlers/issues/117
		gorilla.AllowedOriginValidator(api.OriginValidator(app.Config.ApplicationDomains)),
	)(stack)

	if app.Config.Proxied {
		stack = gorilla.ProxyHeaders(stack)
	}

	return ops.PanicHandler(app.Reporter, stack)
}
