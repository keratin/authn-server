package public

import (
	"net"
	"testing"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/grpc/internal/errors"

	"github.com/stretchr/testify/assert"

	"google.golang.org/grpc/metadata"

	authnpb "github.com/keratin/authn-server/grpc"
	grpctest "github.com/keratin/authn-server/grpc/test"
	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/require"
	context "golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func signupSetup(t *testing.T, app *app.App) (authnpb.SignupServiceClient, func()) {

	app.Config.UsernameIsEmail = true

	srvCtx := context.Background()
	srv := grpctest.NewPublicServer(app)
	conn, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("error creating listener: %s", err)
	}
	authnpb.RegisterSignupServiceServer(srv, signupServiceServer{app})
	go func() {
		grpctest.Serve(srvCtx, srv, conn)
	}()

	clientConn, err := grpc.Dial(conn.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("error dialing: %s", err)
	}
	client := authnpb.NewSignupServiceClient(clientConn)

	return client, func() {
		clientConn.Close()
		srvCtx.Done() // will trigger srv.GracefulStop()
	}
}

func TestSignup(t *testing.T) {

	app := test.App()
	client, teardown := signupSetup(t, app)
	defer teardown()

	type args struct {
		ctx context.Context
		req *authnpb.SignupRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		err     services.FieldErrors
	}{
		{
			name: "Signup with correct input is successfull",
			args: args{
				ctx: context.Background(),
				req: &authnpb.SignupRequest{
					Username: "foo@example.com",
					Password: "0a0b0c0",
				},
			},
			wantErr: false,
		},
		{
			name: "Signup with taken username returns error",
			args: args{
				ctx: context.Background(),
				req: &authnpb.SignupRequest{
					Username: "foo@example.com",
					Password: "0a0b0c0",
				},
			},
			wantErr: true,
			err: services.FieldErrors{
				services.FieldError{
					Field:   "username",
					Message: services.ErrTaken,
				},
			},
		},
		{
			name: "Singup with multiple invalid parameters returns multiple errors",
			args: args{
				ctx: context.Background(),
				req: &authnpb.SignupRequest{
					Username: "foo",
					Password: "a",
				},
			},
			wantErr: true,
			err: services.FieldErrors{
				services.FieldError{
					Field:   "username",
					Message: services.ErrFormatInvalid,
				},
				services.FieldError{
					Field:   "password",
					Message: services.ErrInsecure,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var header metadata.MD
			got, err := client.Signup(tt.args.ctx, tt.args.req, grpc.Header(&header))
			if (err != nil) != tt.wantErr {
				t.Errorf("signupServiceServer.Signup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				s, ok := status.FromError(err)
				if !ok {
					t.Errorf("signupServiceServer.Signup() unrcognized error code: %s, err: %s", s.Code(), s.Err())
					return
				}
				assert.Equal(t, s.Code(), codes.FailedPrecondition)
				for _, detail := range s.Details() {
					br, ok := detail.(*errdetails.BadRequest)
					if !ok {
						t.Errorf("signupServiceServer.Signup() unrcognized error detail type: %T, expected: %T", detail, &errdetails.BadRequest{})
						return
					}
					fes := errors.ToFieldErrors(br)
					assert.Equal(t, tt.err, fes)
				}
				return
			}
			grpctest.AssertSession(t, app.Config, header)
			grpctest.AssertIDTokenResponse(t, got.Result.GetIdToken(), app.KeyStore, app.Config)
		})
	}

}

func TestIsUsernameAvailable(t *testing.T) {

	app := test.App()
	account, err := app.AccountStore.Create("existing@test.com", []byte("bar"))
	require.NoError(t, err)

	client, teardown := signupSetup(t, app)
	defer teardown()

	type args struct {
		ctx context.Context
		req *authnpb.IsUsernameAvailableRequest
	}
	tests := []struct {
		name    string
		s       signupServiceServer
		args    args
		want    *authnpb.IsUsernameAvailableResponseEnvelope
		wantErr bool
		err     services.FieldErrors
	}{
		{
			name: "known username",
			s: signupServiceServer{
				app: app,
			},
			args: args{
				ctx: context.Background(),
				req: &authnpb.IsUsernameAvailableRequest{
					Username: account.Username,
				},
			},
			want:    nil,
			wantErr: true,
			err: services.FieldErrors{
				services.FieldError{
					Field:   "username",
					Message: services.ErrTaken,
				},
			},
		},
		{
			name: "unknown username",
			s: signupServiceServer{
				app: app,
			},
			args: args{
				ctx: context.Background(),
				req: &authnpb.IsUsernameAvailableRequest{
					Username: "unknown@test.com",
				},
			},
			want: &authnpb.IsUsernameAvailableResponseEnvelope{
				Result: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.IsUsernameAvailable(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("signupServiceServer.IsUsernameAvailable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) && tt.wantErr {
				errCode := status.Code(err)
				assert.Equal(t, errCode, codes.FailedPrecondition)
			}
			assert.Equal(t, got, tt.want)
		})
	}
}
