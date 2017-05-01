package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func HealthHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("up"))
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/health", HealthHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", r))
}
