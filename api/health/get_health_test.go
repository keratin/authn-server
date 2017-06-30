package health_test

import (
	"net/http"
	"testing"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/api/health"
	"github.com/keratin/authn-server/api/test"
)

func TestHealth(t *testing.T) {
	app := api.App{
		DbCheck:    func() bool { return true },
		RedisCheck: func() bool { return true },
	}

	res := test.Get("/health", health.Health(&app))

	test.AssertCode(t, res, http.StatusOK)
	test.AssertBody(t, res, `{"http":true,"db":true,"redis":true}`)
}
