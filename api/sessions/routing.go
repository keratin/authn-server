package sessions

import (
	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/lib/route"
)

func Routes(app *api.App) []*route.HandledRoute {
	refererSecurity := route.RefererSecurity(app.Config.ApplicationDomains)

	return []*route.HandledRoute{
		route.Post("/session").
			SecuredWith(refererSecurity).
			Handle(postSession(app)),

		route.Delete("/session").
			SecuredWith(refererSecurity).
			Handle(deleteSession(app)),

		route.Get("/session/refresh").
			SecuredWith(refererSecurity).
			Handle(getSessionRefresh(app)),
	}
}
