package public

import (
	"net"
	"testing"

	"github.com/keratin/authn-server/grpc/internal/errors"
	"github.com/keratin/authn-server/grpc/internal/meta"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/models"
	"github.com/keratin/authn-server/app/tokens/passwordless"
	"github.com/keratin/authn-server/app/tokens/sessions"
	authnpb "github.com/keratin/authn-server/grpc"
	grpctest "github.com/keratin/authn-server/grpc/test"
	"github.com/keratin/authn-server/server/test"
	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func passwordlessServerSetup(t *testing.T, app *app.App) (authnpb.PasswordlessServiceClient, func()) {

	app.Config.UsernameIsEmail = true

	srvCtx := context.Background()
	srv := grpctest.NewPublicServer(app)
	conn, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("error creating listener: %s", err)
	}
	authnpb.RegisterPasswordlessServiceServer(srv, passwordlessServer{app})
	go func() {
		grpctest.Serve(srvCtx, srv, conn)
	}()

	clientConn, err := grpc.Dial(conn.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("error dialing: %s", err)
	}
	client := authnpb.NewPasswordlessServiceClient(clientConn)

	return client, func() {
		clientConn.Close()
		srvCtx.Done() // will trigger srv.GracefulStop()
	}
}

func TestRequestPasswordlessLogin(t *testing.T) {

	app := test.App()
	client, teardown := passwordlessServerSetup(t, app)
	defer teardown()

	t.Run("known account", func(t *testing.T) {
		_, err := app.AccountStore.Create("known@keratin.tech", []byte("pwd"))
		require.NoError(t, err)

		req := &authnpb.RequestPasswordlessLoginRequest{
			Username: "known@keratin.tech",
		}
		_, err = client.RequestPasswordlessLogin(context.TODO(), req)
		require.NoError(t, err)

		// TODO: assert go routine?
	})

	t.Run("unknown account", func(t *testing.T) {
		req := &authnpb.RequestPasswordlessLoginRequest{
			Username: "unknown@keratin.tech",
		}
		_, err := client.RequestPasswordlessLogin(context.TODO(), req)
		require.NoError(t, err)
	})
}

func TestSubmitPasswordlessLogin(t *testing.T) {

	testApp := test.App()
	client, teardown := passwordlessServerSetup(t, testApp)
	defer teardown()

	assertSuccess := func(t *testing.T, res *authnpb.SubmitPasswordlessLoginResponseEnvelope, md metadata.MD, account *models.Account) {
		grpctest.AssertSession(t, testApp.Config, md)
		grpctest.AssertIDTokenResponse(t, res.Result.GetIdToken(), testApp.KeyStore, testApp.Config)
		found, err := testApp.AccountStore.Find(account.ID)
		require.NoError(t, err)
		assert.Equal(t, found.Password, account.Password)
	}
	factory := func(username string, password string) (*models.Account, error) {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), testApp.Config.BcryptCost)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "bcrypt")
		}
		return testApp.AccountStore.Create(username, hash)
	}

	t.Run("valid passwordless token", func(t *testing.T) {
		// given an account
		account, err := factory("valid.token@authn.tech", "oldpwd")
		require.NoError(t, err)

		// given a passwordless token
		token, err := passwordless.New(testApp.Config, account.ID)
		require.NoError(t, err)
		tokenStr, err := token.Sign(testApp.Config.PasswordlessTokenSigningKey)
		require.NoError(t, err)

		req := &authnpb.SubmitPasswordlessLoginRequest{
			Token: tokenStr,
		}

		var header metadata.MD
		// invoking the endpoint
		res, err := client.SubmitPasswordlessLogin(context.TODO(), req, grpc.Header(&header))
		require.NoError(t, err)

		// works
		assertSuccess(t, res, header, account)
	})

	t.Run("invalid passwordless token", func(t *testing.T) {

		req := &authnpb.SubmitPasswordlessLoginRequest{
			Token: "invalid",
		}

		expect := errors.FieldErrors{{"token", "INVALID_OR_EXPIRED"}}

		// invoking the endpoint
		_, err := client.SubmitPasswordlessLogin(context.TODO(), req)
		require.Error(t, err)

		s, ok := status.FromError(err)
		if !ok {
			t.Errorf("signupServiceServer.Signup() unrcognized error code: %s, err: %s", s.Code(), s.Err())
			return
		}
		assert.Equal(t, s.Code(), codes.FailedPrecondition)
		for _, detail := range s.Details() {
			br, ok := detail.(*errdetails.BadRequest)
			if !ok {
				t.Errorf("publicServer.Login() unrcognized error detail type: %T, expected: %T", detail, &errdetails.BadRequest{})
				return
			}
			fes := errors.ToFieldErrors(br)
			assert.Equal(t, expect, fes)
		}
	})

	t.Run("valid session", func(t *testing.T) {
		// given an account
		account, err := factory("valid.session@authn.tech", "oldpwd")
		require.NoError(t, err)

		// given a session
		session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, account.ID)

		// given a passwordless token
		token, err := passwordless.New(testApp.Config, account.ID)
		require.NoError(t, err)
		tokenStr, err := token.Sign(testApp.Config.PasswordlessTokenSigningKey)
		require.NoError(t, err)

		req := &authnpb.SubmitPasswordlessLoginRequest{
			Token: tokenStr,
		}

		md := metadata.Pairs(testApp.Config.SessionCookieName, session.Value)
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		meta.SetSession(ctx, testApp.Config.SessionCookieName, session.Value)

		var header metadata.MD
		// invoking the endpoint
		res, err := client.SubmitPasswordlessLogin(ctx, req, grpc.Header(&header))
		require.NoError(t, err)

		// works
		assertSuccess(t, res, header, account)

		// invalidates old session
		claims, err := sessions.Parse(session.Value, testApp.Config)
		require.NoError(t, err)
		id, err := testApp.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
		require.NoError(t, err)
		assert.Empty(t, id)
	})
}
