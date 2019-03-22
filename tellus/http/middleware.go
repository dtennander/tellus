package http

import (
	"log"
	"net/http"
)

func withLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Logged connection from %s to %s", r.RemoteAddr, r.URL)
		next.ServeHTTP(w, r)
	}
}
