package passwords

import "github.com/keratin/authn-server/api"

func Routes(app *api.App) []*api.HandledRoute {
	refererSecurity := api.RefererSecurity(app.Config.ApplicationDomains)

	routes := []*api.HandledRoute{
		api.Post("/password").
			SecuredWith(refererSecurity).
			Handle(postPassword(app)),
	}

	if app.Config.AppPasswordResetURL != nil {
		routes = append(routes,
			api.Get("/password/reset").
				SecuredWith(refererSecurity).
				Handle(getPasswordReset(app)),
		)
	}

	return routes
}
