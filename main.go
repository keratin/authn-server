package main

import (
	"fmt"
	"os"
	"path"

	"github.com/keratin/authn-server/app"
	"github.com/keratin/authn-server/app/data"
	grpcserver "github.com/keratin/authn-server/grpc/server"
	"github.com/keratin/authn-server/server"
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
	app, err := app.NewApp(cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(fmt.Sprintf("~*~ Keratin AuthN v%s ~*~", VERSION))
	fmt.Println(fmt.Sprintf("AUTHN_URL: %s", cfg.AuthNURL))
	fmt.Println(fmt.Sprintf("PORT: %d", cfg.ServerPort))
	if app.Config.PublicPort != 0 {
		fmt.Println(fmt.Sprintf("PUBLIC_PORT: %d", app.Config.PublicPort))
	}

	if cfg.EnableGRPC {
		grpcserver.Server(app)
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
