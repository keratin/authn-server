package main

import (
	"log"
	"net/http"

	"github.com/keratin/authn/handlers"
)

func main() {
	// set up connections and configuration
	app, err := handlers.NewApp()
	if err != nil {
		panic(err)
	}

	stack := routing(app)

	log.Fatal(http.ListenAndServe(":8000", stack))
}
