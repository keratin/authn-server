package sessions

import (
	"github.com/keratin/authn-server/api"
)

func Routes(app *api.App) []*api.HandledRoute {
	refererSecurity := api.RefererSecurity(app.Config.ApplicationDomains)

	return []*api.HandledRoute{
		api.Post("/session").
			SecuredWith(refererSecurity).
			Handle(postSession(app)),

		api.Delete("/session").
			SecuredWith(refererSecurity).
			Handle(deleteSession(app)),

		api.Get("/session/refresh").
			SecuredWith(refererSecurity).
			Handle(getSessionRefresh(app)),
	}
}
