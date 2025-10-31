package main

import (
	"net/http"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
