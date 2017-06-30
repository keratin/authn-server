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

func main() {
	if len(os.Args) == 1 {
		serve()
	} else if len(os.Args) == 2 && os.Args[1] == "migrate" {
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

	stack := routing(app)

	fmt.Println(fmt.Sprintf("~*~ Keratin AuthN server is ready on %s ~*~", app.Config.AuthNURL))
	log.Fatal(http.ListenAndServe(":8000", stack))
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
%s         - run the server
%s migrate - run migrations
`, exe, exe))
}
