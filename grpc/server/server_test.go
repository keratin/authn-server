package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
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
	"github.com/keratin/authn-server/app/models"
	"github.com/keratin/authn-server/app/services"
	"github.com/keratin/authn-server/app/tokens/identities"
	"github.com/keratin/authn-server/app/tokens/passwordless"
	"github.com/keratin/authn-server/app/tokens/resets"
	"github.com/keratin/authn-server/app/tokens/sessions"
	authngrpc "github.com/keratin/authn-server/grpc"
	"github.com/keratin/authn-server/grpc/internal/errors"
	oauthlib "github.com/keratin/authn-server/lib/oauth"
	"github.com/keratin/authn-server/lib/route"
	"github.com/keratin/authn-server/server/test"
	"github.com/keratin/authn-server/server/views"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
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
	// setup test app
	app := test.App()
	app.DbCheck = func() bool {
		return true
	}
	app.RedisCheck = func() bool {
		return true
	}
	// The ports are hardcoded because the Server() function blackboxes our
	// access to the listeners, so we can't get their address dynamically.
	app.Config.PublicPort = 8080
	app.Config.ServerPort = 9090

	// run against a real database if the test isn't run with -test.short flag
	if !testing.Short() {
		app = setup(t)
	}

	// start a fake oauth provider
	providerServer := httptest.NewServer(test.ProviderApp())
	defer providerServer.Close()
	// configure a client for the fake oauth provider
	providerClient := oauthlib.NewTestProvider(providerServer)

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
				Id: int64(testSubjectID),
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

func TestRESTInterface(t *testing.T) {
	// setup test app
	testApp := test.App() // Can be setup(t)?
	testApp.Config.UsernameIsEmail = true

	// The ports are hardcoded because the Server() function blackboxes our
	// access to the listeners, so we can't get their address dynamically.
	testApp.Config.PublicPort = 8080
	testApp.Config.ServerPort = 9090

	// parent context for servers
	ctx := context.Background()
	rootCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go Server(rootCtx, testApp)

	jar, _ := cookiejar.New(nil)
	privateClient := route.NewClient(
		fmt.Sprintf("http://127.0.0.1:%d", testApp.Config.ServerPort),
	).
		WithClient(&http.Client{
			Jar: jar,
		}).
		Referred(&testApp.Config.ApplicationDomains[0]).
		Authenticated(testApp.Config.AuthUsername, testApp.Config.AuthPassword)

	jar, _ = cookiejar.New(nil)
	publicClient := route.NewClient(
		fmt.Sprintf("http://127.0.0.1:%d", testApp.Config.PublicPort),
	).
		WithClient(&http.Client{
			Jar: jar,
		}).
		Referred(&testApp.Config.ApplicationDomains[0])

	// Give the server some time to be scheduled and run.
	// TODO: This feels icky, but a cleaner way is yet to be figured out.
	time.Sleep(time.Second * 2)

	t.Run("Private REST", testPrivateRESTInterface(testApp, privateClient))
	t.Run("Public REST", testPublicRESTInterface(testApp, publicClient))
}

func testPrivateRESTInterface(testApp *app.App, client *route.Client) func(*testing.T) {
	return func(t *testing.T) {
		t.Run("Private-Only Routes", func(t *testing.T) {
			t.Run("Index", func(t *testing.T) {
				resp, err := client.Get("/")
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)

				buf := &bytes.Buffer{}
				buf.ReadFrom(resp.Body)
				assert.NotZero(t, buf.Len())
				nominal := &bytes.Buffer{}
				views.Root(nominal)
				assert.Equal(t, nominal.Bytes(), buf.Bytes())
				resp.Body.Close()
			})
			t.Run("JWKS", func(t *testing.T) {
				resp, err := client.Get("/jwks")
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)

				buf := &bytes.Buffer{}
				buf.ReadFrom(resp.Body)
				assert.NotZero(t, buf.Len())
				resp.Body.Close()
			})
			t.Run("Configurations", func(t *testing.T) {
				resp, err := client.Get("/configuration")
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)

				buf := &bytes.Buffer{}
				buf.ReadFrom(resp.Body)
				assert.NotZero(t, buf.Len())
				resp.Body.Close()
			})
			t.Run("Server Stats", func(t *testing.T) {
				resp, err := client.Get("/metrics")
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)

				buf := &bytes.Buffer{}
				buf.ReadFrom(resp.Body)
				assert.NotZero(t, buf.Len())
				resp.Body.Close()
			})
			t.Run("Service Stats", func(t *testing.T) {
				resp, err := client.Get("/stats")
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)

				buf := &bytes.Buffer{}
				buf.ReadFrom(resp.Body)
				assert.NotZero(t, buf.Len())
				resp.Body.Close()
			})
			t.Run("Import Account - Locked", func(t *testing.T) {
				username := generateUsername()
				data := url.Values{
					"username": {username},
					"password": {"11aa22!bb33cc"},
					"locked":   {"true"},
				}

				resp, err := client.PostForm("/accounts/import", data)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				defer resp.Body.Close()

				buf := &bytes.Buffer{}
				buf.ReadFrom(resp.Body)
				assert.NotZero(t, buf.Len())
				var response struct {
					Result struct {
						ID int
					}
				}
				assert.NoError(t, json.Unmarshal(buf.Bytes(), &response))
				assert.NotEmpty(t, response)

				account, err := testApp.AccountStore.Find(response.Result.ID)
				require.NoError(t, err)
				assert.Equal(t, username, account.Username)
				assert.True(t, account.Locked)
			})
			t.Run("Import Account - Unlocked", func(t *testing.T) {
				username := generateUsername()
				data := url.Values{
					"username": {username},
					"password": {"11aa22!bb33cc"},
					"locked":   {"false"},
				}

				resp, err := client.PostForm("/accounts/import", data)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				defer resp.Body.Close()

				buf := &bytes.Buffer{}
				buf.ReadFrom(resp.Body)
				assert.NotZero(t, buf.Len())
				var response struct {
					Result struct {
						ID int
					}
				}
				assert.NoError(t, json.Unmarshal(buf.Bytes(), &response))
				assert.NotEmpty(t, response)

				account, err := testApp.AccountStore.Find(response.Result.ID)
				require.NoError(t, err)
				assert.Equal(t, username, account.Username)
				assert.False(t, account.Locked)
			})
			t.Run("Import Account - Absent Lock Marker", func(t *testing.T) {
				username := generateUsername()
				data := url.Values{
					"username": {username},
					"password": {"11aa22!bb33cc"},
				}

				resp, err := client.PostForm("/accounts/import", data)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				defer resp.Body.Close()

				buf := &bytes.Buffer{}
				buf.ReadFrom(resp.Body)
				assert.NotZero(t, buf.Len())
				var response struct {
					Result struct {
						ID int
					}
				}
				assert.NoError(t, json.Unmarshal(buf.Bytes(), &response))
				assert.NotEmpty(t, response)

				account, err := testApp.AccountStore.Find(response.Result.ID)
				require.NoError(t, err)
				assert.Equal(t, username, account.Username)
				assert.False(t, account.Locked)
			})
			t.Run("Get Locked Account", func(t *testing.T) {
				// Prep
				acc, _ := createUser(t, testApp)
				locked, err := testApp.AccountStore.Lock(acc.ID)
				require.NoError(t, err)
				require.True(t, locked)

				// Test
				resp, err := client.Get(fmt.Sprintf("/accounts/%d", acc.ID))
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				defer resp.Body.Close()

				buf := &bytes.Buffer{}
				buf.ReadFrom(resp.Body)
				assert.NotZero(t, buf.Len())
				var response struct {
					Result struct {
						ID       int
						Username string
						Locked   bool
						Deleted  bool
					}
				}
				assert.NoError(t, json.Unmarshal(buf.Bytes(), &response))
				assert.NotEmpty(t, response)
				assert.True(t, response.Result.Locked)
			})
			t.Run("PATCH: Update Account", func(t *testing.T) {
				// Prep
				acc, _ := createUser(t, testApp)

				// Test
				newUsername := generateUsername()
				data := url.Values{
					"username": {newUsername},
				}

				resp, err := client.Patch(fmt.Sprintf("/accounts/%d", acc.ID), data)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()

				account, err := testApp.AccountStore.Find(acc.ID)
				require.NoError(t, err)
				assert.Equal(t, newUsername, account.Username)
			})
			t.Run("PUT: Update Account", func(t *testing.T) {
				// Prep
				acc, _ := createUser(t, testApp)

				// Test
				newUsername := generateUsername()
				data := url.Values{
					"username": {newUsername},
				}

				resp, err := client.Put(fmt.Sprintf("/accounts/%d", acc.ID), data)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()

				account, err := testApp.AccountStore.Find(acc.ID)
				require.NoError(t, err)
				assert.Equal(t, newUsername, account.Username)
			})
			t.Run("PATCH: Lock Account", func(t *testing.T) {
				// Prep
				acc, _ := createUser(t, testApp)

				// Test
				resp, err := client.Patch(fmt.Sprintf("/accounts/%d/lock", acc.ID), nil)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()

				account, err := testApp.AccountStore.Find(acc.ID)
				require.NoError(t, err)
				assert.True(t, account.Locked)
			})
			t.Run("PUT: Unlock Account", func(t *testing.T) {
				// Prep
				acc, _ := createUser(t, testApp)
				locked, err := testApp.AccountStore.Lock(acc.ID)
				require.NoError(t, err)
				require.True(t, locked)

				// Test
				resp, err := client.Put(fmt.Sprintf("/accounts/%d/unlock", acc.ID), nil)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()

				account, err := testApp.AccountStore.Find(acc.ID)
				require.NoError(t, err)
				assert.False(t, account.Locked)
			})
			t.Run("PUT: Lock Account", func(t *testing.T) {
				// Prep
				acc, _ := createUser(t, testApp)

				// Test
				resp, err := client.Put(fmt.Sprintf("/accounts/%d/lock", acc.ID), nil)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()

				account, err := testApp.AccountStore.Find(acc.ID)
				require.NoError(t, err)
				assert.True(t, account.Locked)
			})
			t.Run("PATCH: Unlock Account", func(t *testing.T) {
				// Prep
				acc, _ := createUser(t, testApp)
				locked, err := testApp.AccountStore.Lock(acc.ID)
				require.NoError(t, err)
				require.True(t, locked)

				// Test
				resp, err := client.Patch(fmt.Sprintf("/accounts/%d/unlock", acc.ID), nil)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()

				account, err := testApp.AccountStore.Find(acc.ID)
				require.NoError(t, err)
				assert.False(t, account.Locked)
			})
			t.Run("PUT: Expire Password", func(t *testing.T) {
				// Prep
				acc, _ := createUser(t, testApp)

				// Test
				resp, err := client.Put(fmt.Sprintf("/accounts/%d/expire_password", acc.ID), nil)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()

				account, err := testApp.AccountStore.Find(acc.ID)
				require.NoError(t, err)
				assert.True(t, account.RequireNewPassword)
			})
			t.Run("PATCH: Expire Password", func(t *testing.T) {
				// Prep
				acc, _ := createUser(t, testApp)

				// Test
				resp, err := client.Patch(fmt.Sprintf("/accounts/%d/expire_password", acc.ID), nil)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()

				account, err := testApp.AccountStore.Find(acc.ID)
				require.NoError(t, err)
				assert.True(t, account.RequireNewPassword)
			})
			t.Run("Archive Account", func(t *testing.T) {
				// Prep
				acc, _ := createUser(t, testApp)

				// Test
				resp, err := client.Delete(fmt.Sprintf("/accounts/%d", acc.ID))
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				io.Copy(ioutil.Discard, resp.Body)
				resp.Body.Close()

				acc, err = testApp.AccountStore.Find(acc.ID)
				assert.NoError(t, err)
				assert.True(t, acc.Archived())
			})
		})
		t.Run("Public Routes", testPublicRESTInterface(testApp, client))
	}
}

func testPublicRESTInterface(testApp *app.App, client *route.Client) func(*testing.T) {
	return func(t *testing.T) {
		t.Run("Health", func(t *testing.T) {
			resp, err := client.Get("/health")
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			buf := &bytes.Buffer{}
			buf.ReadFrom(resp.Body)
			assert.NotZero(t, buf.Len())
			resp.Body.Close()
		})
		t.Run("Signup", func(t *testing.T) {
			t.Run("Successful", func(t *testing.T) {
				newUsername := generateUsername()
				data := url.Values{
					"username": {newUsername},
					"password": {"aa11bb22!cc"},
				}

				resp, err := client.PostForm("/accounts", data)
				assert.NoError(t, err)
				if assert.Equal(t, http.StatusCreated, resp.StatusCode) {
					test.AssertSession(t, testApp.Config, resp.Cookies())
					test.AssertIDTokenResponse(t, resp, testApp.KeyStore, testApp.Config)
				}
			})
			t.Run("Missing Username & Password", func(t *testing.T) {
				// Temporary disbale the requirement
				testApp.Config.UsernameIsEmail = false
				defer func() {
					testApp.Config.UsernameIsEmail = true
				}()

				data := url.Values{
					"username": {""},
					"password": {""},
				}

				resp, err := client.PostForm("/accounts", data)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

				test.AssertErrors(t, resp, services.FieldErrors{{"username", "MISSING"}, {"password", "MISSING"}})
			})
			t.Run("Invalid Format", func(t *testing.T) {
				data := url.Values{
					"username": {"test"},
					"password": {"aa11bb22!cc"},
				}

				resp, err := client.PostForm("/accounts", data)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

				test.AssertErrors(t, resp, services.FieldErrors{{"username", "FORMAT_INVALID"}})
			})
			t.Run("Taken Username", func(t *testing.T) {
				// Prep
				acc, _ := createUser(t, testApp)

				// Test
				data := url.Values{
					"username": {acc.Username},
					"password": {"aa11bb22!cc"},
				}

				resp, err := client.PostForm("/accounts", data)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

				test.AssertErrors(t, resp, services.FieldErrors{{"username", "TAKEN"}})
			})
			t.Run("Missing Password", func(t *testing.T) {
				data := url.Values{
					"username": {generateUsername()},
				}

				resp, err := client.PostForm("/accounts", data)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

				test.AssertErrors(t, resp, services.FieldErrors{{"password", "MISSING"}})
			})
			t.Run("Insecure Password", func(t *testing.T) {
				data := url.Values{
					"username": {generateUsername()},
					"password": {"1"},
				}

				resp, err := client.PostForm("/accounts", data)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

				test.AssertErrors(t, resp, services.FieldErrors{{"password", "INSECURE"}})
			})
		})
		t.Run("Username Availability", func(t *testing.T) {
			t.Run("Available", func(t *testing.T) {
				resp, err := client.Get(fmt.Sprintf("/accounts/available?username=%s", generateUsername()))
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			})
			t.Run("Taken", func(t *testing.T) {
				// Prep
				acc, _ := createUser(t, testApp)

				// Test
				resp, err := client.Get(fmt.Sprintf("/accounts/available?username=%s", acc.Username))
				assert.NoError(t, err)
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
				test.AssertErrors(t, resp, services.FieldErrors{{"username", "TAKEN"}})
			})
		})
		t.Run("Login", func(t *testing.T) {
			t.Run("Successful", func(t *testing.T) {
				// Prep
				acc, password := createUser(t, testApp)

				// Test
				resp, err := client.PostForm("/session", url.Values{
					"username": {acc.Username},
					"password": {password},
				})
				assert.NoError(t, err)
				assertSuccessfulSession(t, testApp, resp, acc)
			})
			t.Run("Failed", func(t *testing.T) {
				// Prep
				acc, _ := createUser(t, testApp)

				// Test
				resp, err := client.PostForm("/session", url.Values{
					"username": {acc.Username},
				})
				assert.NoError(t, err)
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

				test.AssertErrors(t, resp, services.FieldErrors{{"credentials", "FAILED"}})
			})
			t.Run("Expired", func(t *testing.T) {
				// Prep
				acc, password := createUser(t, testApp)
				testApp.AccountStore.RequireNewPassword(acc.ID)

				// Test
				resp, err := client.PostForm("/session", url.Values{
					"username": {acc.Username},
					"password": {password},
				})
				assert.NoError(t, err)
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

				test.AssertErrors(t, resp, services.FieldErrors{{"credentials", "EXPIRED"}})
			})
			t.Run("Locked", func(t *testing.T) {
				// Prep
				acc, password := createUser(t, testApp)
				locked, err := testApp.AccountStore.Lock(acc.ID)
				require.NoError(t, err)
				require.True(t, locked)

				// Test
				resp, err := client.PostForm("/session", url.Values{
					"username": {acc.Username},
					"password": {password},
				})
				assert.NoError(t, err)
				assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

				test.AssertErrors(t, resp, services.FieldErrors{{"account", "LOCKED"}})
			})
		})
		t.Run("Refresh Session", func(t *testing.T) {
			t.Run("Successful", func(t *testing.T) {
				// Prep
				acc, _ := createUser(t, testApp)

				// Test
				existingSession := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, acc.ID)
				resp, err := client.WithCookie(existingSession).Get("/session/refresh")
				assert.NoError(t, err)

				if assert.Equal(t, http.StatusCreated, resp.StatusCode) {
					test.AssertIDTokenResponse(t, resp, testApp.KeyStore, testApp.Config)
				}
			})
			t.Run("Failed", func(t *testing.T) {
				// Lifted from: servers/handlers/get_session_refresh_test.go#TestGetSessionRefreshFailure
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

					client := client.WithCookie(existingSession)
					res, err := client.Get("/session/refresh")
					require.NoError(t, err)

					assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
				}
			})
		})
		t.Run("Logout", func(t *testing.T) {
			t.Run("Successful", func(t *testing.T) {
				// Prep
				acc, _ := createUser(t, testApp)
				session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, acc.ID)

				// token exists
				claims, err := sessions.Parse(session.Value, testApp.Config)
				require.NoError(t, err)
				id, err := testApp.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
				require.NoError(t, err)
				assert.NotEmpty(t, id)

				// Test
				res, err := client.WithCookie(session).Delete("/session")
				assert.NoError(t, err)

				// request always succeeds
				assert.Equal(t, http.StatusOK, res.StatusCode)

				// token no longer exists
				id, err = testApp.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
				require.NoError(t, err)
				assert.Empty(t, id)
			})
			t.Run("Failure", func(t *testing.T) {
				// Prep
				badCfg := &app.Config{
					AuthNURL:           testApp.Config.AuthNURL,
					SessionCookieName:  testApp.Config.SessionCookieName,
					SessionSigningKey:  []byte("wrong"),
					ApplicationDomains: testApp.Config.ApplicationDomains,
				}
				session := test.CreateSession(testApp.RefreshTokenStore, badCfg, 123)

				// Test
				res, err := client.WithCookie(session).Delete("/session")
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, res.StatusCode)
			})
		})
		t.Run("Passwordless", func(t *testing.T) {
			t.Run("Request", func(t *testing.T) {
				t.Run("Known Account", func(t *testing.T) {
					acc, _ := createUser(t, testApp)

					res, err := client.Get("/session/token?username=" + acc.Username)
					require.NoError(t, err)
					assert.Equal(t, http.StatusOK, res.StatusCode)
				})
				t.Run("Unknown Account", func(t *testing.T) {
					res, err := client.Get("/session/token?username=" + generateUsername())
					require.NoError(t, err)
					assert.Equal(t, http.StatusOK, res.StatusCode)
				})
			})
			t.Run("Submit", func(t *testing.T) {
				t.Run("Successful - Valid Token", func(t *testing.T) {
					acc, _ := createUser(t, testApp)

					// given a passwordless token
					token, err := passwordless.New(testApp.Config, acc.ID)
					require.NoError(t, err)
					tokenStr, err := token.Sign(testApp.Config.PasswordlessTokenSigningKey)
					require.NoError(t, err)

					// invoking the endpoint
					res, err := client.PostForm("/session/token", url.Values{
						"token": []string{tokenStr},
					})
					require.NoError(t, err)

					// works
					assertSuccessfulSession(t, testApp, res, acc)
				})
				t.Run("Failure - Invalid Token", func(t *testing.T) {
					// invoking the endpoint
					res, err := client.PostForm("/session/token", url.Values{
						"token": []string{"invalid"},
					})
					require.NoError(t, err)

					// does not work
					assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
					test.AssertErrors(t, res, services.FieldErrors{{"token", "INVALID_OR_EXPIRED"}})
				})
				t.Run("Successful - Valid Session", func(t *testing.T) {
					acc, _ := createUser(t, testApp)

					// given a session
					session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, acc.ID)

					// given a passwordless token
					token, err := passwordless.New(testApp.Config, acc.ID)
					require.NoError(t, err)
					tokenStr, err := token.Sign(testApp.Config.PasswordlessTokenSigningKey)
					require.NoError(t, err)

					// invoking the endpoint
					res, err := client.WithCookie(session).PostForm("/session/token", url.Values{
						"token": []string{tokenStr},
					})
					require.NoError(t, err)

					// works
					assertSuccessfulSession(t, testApp, res, acc)

					// invalidates old session
					claims, err := sessions.Parse(session.Value, testApp.Config)
					require.NoError(t, err)
					id, err := testApp.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
					require.NoError(t, err)
					assert.Empty(t, id)
				})
			})
		})
		t.Run("Change Password", func(t *testing.T) {
			// Lifted from server/handlers/post_password_test.go

			t.Run("Successful - Valid Reset Token", func(t *testing.T) {
				// Prep
				acc, _ := createUser(t, testApp)
				token, err := resets.New(testApp.Config, acc.ID, acc.PasswordChangedAt)
				require.NoError(t, err)
				tokenStr, err := token.Sign(testApp.Config.ResetSigningKey)
				require.NoError(t, err)

				// Test
				res, err := client.PostForm("/password", url.Values{
					"token":    {tokenStr},
					"password": {"0a0b!c0d0"},
				})
				assert.NoError(t, err)

				assertSuccessfulSession(t, testApp, res, acc)
				assertChangedPassword(t, testApp, acc)
			})
			t.Run("Failure - Invalid Reset Token", func(t *testing.T) {
				res, err := client.PostForm("/password", url.Values{
					"token":    {"invalid"},
					"password": {"0a0b!c0d0"},
				})
				assert.NoError(t, err)
				assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
				test.AssertErrors(t, res, services.FieldErrors{{"token", "INVALID_OR_EXPIRED"}})
			})
			t.Run("Successful - Valid Session", func(t *testing.T) {
				acc, password := createUser(t, testApp)

				// given a session
				session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, acc.ID)

				// invoking the endpoint
				res, err := client.WithCookie(session).PostForm("/password", url.Values{
					"currentPassword": {password},
					"password":        {"0a0b0c0d0"},
				})
				require.NoError(t, err)

				// works
				assertSuccessfulSession(t, testApp, res, acc)
				assertChangedPassword(t, testApp, acc)

				// invalidates old session
				claims, err := sessions.Parse(session.Value, testApp.Config)
				require.NoError(t, err)
				id, err := testApp.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
				require.NoError(t, err)
				assert.Empty(t, id)
			})
			t.Run("Failure - Valid Session & Insecure Password", func(t *testing.T) {
				acc, password := createUser(t, testApp)

				// given a session
				session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, acc.ID)

				// invoking the endpoint
				res, err := client.WithCookie(session).PostForm("/password", url.Values{
					"currentPassword": {password},
					"password":        {"a"},
				})
				require.NoError(t, err)

				assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
				test.AssertErrors(t, res, services.FieldErrors{{"password", "INSECURE"}})
			})
			t.Run("Failure - Valid Session & Invalid Current Password", func(t *testing.T) {
				acc, _ := createUser(t, testApp)

				// given a session
				session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, acc.ID)

				// invoking the endpoint
				res, err := client.WithCookie(session).PostForm("/password", url.Values{
					"currentPassword": {"wrong"},
					"password":        {"0a0b0c0d0"},
				})
				require.NoError(t, err)

				assert.Equal(t, http.StatusUnprocessableEntity, res.StatusCode)
				test.AssertErrors(t, res, services.FieldErrors{{"credentials", "FAILED"}})
			})
			t.Run("Failure - Invalid Session", func(t *testing.T) {
				session := &http.Cookie{
					Name:  testApp.Config.SessionCookieName,
					Value: "invalid",
				}

				res, err := client.WithCookie(session).PostForm("/password", url.Values{
					"currentPassword": {"oldpwd"},
					"password":        {"0a0b0c0d0"},
				})
				require.NoError(t, err)

				assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
			})
			t.Run("Successful - Valid Token & Session", func(t *testing.T) {
				// Token account
				tokenAccount, password := createUser(t, testApp)

				token, err := resets.New(testApp.Config, tokenAccount.ID, tokenAccount.PasswordChangedAt)
				require.NoError(t, err)
				tokenStr, err := token.Sign(testApp.Config.ResetSigningKey)
				require.NoError(t, err)

				// given another account
				sessionAccount, _ := createUser(t, testApp)
				require.NoError(t, err)
				// with a session
				session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, sessionAccount.ID)

				// invoking the endpoint
				res, err := client.WithCookie(session).PostForm("/password", url.Values{
					"token":           {tokenStr},
					"currentPassword": {password},
					"password":        {"0a0b0c0d0"},
				})
				require.NoError(t, err)

				// works
				assertSuccessfulSession(t, testApp, res, tokenAccount)
				assertChangedPassword(t, testApp, tokenAccount)
			})
		})
		t.Run("Password Reset", func(t *testing.T) {
			t.Run("Known Account", func(t *testing.T) {
				acc, _ := createUser(t, testApp)

				res, err := client.Get(fmt.Sprintf("/password/reset?username=%s", acc.Username))
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, res.StatusCode)
			})
			t.Run("Unknown Account", func(t *testing.T) {
				res, err := client.Get(fmt.Sprintf("/password/reset?username=%s", generateUsername()))
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, res.StatusCode)
			})
		})
		//TODO: Add OAuth tests
	}
}

func TestGRPCInterface(t *testing.T) {
	// setup test app
	testApp := test.App()
	testApp.DbCheck = func() bool {
		return true
	}
	testApp.RedisCheck = func() bool {
		return true
	}
	// The ports are hardcoded because the Server() function blackboxes our
	// access to the listeners, so we can't get their address dynamically.
	testApp.Config.PublicPort = 8080
	testApp.Config.ServerPort = 9090

	// run against a real database if the test isn't run with -test.short flag
	if !testing.Short() {
		testApp = setup(t)
	}

	// We still want the username to be an email for testing purposes
	testApp.Config.UsernameIsEmail = true

	// parent context for servers
	ctx := context.Background()
	rootCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go Server(rootCtx, testApp)

	publicClientConn, err := grpc.DialContext(ctx, fmt.Sprintf("localhost:%d", testApp.Config.PublicPort), grpc.WithInsecure(), grpc.WithBlock())
	require.NoError(t, err)

	privateClientConn, err := grpc.DialContext(ctx, fmt.Sprintf("localhost:%d", testApp.Config.ServerPort), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithPerRPCCredentials(basicAuth{
		username: testApp.Config.AuthUsername,
		password: testApp.Config.AuthPassword,
	}))
	require.NoError(t, err)

	// Give the server some time to be scheduled and run.
	// TODO: This feels icky, but a cleaner way is yet to be figured out.
	time.Sleep(time.Second * 2)

	t.Run("Private", testPrivateGRPCInterface(testApp, privateClientConn))
	t.Run("Public", testPublicGRPCInterface(testApp, publicClientConn))
}

func testPrivateGRPCInterface(testApp *app.App, client *grpc.ClientConn) func(*testing.T) {
	return func(t *testing.T) {
		t.Run("Private-Only Services", func(t *testing.T) {
			t.Run("UnsecuredAdminAuthN Service", func(t *testing.T) {
				svcClient := authngrpc.NewUnsecuredAdminAuthNClient(client)
				t.Run("JWKS", func(t *testing.T) {
					resp, err := svcClient.JWKS(context.Background(), &authngrpc.JWKSRequest{})
					assert.NoError(t, err)

					assert.NotNil(t, resp)
					assert.NotZero(t, len(resp.GetKeys()))
				})
				t.Run("Configurations", func(t *testing.T) {
					resp, err := svcClient.ServiceConfiguration(context.Background(), &authngrpc.ServiceConfigurationRequest{})
					assert.NoError(t, err)

					assert.NotEmpty(t, resp)
				})
			})
			t.Run("AuthNActives Service", func(t *testing.T) {
				if testApp.Actives == nil {
					t.Skip()
				}
				svcClient := authngrpc.NewAuthNActivesClient(client)
				t.Run("Service Stats", func(t *testing.T) {
					resp, err := svcClient.ServiceStats(context.Background(), &authngrpc.ServiceStatsRequest{})
					assert.NoError(t, err)
					assert.NotEmpty(t, resp)
				})
			})

			t.Run("SecuredAdminAuthN Service", func(t *testing.T) {
				svcClient := authngrpc.NewSecuredAdminAuthNClient(client)
				t.Run("Import Account - Locked", func(t *testing.T) {
					username := generateUsername()

					resp, err := svcClient.ImportAccount(context.Background(), &authngrpc.ImportAccountRequest{
						Username: username,
						Password: "11aa22!bb33cc",
						Locked:   true,
					})
					assert.NoError(t, err)
					assert.NotEmpty(t, resp)

					account, err := testApp.AccountStore.Find(int(resp.GetResult().GetId()))
					require.NoError(t, err)
					assert.Equal(t, username, account.Username)
					assert.True(t, account.Locked)
				})
				t.Run("Import Account - Unlocked", func(t *testing.T) {
					username := generateUsername()
					resp, err := svcClient.ImportAccount(context.Background(), &authngrpc.ImportAccountRequest{
						Username: username,
						Password: "11aa22!bb33cc",
						Locked:   false,
					})
					assert.NoError(t, err)
					assert.NotEmpty(t, resp)

					account, err := testApp.AccountStore.Find(int(resp.Result.Id))
					require.NoError(t, err)
					assert.Equal(t, username, account.Username)
					assert.False(t, account.Locked)
				})
				t.Run("Import Account - Absent Lock Marker", func(t *testing.T) {
					username := generateUsername()
					resp, err := svcClient.ImportAccount(context.Background(), &authngrpc.ImportAccountRequest{
						Username: username,
						Password: "11aa22!bb33cc",
					})
					assert.NoError(t, err)
					assert.NotEmpty(t, resp)

					account, err := testApp.AccountStore.Find(int(resp.Result.Id))
					require.NoError(t, err)
					assert.Equal(t, username, account.Username)
					assert.False(t, account.Locked)
				})
				t.Run("Get Locked Account", func(t *testing.T) {
					// Prep
					acc, _ := createUser(t, testApp)
					locked, err := testApp.AccountStore.Lock(acc.ID)
					require.NoError(t, err)
					require.True(t, locked)

					// Test
					resp, err := svcClient.GetAccount(context.Background(), &authngrpc.GetAccountRequest{
						Id: int64(acc.ID),
					})
					assert.NoError(t, err)

					assert.True(t, resp.Result.Locked)
				})
				t.Run("Update Account", func(t *testing.T) {
					// Prep
					acc, _ := createUser(t, testApp)

					// Test
					username := generateUsername()
					_, err := svcClient.UpdateAccount(context.Background(), &authngrpc.UpdateAccountRequest{
						Id:       int64(acc.ID),
						Username: username,
					})
					assert.NoError(t, err)

					account, err := testApp.AccountStore.Find(acc.ID)
					require.NoError(t, err)
					assert.Equal(t, username, account.Username)
				})
				t.Run("Lock Account", func(t *testing.T) {
					// Prep
					acc, _ := createUser(t, testApp)

					// Test
					_, err := svcClient.LockAccount(context.Background(), &authngrpc.LockAccountRequest{
						Id: int64(acc.ID),
					})
					assert.NoError(t, err)

					account, err := testApp.AccountStore.Find(acc.ID)
					require.NoError(t, err)
					assert.True(t, account.Locked)
				})
				t.Run("Unlock Account", func(t *testing.T) {
					// Prep
					acc, _ := createUser(t, testApp)
					locked, err := testApp.AccountStore.Lock(acc.ID)
					require.NoError(t, err)
					require.True(t, locked)

					// Test
					_, err = svcClient.UnlockAccount(context.Background(), &authngrpc.UnlockAccountRequest{
						Id: int64(acc.ID),
					})
					assert.NoError(t, err)

					account, err := testApp.AccountStore.Find(acc.ID)
					require.NoError(t, err)
					assert.False(t, account.Locked)
				})
				t.Run("Expire Password", func(t *testing.T) {
					// Prep
					acc, _ := createUser(t, testApp)

					// Test
					_, err := svcClient.ExpirePassword(context.Background(), &authngrpc.ExpirePasswordRequest{
						Id: int64(acc.ID),
					})
					assert.NoError(t, err)

					account, err := testApp.AccountStore.Find(acc.ID)
					require.NoError(t, err)
					assert.True(t, account.RequireNewPassword)
				})
				t.Run("Archive Account", func(t *testing.T) {
					// Prep
					acc, _ := createUser(t, testApp)

					// Test
					_, err := svcClient.ArchiveAccount(context.Background(), &authngrpc.ArchiveAccountRequest{
						Id: int64(acc.ID),
					})
					assert.NoError(t, err)

					acc, err = testApp.AccountStore.Find(acc.ID)
					assert.NoError(t, err)
					assert.True(t, acc.Archived())
				})
			})
		})
		t.Run("Public Routes", testPublicGRPCInterface(testApp, client))
	}
}

func testPublicGRPCInterface(testApp *app.App, client *grpc.ClientConn) func(*testing.T) {
	return func(t *testing.T) {
		t.Run("SignupService Service", func(t *testing.T) {
			svcClient := authngrpc.NewSignupServiceClient(client)
			t.Run("Signup", func(t *testing.T) {
				t.Run("Successful", func(t *testing.T) {
					username := generateUsername()
					var header metadata.MD
					resp, err := svcClient.Signup(context.Background(), &authngrpc.SignupRequest{
						Username: username,
						Password: "aa11bb22!cc",
					}, grpc.Header(&header))

					assert.NoError(t, err)

					assert.NotZero(t, len(header.Get(testApp.Config.SessionCookieName)))
					test.AssertGRPCSession(t, testApp.Config, header)
					test.AssertGRPCIDTokenResponse(t, resp.GetResult().GetIdToken(), testApp.KeyStore, testApp.Config)

				})
				t.Run("Missing Username & Password", func(t *testing.T) {
					// Temporary disbale the requirement
					testApp.Config.UsernameIsEmail = false
					defer func() {
						testApp.Config.UsernameIsEmail = true
					}()

					_, err := svcClient.Signup(context.Background(), &authngrpc.SignupRequest{})
					assert.Error(t, err)

					st := status.Convert(err)
					assert.Equal(t, codes.FailedPrecondition, st.Code())

					errorDetails := st.Details()[0].(*errdetails.BadRequest)
					fes := errors.ToFieldErrors(errorDetails)
					assert.EqualValues(t, services.FieldErrors{{"username", "MISSING"}, {"password", "MISSING"}}, fes)
				})
				t.Run("Invalid Format", func(t *testing.T) {

					_, err := svcClient.Signup(context.Background(), &authngrpc.SignupRequest{
						Username: "test",
						Password: "aa11bb22!cc",
					})
					assert.Error(t, err)
					st := status.Convert(err)
					assert.Equal(t, codes.FailedPrecondition, st.Code())

					errorDetails := st.Details()[0].(*errdetails.BadRequest)
					fes := errors.ToFieldErrors(errorDetails)
					assert.EqualValues(t, services.FieldErrors{{"username", "FORMAT_INVALID"}}, fes)
				})
				t.Run("Taken Username", func(t *testing.T) {
					// Prep
					acc, _ := createUser(t, testApp)

					// Test
					_, err := svcClient.Signup(context.Background(), &authngrpc.SignupRequest{
						Username: acc.Username,
						Password: "aa11bb22!cc",
					})
					assert.Error(t, err)
					st := status.Convert(err)
					assert.Equal(t, codes.FailedPrecondition, st.Code())
					errorDetails := st.Details()[0].(*errdetails.BadRequest)
					fes := errors.ToFieldErrors(errorDetails)
					assert.EqualValues(t, services.FieldErrors{{"username", "TAKEN"}}, fes)
				})
				t.Run("Missing Password", func(t *testing.T) {
					_, err := svcClient.Signup(context.Background(), &authngrpc.SignupRequest{
						Username: generateUsername(),
					})
					assert.Error(t, err)
					st := status.Convert(err)
					assert.Equal(t, codes.FailedPrecondition, st.Code())

					errorDetails := st.Details()[0].(*errdetails.BadRequest)
					fes := errors.ToFieldErrors(errorDetails)
					assert.EqualValues(t, services.FieldErrors{{"password", "MISSING"}}, fes)
				})
				t.Run("Insecure Password", func(t *testing.T) {
					_, err := svcClient.Signup(context.Background(), &authngrpc.SignupRequest{
						Username: generateUsername(),
						Password: "1",
					}) // client.PostForm("/accounts", data)
					assert.Error(t, err)
					st := status.Convert(err)
					assert.Equal(t, codes.FailedPrecondition, st.Code())

					errorDetails := st.Details()[0].(*errdetails.BadRequest)
					fes := errors.ToFieldErrors(errorDetails)
					assert.EqualValues(t, services.FieldErrors{{"password", "INSECURE"}}, fes)
				})
			})
			t.Run("IsUsernameAvailable", func(t *testing.T) {
				t.Run("Available", func(t *testing.T) {
					resp, err := svcClient.IsUsernameAvailable(context.Background(), &authngrpc.IsUsernameAvailableRequest{
						Username: generateUsername(),
					})
					assert.NoError(t, err)
					assert.True(t, resp.GetResult())
				})
				t.Run("Taken", func(t *testing.T) {
					// Prep
					acc, _ := createUser(t, testApp)

					// Test
					_, err := svcClient.IsUsernameAvailable(context.Background(), &authngrpc.IsUsernameAvailableRequest{
						Username: acc.Username,
					})
					assert.Error(t, err)
					st := status.Convert(err)
					assert.Equal(t, codes.FailedPrecondition, st.Code())

					errorDetails := st.Details()[0].(*errdetails.BadRequest)
					fes := errors.ToFieldErrors(errorDetails)
					assert.EqualValues(t, services.FieldErrors{{"username", "TAKEN"}}, fes)
				})
			})
		})
		t.Run("PublicAuthN", func(t *testing.T) {
			svcClient := authngrpc.NewPublicAuthNClient(client)
			t.Run("Health", func(t *testing.T) {
				resp, err := svcClient.HealthCheck(context.Background(), &authngrpc.HealthCheckRequest{})
				assert.NoError(t, err)
				assert.NotEmpty(t, resp)
			})
			t.Run("Login", func(t *testing.T) {
				t.Run("Successful", func(t *testing.T) {
					// Prep
					acc, password := createUser(t, testApp)

					// Test
					var header metadata.MD
					resp, err := svcClient.Login(context.Background(), &authngrpc.LoginRequest{
						Username: acc.Username,
						Password: password,
					}, grpc.Header(&header))
					assert.NoError(t, err)
					test.AssertGRPCSession(t, testApp.Config, header)
					test.AssertGRPCIDTokenResponse(t, resp.GetResult().GetIdToken(), testApp.KeyStore, testApp.Config)
				})
				t.Run("Failed", func(t *testing.T) {
					// Prep
					acc, _ := createUser(t, testApp)

					// Test
					_, err := svcClient.Login(context.Background(), &authngrpc.LoginRequest{
						Username: acc.Username,
					})
					assert.Error(t, err)
					st := status.Convert(err)
					assert.Equal(t, codes.FailedPrecondition, st.Code())

					errorDetails := st.Details()[0].(*errdetails.BadRequest)
					fes := errors.ToFieldErrors(errorDetails)
					assert.EqualValues(t, services.FieldErrors{{"credentials", "FAILED"}}, fes)
				})
				t.Run("Expired", func(t *testing.T) {
					// Prep
					acc, password := createUser(t, testApp)
					testApp.AccountStore.RequireNewPassword(acc.ID)

					// Test
					_, err := svcClient.Login(context.Background(), &authngrpc.LoginRequest{
						Username: acc.Username,
						Password: password,
					})
					assert.Error(t, err)
					st := status.Convert(err)
					assert.Equal(t, codes.FailedPrecondition, st.Code())

					errorDetails := st.Details()[0].(*errdetails.BadRequest)
					fes := errors.ToFieldErrors(errorDetails)
					assert.EqualValues(t, services.FieldErrors{{"credentials", "EXPIRED"}}, fes)
				})
				t.Run("Locked", func(t *testing.T) {
					// Prep
					acc, password := createUser(t, testApp)
					locked, err := testApp.AccountStore.Lock(acc.ID)
					require.NoError(t, err)
					require.True(t, locked)

					// Test
					_, err = svcClient.Login(context.Background(), &authngrpc.LoginRequest{
						Username: acc.Username,
						Password: password,
					})
					assert.Error(t, err)
					st := status.Convert(err)
					assert.Equal(t, codes.FailedPrecondition, st.Code())

					errorDetails := st.Details()[0].(*errdetails.BadRequest)
					fes := errors.ToFieldErrors(errorDetails)
					assert.EqualValues(t, services.FieldErrors{{"account", "LOCKED"}}, fes)
				})
			})
			t.Run("Refresh Session", func(t *testing.T) {
				t.Run("Successful", func(t *testing.T) {
					// Prep
					acc, _ := createUser(t, testApp)
					ctx := context.Background()

					// Test
					existingSession := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, acc.ID)
					ctx = metadata.AppendToOutgoingContext(ctx, testApp.Config.SessionCookieName, existingSession.Value)
					resp, err := svcClient.RefreshSession(ctx, &authngrpc.RefreshSessionRequest{})
					assert.NoError(t, err)

					test.AssertGRPCIDTokenResponse(t, resp.GetResult().GetIdToken(), testApp.KeyStore, testApp.Config)
				})
				t.Run("Failed", func(t *testing.T) {
					// Lifted from: servers/handlers/get_session_refresh_test.go#TestGetSessionRefreshFailure
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

						ctx := context.Background()
						ctx = metadata.AppendToOutgoingContext(ctx, tcCfg.SessionCookieName, existingSession.Value)
						_, err := svcClient.RefreshSession(ctx, &authngrpc.RefreshSessionRequest{}) // client.Get("/session/refresh")
						require.Error(t, err)
						st := status.Convert(err)
						assert.Equal(t, codes.Unauthenticated, st.Code())
					}
				})
			})
			t.Run("Logout", func(t *testing.T) {
				t.Run("Successful", func(t *testing.T) {
					// Prep
					acc, _ := createUser(t, testApp)
					session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, acc.ID)

					// token exists
					claims, err := sessions.Parse(session.Value, testApp.Config)
					require.NoError(t, err)
					id, err := testApp.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
					require.NoError(t, err)
					assert.NotEmpty(t, id)

					// Test
					ctx := context.Background()
					ctx = metadata.AppendToOutgoingContext(ctx, testApp.Config.SessionCookieName, session.Value)
					_, err = svcClient.Logout(ctx, &authngrpc.LogoutRequest{})
					assert.NoError(t, err)

					// token no longer exists
					id, err = testApp.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
					require.NoError(t, err)
					assert.Empty(t, id)
				})
				t.Run("Failure", func(t *testing.T) {
					// Prep
					badCfg := &app.Config{
						AuthNURL:           testApp.Config.AuthNURL,
						SessionCookieName:  testApp.Config.SessionCookieName,
						SessionSigningKey:  []byte("wrong"),
						ApplicationDomains: testApp.Config.ApplicationDomains,
					}
					session := test.CreateSession(testApp.RefreshTokenStore, badCfg, 123)

					// Test
					ctx := context.Background()
					ctx = metadata.AppendToOutgoingContext(ctx, testApp.Config.SessionCookieName, session.Value)
					_, err := svcClient.Logout(ctx, &authngrpc.LogoutRequest{})

					// This method never returns an error
					assert.NoError(t, err)
				})
			})
			t.Run("Change Password", func(t *testing.T) {
				// Lifted from server/handlers/post_password_test.go

				t.Run("Successful - Valid Reset Token", func(t *testing.T) {
					// Prep
					acc, _ := createUser(t, testApp)
					token, err := resets.New(testApp.Config, acc.ID, acc.PasswordChangedAt)
					require.NoError(t, err)
					tokenStr, err := token.Sign(testApp.Config.ResetSigningKey)
					require.NoError(t, err)

					// Test
					var header metadata.MD
					res, err := svcClient.ChangePassword(context.Background(), &authngrpc.ChangePasswordRequest{
						Token:    tokenStr,
						Password: "0a0b!c0d0",
					}, grpc.Header(&header))
					assert.NoError(t, err)

					assertSuccessfulGRPCSession(t, testApp, res.GetResult().GetIdToken(), header, acc)
					assertChangedPassword(t, testApp, acc)
				})
				t.Run("Failure - Invalid Reset Token", func(t *testing.T) {
					_, err := svcClient.ChangePassword(context.Background(), &authngrpc.ChangePasswordRequest{
						Token:    "invalid",
						Password: "0a0b!c0d0",
					})

					assert.Error(t, err)
					st := status.Convert(err)
					assert.Equal(t, codes.FailedPrecondition, st.Code())
					fes := errors.ToFieldErrors(st.Details()[0].(*errdetails.BadRequest))
					assert.EqualValues(t, services.FieldErrors{{"token", "INVALID_OR_EXPIRED"}}, fes)
				})
				t.Run("Successful - Valid Session", func(t *testing.T) {
					acc, password := createUser(t, testApp)

					// given a session
					session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, acc.ID)

					ctx := metadata.AppendToOutgoingContext(context.Background(), testApp.Config.SessionCookieName, session.Value)
					var header metadata.MD

					// invoking the endpoint
					res, err := svcClient.ChangePassword(ctx, &authngrpc.ChangePasswordRequest{
						CurrentPassword: password,
						Password:        "0a0b0c0d0",
					}, grpc.Header(&header))
					assert.NoError(t, err)

					// works
					assertSuccessfulGRPCSession(t, testApp, res.GetResult().GetIdToken(), header, acc)
					assertChangedPassword(t, testApp, acc)

					// invalidates old session
					claims, err := sessions.Parse(session.Value, testApp.Config)
					require.NoError(t, err)
					id, err := testApp.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
					require.NoError(t, err)
					assert.Empty(t, id)
				})
				t.Run("Failure - Valid Session & Insecure Password", func(t *testing.T) {
					acc, password := createUser(t, testApp)

					// given a session
					session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, acc.ID)
					ctx := metadata.AppendToOutgoingContext(context.Background(), testApp.Config.SessionCookieName, session.Value)

					// invoking the endpoint
					_, err := svcClient.ChangePassword(ctx, &authngrpc.ChangePasswordRequest{
						CurrentPassword: password,
						Password:        "a",
					})
					require.Error(t, err)
					st := status.Convert(err)
					assert.Equal(t, codes.FailedPrecondition, st.Code())
					fes := errors.ToFieldErrors(st.Details()[0].(*errdetails.BadRequest))
					assert.EqualValues(t, services.FieldErrors{{"password", "INSECURE"}}, fes)
				})
				t.Run("Failure - Valid Session & Invalid Current Password", func(t *testing.T) {
					acc, _ := createUser(t, testApp)

					// given a session
					session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, acc.ID)
					ctx := metadata.AppendToOutgoingContext(context.Background(), testApp.Config.SessionCookieName, session.Value)

					// invoking the endpoint
					_, err := svcClient.ChangePassword(ctx, &authngrpc.ChangePasswordRequest{
						CurrentPassword: "wrong",
						Password:        "0a0b0c0d0",
					})
					require.Error(t, err)

					st := status.Convert(err)
					assert.Equal(t, codes.FailedPrecondition, st.Code())
					fes := errors.ToFieldErrors(st.Details()[0].(*errdetails.BadRequest))
					assert.EqualValues(t, services.FieldErrors{{"credentials", "FAILED"}}, fes)
				})
				t.Run("Failure - Invalid Session", func(t *testing.T) {
					ctx := metadata.AppendToOutgoingContext(context.Background(), testApp.Config.SessionCookieName, "invalid")
					_, err := svcClient.ChangePassword(ctx, &authngrpc.ChangePasswordRequest{
						CurrentPassword: "oldpwd",
						Password:        "0a0b0c0d0",
					})
					require.Error(t, err)

					st := status.Convert(err)
					assert.Equal(t, codes.Unauthenticated, st.Code())
				})
				t.Run("Successful - Valid Token & Session", func(t *testing.T) {
					// Token account
					tokenAccount, password := createUser(t, testApp)

					token, err := resets.New(testApp.Config, tokenAccount.ID, tokenAccount.PasswordChangedAt)
					require.NoError(t, err)
					tokenStr, err := token.Sign(testApp.Config.ResetSigningKey)
					require.NoError(t, err)

					// given another account
					sessionAccount, _ := createUser(t, testApp)
					require.NoError(t, err)
					// with a session
					session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, sessionAccount.ID)

					ctx := metadata.AppendToOutgoingContext(context.Background(), testApp.Config.SessionCookieName, session.Value)
					var header metadata.MD

					// invoking the endpoint
					res, err := svcClient.ChangePassword(ctx, &authngrpc.ChangePasswordRequest{
						Token:           tokenStr,
						CurrentPassword: password,
						Password:        "0a0b0c0d0",
					}, grpc.Header(&header))
					require.NoError(t, err)

					// works
					assertSuccessfulGRPCSession(t, testApp, res.GetResult().GetIdToken(), header, tokenAccount)
					assertChangedPassword(t, testApp, tokenAccount)
				})
			})
		})
		t.Run("Password Reset", func(t *testing.T) {
			svcClient := authngrpc.NewPasswordResetServiceClient(client)
			t.Run("Known Account", func(t *testing.T) {
				acc, _ := createUser(t, testApp)

				_, err := svcClient.RequestPasswordReset(context.Background(), &authngrpc.PasswordResetRequest{
					Username: acc.Username,
				})
				assert.NoError(t, err)
			})
			t.Run("Unknown Account", func(t *testing.T) {
				_, err := svcClient.RequestPasswordReset(context.Background(), &authngrpc.PasswordResetRequest{
					Username: generateUsername(),
				})
				assert.NoError(t, err)
			})
		})
		t.Run("Passwordless", func(t *testing.T) {
			svcClient := authngrpc.NewPasswordlessServiceClient(client)
			t.Run("Request", func(t *testing.T) {
				t.Run("Known Account", func(t *testing.T) {
					acc, _ := createUser(t, testApp)

					_, err := svcClient.RequestPasswordlessLogin(context.Background(), &authngrpc.RequestPasswordlessLoginRequest{
						Username: acc.Username,
					})
					require.NoError(t, err)
				})
				t.Run("Unknown Account", func(t *testing.T) {
					_, err := svcClient.RequestPasswordlessLogin(context.Background(), &authngrpc.RequestPasswordlessLoginRequest{
						Username: generateUsername(),
					})
					require.NoError(t, err)
				})
			})
			t.Run("Submit", func(t *testing.T) {
				t.Run("Successful - Valid Token", func(t *testing.T) {
					acc, _ := createUser(t, testApp)

					// given a passwordless token
					token, err := passwordless.New(testApp.Config, acc.ID)
					require.NoError(t, err)
					tokenStr, err := token.Sign(testApp.Config.PasswordlessTokenSigningKey)
					require.NoError(t, err)

					var header metadata.MD
					// invoking the endpoint
					res, err := svcClient.SubmitPasswordlessLogin(context.Background(), &authngrpc.SubmitPasswordlessLoginRequest{
						Token: tokenStr,
					}, grpc.Header(&header))
					require.NoError(t, err)

					// works
					assertSuccessfulGRPCSession(t, testApp, res.GetResult().GetIdToken(), header, acc)
				})
				t.Run("Failure - Invalid Token", func(t *testing.T) {
					// invoking the endpoint
					_, err := svcClient.SubmitPasswordlessLogin(context.Background(), &authngrpc.SubmitPasswordlessLoginRequest{
						Token: "invalid",
					})
					assert.Error(t, err)

					st := status.Convert(err)

					// does not work
					assert.Equal(t, codes.FailedPrecondition, st.Code())

					fes := errors.ToFieldErrors(st.Details()[0].(*errdetails.BadRequest))
					assert.EqualValues(t, services.FieldErrors{{"token", "INVALID_OR_EXPIRED"}}, fes)
				})
				t.Run("Successful - Valid Session", func(t *testing.T) {
					acc, _ := createUser(t, testApp)

					// given a session
					session := test.CreateSession(testApp.RefreshTokenStore, testApp.Config, acc.ID)

					// given a passwordless token
					token, err := passwordless.New(testApp.Config, acc.ID)
					require.NoError(t, err)
					tokenStr, err := token.Sign(testApp.Config.PasswordlessTokenSigningKey)
					require.NoError(t, err)

					ctx := metadata.AppendToOutgoingContext(context.Background(), testApp.Config.SessionCookieName, session.Value)
					var header metadata.MD

					// invoking the endpoint
					res, err := svcClient.SubmitPasswordlessLogin(ctx, &authngrpc.SubmitPasswordlessLoginRequest{
						Token: tokenStr,
					}, grpc.Header(&header))
					require.NoError(t, err)

					// works
					assertSuccessfulGRPCSession(t, testApp, res.GetResult().GetIdToken(), header, acc)

					// invalidates old session
					claims, err := sessions.Parse(session.Value, testApp.Config)
					require.NoError(t, err)
					id, err := testApp.RefreshTokenStore.Find(models.RefreshToken(claims.Subject))
					require.NoError(t, err)
					assert.Empty(t, id)
				})
			})
		})
	}
}

func assertSuccessfulSession(t *testing.T, testApp *app.App, res *http.Response, account *models.Account) {
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	test.AssertSession(t, testApp.Config, res.Cookies())
	test.AssertIDTokenResponse(t, res, testApp.KeyStore, testApp.Config)
}

func assertSuccessfulGRPCSession(t *testing.T, testApp *app.App, idToken string, header metadata.MD, account *models.Account) {
	test.AssertGRPCSession(t, testApp.Config, header)
	test.AssertGRPCIDTokenResponse(t, idToken, testApp.KeyStore, testApp.Config)
}

func assertChangedPassword(t *testing.T, testApp *app.App, account *models.Account) {
	found, err := testApp.AccountStore.Find(account.ID)
	require.NoError(t, err)
	assert.NotEqual(t, found.Password, account.Password)
}

const letters = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`

// generateUsername returns a username of 5 random letters appended with `@example.com`
func generateUsername() string {
	u := strings.Builder{}
	for i := 0; i < 5; i++ {
		u.WriteByte(letters[rand.Intn(len(letters))])
	}
	return fmt.Sprintf("%s@example.com", u.String())
}

// createUser creates a user and returns the resulting account and plaintext password
func createUser(t *testing.T, app *app.App) (*models.Account, string) {
	username := generateUsername()
	password := "aa11bb!cc"
	b, _ := bcrypt.GenerateFromPassword([]byte(password), app.Config.BcryptCost)
	acc, err := app.AccountStore.Create(username, b)
	require.NoError(t, err)
	return acc, password
}
