package public

import (
	"net"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pkgerrors "github.com/pkg/errors"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	authnpb "github.com/keratin/authn-server/grpc"
	"github.com/keratin/authn-server/grpc/internal/errors"
	"github.com/keratin/authn-server/grpc/internal/meta"
)

// Compile-time check
var _ authnpb.PublicAuthNServer = publicServer{}

type publicServer struct {
	app *app.App
}

func RunPublicGRPC(ctx context.Context, app *app.App, l net.Listener) error {
	opts := meta.PublicServerOptions(app)
	srv := grpc.NewServer(opts...)

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

func RegisterPublicGRPCMethods(srv *grpc.Server, app *app.App) {
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

func (s publicServer) Login(ctx context.Context, req *authnpb.LoginRequest) (*authnpb.LoginResponseEnvelope, error) {
	account, err := services.CredentialsVerifier(
		s.app.AccountStore,
		s.app.Config,
		req.GetUsername(),
		req.GetPassword(),
	)
	if err != nil {
		if fe, ok := err.(services.FieldErrors); ok {
			return nil, errors.ToStatusErrorWithDetails(fe, codes.FailedPrecondition).Err()
		}
		panic(err)
	}

	sessionToken, identityToken, err := services.SessionCreator(
		s.app.AccountStore, s.app.RefreshTokenStore, s.app.KeyStore, s.app.Actives, s.app.Config, s.app.Reporter,
		account.ID, &s.app.Config.ApplicationDomains[0], meta.GetRefreshToken(ctx),
	)
	if err != nil {
		panic(err)
	}

	// Return the signed session in a metadata
	meta.SetSession(ctx, s.app.Config.SessionCookieName, sessionToken)

	// Return the signed identity token in the body
	return &authnpb.LoginResponseEnvelope{
		Result: &authnpb.LoginResponse{
			IdToken: identityToken,
		},
	}, nil
}

func (s publicServer) RefreshSession(ctx context.Context, _ *authnpb.RefreshSessionRequest) (*authnpb.RefreshSessionResponseEnvelope, error) {

	// check for valid session with live token
	accountID := meta.GetSessionAccountID(ctx)
	if accountID == 0 {
		return nil, status.Error(codes.Unauthenticated, "account not found")
	}

	identityToken, err := services.SessionRefresher(
		s.app.RefreshTokenStore, s.app.KeyStore, s.app.Actives, s.app.Config, s.app.Reporter,
		meta.GetSession(ctx), accountID, &s.app.Config.ApplicationDomains[0],
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

	err := services.SessionEnder(s.app.RefreshTokenStore, meta.GetRefreshToken(ctx))
	if err != nil {
		s.app.Reporter.ReportError(err)
	}

	meta.SetSession(ctx, s.app.Config.SessionCookieName, "")

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
		accountID = meta.GetSessionAccountID(ctx)
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
		accountID, &s.app.Config.ApplicationDomains[0], meta.GetRefreshToken(ctx),
	)
	if err != nil {
		panic(err)
	}

	// Return the signed session in a cookie
	meta.SetSession(ctx, s.app.Config.SessionCookieName, sessionToken)

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
