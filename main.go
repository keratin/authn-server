package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
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

	cfg, err := config.ReadEnv()
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

func serve(cfg *config.Config) {
	app, err := api.NewApp(cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(fmt.Sprintf("~*~ Keratin AuthN v%s ~*~", VERSION))
	fmt.Println(fmt.Sprintf("AUTHN_URL: %s", cfg.AuthNURL))
	fmt.Println(fmt.Sprintf("PORT: %d", cfg.ServerPort))

	if cfg.PublicPort != 0 {
		go func() {
			fmt.Println(fmt.Sprintf("PUBLIC_PORT: %d", cfg.PublicPort))
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.PublicPort), publicRouter(app)))
		}()
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.ServerPort), router(app)))
}

func migrate(cfg *config.Config) {
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
