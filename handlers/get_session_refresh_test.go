package handlers_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/keratin/authn-server/handlers"
	"github.com/keratin/authn-server/tokens/sessions"
)

func TestGetSessionRefreshSuccess(t *testing.T) {
	app := testApp()

	account_id := 82594
	session, err := sessions.New(app.RefreshTokenStore, app.Config, account_id)
	if err != nil {
		panic(err)
	}

	injectSessionContext := func(req *http.Request) *http.Request {
		ctx := req.Context()
		ctx = context.WithValue(ctx, handlers.SessionKey, session)
		ctx = context.WithValue(ctx, handlers.AccountIDKey, account_id)
		return req.WithContext(ctx)
	}

	res := get("/session/refresh", app.GetSessionRefresh, injectSessionContext)

	assertCode(t, res, http.StatusCreated)
	assertIdTokenResponse(t, res, app.Config)
}
