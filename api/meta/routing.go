package meta

import (
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/lib/route"
)

func Routes(app *api.App) []*route.HandledRoute {
	authentication := route.BasicAuthSecurity(app.Config.AuthUsername, app.Config.AuthPassword, "Private AuthN Realm")

	return []*route.HandledRoute{
		route.Get("/").
			SecuredWith(route.Unsecured()).
			Handle(getRoot(app)),
		route.Get("/health").
			SecuredWith(route.Unsecured()).
			Handle(getHealth(app)),
		route.Get("/jwks").
			SecuredWith(route.Unsecured()).
			Handle(getJWKs(app)),
		route.Get("/configuration").
			SecuredWith(route.Unsecured()).
			Handle(getConfiguration(app)),
		route.Get("/stats").
			SecuredWith(authentication).
			Handle(getStats(app)),
	}
}
