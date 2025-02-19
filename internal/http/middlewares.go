package http

import (
	"net/http"
)

func (s *Server) adminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		if token == "" {
			jsonUnauthorized(w, "Authorization header with token is required.")
			return
		}

		if token != s.config.SecretKey {
			jsonUnauthorized(w, "Invalid authorization token.")
			return
		}

		next.ServeHTTP(w, r)
	})
}
