package main

import (
	"net/http"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	err := htmlRoot(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
