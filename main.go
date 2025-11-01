package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var sessionManager *scs.SessionManager

var BaseURL, ClientID, ClientSecret string

func main() {
	BaseURL = os.Getenv("BASE_URL")
	if BaseURL == "" {
		log.Println("You must specify a non-empty BASE_URL")
		os.Exit(1)
	}
	ClientID = os.Getenv("CLIENT_ID")
	if ClientID == "" {
		log.Println("You must specify a non-empty CLIENT_ID")
		os.Exit(1)
	}
	ClientSecret = os.Getenv("CLIENT_SECRET")
	if ClientSecret == "" {
		log.Println("You must specify a non-empty CLIENT_SECRET")
		os.Exit(1)
	}

	log.Println("Starting...")
	go loadBabbler()
	initTokenDatabase()
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(sessionManager.LoadAndSave)

	r.Get("/", getRoot)
	r.Post("/authenticate", postAuthenticate)
	r.Get("/babble/*", handleBabbleRequest)
	r.Get("/dashboard", getDashboard)
	r.Get("/favicon.ico", getFavicon)
	r.Post("/login", postAuthenticate)
	r.Get("/oauth-begin", getOAuthBegin)
	r.Get("/oauth-callback", getOAuthCallback)
	log.Println("Serving on :9001")
	http.ListenAndServe(":9001", r)
}
