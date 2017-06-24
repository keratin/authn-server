package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/keratin/authn-server/handlers"
)

func main() {
	// set up connections and configuration
	app, err := handlers.NewApp()
	if err != nil {
		panic(err)
	}

	stack := routing(app)

	fmt.Println("~*~ Keratin AuthN server is ready ~*~")
	log.Fatal(http.ListenAndServe(":8000", stack))
}
