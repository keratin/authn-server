package main

import (
	"fmt"
	"os"
	"path"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data"
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
		os.Stderr.WriteString("unexpected invocation\n")
		usage()
		os.Exit(2)
	}
}

func serve(cfg *app.Config) {
	fmt.Printf("~*~ Keratin AuthN v%s ~*~\n", VERSION)

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

	fmt.Printf("AUTHN_URL: %s\n", cfg.AuthNURL)
	fmt.Printf("PORT: %d\n", cfg.ServerPort)
	if app.Config.PublicPort != 0 {
		fmt.Printf("PUBLIC_PORT: %d\n", app.Config.PublicPort)
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
	fmt.Printf(`
Usage:
%s server  - run the server (default)
%s migrate - run migrations

`, exe, exe)
}
