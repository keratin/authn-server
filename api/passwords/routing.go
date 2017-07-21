package passwords

import "github.com/keratin/authn-server/api"

func Routes(app *api.App) []*api.HandledRoute {
	refererSecurity := api.RefererSecurity(app.Config.ApplicationDomains)

	return []*api.HandledRoute{
		api.Post("/password").
			SecuredWith(refererSecurity).
			Handle(postPassword(app)),
	}
}
