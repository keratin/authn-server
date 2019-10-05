package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data"
	grpcserver "github.com/keratin/authn-server/grpc/server"
	"github.com/keratin/authn-server/server"
	"github.com/sirupsen/logrus"
)

// VERSION is a value injected at build time with ldflags
var VERSION string

func main() {
	var cmd string
	if len(os.Args) == 1 {
		cmd = "server"
	} else {
		cmd = os.Args[1]
	}

	cfg, err := app.ReadEnv()
	if err != nil {
		fmt.Println(err)
		fmt.Println("\nsee: https://github.com/keratin/authn-server/blob/master/docs/config.md")
		return
	}

	if cmd == "server" {
		serve(cfg)
	} else if cmd == "migrate" {
		migrate(cfg)
	} else {
		os.Stderr.WriteString(fmt.Sprintf("unexpected invocation\n"))
		usage()
		os.Exit(2)
	}
}

func serve(cfg *app.Config) {
	fmt.Println(fmt.Sprintf("~*~ Keratin AuthN v%s ~*~", VERSION))

	// Default logger
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}
	logger.Level = logrus.DebugLevel
	logger.Out = os.Stdout

	app, err := app.NewApp(cfg, logger)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(fmt.Sprintf("AUTHN_URL: %s", cfg.AuthNURL))
	fmt.Println(fmt.Sprintf("PORT: %d", cfg.ServerPort))
	if app.Config.PublicPort != 0 {
		fmt.Println(fmt.Sprintf("PUBLIC_PORT: %d", app.Config.PublicPort))
	}

	if cfg.EnableGRPC {
		ctx, cancel := context.WithCancel(context.Background())
		gracefulStop := make(chan os.Signal)
		signal.Notify(gracefulStop, os.Interrupt, os.Kill)
		go func() {
			grpcserver.Server(ctx, app)
		}()
		<-gracefulStop
		log.Println("kill signal received")
		log.Println("gracefully stopping")
		cancel()
		return
	}
	server.Server(app)
}

func migrate(cfg *app.Config) {
	fmt.Println("Running migrations.")
	err := data.MigrateDB(cfg.DatabaseURL)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Migrations complete.")
	}
}

func usage() {
	exe := path.Base(os.Args[0])
	fmt.Println(fmt.Sprintf(`
Usage:
%s server  - run the server (default)
%s migrate - run migrations
`, exe, exe))
}
