package private

import (
	"fmt"
	"testing"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/keratin/authn-server/lib/oauth"
	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/require"
	"github.com/test-go/testify/assert"
	"golang.org/x/oauth2"
)

// corre
var pathMethods = map[string]bool{
	// correlated with public routes
	"[GET]:/health":             true,
	"[POST]:/accounts":          true,
	"[GET]:/accounts/available": true,
	"[DELETE]:/session":         true,
	"[POST]:/session":           true,
	"[GET]:/session/refresh":    true,
	"[GET]:/session/token":      true,
	"[POST]:/session/token":     true,
	"[POST]:/password":          true,
	"[GET]:/password/reset":     true,
	"[GET]:/oauth/test":         true,
	"[GET]:/oauth/test/return":  true,

	// private-only routes
	"[GET]:/stats":                                  true,
	"[GET]:/jwks":                                   true,
	"[GET]:/configuration":                          true,
	"[GET]:/metrics":                                true,
	"[POST]:/accounts/import":                       true,
	"[GET]:/accounts/{id:[0-9]+}":                   true,
	"[PATCH]:/accounts/{id:[0-9]+}":                 true,
	"[PATCH]:/accounts/{id:[0-9]+}/lock":            true,
	"[PATCH]:/accounts/{id:[0-9]+}/unlock":          true,
	"[PATCH]:/accounts/{id:[0-9]+}/expire_password": true,
	"[PUT]:/accounts/{id:[0-9]+}":                   true,
	"[PUT]:/accounts/{id:[0-9]+}/lock":              true,
	"[PUT]:/accounts/{id:[0-9]+}/unlock":            true,
	"[PUT]:/accounts/{id:[0-9]+}/expire_password":   true,
	"[DELETE]:/accounts/{id:[0-9]+}":                true,
}

func TestRegisterRoutes(t *testing.T) {
	testOAuthProvider := oauth.NewProvider(&oauth2.Config{}, func(t *oauth2.Token) (*oauth.UserInfo, error) {
		return &oauth.UserInfo{
			ID:    "test@example.com",
			Email: "test@example.com",
		}, nil
	})
	t.Run("App with SignupEnabled & AppPasswordResetURL & Actives enabled has all routes registered", func(t *testing.T) {
		gmux := runtime.NewServeMux()
		router := mux.NewRouter()
		app := test.App()
		app.Config.EnableSignup = true
		app.OauthProviders["test"] = *testOAuthProvider
		RegisterRoutes(router, app, gmux)

		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			methods, err := route.GetMethods()
			require.NoError(t, err)
			pathtpl, err := route.GetPathTemplate()
			require.NoError(t, err)
			assert.True(t, pathMethods[fmt.Sprintf("%+v:%s", methods, pathtpl)])
			return nil
		})
	})
	t.Run("App with SignupEnabled disabled doesn't have /accounts and /accounts/available routes registered", func(t *testing.T) {
		gmux := runtime.NewServeMux()
		router := mux.NewRouter()
		app := test.App()
		app.Config.EnableSignup = false
		RegisterRoutes(router, app, gmux)

		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			methods, err := route.GetMethods()
			require.NoError(t, err)
			pathtpl, err := route.GetPathTemplate()
			require.NoError(t, err)
			pathKey := fmt.Sprintf("%+v:%s", methods, pathtpl)
			assert.NotEqual(t, "[POST]:/accounts", pathKey)
			assert.NotEqual(t, "[GET]:/accounts/available", pathKey)
			return nil
		})
	})
	t.Run("App with AppPasswordlessTokenURL nil doesn't have {GET,POST}:/session/token routes registered", func(t *testing.T) {
		gmux := runtime.NewServeMux()
		router := mux.NewRouter()
		app := test.App()
		app.Config.AppPasswordlessTokenURL = nil
		RegisterRoutes(router, app, gmux)

		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			methods, err := route.GetMethods()
			require.NoError(t, err)
			pathtpl, err := route.GetPathTemplate()
			require.NoError(t, err)
			pathKey := fmt.Sprintf("%+v:%s", methods, pathtpl)
			assert.NotEqual(t, "[POST]:/session/token", pathKey)
			assert.NotEqual(t, "[GET]:/session/token", pathKey)
			return nil
		})
	})
	t.Run("App with AppPasswordResetURL nil doesn't have /password/reset route registered", func(t *testing.T) {
		gmux := runtime.NewServeMux()
		router := mux.NewRouter()
		app := test.App()
		app.Config.AppPasswordResetURL = nil
		RegisterRoutes(router, app, gmux)

		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			methods, err := route.GetMethods()
			require.NoError(t, err)
			pathtpl, err := route.GetPathTemplate()
			require.NoError(t, err)
			pathKey := fmt.Sprintf("%+v:%s", methods, pathtpl)
			assert.NotEqual(t, "[GET]:/password/reset", pathKey)
			return nil
		})
	})
	t.Run("App with Actives nil doesn't have /stats route registered", func(t *testing.T) {
		gmux := runtime.NewServeMux()
		router := mux.NewRouter()
		app := test.App()
		app.Actives = nil
		RegisterRoutes(router, app, gmux)

		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			methods, err := route.GetMethods()
			require.NoError(t, err)
			pathtpl, err := route.GetPathTemplate()
			require.NoError(t, err)
			pathKey := fmt.Sprintf("%+v:%s", methods, pathtpl)
			assert.NotEqual(t, "[GET]:/stats", pathKey)
			return nil
		})
	})
}
