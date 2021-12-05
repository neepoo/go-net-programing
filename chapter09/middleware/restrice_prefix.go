package middleware

import (
	"net/http"
	"path"
	"strings"
)

func RestrictPrefix(prefix string, next http.Handler) http.Handler {
	return http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, p := range strings.Split(path.Clean(r.URL.Path), "/") {
			if strings.HasSuffix(p, prefix) {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
		}
		next.ServeHTTP(w, r)
	}))
}