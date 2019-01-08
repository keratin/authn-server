package sessions

import (
	"context"
	"net/http"
	"sync"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/models"
	"github.com/keratin/authn-server/app/tokens/sessions"
	"github.com/pkg/errors"
)

type sessionKey int
type accountIDKey int

func Middleware(app *app.App) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var session *sessions.Claims
			var parseOnce sync.Once
			parse := func() *sessions.Claims {
				parseOnce.Do(func() {
					cookie, err := r.Cookie(app.Config.SessionCookieName)
					if err == http.ErrNoCookie {
						return
					} else if err != nil {
						app.Reporter.ReportRequestError(errors.Wrap(err, "Cookie"), r)
						return
					}

					session, err = sessions.Parse(cookie.Value, app.Config)
					if err != nil {
						app.Reporter.ReportRequestError(errors.Wrap(err, "Parse"), r)
					}
				})

				return session
			}

			var accountID int
			var lookupOnce sync.Once
			lookup := func() int {
				lookupOnce.Do(func() {
					var err error
					session := parse()
					if session == nil {
						return
					}

					accountID, err = app.RefreshTokenStore.Find(models.RefreshToken(session.Subject))
					if err != nil {
						app.Reporter.ReportRequestError(errors.Wrap(err, "Find"), r)
					}
				})

				return accountID
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, sessionKey(0), parse)
			ctx = context.WithValue(ctx, accountIDKey(0), lookup)

			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
