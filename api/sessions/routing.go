package sessions

import (
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/lib/route"
)

func Routes(app *api.App) []*route.HandledRoute {
	originSecurity := route.OriginSecurity(app.Config.ApplicationDomains)

	return []*route.HandledRoute{
		route.Post("/session").
			SecuredWith(originSecurity).
			Handle(postSession(app)),

		route.Delete("/session").
			SecuredWith(originSecurity).
			Handle(deleteSession(app)),

		route.Get("/session/refresh").
			SecuredWith(originSecurity).
			Handle(getSessionRefresh(app)),
	}
}
