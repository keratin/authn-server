package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/keratin/authn-server/app"
)

func Server(app *app.App) {
	if app.Config.PublicPort != 0 {
		go func() {
			fmt.Printf("PUBLIC_PORT: %d\n", app.Config.PublicPort)
			publicServer := &http.Server{
				Addr:              fmt.Sprintf(":%d", app.Config.PublicPort),
				Handler:           PublicRouter(app),
				ReadHeaderTimeout: 10 * time.Second,
			}
			log.Fatal(publicServer.ListenAndServe())
		}()
	}

	privateServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.Config.ServerPort),
		Handler:           Router(app),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Fatal(privateServer.ListenAndServe())
}
