package handlers_test

import (
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPostPasswordScore(t *testing.T) {
	test.App()
	app := test.App()
	server := test.Server(app)
	defer server.Close()

	client := route.NewClient(server.URL).Referred(&app.Config.ApplicationDomains[0])

	testCases := []struct {
		Description string
		Password     string
		Result      map[string]interface{} // expected JSON response
	}{
		{Description: "Should return score 1", Password: "strongerPassword", Result: map[string]interface{}{"score": 1, "requiredScore": 2}},
		{Description: "Should return score 3", Password: "strongerPassword17", Result: map[string]interface{}{"score": 3, "requiredScore": 2}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			res, err := client.PostJSON("/password/score", map[string]interface{}{"password": testCase.Password})

			require.NoError(t, err)
			test.AssertData(t, res, testCase.Result)
		})
	}

	t.Run("Should accuse missing password", func(t *testing.T) {
		res, err := client.PostJSON("/password/score?password=", map[string]interface{}{})

		require.NoError(t, err)
		test.AssertErrors(t, res, services.FieldErrors{services.FieldError{
			Field:   "password",
			Message: services.ErrMissing,
		}})
	})
}

