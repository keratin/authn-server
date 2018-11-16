package grpc

import (
	"net"

	"github.com/sirupsen/logrus"

	"github.com/gogo/protobuf/types"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"google.golang.org/grpc/credentials"
)

// Compile-time check
var _ AuthNServer = Server{}

type Server struct {
	lis   net.Listener
	creds credentials.TransportCredentials
}

// NewServer returns a new gRPC server
func NewServer(lis net.Listener, creds credentials.TransportCredentials) Server {
	return Server{
		lis:   lis,
		creds: creds,
	}
}

func (s Server) RunGRPC(ctx context.Context) {

	srv := grpc.NewServer(
		grpc.Creds(s.creds),
		grpc.UnaryInterceptor(logInterceptor),
	)

	RegisterAuthNServer(srv, s)

	logrus.Infof("gRPC Listening on %s", s.lis.Addr().String())

	go func() {
		if err := srv.Serve(s.lis); err != nil {
			logrus.Printf("serve error: %s", err)
			return
		}
	}()
	<-ctx.Done()
	srv.Stop()
}

func logInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logrus.Infof("calling method: %s", info.FullMethod)

	res, err := handler(ctx, req)
	if err != nil {
		logrus.Errorf("error from method: %s", err)
	}
	return res, err
}

func (s Server) Signup(ctx context.Context, req *SignupRequest) (*SignupResponseEnvelope, error) {
	return nil, nil
}

func (s Server) GetAccount(ctx context.Context, req *GetAccountRequest) (*GetAccountResponseEnvelope, error) {
	return nil, nil
}

func (s Server) UpdateAccount(ctx context.Context, req *UpdateAccountRequest) (*types.Empty, error) {
	return nil, nil
}

func (s Server) IsUsernameAvailable(ctx context.Context, req *IsUsernameAvailableRequest) (*IsUsernameAvailableResponseEnvelope, error) {
	return nil, nil
}

func (s Server) LockAccount(ctx context.Context, req *LockAccountRequest) (*types.Empty, error) {
	return nil, nil
}

func (s Server) UnlockAcount(ctx context.Context, req *UnlockAccountRequest) (*types.Empty, error) {
	return nil, nil
}

func (s Server) ArchiveAccount(ctx context.Context, req *ArchiveAccountRequest) (*types.Empty, error) {
	return nil, nil
}

func (s Server) ImportAccount(ctx context.Context, req *ImportAccountRequst) (*ImportAccountResponseEnvelope, error) {
	return nil, nil
}

// Session Management
func (s Server) Login(ctx context.Context, req *LoginRequest) (*SignupResponse, error) {
	return nil, nil
}

func (s Server) RefreshSession(ctx context.Context, req *types.Empty) (*SignupResponse, error) {
	return nil, nil
}

func (s Server) Logout(ctx context.Context, req *types.Empty) (*types.Empty, error) {
	return nil, nil
}

// Password Management
func (s Server) RequestPasswordReset(ctx context.Context, req *PasswordResetRequest) (*types.Empty, error) {
	return nil, nil
}

func (s Server) ChangePassword(ctx context.Context, req *ChangePasswordRequest) (*SignupResponse, error) {
	return nil, nil
}

func (s Server) ExpirePassword(ctx context.Context, req *ExpirePasswordRequest) (*types.Empty, error) {
	return nil, nil
}

// OAuth
func (s Server) BeginOAuth(ctx context.Context, req *BeginOAuthRequest) (*BeginOAuthResponse, error) {
	return nil, nil
}

func (s Server) OAuthReturn(ctx context.Context, req *OAuthReturnRequest) (*OAuthReturnResponse, error) {
	return nil, nil
}

// Config
func (s Server) ServiceConfiguration(ctx context.Context, req *types.Empty) (*Configuration, error) {
	return nil, nil
}

func (s Server) JWKS(ctx context.Context, req *types.Empty) (*JWKSResponse, error) {
	return nil, nil
}

func (s Server) ServiceStats(ctx context.Context, req *types.Empty) (*ServiceStatsResponse, error) {
	return nil, nil
}

func (s Server) ServerStats(ctx context.Context, req *types.Empty) (*types.Empty, error) {
	return nil, nil
}

func (s Server) HealthCheck(ctx context.Context, req *types.Empty) (*HealthCheckResponse, error) {
	return nil, nil
}
