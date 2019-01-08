package server

import (
	"fmt"
	"github.com/keratin/authn-server/app"
	"log"
	"net/http"
)

func Server(app *app.App) {
	if app.Config.PublicPort != 0 {
		go func() {
			fmt.Println(fmt.Sprintf("PUBLIC_PORT: %d", app.Config.PublicPort))
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", app.Config.PublicPort), PublicRouter(app)))
		}()
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", app.Config.ServerPort), Router(app)))
}
