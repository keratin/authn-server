package health

import "github.com/keratin/authn-server/api"

func Routes(app *api.App) []*api.HandledRoute {
	return []*api.HandledRoute{
		api.Get("/health").
			SecuredWith(api.Unsecured()).
			Handle(getHealth(app)),
	}
}
