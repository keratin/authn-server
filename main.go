package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/keratin/authn-server/api"
	"github.com/keratin/authn-server/config"
	"github.com/keratin/authn-server/data"
	"github.com/namsral/flag"
)

func main() {
	port := flag.Int("port", 0, "Optional port for server. Default value is from AUTHN_URL.")
	flag.Parse()

	args := flag.Args()

	var cmd string
	if len(args) == 0 {
		cmd = "server"
	} else {
		cmd = args[0]
	}

	if cmd == "server" {
		serve(port)
	} else if cmd == "migrate" {
		migrate()
	} else {
		os.Stderr.WriteString(fmt.Sprintf("unexpected invocation\n"))
		usage()
		os.Exit(2)
	}
}

func serve(port *int) {
	cfg, err := config.ReadFlags()
	if err != nil {
		panic(err)
	}
	// set up connections and configuration
	app, err := api.NewApp(cfg)
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
	cfg, err := config.ReadFlags()
	if err != nil {
		panic(err)
	}
	fmt.Println("Running migrations.")
	err = data.MigrateDB(cfg.DatabaseURL)
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
