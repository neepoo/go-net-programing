package middleware

import (
	"log"
	"net/http"
	"time"
)

func MiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodTrace {
			// But in some cases
			//that may not be proper, and the middleware should block the next handler
			//and respond to the client itself
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}

		w.Header().Set("X-Content-Type-Options", "nosniff")

		start := time.Now()
		// In most cases, middleware calls the given handler
		next.ServeHTTP(w, r)
		log.Printf("Mext handler duration %v", time.Now().Sub(start))
	})
}
