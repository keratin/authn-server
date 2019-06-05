package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/app/tokens/identities"
	"github.com/keratin/authn-server/app/tokens/sessions"
	authngrpc "github.com/keratin/authn-server/grpc"
	oauthlib "github.com/keratin/authn-server/lib/oauth"
	"github.com/keratin/authn-server/server/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"gopkg.in/square/go-jose.v2/jwt"
)

type basicAuth struct {
	username        string
	password        string
	secureTransport bool
}

func (b basicAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	auth := b.username + ":" + b.password
	enc := base64.StdEncoding.EncodeToString([]byte(auth))
	return map[string]string{
		"authorization": "basic " + enc,
	}, nil
}

func (b basicAuth) RequireTransportSecurity() bool {
	return b.secureTransport
}

type tokenResponse struct {
	Result struct {
		IDToken string `json:"id_token"`
	} `json:"result"`
}

type errorResponse struct {
	Errors services.FieldErrors
}

type account struct {
	ID       int64
	Username string
	Locked   bool
	Deleted  bool
}

func setup(t *testing.T) *app.App {
	configMap := map[string]string{
		"SECRET_KEY_BASE":            "TestKey",
		"DATABASE_URL":               os.Getenv("TEST_POSTGRES_URL"),
		"REDIS_URL":                  os.Getenv("TEST_REDIS_URL"),
		"AUTHN_URL":                  "http://authn.example.com",
		"USERNAME_IS_EMAIL":          "true",
		"APP_DOMAINS":                "test.com",
		"ENABLE_SIGNUP":              "true",
		"APP_PASSWORD_RESET_URL":     "http://app.example.com",
		"APP_PASSWORDLESS_TOKEN_URL": "http://app.example.com",
		"PUBLIC_PORT":                "8080",
		"PORT":                       "9090",
		"HTTP_AUTH_USERNAME":         "username",
		"HTTP_AUTH_PASSWORD":         "password",
	}

	for k, v := range configMap {
		err := os.Setenv(k, v)
		require.NoError(t, err)
	}

	cfg, err := app.ReadEnv()
	require.NoError(t, err)

	err = data.MigrateDB(cfg.DatabaseURL)
	require.NoError(t, err)

	app, err := app.NewApp(cfg)
	require.NoError(t, err)

	return app
}

// TestServer is end-to-end test of gRPC and REST services and cross-interface verification.
// The goal here is to ensure both interfaces function simultaneously and consistently.
func TestServer(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// start a fake oauth provider
	providerServer := httptest.NewServer(test.ProviderApp())
	defer providerServer.Close()
	// configure a client for the fake oauth provider
	providerClient := oauthlib.NewTestProvider(providerServer)

	app := setup(t)
	app.OauthProviders["test"] = *providerClient

	ctx, cancel := context.WithCancel(context.Background())
	go Server(ctx, app)
	defer cancel()

	// to use with REST API
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	httpClient := &http.Client{
		Jar: jar,
	}
	setHeaders := func(req *http.Request) {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Origin", "http://test.com")
	}
	privateURLBase := fmt.Sprintf("http://localhost:%d", app.Config.ServerPort)
	publicURLBase := fmt.Sprintf("http://localhost:%d", app.Config.PublicPort)

	publicSrvcConn, err := grpc.DialContext(ctx, fmt.Sprintf("localhost:%d", app.Config.PublicPort), grpc.WithInsecure())
	require.NoError(t, err)
	defer publicSrvcConn.Close()

	signupClient := authngrpc.NewSignupServiceClient(publicSrvcConn)
	publicClient := authngrpc.NewPublicAuthNClient(publicSrvcConn)

	privateSrvcConn, err := grpc.DialContext(ctx, fmt.Sprintf("localhost:%d", app.Config.ServerPort), grpc.WithInsecure(), grpc.WithPerRPCCredentials(basicAuth{
		username: app.Config.AuthUsername,
		password: app.Config.AuthPassword,
	}))
	require.NoError(t, err)
	securedClient := authngrpc.NewSecuredAdminAuthNClient(privateSrvcConn)
	unsecuredClient := authngrpc.NewUnsecuredAdminAuthNClient(privateSrvcConn)
	activesClient := authngrpc.NewAuthNActivesClient(privateSrvcConn)

	// the ID of the user against who tests will be carried
	var testSubjectID int

	t.Run("Queries' results across interfaces match", func(t *testing.T) {
		nameAvailable, err := signupClient.IsUsernameAvailable(context.Background(), &authngrpc.IsUsernameAvailableRequest{
			Username: "test@example.com",
		})
		assert.NoError(t, err)
		assert.NotEmpty(t, nameAvailable)
		assert.True(t, nameAvailable.Result)

		data := url.Values{
			"username": {"test@example.com"},
		}

		req, err := http.NewRequest("GET", publicURLBase+"/accounts/available", strings.NewReader(data.Encode()))
		assert.NoError(t, err)
		setHeaders(req)

		res, err := httpClient.Do(req)
		assert.NoError(t, err)
		defer res.Body.Close()
		var response struct {
			Result bool
		}
		body := test.ReadBody(res)
		err = json.Unmarshal(body, &response)
		assert.NoError(t, err)
		assert.EqualValues(t, response.Result, nameAvailable.Result)
	})

	t.Run("Signup via gRPC is successful and discoverable via REST", func(t *testing.T) {
		t.Run("gRPC signup with available username is successful and returns refresh token in metadata", func(t *testing.T) {
			var header metadata.MD
			signupResponse, err := signupClient.Signup(ctx, &authngrpc.SignupRequest{
				Username: "test@example.com",
				Password: "11aa22!bb33cc",
			}, grpc.Header(&header))

			// response received without error
			assert.NoError(t, err)
			assert.NotEmpty(t, signupResponse)

			// Valid refresh token included in header tagged by the cookie name
			assert.Len(t, header.Get(app.Config.SessionCookieName), 1)
			refresh := header.Get(app.Config.SessionCookieName)[0]
			_, err = sessions.Parse(refresh, app.Config)
			assert.NoError(t, err)

			// Valid access/identity token is received
			id, err := jwt.ParseSigned(signupResponse.Result.IdToken)
			assert.NoError(t, err)
			claims := identities.Claims{}
			err = id.Claims(app.KeyStore.Key().Public(), &claims)
			if assert.NoError(t, err) {
				// check that the JWT contains nice things
				assert.Equal(t, app.Config.AuthNURL.String(), claims.Issuer)
			}
			testSubjectID, err = strconv.Atoi(claims.Subject)

			tokens, err := app.RefreshTokenStore.FindAll(testSubjectID)
			assert.NoError(t, err)
			assert.Len(t, tokens, 1)
		})

		t.Run("Get Account of a user registered via gRPC is discoverable via REST", func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf(privateURLBase+"/accounts/%d", testSubjectID), nil)
			assert.NoError(t, err)
			req.SetBasicAuth(app.Config.AuthUsername, app.Config.AuthPassword)

			res, err := httpClient.Do(req)
			assert.NoError(t, err)
			defer res.Body.Close()

			var response struct {
				Result account
			}

			body := test.ReadBody(res)
			err = json.Unmarshal(body, &response)
			assert.NoError(t, err)
			assert.Equal(t, account{
				ID:       int64(testSubjectID),
				Username: "test@example.com",
				Locked:   false,
				Deleted:  false,
			}, response.Result)
		})
	})

	t.Run("User login, refresh, and `logout` via REST are successful", func(t *testing.T) {
		t.Run("User with already active session via gRPC get new separate session via REST", func(t *testing.T) {
			data := url.Values{
				"username": {"test@example.com"},
				"password": {"11aa22!bb33cc"},
			}
			req, err := http.NewRequest("POST", publicURLBase+"/session", strings.NewReader(data.Encode()))
			assert.NoError(t, err)
			setHeaders(req)

			res, err := httpClient.Do(req)
			assert.NoError(t, err)

			body := test.ReadBody(res)

			var response tokenResponse

			err = json.Unmarshal(body, &response)
			assert.NoError(t, err)
			assert.NotEmpty(t, response.Result)
			token, err := jwt.ParseSigned(response.Result.IDToken)
			assert.NoError(t, err)
			claims := identities.Claims{}
			err = token.Claims(app.KeyStore.Key().Public(), &claims)
			assert.NoError(t, err)

			// There should be 2 active sessions in the store: 1 via gRPC, 1 via REST
			rSessions, err := app.RefreshTokenStore.FindAll(testSubjectID)
			assert.NoError(t, err)
			assert.Len(t, rSessions, 2)

			u, err := url.Parse(publicURLBase)
			assert.NoError(t, err)
			cookies := httpClient.Jar.Cookies(u)
			assert.NotEmpty(t, cookies)
		})
		t.Run("User with already active session can refresh against public URL", func(t *testing.T) {
			// Without the wait, the newly generated token will be exactly the same.
			time.Sleep(time.Second * 1)
			req, err := http.NewRequest("GET", publicURLBase+"/session/refresh", nil)
			assert.NoError(t, err)
			setHeaders(req)

			res, err := httpClient.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusCreated, res.StatusCode)

			body := test.ReadBody(res)

			var response tokenResponse

			err = json.Unmarshal(body, &response)
			assert.NoError(t, err)
			assert.NotEmpty(t, response.Result)
			token, err := jwt.ParseSigned(response.Result.IDToken)
			assert.NoError(t, err)
			claims := identities.Claims{}
			err = token.Claims(app.KeyStore.Key().Public(), &claims)
			assert.NoError(t, err)

			rSessions, err := app.RefreshTokenStore.FindAll(testSubjectID)
			assert.NoError(t, err)
			assert.Len(t, rSessions, 2)

			u, err := url.Parse(publicURLBase)
			assert.NoError(t, err)
			cookies := httpClient.Jar.Cookies(u)
			assert.NotEmpty(t, cookies)
		})

		// This is a special case due to the nature of the canoncal host being "localhost"
		t.Run("User with already active session can refresh against private URL", func(t *testing.T) {
			time.Sleep(time.Second * 1)
			req, err := http.NewRequest("GET", privateURLBase+"/session/refresh", nil)
			assert.NoError(t, err)
			setHeaders(req)

			res, err := httpClient.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusCreated, res.StatusCode)

			body := test.ReadBody(res)

			var response tokenResponse

			err = json.Unmarshal(body, &response)
			assert.NoError(t, err)
			assert.NotEmpty(t, response.Result)
			token, err := jwt.ParseSigned(response.Result.IDToken)
			assert.NoError(t, err)
			claims := identities.Claims{}
			err = token.Claims(app.KeyStore.Key().Public(), &claims)
			assert.NoError(t, err)

			rSessions, err := app.RefreshTokenStore.FindAll(testSubjectID)
			assert.NoError(t, err)
			assert.Len(t, rSessions, 2)

			u, err := url.Parse(privateURLBase)
			assert.NoError(t, err)
			cookies := httpClient.Jar.Cookies(u)
			assert.NotEmpty(t, cookies)
		})

		t.Run("User with already active session can refresh against private URL", func(t *testing.T) {
			req, err := http.NewRequest("DELETE", publicURLBase+"/session", nil)
			assert.NoError(t, err)
			setHeaders(req)

			res, err := httpClient.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			rSessions, err := app.RefreshTokenStore.FindAll(testSubjectID)
			assert.NoError(t, err)
			assert.Len(t, rSessions, 1)
		})
	})

	t.Run("(Un)locking operations on users succeed across interfaces", func(t *testing.T) {
		t.Run("Lock Account via REST interface reflects on Get Account on gRPC", func(t *testing.T) {
			req, err := http.NewRequest("PATCH", fmt.Sprintf(privateURLBase+"/accounts/%d/lock", testSubjectID), nil)
			assert.NoError(t, err)
			req.SetBasicAuth(app.Config.AuthUsername, app.Config.AuthPassword)

			res, err := httpClient.Do(req)
			assert.NoError(t, err)
			assert.Equalf(t, http.StatusOK, res.StatusCode, "account lock status code doesn't match expectation")
			io.Copy(ioutil.Discard, res.Body)
			assert.NoError(t, res.Body.Close())

			getActResponse, err := securedClient.GetAccount(ctx, &authngrpc.GetAccountRequest{
				Id: 1,
			})
			assert.NoError(t, err)
			assert.Equal(t, &authngrpc.GetAccountResponse{
				Id:       int64(testSubjectID),
				Username: "test@example.com",
				Locked:   true,
				Deleted:  false,
			}, getActResponse.Result)

			// locking account clears all active sessions
			rSessions, err := app.RefreshTokenStore.FindAll(testSubjectID)
			assert.NoError(t, err)
			assert.Len(t, rSessions, 0)
		})
		t.Run("Unlock Account via gRPC interface reflects on Get Account via REST", func(t *testing.T) {
			unlockRes, err := securedClient.UnlockAccount(ctx, &authngrpc.UnlockAccountRequest{
				Id: int64(testSubjectID),
			})
			assert.NoError(t, err)
			assert.Equal(t, &authngrpc.UnlockAccountResponse{}, unlockRes)

			req, err := http.NewRequest("GET", fmt.Sprintf(privateURLBase+"/accounts/%d", testSubjectID), nil)
			assert.NoError(t, err)
			req.SetBasicAuth(app.Config.AuthUsername, app.Config.AuthPassword)

			res, err := httpClient.Do(req)
			assert.NoError(t, err)
			defer res.Body.Close()

			var response struct {
				Result account
			}

			body := test.ReadBody(res)
			err = json.Unmarshal(body, &response)
			assert.NoError(t, err)
			assert.Equal(t, account{
				ID:       int64(testSubjectID),
				Username: "test@example.com",
				Locked:   false,
				Deleted:  false,
			}, response.Result)
		})
	})

	t.Run("Update account coverage", func(t *testing.T) {
		t.Run("Update account with invalid username format returns error", func(t *testing.T) {
			res, err := securedClient.UpdateAccount(context.Background(), &authngrpc.UpdateAccountRequest{
				Id:       int64(testSubjectID),
				Username: "user1",
			})
			assert.Nil(t, res)
			assert.Error(t, err)
			code, ok := status.FromError(err)
			assert.True(t, ok)
			assert.Equal(t, codes.FailedPrecondition, code.Code())

			acc, err := app.AccountStore.Find(testSubjectID)
			assert.NoError(t, err)
			assert.Equal(t, "test@example.com", acc.Username)
		})
		t.Run("Update account with invalid user ID returns error", func(t *testing.T) {
			data := url.Values{
				"username": {"test@example.net"},
			}
			// Account with ID 5300 does not exist
			req, err := http.NewRequest("PUT", privateURLBase+"/accounts/5300", strings.NewReader(data.Encode()))
			assert.NoError(t, err)
			setHeaders(req)
			req.SetBasicAuth(app.Config.AuthUsername, app.Config.AuthPassword)

			res, err := httpClient.Do(req)
			assert.NoError(t, err)

			body := test.ReadBody(res)

			var errResponse errorResponse
			err = json.Unmarshal(body, &errResponse)
			assert.NoError(t, err)
			assert.Len(t, errResponse.Errors, 1)
			assert.Equal(t, services.ErrNotFound, errResponse.Errors[0].Message)

			// The username shouldn't have changed
			acc, err := app.AccountStore.Find(testSubjectID)
			assert.NoError(t, err)
			assert.Equal(t, "test@example.com", acc.Username)
		})
		t.Run("Update account via REST PATCH is successful", func(t *testing.T) {
			data := url.Values{
				"username": {"test@example.net"},
			}

			req, err := http.NewRequest(
				"PATCH",
				fmt.Sprintf(privateURLBase+"/accounts/%d", testSubjectID),
				strings.NewReader(data.Encode()),
			)
			assert.NoError(t, err)
			setHeaders(req)
			req.SetBasicAuth(app.Config.AuthUsername, app.Config.AuthPassword)

			res, err := httpClient.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			acc, err := app.AccountStore.Find(testSubjectID)
			assert.NoError(t, err)
			assert.Equal(t, "test@example.net", acc.Username)
		})
		t.Run("Update account via REST PUT is successful", func(t *testing.T) {
			data := url.Values{
				"username": {"test@example.io"},
			}

			req, err := http.NewRequest(
				"PUT",
				fmt.Sprintf(privateURLBase+"/accounts/%d", testSubjectID),
				strings.NewReader(data.Encode()),
			)
			assert.NoError(t, err)
			setHeaders(req)
			req.SetBasicAuth(app.Config.AuthUsername, app.Config.AuthPassword)

			res, err := httpClient.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			acc, err := app.AccountStore.Find(testSubjectID)
			assert.NoError(t, err)
			assert.Equal(t, "test@example.io", acc.Username)
		})
		t.Run("Update account via gRPC is successful", func(t *testing.T) {
			res, err := securedClient.UpdateAccount(context.Background(), &authngrpc.UpdateAccountRequest{
				Id:       int64(testSubjectID),
				Username: "test@example.gov",
			})
			assert.NoError(t, err)
			assert.NotNil(t, res)

			acc, err := app.AccountStore.Find(testSubjectID)
			assert.NoError(t, err)
			assert.Equal(t, "test@example.gov", acc.Username)
		})
	})
	t.Run("Calls to get JWKs are successful", func(t *testing.T) {
		t.Run("REST call to get JWKs is successful", func(t *testing.T) {
			res, err := httpClient.Get(privateURLBase + "/jwks")
			assert.NoError(t, err)

			body := test.ReadBody(res)

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, []string{"application/json"}, res.Header["Content-Type"])
			assert.NotEmpty(t, body)
		})
		t.Run("gRPC call to get JWKs is successful", func(t *testing.T) {
			res, err := unsecuredClient.JWKS(context.Background(), &authngrpc.JWKSRequest{})
			assert.NoError(t, err)
			assert.NotEmpty(t, res)
		})
	})
	t.Run("Calls for Health Check are successful", func(t *testing.T) {
		t.Run("REST call to get Health Check is successful", func(t *testing.T) {
			res, err := httpClient.Get(publicURLBase + "/health")
			assert.NoError(t, err)

			body := test.ReadBody(res)

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, []string{"application/json"}, res.Header["Content-Type"])
			assert.NotEmpty(t, body)
		})
		t.Run("gRPC call to get Health Check is successful", func(t *testing.T) {
			res, err := publicClient.HealthCheck(context.Background(), &authngrpc.HealthCheckRequest{})
			assert.NoError(t, err)
			assert.NotEmpty(t, res)
		})
	})
	t.Run("Call to /metrics is successful", func(t *testing.T) {
		res, err := httpClient.Get(privateURLBase + "/metrics")
		assert.NoError(t, err)
		assert.NotZero(t, res.ContentLength)
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	})
	t.Run("Service Stats are available when Redis is available", func(t *testing.T) {
		if !app.RedisCheck() {
			t.Skip("Redis is not available")
		}
		t.Run("REST call to Service Stats is successful", func(t *testing.T) {
			res, err := httpClient.Get(privateURLBase + "/stats")
			assert.NoError(t, err)
			assert.NotZero(t, res.ContentLength)
			io.Copy(ioutil.Discard, res.Body)
			res.Body.Close()
		})
		t.Run("gRPC call to Service Stats is successful", func(t *testing.T) {
			res, err := activesClient.ServiceStats(context.Background(), &authngrpc.ServiceStatsRequest{})
			assert.NoError(t, err)
			assert.NotEmpty(t, res)
		})
	})
}
