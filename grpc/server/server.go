package server

import (
	"context"
	"crypto/tls"

	"google.golang.org/grpc/credentials"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"

	"net"

	"github.com/gorilla/mux"
	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/grpc/internal/errors"
	"github.com/keratin/authn-server/grpc/private"
	"github.com/keratin/authn-server/grpc/public"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func init() {
	runtime.HTTPError = errors.CustomHTTPError
}

// RunPrivateService starts a gRPC server for the private API and accompanying gRPC-Gateway server
func RunPrivateService(ctx context.Context, app *app.App, grpcListener net.Listener, httpListener net.Listener) error {

	privateRouter := mux.NewRouter()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return private.RunPrivateGRPC(ctx, app, grpcListener)
	})

	connCreds := grpc.WithInsecure()
	if app.Config.ClientCA != nil {
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{app.Config.Certificate},
			ClientCAs:          app.Config.ClientCA,
			ClientAuth:         tls.RequireAndVerifyClientCert,
			InsecureSkipVerify: app.Config.TLSSkipVerify,
		}
		connCreds = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	}

	privClientConn, err := grpc.DialContext(ctx, grpcListener.Addr().String(), connCreds)
	if err != nil {
		log.Fatal(err)
	}

	g.Go(func() error {
		return private.RunPrivateGateway(ctx, app, privateRouter, privClientConn, httpListener)
	})

	return g.Wait()
}

// RunPublicService starts a gRPC server for the public API and accompanying gRPC-Gateway server
func RunPublicService(ctx context.Context, app *app.App, grpcListener net.Listener, httpListener net.Listener) error {

	publicRouter := mux.NewRouter()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return public.RunPublicGRPC(ctx, app, grpcListener)
	})

	connCreds := grpc.WithInsecure()
	if app.Config.ClientCA != nil {
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{app.Config.Certificate},
			ClientCAs:          app.Config.ClientCA,
			ClientAuth:         tls.RequireAndVerifyClientCert,
			InsecureSkipVerify: app.Config.TLSSkipVerify,
		}
		connCreds = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	}

	clientConn, err := grpc.DialContext(ctx, grpcListener.Addr().String(), connCreds)
	if err != nil {
		log.Fatal(err)
	}

	g.Go(func() error {
		return public.RunPublicGateway(ctx, app, publicRouter, clientConn, httpListener)
	})

	return g.Wait()
}
