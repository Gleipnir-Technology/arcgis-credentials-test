package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	client_id := os.Getenv("CLIENT_ID")
	if client_id == "" {
		log.Println("You must specify a non-empty CLIENT_ID")
		os.Exit(1)
	}
	client_secret := os.Getenv("CLIENT_SECRET")
	if client_secret == "" {
		log.Println("You must specify a non-empty CLIENT_SECRET")
		os.Exit(1)
	}

	log.Println("Starting...")
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	log.Println("Serving on :9001")
	http.ListenAndServe(":9001", r)
}
