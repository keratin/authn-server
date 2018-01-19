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

var VERSION string

func main() {
	var cmd string
	if len(os.Args) == 1 {
		cmd = "server"
	} else {
		cmd = os.Args[1]
	}

	if cmd == "server" {
		serve()
	} else if cmd == "migrate" {
		migrate()
	} else {
		os.Stderr.WriteString(fmt.Sprintf("unexpected invocation\n"))
		usage()
		os.Exit(2)
	}
}

func serve() {
	// set up connections and configuration
	app, err := api.NewApp()
	if err != nil {
		panic(err)
	}

	if app.Config.PublicPort != 0 {
		go func() {
			fmt.Println(fmt.Sprintf("~*~ Keratin AuthN server v%s is listening to public routes on %s (%d) ~*~", VERSION, app.Config.AuthNURL, app.Config.PublicPort))
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", app.Config.PublicPort), publicRouter(app)))
		}()
	}

	fmt.Println(fmt.Sprintf("~*~ Keratin AuthN server v%s is listening on %s (%d) ~*~", VERSION, app.Config.AuthNURL, app.Config.ServerPort))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", app.Config.ServerPort), router(app)))
}

func migrate() {
	cfg := config.ReadEnv()
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
