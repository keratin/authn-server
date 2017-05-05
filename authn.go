package main

import (
	"log"
	"net/http"
	"os"

	gorilla "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/keratin/authn/handlers"
)

func stubHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("not implemented"))
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", stubHandler).Methods("GET")

	r.HandleFunc("/accounts", stubHandler).Methods("POST")
	r.HandleFunc("/accounts/import", stubHandler).Methods("POST")
	r.HandleFunc("/accounts/available", stubHandler).Methods("GET")

	r.HandleFunc("/accounts/{account_id:[0-9]+}", stubHandler).Methods("DELETE")
	r.HandleFunc("/accounts/{account_id:[0-9]+}/lock", stubHandler).Methods("PUT", "PATCH")
	r.HandleFunc("/accounts/{account_id:[0-9]+}/unlock", stubHandler).Methods("PUT", "PATCH")

	r.HandleFunc("/session", stubHandler).Methods("POST")
	r.HandleFunc("/session", stubHandler).Methods("DELETE")
	r.HandleFunc("/session/refresh", stubHandler).Methods("GET")

	r.HandleFunc("/password", stubHandler).Methods("POST")
	r.HandleFunc("/password/reset", stubHandler).Methods("GET")

	r.HandleFunc("/configuration", stubHandler).Methods("GET")
	r.HandleFunc("/jwks", stubHandler).Methods("GET")

	r.HandleFunc("/stats", stubHandler).Methods("GET")
	r.HandleFunc("/health", handlers.Health).Methods("GET")

	app := gorilla.RecoveryHandler()(gorilla.CombinedLoggingHandler(os.Stdout, r))

	log.Fatal(http.ListenAndServe(":8000", app))
}
