package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
)

func main() {
	port := flag.Int("port", 0, "Optional port for server. Default value is from AUTHN_URL.")
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		serve(port)
	} else if len(args) == 1 && args[0] == "migrate" {
		migrate()
	} else {
		os.Stderr.WriteString(fmt.Sprintf("unexpected invocation\n"))
		usage()
		os.Exit(2)
	}
}

func serve(port *int) {
	// set up connections and configuration
	app, err := api.NewApp()
	if err != nil {
		panic(err)
	}

	var listen string
	if port != nil && *port > 0 {
		listen = strconv.Itoa(*port)
	} else {
		listen = app.Config.AuthNURL.Port()
	}

	fmt.Println(fmt.Sprintf("~*~ Keratin AuthN server is ready on %s ~*~", app.Config.AuthNURL))
	log.Fatal(http.ListenAndServe(":"+listen, router(app)))
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
