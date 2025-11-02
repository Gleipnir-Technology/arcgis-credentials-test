package main

import (
	"log"
	"net/http"
)

func getDashboard(w http.ResponseWriter, r *http.Request) {
	username := sessionManager.GetString(r.Context(), "username")
	if username == "" {
		log.Println("Redirecting from dashboard since we don't have a username in this session")
		http.Redirect(w, r, BaseURL+"/", http.StatusFound)
		return
	}
	token, ok := TokenDatabase[username]
	if !ok {
		log.Printf("Redirecting from dashboard since we don't have a session for '%s'\n", username)
		http.Redirect(w, r, BaseURL+"/", http.StatusFound)
		return
	}
	tryPortal(token.AccessToken)
	search, err := findFieldseeker(token.AccessToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Search: %s", search)

	err = htmlDashboard(w, r.URL.Path, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "image/x-icon")

	http.ServeFile(w, r, "favicon.ico")
}

func getOAuthBegin(w http.ResponseWriter, r *http.Request) {
	log.Println("Getting ArcGIS login")

	expiration := 60
	authURL := buildArcGISAuthURL(ClientID, redirectURL(), expiration)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func getOAuthCallback(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling oauth callback")
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Access code is empty", http.StatusBadRequest)
		return
	}
	log.Printf("Got oauth access code '%s'. Getting an access token", code)
	token, err := handleAccessCode(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sessionManager.Put(r.Context(), "username", token.Username)
	http.Redirect(w, r, BaseURL+"/dashboard", http.StatusFound)
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
