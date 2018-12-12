package api

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/api/accounts"
	"github.com/keratin/authn-server/api/cors"
	"github.com/keratin/authn-server/api/meta"
	"github.com/keratin/authn-server/api/oauth"
	"github.com/keratin/authn-server/api/passwords"
	"github.com/keratin/authn-server/api/sessions"
	"github.com/keratin/authn-server/api/sessionz"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/ops"
)

func router(app *app.App) http.Handler {
	r := mux.NewRouter()
	route.Attach(r, app.Config.MountedPath, meta.Routes(app)...)
	route.Attach(r, app.Config.MountedPath, accounts.Routes(app)...)
	route.Attach(r, app.Config.MountedPath, sessions.Routes(app)...)
	route.Attach(r, app.Config.MountedPath, passwords.Routes(app)...)
	route.Attach(r, app.Config.MountedPath, oauth.Routes(app)...)

	return wrapRouter(r, app)
}

func publicRouter(app *app.App) http.Handler {
	r := mux.NewRouter()
	route.Attach(r, app.Config.MountedPath, meta.PublicRoutes(app)...)
	route.Attach(r, app.Config.MountedPath, accounts.PublicRoutes(app)...)
	route.Attach(r, app.Config.MountedPath, sessions.PublicRoutes(app)...)
	route.Attach(r, app.Config.MountedPath, passwords.PublicRoutes(app)...)
	route.Attach(r, app.Config.MountedPath, oauth.PublicRoutes(app)...)

	return wrapRouter(r, app)
}

func wrapRouter(r *mux.Router, app *app.App) http.Handler {
	stack := handlers.CombinedLoggingHandler(os.Stdout, r)
	stack = sessionz.Middleware(app)(stack)
	stack = cors.Middleware(app)(stack)

	if app.Config.Proxied {
		stack = handlers.ProxyHeaders(stack)
	}

	return ops.PanicHandler(app.Reporter, stack)
}
