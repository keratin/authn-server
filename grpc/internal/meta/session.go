package meta

import (
	"context"
	"sync"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/models"
	"github.com/keratin/authn-server/app/tokens/sessions"
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type sessionKey int
type accountIDKey int

func sessionInterceptor(app *app.App) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var session *sessions.Claims
		var parseOnce sync.Once
		parse := func() *sessions.Claims {
			parseOnce.Do(func() {
				md, ok := metadata.FromIncomingContext(ctx)
				if !ok {
					return
				}
				cookies := md.Get(app.Config.SessionCookieName)
				if len(cookies) == 0 {
					return
				}
				var err error
				session, err = sessions.Parse(cookies[0], app.Config)
				if err != nil {
					app.Reporter.ReportGRPCError(pkgerrors.Wrap(err, "Parse"), info, req)
					return
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
					app.Reporter.ReportGRPCError(pkgerrors.Wrap(err, "Find"), info, req)
				}
			})
			return accountID
		}
		ctx = context.WithValue(ctx, sessionKey(0), parse)
		ctx = context.WithValue(ctx, accountIDKey(0), lookup)
		return handler(ctx, req)
	}
}

// GetRefreshToken extracts the refresh token from the context.
func GetRefreshToken(ctx context.Context) *models.RefreshToken {
	claims := GetSession(ctx)
	if claims != nil {
		token := models.RefreshToken(claims.Subject)
		return &token
	}
	return nil
}

// GetSession extracts the session claim from the request context.
func GetSession(ctx context.Context) *sessions.Claims {
	fn, ok := ctx.Value(sessionKey(0)).(func() *sessions.Claims)
	if ok {
		return fn()
	}
	return nil
}

// SetSession sends the session claim in the gRPC response header.
func SetSession(ctx context.Context, cookieName string, val string) {
	// create and send header
	header := metadata.Pairs(cookieName, val)
	grpc.SendHeader(ctx, header)
}

// GetSessionAccountID extracts the account ID stored in the request context.
func GetSessionAccountID(ctx context.Context) int {
	fn, ok := ctx.Value(accountIDKey(0)).(func() int)
	if ok {
		return fn()
	}
	return 0
}
