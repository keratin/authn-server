package public

import (
	"fmt"
	"net"
	"sync"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pkgerrors "github.com/pkg/errors"

	"github.com/keratin/authn-server/api"
	authnpb "github.com/keratin/authn-server/grpc"
	"github.com/keratin/authn-server/grpc/internal/errors"
	"github.com/keratin/authn-server/models"
	"github.com/keratin/authn-server/services"
	"github.com/keratin/authn-server/tokens/sessions"
)

// Compile-time check
var _ authnpb.PublicAuthNServer = publicServer{}

type sessionKey int
type accountIDKey int

type publicServer struct {
	app *api.App
}

func RunPublicGRPC(ctx context.Context, app *api.App, l net.Listener) error {
	srv := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			logInterceptor,
			sessionInterceptor(app),
		),
	)

	RegisterPublicGRPCMethods(srv, app)

	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()

	if err := srv.Serve(l); err != nil {
		log.Printf("serve error: %s", err)
		return err
	}
	return nil
}

func RegisterPublicGRPCMethods(srv *grpc.Server, app *api.App) {
	authnpb.RegisterPublicAuthNServer(srv, publicServer{
		app: app,
	})

	if app.Config.EnableSignup {
		authnpb.RegisterSignupServiceServer(srv, signupServiceServer{
			app: app,
		})
	}

	if app.Config.AppPasswordResetURL != nil {
		authnpb.RegisterPasswordResetServiceServer(srv, passwordResetServer{
			app: app,
		})
	}

	if app.Config.AppPasswordlessTokenURL != nil {
		authnpb.RegisterPasswordlessServiceServer(srv, passwordlessServer{
			app: app,
		})
	}
}

func sessionInterceptor(app *api.App) grpc.UnaryServerInterceptor {
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
					app.Reporter.ReportError(pkgerrors.Wrap(err, "Parse"))
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
					app.Reporter.ReportError(pkgerrors.Wrap(err, "Find"))
				}
			})
			return accountID
		}
		ctx = context.WithValue(ctx, sessionKey(0), parse)
		ctx = context.WithValue(ctx, accountIDKey(0), lookup)
		return handler(ctx, req)
	}
}

func logInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Infof("calling method: %s", info.FullMethod)

	res, err := handler(ctx, req)
	if err != nil {
		log.Errorf("error from method: %s", err)
		log.Errorf("error type: %T", err)
	}
	return res, err
}

func (s publicServer) Login(ctx context.Context, req *authnpb.LoginRequest) (*authnpb.LoginResponseEnvelope, error) {
	account, err := services.CredentialsVerifier(
		s.app.AccountStore,
		s.app.Config,
		req.GetUsername(),
		req.GetPassword(),
	)
	if err != nil {
		if fe, ok := err.(services.FieldErrors); ok {
			br := &errdetails.BadRequest{}
			for _, fe := range fe {
				br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
					Field:       fe.Field,
					Description: fe.Message,
				})
			}
			statusError := status.New(codes.FailedPrecondition, fe.Error())
			statusEr, e := statusError.WithDetails(br)
			if e != nil {
				panic(fmt.Sprintf("Unexpected error attaching metadata: %v", e))
			}
			return nil, statusEr.Err()
		}
		panic(err)
	}

	sessionToken, identityToken, err := services.SessionCreator(
		s.app.AccountStore, s.app.RefreshTokenStore, s.app.KeyStore, s.app.Actives, s.app.Config, s.app.Reporter,
		account.ID, &s.app.Config.ApplicationDomains[0], getRefreshToken(ctx),
	)
	if err != nil {
		panic(err)
	}

	// Return the signed session in a metadata
	setSession(ctx, s.app.Config.SessionCookieName, sessionToken)

	// Return the signed identity token in the body
	return &authnpb.LoginResponseEnvelope{
		Result: &authnpb.LoginResponse{
			IdToken: identityToken,
		},
	}, nil
}

func (s publicServer) RefreshSession(ctx context.Context, _ *authnpb.RefreshSessionRequest) (*authnpb.RefreshSessionResponseEnvelope, error) {

	// check for valid session with live token
	accountID := getSessionAccountID(ctx)
	if accountID == 0 {
		return nil, status.Error(codes.Unauthenticated, "account not found")
	}

	identityToken, err := services.SessionRefresher(
		s.app.RefreshTokenStore, s.app.KeyStore, s.app.Actives, s.app.Config, s.app.Reporter,
		getSession(ctx), accountID, &s.app.Config.ApplicationDomains[0],
	)
	if err != nil {
		panic(pkgerrors.Wrap(err, "IdentityForSession"))
	}

	return &authnpb.RefreshSessionResponseEnvelope{
		Result: &authnpb.RefreshSessionResponse{
			IdToken: identityToken,
		},
	}, nil
}

func (s publicServer) Logout(ctx context.Context, _ *authnpb.LogoutRequest) (*authnpb.LogoutResponse, error) {

	err := services.SessionEnder(s.app.RefreshTokenStore, getRefreshToken(ctx))
	if err != nil {
		s.app.Reporter.ReportError(err)
	}

	setSession(ctx, s.app.Config.SessionCookieName, "")

	return &authnpb.LogoutResponse{}, nil
}

func (s publicServer) ChangePassword(ctx context.Context, req *authnpb.ChangePasswordRequest) (*authnpb.ChangePasswordResponseEnvelope, error) {

	var err error
	var accountID int
	if req.GetToken() != "" {
		accountID, err = services.PasswordResetter(
			s.app.AccountStore,
			s.app.Reporter,
			s.app.Config,
			req.GetToken(),
			req.GetPassword(),
		)
	} else {
		accountID = getSessionAccountID(ctx)
		if accountID == 0 {
			return nil, status.Error(codes.Unauthenticated, "account")
		}
		err = services.PasswordChanger(
			s.app.AccountStore,
			s.app.Reporter,
			s.app.Config,
			accountID,
			req.GetCurrentPassword(),
			req.GetPassword(),
		)
	}

	if err != nil {
		if fe, ok := err.(services.FieldErrors); ok {
			return nil, errors.ToStatusErrorWithDetails(fe, codes.FailedPrecondition).Err()
		}
		panic(err)
	}

	sessionToken, identityToken, err := services.SessionCreator(
		s.app.AccountStore, s.app.RefreshTokenStore, s.app.KeyStore, s.app.Actives, s.app.Config, s.app.Reporter,
		accountID, &s.app.Config.ApplicationDomains[0], getRefreshToken(ctx),
	)
	if err != nil {
		panic(err)
	}

	// Return the signed session in a cookie
	setSession(ctx, s.app.Config.SessionCookieName, sessionToken)

	// Return the signed identity token in the body
	return &authnpb.ChangePasswordResponseEnvelope{
		Result: &authnpb.ChangePasswordResponse{
			IdToken: identityToken,
		},
	}, nil
}

func (s publicServer) HealthCheck(context.Context, *authnpb.HealthCheckRequest) (*authnpb.HealthCheckResponse, error) {
	return &authnpb.HealthCheckResponse{
		Http:  true,
		Redis: s.app.RedisCheck(),
		Db:    s.app.DbCheck(),
	}, nil
}

func getRefreshToken(ctx context.Context) *models.RefreshToken {
	claims := getSession(ctx)
	if claims != nil {
		token := models.RefreshToken(claims.Subject)
		return &token
	}
	return nil
}

func getSession(ctx context.Context) *sessions.Claims {
	fn, ok := ctx.Value(sessionKey(0)).(func() *sessions.Claims)
	if ok {
		return fn()
	}
	return nil
}

func setSession(ctx context.Context, cookieName string, val string) {
	// create and send header
	header := metadata.Pairs(cookieName, val)
	grpc.SendHeader(ctx, header)
}

func getSessionAccountID(ctx context.Context) int {
	fn, ok := ctx.Value(accountIDKey(0)).(func() int)
	if ok {
		return fn()
	}
	return 0
}
