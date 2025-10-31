package main

import (
	"log"
	"net/http"
)

func getFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "image/x-icon")

	http.ServeFile(w, r, "favicon.ico")
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	err := htmlRoot(w, r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func postAuthenticate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	log.Printf("Doing login with username '%s' and password '%s'\n", username, password)
}
