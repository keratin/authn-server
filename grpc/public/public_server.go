package public

import (
	"golang.org/x/net/context"
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

func (s publicServer) Login(ctx context.Context, req *authnpb.LoginRequest) (*authnpb.LoginResponse, error) {
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
		account.ID, meta.MatchedDomain(ctx), meta.GetRefreshToken(ctx),
	)
	if err != nil {
		panic(err)
	}

	// Return the signed session in a metadata
	meta.SetSession(ctx, s.app.Config.SessionCookieName, sessionToken)

	// Return the signed identity token in the body
	return &authnpb.LoginResponse{
		IdToken: identityToken,
	}, nil
}

func (s publicServer) RefreshSession(ctx context.Context, _ *authnpb.RefreshSessionRequest) (*authnpb.RefreshSessionResponse, error) {

	// check for valid session with live token
	accountID := meta.GetSessionAccountID(ctx)
	if accountID == 0 {
		return nil, status.Error(codes.Unauthenticated, "account not found")
	}

	identityToken, err := services.SessionRefresher(
		s.app.RefreshTokenStore, s.app.KeyStore, s.app.Actives, s.app.Config, s.app.Reporter,
		meta.GetSession(ctx), accountID, meta.MatchedDomain(ctx),
	)
	if err != nil {
		panic(pkgerrors.Wrap(err, "IdentityForSession"))
	}

	return &authnpb.RefreshSessionResponse{
		IdToken: identityToken,
	}, nil
}

func (s publicServer) Logout(ctx context.Context, req *authnpb.LogoutRequest) (*authnpb.LogoutResponse, error) {

	err := services.SessionEnder(s.app.RefreshTokenStore, meta.GetRefreshToken(ctx))
	if err != nil {
		info := meta.GetUnaryServerInfo(ctx)
		s.app.Reporter.ReportGRPCError(err, info, req)
	}

	meta.SetSession(ctx, s.app.Config.SessionCookieName, "")

	return &authnpb.LogoutResponse{}, nil
}

func (s publicServer) ChangePassword(ctx context.Context, req *authnpb.ChangePasswordRequest) (*authnpb.ChangePasswordResponse, error) {

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
		accountID, meta.MatchedDomain(ctx), meta.GetRefreshToken(ctx),
	)
	if err != nil {
		panic(err)
	}

	// Return the signed session in a cookie
	meta.SetSession(ctx, s.app.Config.SessionCookieName, sessionToken)

	// Return the signed identity token in the body
	return &authnpb.ChangePasswordResponse{
		IdToken: identityToken,
	}, nil
}

func (s publicServer) HealthCheck(context.Context, *authnpb.HealthCheckRequest) (*authnpb.HealthCheckResponse, error) {
	return &authnpb.HealthCheckResponse{
		Http:  true,
		Redis: s.app.RedisCheck(),
		Db:    s.app.DbCheck(),
	}, nil
}
