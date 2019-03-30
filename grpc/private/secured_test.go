package private

import (
	"net"
	"testing"

	"github.com/keratin/authn-server/app/services"

	"github.com/keratin/authn-server/grpc/internal/errors"

	"github.com/keratin/authn-server/app"
	authnpb "github.com/keratin/authn-server/grpc"
	grpctest "github.com/keratin/authn-server/grpc/test"
	"github.com/keratin/authn-server/server/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
)

func securedServerSetup(t *testing.T, app *app.App) (authnpb.SecuredAdminAuthNClient, func()) {

	srvCtx := context.Background()
	srv := grpc.NewServer(
		grpc.Creds(credentials.NewServerTLSFromCert(&grpctest.Cert)),
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logrus.StandardLogger())),
			grpc_prometheus.UnaryServerInterceptor,
			// the default authentication is none
			grpc_auth.UnaryServerInterceptor(func(ctx context.Context) (context.Context, error) {
				return ctx, nil
			}),
		),
	) // grpctest.NewPrivateServer(app)
	conn, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("error creating listener: %s", err)
	}

	authnpb.RegisterSecuredAdminAuthNServer(srv, securedServer{app: app, matcher: func(username, password string) bool { return true }})
	go func() {
		grpctest.Serve(srvCtx, srv, conn)
	}()

	// error handling omitted

	clientConn, err := grpc.Dial(
		conn.Addr().String(),
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(grpctest.CertPool, "")),
		grpc.WithPerRPCCredentials(basicAuth{
			username: app.Config.AuthUsername,
			password: app.Config.AuthPassword,
		}),
	)
	if err != nil {
		t.Fatalf("error dialing: %s", err)
	}
	client := authnpb.NewSecuredAdminAuthNClient(clientConn)

	return client, func() {
		clientConn.Close()
		srvCtx.Done() // will trigger srv.GracefulStop()
	}
}

func TestGetAccount(t *testing.T) {

	app := test.App()
	client, teardown := securedServerSetup(t, app)
	defer teardown()

	account, err := app.AccountStore.Create("unlocked@test.com", []byte("bar"))
	require.NoError(t, err)

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "unknown account",
			testFunc: func(t *testing.T) {
				req := &authnpb.GetAccountRequest{
					Id: 999999,
				}
				res, err := client.GetAccount(context.TODO(), req)
				require.Error(t, err)
				require.Nil(t, res)
				s, ok := status.FromError(err)
				if !ok {
					t.Errorf("securedServer.GetAccount() unrcognized error code: %s, err: %s", s.Code(), s.Err())
					return
				}
				assert.Equal(t, s.Code(), codes.NotFound)

				expectedErr := errors.FieldErrors{
					{
						Field:   "account",
						Message: services.ErrNotFound,
					},
				}

				for _, detail := range s.Details() {
					br, ok := detail.(*errdetails.BadRequest)
					if !ok {
						t.Errorf("securedServer.GetAccount() unrcognized error detail type: %T, expected: %T", detail, &errdetails.BadRequest{})
						return
					}
					fes := errors.ToFieldErrors(br)
					assert.Equal(t, expectedErr, fes)
				}
			},
		},
		{
			name: "valid account",
			testFunc: func(t *testing.T) {
				req := &authnpb.GetAccountRequest{
					Id: int64(account.ID),
				}
				res, err := client.GetAccount(context.TODO(), req)
				require.NoError(t, err)

				require.NotNil(t, res.GetResult())
				assert.Equal(t, account.Username, res.GetResult().GetUsername())
				assert.Equal(t, int64(account.ID), res.GetResult().GetId())
				assert.Equal(t, false, res.GetResult().GetLocked())
				assert.Equal(t, false, res.GetResult().GetDeleted())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func TestUpdateAccount(t *testing.T) {

	app := test.App()
	app.Config.UsernameIsEmail = false
	client, teardown := securedServerSetup(t, app)
	defer teardown()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "unknown account",
			testFunc: func(t *testing.T) {
				req := &authnpb.UpdateAccountRequest{
					Id:       999999,
					Username: "irrelevant",
				}
				res, err := client.UpdateAccount(context.TODO(), req)
				require.Error(t, err)
				require.Nil(t, res)
				s, ok := status.FromError(err)
				if !ok {
					t.Errorf("securedServer.UpdateAccount() unrcognized error code: %s, err: %s", s.Code(), s.Err())
					return
				}
				assert.Equal(t, s.Code(), codes.NotFound)

				expectedErr := errors.FieldErrors{
					errors.FieldError{
						Field:   "account",
						Message: services.ErrNotFound,
					},
				}

				for _, detail := range s.Details() {
					br, ok := detail.(*errdetails.BadRequest)
					if !ok {
						t.Errorf("securedServer.UpdateAccount() unrcognized error detail type: %T, expected: %T", detail, &errdetails.BadRequest{})
						return
					}
					fes := errors.ToFieldErrors(br)
					assert.Equal(t, expectedErr, fes)
				}
			},
		},
		{
			name: "existing account",
			testFunc: func(t *testing.T) {
				account, err := app.AccountStore.Create("one@test.com", []byte("bar"))
				require.NoError(t, err)

				req := &authnpb.UpdateAccountRequest{
					Id:       int64(account.ID),
					Username: "newname",
				}
				res, err := client.UpdateAccount(context.TODO(), req)
				require.NoError(t, err)
				assert.NotNil(t, res)

				account, err = app.AccountStore.Find(account.ID)
				require.NoError(t, err)
				assert.Equal(t, "newname", account.Username)
			},
		},
		{
			name: "bad username",
			testFunc: func(t *testing.T) {
				account, err := app.AccountStore.Create("two@test.com", []byte("bar"))
				require.NoError(t, err)

				req := &authnpb.UpdateAccountRequest{
					Id:       int64(account.ID),
					Username: "",
				}

				res, err := client.UpdateAccount(context.TODO(), req)
				require.Error(t, err)
				require.Nil(t, res)
				s, ok := status.FromError(err)
				if !ok {
					t.Errorf("securedServer.UpdateAccount() unrcognized error code: %s, err: %s", s.Code(), s.Err())
					return
				}
				assert.Equal(t, s.Code(), codes.FailedPrecondition)

				expectedErr := errors.FieldErrors{
					errors.FieldError{
						Field:   "username",
						Message: services.ErrMissing,
					},
				}

				for _, detail := range s.Details() {
					br, ok := detail.(*errdetails.BadRequest)
					if !ok {
						t.Errorf("securedServer.UpdateAccount() unrcognized error detail type: %T, expected: %T", detail, &errdetails.BadRequest{})
						return
					}
					fes := errors.ToFieldErrors(br)
					assert.Equal(t, expectedErr, fes)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func TestLockAccount(t *testing.T) {
	app := test.App()
	app.Config.UsernameIsEmail = false
	client, teardown := securedServerSetup(t, app)
	defer teardown()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "unknown account",
			testFunc: func(t *testing.T) {
				req := &authnpb.LockAccountRequest{
					Id: 999999,
				}
				res, err := client.LockAccount(context.TODO(), req)
				assert.Error(t, err)
				assert.Nil(t, res)

				s, ok := status.FromError(err)
				if !ok {
					t.Errorf("securedServer.LockAccount() unrcognized error code: %s, err: %s", s.Code(), s.Err())
					return
				}
				assert.Equal(t, s.Code(), codes.NotFound)

				expectedErr := errors.FieldErrors{
					errors.FieldError{
						Field:   "account",
						Message: services.ErrNotFound,
					},
				}

				for _, detail := range s.Details() {
					br, ok := detail.(*errdetails.BadRequest)
					if !ok {
						t.Errorf("securedServer.LockAccount() unrcognized error detail type: %T, expected: %T", detail, &errdetails.BadRequest{})
						return
					}
					fes := errors.ToFieldErrors(br)
					assert.Equal(t, expectedErr, fes)
				}
			},
		},
		{
			name: "unlocked account",
			testFunc: func(t *testing.T) {
				account, err := app.AccountStore.Create("unlocked@test.com", []byte("bar"))
				require.NoError(t, err)

				req := &authnpb.LockAccountRequest{
					Id: int64(account.ID),
				}

				res, err := client.LockAccount(context.TODO(), req)
				require.NoError(t, err)
				assert.NotNil(t, res)

				account, err = app.AccountStore.Find(account.ID)
				require.NoError(t, err)
				assert.True(t, account.Locked)
			},
		},
		{
			name: "locked account",
			testFunc: func(t *testing.T) {
				account, err := app.AccountStore.Create("locked@test.com", []byte("bar"))
				require.NoError(t, err)
				app.AccountStore.Lock(account.ID)

				req := &authnpb.LockAccountRequest{
					Id: int64(account.ID),
				}

				res, err := client.LockAccount(context.TODO(), req) //client.Patch(fmt.Sprintf("/accounts/%v/lock", account.ID), url.Values{})
				require.NoError(t, err)
				assert.NotNil(t, res)

				account, err = app.AccountStore.Find(account.ID)
				require.NoError(t, err)
				assert.True(t, account.Locked)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func TestUnlockAccount(t *testing.T) {
	app := test.App()
	app.Config.UsernameIsEmail = false
	client, teardown := securedServerSetup(t, app)
	defer teardown()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "unknown account",
			testFunc: func(t *testing.T) {
				req := &authnpb.UnlockAccountRequest{
					Id: 999999,
				}
				res, err := client.UnlockAccount(context.TODO(), req)
				assert.Error(t, err)
				assert.Nil(t, res)

				s, ok := status.FromError(err)
				if !ok {
					t.Errorf("securedServer.UnlockAccount() unrcognized error code: %s, err: %s", s.Code(), s.Err())
					return
				}
				assert.Equal(t, s.Code(), codes.NotFound)

				expectedErr := errors.FieldErrors{
					errors.FieldError{
						Field:   "account",
						Message: services.ErrNotFound,
					},
				}

				for _, detail := range s.Details() {
					br, ok := detail.(*errdetails.BadRequest)
					if !ok {
						t.Errorf("securedServer.UnlockAccount() unrcognized error detail type: %T, expected: %T", detail, &errdetails.BadRequest{})
						return
					}
					fes := errors.ToFieldErrors(br)
					assert.Equal(t, expectedErr, fes)
				}
			},
		},
		{
			name: "unlocked account",
			testFunc: func(t *testing.T) {
				account, err := app.AccountStore.Create("unlocked@test.com", []byte("bar"))
				require.NoError(t, err)

				req := &authnpb.UnlockAccountRequest{
					Id: int64(account.ID),
				}

				res, err := client.UnlockAccount(context.TODO(), req)
				require.NoError(t, err)
				assert.NotNil(t, res)

				account, err = app.AccountStore.Find(account.ID)
				require.NoError(t, err)
				assert.False(t, account.Locked)
			},
		},
		{
			name: "locked account",
			testFunc: func(t *testing.T) {
				account, err := app.AccountStore.Create("locked@test.com", []byte("bar"))
				require.NoError(t, err)
				app.AccountStore.Lock(account.ID)

				req := &authnpb.UnlockAccountRequest{
					Id: int64(account.ID),
				}

				res, err := client.UnlockAccount(context.TODO(), req) //client.Patch(fmt.Sprintf("/accounts/%v/lock", account.ID), url.Values{})
				require.NoError(t, err)
				assert.NotNil(t, res)

				account, err = app.AccountStore.Find(account.ID)
				require.NoError(t, err)
				assert.False(t, account.Locked)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func TestArchiveAccount(t *testing.T) {
	app := test.App()
	app.Config.UsernameIsEmail = false
	client, teardown := securedServerSetup(t, app)
	defer teardown()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "unknown account",
			testFunc: func(t *testing.T) {
				req := &authnpb.ArchiveAccountRequest{
					Id: 999999,
				}
				res, err := client.ArchiveAccount(context.TODO(), req)
				assert.Error(t, err)
				assert.Nil(t, res)

				s, ok := status.FromError(err)
				if !ok {
					t.Errorf("securedServer.ArchiveAccount() unrcognized error code: %s, err: %s", s.Code(), s.Err())
					return
				}
				assert.Equal(t, s.Code(), codes.FailedPrecondition)

				expectedErr := errors.FieldErrors{
					errors.FieldError{
						Field:   "account",
						Message: services.ErrNotFound,
					},
				}

				for _, detail := range s.Details() {
					br, ok := detail.(*errdetails.BadRequest)
					if !ok {
						t.Errorf("securedServer.ArchiveAccount() unrcognized error detail type: %T, expected: %T", detail, &errdetails.BadRequest{})
						return
					}
					fes := errors.ToFieldErrors(br)
					assert.Equal(t, expectedErr, fes)
				}
			},
		},
		{
			name: "unarchived account",
			testFunc: func(t *testing.T) {
				account, err := app.AccountStore.Create("unlocked@test.com", []byte("bar"))
				require.NoError(t, err)

				req := &authnpb.ArchiveAccountRequest{
					Id: int64(account.ID),
				}

				res, err := client.ArchiveAccount(context.TODO(), req) //client.Delete(fmt.Sprintf("/accounts/%v", account.ID))
				require.NoError(t, err)
				assert.NotNil(t, res)

				account, err = app.AccountStore.Find(account.ID)
				require.NoError(t, err)
				assert.NotEmpty(t, account.DeletedAt)
			},
		},
		{
			name: "archived account",
			testFunc: func(t *testing.T) {
				account, err := app.AccountStore.Create("locked@test.com", []byte("bar"))
				require.NoError(t, err)
				app.AccountStore.Archive(account.ID)

				req := &authnpb.ArchiveAccountRequest{
					Id: int64(account.ID),
				}

				res, err := client.ArchiveAccount(context.TODO(), req)
				require.NoError(t, err)
				assert.NotNil(t, res)

				account, err = app.AccountStore.Find(account.ID)
				require.NoError(t, err)
				assert.NotEmpty(t, account.DeletedAt)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func TestImportAccount(t *testing.T) {
	app := test.App()
	app.Config.UsernameIsEmail = false
	client, teardown := securedServerSetup(t, app)
	defer teardown()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "importing someone",
			testFunc: func(t *testing.T) {
				req := &authnpb.ImportAccountRequst{
					Username: "someone@app.com",
					Password: "secret",
				}
				res, err := client.ImportAccount(context.TODO(), req)
				require.NoError(t, err)

				account, err := app.AccountStore.FindByUsername("someone@app.com")
				require.NoError(t, err)
				assert.Equal(t, res.Result.Id, int64(account.ID))
			},
		},
		{
			name: "importing a locked user",
			testFunc: func(t *testing.T) {
				req := &authnpb.ImportAccountRequst{
					Username: "locked@app.com",
					Password: "secret",
					Locked:   true,
				}

				res, err := client.ImportAccount(context.TODO(), req)
				require.NoError(t, err)

				account, err := app.AccountStore.FindByUsername("locked@app.com")
				require.NoError(t, err)
				assert.Equal(t, res.Result.Id, int64(account.ID))
				assert.True(t, account.Locked)
			},
		},
		{
			name: "importing an unlocked user",
			testFunc: func(t *testing.T) {
				req := &authnpb.ImportAccountRequst{
					Username: "unlocked@app.com",
					Password: "secret",
					Locked:   false,
				}

				res, err := client.ImportAccount(context.TODO(), req)
				require.NoError(t, err)

				account, err := app.AccountStore.FindByUsername("unlocked@app.com")
				require.NoError(t, err)
				assert.Equal(t, res.Result.Id, int64(account.ID))
				assert.False(t, account.Locked)
			},
		},
		{
			name: "importing an invalid user",
			testFunc: func(t *testing.T) {
				req := &authnpb.ImportAccountRequst{
					Username: "unlocked@app.com",
					Password: "",
				}

				res, err := client.ImportAccount(context.TODO(), req)
				assert.Error(t, err)
				assert.Nil(t, res)

				s, ok := status.FromError(err)
				if !ok {
					t.Errorf("securedServer.ImportAccount() unrcognized error code: %s, err: %s", s.Code(), s.Err())
					return
				}
				assert.Equal(t, s.Code(), codes.FailedPrecondition)

				expectedErr := errors.FieldErrors{
					errors.FieldError{
						Field:   "password",
						Message: services.ErrMissing,
					},
				}

				for _, detail := range s.Details() {
					br, ok := detail.(*errdetails.BadRequest)
					if !ok {
						t.Errorf("securedServer.ImportAccount() unrcognized error detail type: %T, expected: %T", detail, &errdetails.BadRequest{})
						return
					}
					fes := errors.ToFieldErrors(br)
					assert.Equal(t, expectedErr, fes)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}

func TestExpirePassword(t *testing.T) {
	app := test.App()
	app.Config.UsernameIsEmail = false
	client, teardown := securedServerSetup(t, app)
	defer teardown()

	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "unknown account",
			testFunc: func(t *testing.T) {
				req := &authnpb.ExpirePasswordRequest{
					Id: 999999,
				}

				res, err := client.ExpirePassword(context.TODO(), req)
				require.Error(t, err)
				assert.Nil(t, res)

				s, ok := status.FromError(err)
				if !ok {
					t.Errorf("securedServer.ExpirePassword() unrcognized error code: %s, err: %s", s.Code(), s.Err())
					return
				}
				assert.Equal(t, s.Code(), codes.NotFound)

				expectedErr := errors.FieldErrors{
					errors.FieldError{
						Field:   "account",
						Message: services.ErrNotFound,
					},
				}

				for _, detail := range s.Details() {
					br, ok := detail.(*errdetails.BadRequest)
					if !ok {
						t.Errorf("securedServer.ExpirePassword() unrcognized error detail type: %T, expected: %T", detail, &errdetails.BadRequest{})
						return
					}
					fes := errors.ToFieldErrors(br)
					assert.Equal(t, expectedErr, fes)
				}
			},
		},
		{
			name: "active account",
			testFunc: func(t *testing.T) {
				account, err := app.AccountStore.Create("active@test.com", []byte("bar"))
				require.NoError(t, err)

				req := &authnpb.ExpirePasswordRequest{
					Id: int64(account.ID),
				}
				res, err := client.ExpirePassword(context.TODO(), req)
				require.NoError(t, err)
				assert.NotNil(t, res)

				account, err = app.AccountStore.Find(account.ID)
				require.NoError(t, err)
				assert.True(t, account.RequireNewPassword)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.testFunc(t)
		})
	}
}
