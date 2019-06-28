package public

import (
	"net"
	"net/url"
	"testing"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/models"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/app/tokens/resets"
	"github.com/keratin/authn-server/app/tokens/sessions"
	authnpb "github.com/keratin/authn-server/grpc"
	"github.com/keratin/authn-server/grpc/internal/errors"
	"github.com/keratin/authn-server/grpc/internal/meta"
	grpctest "github.com/keratin/authn-server/grpc/test"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
	pkgerrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func publicServerSetup(t *testing.T, app *app.App) (authnpb.PublicAuthNClient, func()) {

	app.Config.UsernameIsEmail = true

	srvCtx := context.Background()
	srv := grpctest.NewPublicServer(app)
	conn, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("error creating listener: %s", err)
	}
	authnpb.RegisterPublicAuthNServer(srv, publicServer{app})
	go func() {
		grpctest.Serve(srvCtx, srv, conn)
	}()

	clientConn, err := grpc.Dial(conn.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("error dialing: %s", err)
	}
	client := authnpb.NewPublicAuthNClient(clientConn)

	return client, func() {
		clientConn.Close()
		srvCtx.Done() // will trigger srv.GracefulStop()
	}
}

func TestLogin(t *testing.T) {

	app := test.App()
	client, teardown := publicServerSetup(t, app)
	defer teardown()

	t.Run("existing user can login", func(t *testing.T) {
		//setup
		b, err := bcrypt.GenerateFromPassword([]byte("bar"), 4)
		require.NoError(t, err)
		_, err = app.AccountStore.Create("foo1", b)
		require.NoError(t, err)

		req := &authnpb.LoginRequest{
			Username: "foo1",
			Password: "bar",
		}

		var header metadata.MD
		got, err := client.Login(context.TODO(), req, grpc.Header(&header))
		if err != nil {
			t.Errorf("publicServer.Login() error = %v", err)
			return
		}

		grpctest.AssertSession(t, app.Config, header)
		grpctest.AssertIDTokenResponse(t, got.Result.GetIdToken(), app.KeyStore, app.Config)
	})

	t.Run("TestPostSessionSuccessWithSession", func(t *testing.T) {

		b, _ := bcrypt.GenerateFromPassword([]byte("bar"), 4)
		app.AccountStore.Create("foo2", b)

		accountID := 8642
		session := test.CreateSession(app.RefreshTokenStore, app.Config, accountID)

		// before
		refreshTokens, err := app.RefreshTokenStore.FindAll(accountID)
		require.NoError(t, err)
		refreshToken := refreshTokens[0]

		req := &authnpb.LoginRequest{
			Username: "foo1",
			Password: "bar",
		}

		md := metadata.Pairs(app.Config.SessionCookieName, session.Value)
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		meta.SetSession(ctx, app.Config.SessionCookieName, session.Value)

		var header metadata.MD
		got, err := client.Login(ctx, req, grpc.Header(&header))
		if err != nil {
			t.Errorf("publicServer.Login() error = %v", err)
			return
		}

		grpctest.AssertSession(t, app.Config, header)
		grpctest.AssertIDTokenResponse(t, got.Result.GetIdToken(), app.KeyStore, app.Config)

		// after
		id, err := app.RefreshTokenStore.Find(refreshToken)
		require.NoError(t, err)
		assert.Empty(t, id)
	})

	t.Run("TestPostSessionFailure", func(t *testing.T) {
		req := &authnpb.LoginRequest{
			Username: "",
			Password: "",
		}
		expect := services.FieldErrors{
			services.FieldError{
				Field:   "credentials",
				Message: services.ErrFailed,
			},
		}

		got, err := client.Login(context.TODO(), req)
		if err == nil {
			t.Errorf("publicServer.Login() expected error; received none")
			return
		}
		assert.Nil(t, got)

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
}

func TestRefreshSession(t *testing.T) {

	testApp := test.App()
	client, teardown := publicServerSetup(t, testApp)
	defer teardown()

	t.Run("Session Refresh Success", func(t *testing.T) {
		accountID := 82594
		existingSession := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, accountID)

		req := &authnpb.RefreshSessionRequest{}
		md := metadata.Pairs(testApp.Config.SessionCookieName, existingSession.Value)
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		meta.SetSession(ctx, testApp.Config.SessionCookieName, existingSession.Value)

		res, err := client.RefreshSession(ctx, req)
		require.NoError(t, err)
		t.Log(res.Result.GetIdToken())
		grpctest.AssertIDTokenResponse(t, res.Result.GetIdToken(), testApp.KeyStore, testApp.Config)
	})

	t.Run("TestGetSessionRefreshFailure", func(t *testing.T) {

		testCases := []struct {
			signingKey []byte
			liveToken  bool
		}{
			// cookie with the wrong signature
			{[]byte("wrong"), true},
			// cookie with a revoked refresh token
			{testApp.Config.SessionSigningKey, false},
		}

		for idx, tc := range testCases {
			tcCfg := &app.Config{
				AuthNURL:           testApp.Config.AuthNURL,
				SessionCookieName:  testApp.Config.SessionCookieName,
				SessionSigningKey:  tc.signingKey,
				ApplicationDomains: []route.Domain{{Hostname: "test.com"}},
			}
			existingSession := test.CreateSession(testApp.RefreshTokenStore, tcCfg, idx+100)
			if !tc.liveToken {
				test.RevokeSession(testApp.RefreshTokenStore, testApp.Config, existingSession)
			}

			res, err := client.RefreshSession(context.TODO(), &authnpb.RefreshSessionRequest{})
			assert.Nil(t, res)
			assert.NotNil(t, err)

			if st, ok := status.FromError(err); ok {
				assert.Equal(t, codes.Unauthenticated, st.Code())
			} else {
				t.Errorf("error doesn't conform to gRPC status codes")
			}
		}
	})
}

func TestLogout(t *testing.T) {

	testApp := test.App()
	client, teardown := publicServerSetup(t, testApp)
	defer teardown()

	t.Run("TestDeleteSessionSuccess", func(t *testing.T) {
		accountID := 514628
		session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, accountID)

		// token exists
		claims, err := sessions.Parse(session.Value, testApp.Config)
		require.NoError(t, err)
		id, err := testApp.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
		require.NoError(t, err)
		assert.NotEmpty(t, id)

		md := metadata.Pairs(testApp.Config.SessionCookieName, session.Value)
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		meta.SetSession(ctx, testApp.Config.SessionCookieName, session.Value)

		_, err = client.Logout(ctx, &authnpb.LogoutRequest{})
		require.NoError(t, err)

		// token no longer exists
		id, err = testApp.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
		require.NoError(t, err)
		assert.Empty(t, id)
	})

	t.Run("TestDeleteSessionFailure", func(t *testing.T) {
		badCfg := &app.Config{
			AuthNURL:           testApp.Config.AuthNURL,
			SessionCookieName:  testApp.Config.SessionCookieName,
			SessionSigningKey:  []byte("wrong"),
			ApplicationDomains: testApp.Config.ApplicationDomains,
		}
		session := test.CreateSession(testApp.RefreshTokenStore, badCfg, 123)

		md := metadata.Pairs(testApp.Config.SessionCookieName, session.Value)
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		meta.SetSession(ctx, testApp.Config.SessionCookieName, session.Value)

		_, err := client.Logout(ctx, &authnpb.LogoutRequest{})
		require.NoError(t, err)
	})

	t.Run("TestDeleteSessionWithoutSession", func(t *testing.T) {
		_, err := client.Logout(context.TODO(), &authnpb.LogoutRequest{})
		require.NoError(t, err)
	})
}

func TestChangePassword(t *testing.T) {

	testApp := test.App()
	client, teardown := publicServerSetup(t, testApp)
	defer teardown()

	assertSuccess := func(t *testing.T, res *authnpb.ChangePasswordResponseEnvelope, md metadata.MD, account *models.Account) {
		grpctest.AssertSession(t, testApp.Config, md)
		grpctest.AssertIDTokenResponse(t, res.Result.GetIdToken(), testApp.KeyStore, testApp.Config)
		found, err := testApp.AccountStore.Find(account.ID)
		require.NoError(t, err)
		assert.NotEqual(t, found.Password, account.Password)
	}
	factory := func(username string, password string) (*models.Account, error) {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), testApp.Config.BcryptCost)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "bcrypt")
		}
		return testApp.AccountStore.Create(username, hash)
	}

	t.Run("valid reset token", func(t *testing.T) {
		// given an account
		account, err := factory("valid.token@authn.tech", "oldpwd")
		require.NoError(t, err)

		// given a reset token
		token, err := resets.New(testApp.Config, account.ID, account.PasswordChangedAt)
		require.NoError(t, err)
		tokenStr, err := token.Sign(testApp.Config.ResetSigningKey)
		require.NoError(t, err)

		req := &authnpb.ChangePasswordRequest{
			Token:    tokenStr,
			Password: "0a0b0c0d0",
		}
		var header metadata.MD
		res, err := client.ChangePassword(context.TODO(), req, grpc.Header(&header))
		require.NoError(t, err)

		// works
		assertSuccess(t, res, header, account)
	})

	t.Run("invalid reset token", func(t *testing.T) {

		req := &authnpb.ChangePasswordRequest{
			Token:    "invalid",
			Password: "0a0b0c0d0",
		}
		_, err := client.ChangePassword(context.TODO(), req)

		require.NotNil(t, err)

		if errStatus, ok := status.FromError(err); ok {
			assert.Equal(t, codes.FailedPrecondition, errStatus.Code())

			for _, detail := range errStatus.Details() {
				br, ok := detail.(*errdetails.BadRequest)
				if !ok {
					t.Errorf("publicServer.ChangePassword() unrcognized error detail type: %T, expected: %T", detail, &errdetails.BadRequest{})
					return
				}
				fes := errors.ToFieldErrors(br)
				assert.Equal(t, services.FieldErrors{{Field: "token", Message: services.ErrInvalidOrExpired}}, fes)
			}
		} else {
			t.Error("unexpected error type")
		}
	})

	t.Run("valid session", func(t *testing.T) {
		// given an account
		account, err := factory("valid.session@authn.tech", "oldpwd")
		require.NoError(t, err)

		// given a session
		session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, account.ID)

		req := &authnpb.ChangePasswordRequest{
			CurrentPassword: "oldpwd",
			Password:        "0a0b0c0d0",
		}

		md := metadata.Pairs(testApp.Config.SessionCookieName, session.Value)
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		meta.SetSession(ctx, testApp.Config.SessionCookieName, session.Value)

		var header metadata.MD
		// invoking the endpoint
		res, err := client.ChangePassword(ctx, req, grpc.Header(&header))

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

	t.Run("valid session and bad password", func(t *testing.T) {
		// given an account
		account, err := factory("bad.password@authn.tech", "oldpwd")
		require.NoError(t, err)

		// given a session
		session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, account.ID)

		req := &authnpb.ChangePasswordRequest{
			CurrentPassword: "oldpwd",
			Password:        "a",
		}

		md := metadata.Pairs(testApp.Config.SessionCookieName, session.Value)
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		meta.SetSession(ctx, testApp.Config.SessionCookieName, session.Value)

		_, err = client.ChangePassword(ctx, req)

		require.NotNil(t, err)

		if errStatus, ok := status.FromError(err); ok {
			assert.Equal(t, codes.FailedPrecondition, errStatus.Code())

			for _, detail := range errStatus.Details() {
				br, ok := detail.(*errdetails.BadRequest)
				if !ok {
					t.Errorf("publicServer.ChangePassword() unrcognized error detail type: %T, expected: %T", detail, &errdetails.BadRequest{})
					return
				}
				fes := errors.ToFieldErrors(br)
				assert.Equal(t, services.FieldErrors{{Field: "password", Message: services.ErrInsecure}}, fes)
			}
		} else {
			t.Error("unexpected error type")
		}
	})

	t.Run("valid session and bad currentPassword", func(t *testing.T) {
		// given an account
		account, err := factory("bad.currentPassword@authn.tech", "oldpwd")
		require.NoError(t, err)

		// given a session
		session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, account.ID)

		md := metadata.Pairs(testApp.Config.SessionCookieName, session.Value)
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		meta.SetSession(ctx, testApp.Config.SessionCookieName, session.Value)

		req := &authnpb.ChangePasswordRequest{
			CurrentPassword: "wrong",
			Password:        "0a0b0c0d0",
		}

		_, err = client.ChangePassword(ctx, req)

		require.NotNil(t, err)

		if errStatus, ok := status.FromError(err); ok {
			assert.Equal(t, codes.FailedPrecondition, errStatus.Code())

			for _, detail := range errStatus.Details() {
				br, ok := detail.(*errdetails.BadRequest)
				if !ok {
					t.Errorf("publicServer.ChangePassword() unrcognized error detail type: %T, expected: %T", detail, &errdetails.BadRequest{})
					return
				}
				fes := errors.ToFieldErrors(br)
				assert.Equal(t, services.FieldErrors{{Field: "credentials", Message: services.ErrFailed}}, fes)
			}
		} else {
			t.Error("unexpected error type")
		}
	})

	t.Run("invalid session", func(t *testing.T) {
		md := metadata.Pairs(testApp.Config.SessionCookieName, "invalid")
		ctx := metadata.NewOutgoingContext(context.Background(), md)

		req := &authnpb.ChangePasswordRequest{
			CurrentPassword: "oldpwd",
			Password:        "0a0b0c0d0",
		}

		_, err := client.ChangePassword(ctx, req)
		require.Error(t, err)

		if errorStatus, ok := status.FromError(err); ok {
			assert.Equal(t, codes.Unauthenticated, errorStatus.Code())
		} else {
			t.Error("unexpected error type")
		}
	})

	t.Run("token AND session", func(t *testing.T) {
		// given an account
		tokenAccount, err := factory("token@authn.tech", "oldpwd")
		require.NoError(t, err)
		// with a reset token
		token, err := resets.New(testApp.Config, tokenAccount.ID, tokenAccount.PasswordChangedAt)
		require.NoError(t, err)
		tokenStr, err := token.Sign(testApp.Config.ResetSigningKey)
		require.NoError(t, err)

		// given another account
		sessionAccount, err := factory("session@authn.tech", "oldpwd")
		require.NoError(t, err)

		// with a session
		session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, sessionAccount.ID)

		req := &authnpb.ChangePasswordRequest{
			Token:           tokenStr,
			CurrentPassword: "oldpwd",
			Password:        "0a0b0c0d0",
		}

		md := metadata.Pairs(testApp.Config.SessionCookieName, session.Value)
		ctx := metadata.NewOutgoingContext(context.Background(), md)
		meta.SetSession(ctx, testApp.Config.SessionCookieName, session.Value)

		var header metadata.MD
		res, err := client.ChangePassword(ctx, req, grpc.Header(&header))
		require.NoError(t, err)

		// works
		assertSuccess(t, res, header, tokenAccount)
	})
}

func TestHealthCheck(t *testing.T) {

	testApp := test.App()
	testApp.DbCheck = func() bool { return true }
	testApp.RedisCheck = func() bool { return true }
	client, teardown := publicServerSetup(t, testApp)
	defer teardown()

	res, err := client.HealthCheck(context.TODO(), &authnpb.HealthCheckRequest{})
	require.NoError(t, err)

	assert.Equal(t, `&HealthCheckResponse{Http:true,Db:true,Redis:true,}`, res.String())
}

func TestRegisterPublicGRPCMethods(t *testing.T) {

	collectServices := func(sInfo map[string]grpc.ServiceInfo) []string {
		s := []string{}
		for k := range sInfo {
			s = append(s, k)
		}
		return s
	}

	serviceMethods := map[string][]grpc.MethodInfo{
		"keratin.authn.SignupService": []grpc.MethodInfo{
			{
				Name: "Signup",
			},
			{
				Name: "IsUsernameAvailable",
			},
		},
		"keratin.authn.PasswordResetService": []grpc.MethodInfo{
			{
				Name: "RequestPasswordReset",
			},
		},
		"keratin.authn.PasswordlessService": []grpc.MethodInfo{
			{
				Name: "RequestPasswordlessLogin",
			},
			{
				Name: "SubmitPasswordlessLogin",
			},
		},
		"keratin.authn.PublicAuthN": []grpc.MethodInfo{
			{
				Name: "Login",
			},
			{
				Name: "Logout",
			},
			{
				Name: "RefreshSession",
			},
			{
				Name: "ChangePassword",
			},
			{
				Name: "HealthCheck",
			},
		},
	}
	_ = serviceMethods
	type args struct {
		server *grpc.Server
		app    func() *app.App
	}

	tcs := []struct {
		name             string
		args             args
		validate         func(t *testing.T)
		expectedServices []string
	}{
		{
			name: "app with all services configured results in gRPC server containing all services",
			args: args{
				server: grpc.NewServer(),
				app: func() *app.App {
					app := test.App()
					app.Config.EnableSignup = true

					passwordResetURL, _ := url.Parse("http://example.com/password/reset")
					app.Config.AppPasswordResetURL = passwordResetURL

					passowrdlessTokenURL, _ := url.Parse("http://example.com/passwordless")
					app.Config.AppPasswordlessTokenURL = passowrdlessTokenURL

					return app
				},
			},
			expectedServices: []string{"keratin.authn.PublicAuthN", "keratin.authn.SignupService", "keratin.authn.PasswordResetService", "keratin.authn.PasswordlessService"},
		},
		{
			name: "server has only public API when signup is not enabled, password reset URL is not set, and passwordless URL is not set",
			args: args{
				server: grpc.NewServer(),
				app: func() *app.App {
					app := test.App()
					app.Config.EnableSignup = false
					app.Config.AppPasswordResetURL = nil
					app.Config.AppPasswordlessTokenURL = nil

					return app
				},
			},
			expectedServices: []string{"keratin.authn.PublicAuthN"},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			RegisterPublicGRPCMethods(tc.args.server, tc.args.app())
			serverServices := tc.args.server.GetServiceInfo()
			servicesNames := collectServices(serverServices)

			// registered services match expected services
			assert.ElementsMatch(t, tc.expectedServices, servicesNames)

			// methods of expected services match actual methods of services
			for _, v := range tc.expectedServices {
				assert.ElementsMatch(t, serviceMethods[v], serverServices[v].Methods)
			}
		})
	}
}
