package handlers_test

import (
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
)

func TestPostPasswordScore(t *testing.T) {
	test.App()
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0])

	t.Run("Should successfully score password using JSON request", func(t *testing.T) {
		res, err := client.PostJSON("/password/score", map[string]interface{}{"password": "aSmallPassword"})

		require.NoError(t, err)
		test.AssertData(t, res, map[string]interface{}{"score": 3, "requiredScore": 2})
	})

	t.Run("Should successfully score password using form request", func(t *testing.T) {
		res, err := client.PostForm("/password/score", url.Values{"password": []string{"anotherBetterPassword!"}})

		require.NoError(t, err)
		test.AssertData(t, res, map[string]interface{}{"score": 4, "requiredScore": 2})
	})

	t.Run("Should accuse missing password", func(t *testing.T) {
		res, err := client.PostJSON("/password/score?password=", map[string]interface{}{})

		require.NoError(t, err)
		test.AssertErrors(t, res, services.FieldErrors{services.FieldError{
			Field:   "password",
			Message: services.ErrMissing,
		}})
	})

}

