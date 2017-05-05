package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/keratin/authn/handlers"
)

func StubHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("not implemented"))
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", StubHandler).Methods("GET")

	r.HandleFunc("/accounts", StubHandler).Methods("POST")
	r.HandleFunc("/accounts/import", StubHandler).Methods("POST")
	r.HandleFunc("/accounts/available", StubHandler).Methods("GET")

	r.HandleFunc("/accounts/{account_id:[0-9]+}", StubHandler).Methods("DELETE")
	r.HandleFunc("/accounts/{account_id:[0-9]+}/lock", StubHandler).Methods("PUT", "PATCH")
	r.HandleFunc("/accounts/{account_id:[0-9]+}/unlock", StubHandler).Methods("PUT", "PATCH")

	r.HandleFunc("/session", StubHandler).Methods("POST")
	r.HandleFunc("/session", StubHandler).Methods("DELETE")
	r.HandleFunc("/session/refresh", StubHandler).Methods("GET")

	r.HandleFunc("/password", StubHandler).Methods("POST")
	r.HandleFunc("/password/reset", StubHandler).Methods("GET")

	r.HandleFunc("/configuration", StubHandler).Methods("GET")
	r.HandleFunc("/jwks", StubHandler).Methods("GET")

	r.HandleFunc("/stats", StubHandler).Methods("GET")
	r.HandleFunc("/health", handlers.Health).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", r))
}
