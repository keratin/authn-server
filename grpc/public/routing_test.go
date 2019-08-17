package public

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/keratin/authn-server/lib/oauth"
	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/require"
	"github.com/test-go/testify/assert"
	"golang.org/x/oauth2"
)

// Despite some routes having multiple methods,
// they are considered different from the *mux.Router perspective
// because they are registered newly for each method and the methods
// they support are not collected into a single slice. This approach
// is a workaround.
var pathMethods = map[string]bool{
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
}

func copyMap(src map[string]bool) map[string]bool {
	dst := make(map[string]bool)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func TestRegisterRoutes(t *testing.T) {
	testOAuthProvider := oauth.NewProvider(&oauth2.Config{}, func(t *oauth2.Token) (*oauth.UserInfo, error) {
		return &oauth.UserInfo{
			ID:    "test@example.com",
			Email: "test@example.com",
		}, nil
	})
	t.Run("All routes are registered when SignupEnabled is true, AppPasswordResetURL & AppPasswordlessTokenURL are not nil, and OAuth provider 'test' is configured", func(t *testing.T) {
		gmux := runtime.NewServeMux()
		router := mux.NewRouter()
		app := test.App()
		app.Config.EnableSignup = true
		app.OauthProviders["test"] = *testOAuthProvider
		RegisterRoutes(router, app, gmux)

		// to collect all the registered routes of *mux.Router
		registered := make(map[string]bool)
		expected := copyMap(pathMethods)

		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			methods, err := route.GetMethods()
			require.NoError(t, err)
			pathtpl, err := route.GetPathTemplate()
			require.NoError(t, err)

			registered[fmt.Sprintf("%+v:%s", methods, pathtpl)] = true
			return nil
		})
		assert.EqualValues(t, expected, registered)
	})
	t.Run("/accounts and /accounts/available routes are not registered when SignupEnabled is false", func(t *testing.T) {
		gmux := runtime.NewServeMux()
		router := mux.NewRouter()
		app := test.App()
		app.Config.EnableSignup = false
		app.OauthProviders["test"] = *testOAuthProvider
		RegisterRoutes(router, app, gmux)

		// to collect all the registered routes of *mux.Router
		registered := make(map[string]bool)
		expected := copyMap(pathMethods)

		// These 2 routes shouldn't be registered when signup is disabled
		delete(expected, "[POST]:/accounts")
		delete(expected, "[GET]:/accounts/available")

		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			methods, err := route.GetMethods()
			require.NoError(t, err)
			pathtpl, err := route.GetPathTemplate()
			require.NoError(t, err)
			pathKey := fmt.Sprintf("%+v:%s", methods, pathtpl)
			assert.NotEqual(t, "[POST]:/accounts", pathKey)
			assert.NotEqual(t, "[GET]:/accounts/available", pathKey)

			registered[pathKey] = true
			return nil
		})
		assert.EqualValues(t, expected, registered)
	})
	t.Run("{GET,POST}:/session/token routes are not registered when AppPasswordlessTokenURL is not configured", func(t *testing.T) {
		gmux := runtime.NewServeMux()
		router := mux.NewRouter()
		app := test.App()
		app.Config.AppPasswordlessTokenURL = nil
		app.OauthProviders["test"] = *testOAuthProvider
		RegisterRoutes(router, app, gmux)

		// to collect all the registered routes of *mux.Router
		registered := make(map[string]bool)
		expected := copyMap(pathMethods)

		// Passwordless authentication related routes
		delete(expected, "[POST]:/session/token")
		delete(expected, "[GET]:/session/token")

		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			methods, err := route.GetMethods()
			require.NoError(t, err)
			pathtpl, err := route.GetPathTemplate()
			require.NoError(t, err)
			pathKey := fmt.Sprintf("%+v:%s", methods, pathtpl)
			assert.NotEqual(t, "[POST]:/session/token", pathKey)
			assert.NotEqual(t, "[GET]:/session/token", pathKey)

			registered[pathKey] = true
			return nil
		})
		assert.EqualValues(t, expected, registered)
	})
	t.Run("/password/reset is not registered when AppPasswordResetURL is not configured", func(t *testing.T) {
		gmux := runtime.NewServeMux()
		router := mux.NewRouter()
		app := test.App()
		app.Config.AppPasswordResetURL = nil
		app.OauthProviders["test"] = *testOAuthProvider
		RegisterRoutes(router, app, gmux)

		// to collect all the registered routes of *mux.Router
		registered := make(map[string]bool)
		expected := copyMap(pathMethods)

		// Remove the route responsible for handling password reset requests
		delete(expected, "[GET]:/password/reset")

		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			methods, err := route.GetMethods()
			require.NoError(t, err)
			pathtpl, err := route.GetPathTemplate()
			require.NoError(t, err)
			pathKey := fmt.Sprintf("%+v:%s", methods, pathtpl)
			assert.NotEqual(t, "[GET]:/password/reset", pathKey)

			registered[pathKey] = true
			return nil
		})
		assert.EqualValues(t, expected, registered)
	})
	t.Run("/oauth prefixed routes are not registered when no OAuth providers are configured", func(t *testing.T) {
		gmux := runtime.NewServeMux()
		router := mux.NewRouter()
		app := test.App()
		require.Len(t, app.OauthProviders, 0)
		RegisterRoutes(router, app, gmux)

		// to collect all the registered routes of *mux.Router
		registered := make(map[string]bool)
		expected := copyMap(pathMethods)

		for k := range expected {
			if strings.Contains(k, "oauth") {
				delete(expected, k)
			}
		}

		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			methods, err := route.GetMethods()
			require.NoError(t, err)
			pathtpl, err := route.GetPathTemplate()
			require.NoError(t, err)
			assert.False(t, strings.HasPrefix(pathtpl, "/oauth"))

			pathKey := fmt.Sprintf("%+v:%s", methods, pathtpl)
			registered[pathKey] = true
			return nil
		})
		assert.EqualValues(t, expected, registered)
	})
}
