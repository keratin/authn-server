package gateway

import (
	"net/http"
	"os"

	gorilla "github.com/gorilla/handlers"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/ops"
	"github.com/keratin/authn-server/server/cors"
	"github.com/keratin/authn-server/server/sessions"
)

func WrapRouter(r http.Handler, app *app.App) http.Handler {
	stack := gorilla.CombinedLoggingHandler(os.Stdout, r)

	stack = sessions.Middleware(app)(stack)

	stack = gorilla.CORS(
		gorilla.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE"}),
		gorilla.AllowCredentials(),
		gorilla.AllowedOrigins([]string{}), // see: https://github.com/gorilla/handlers/issues/117
		gorilla.AllowedOriginValidator(cors.OriginValidator(app.Config.ApplicationDomains)),
	)(stack)

	if app.Config.Proxied {
		stack = gorilla.ProxyHeaders(stack)
	}

	return ops.PanicHandler(app.Reporter, stack)
}
