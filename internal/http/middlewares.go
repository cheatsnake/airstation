package http

import (
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func (s *Server) jwtAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("jwt")
		if err != nil {
			jsonUnauthorized(w, "Unauthorized, access denied.")
			return
		}

		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])

			}
			return []byte(s.config.JWTSign), nil
		})

		if err != nil || !token.Valid {
			jsonUnauthorized(w, "Invalid token.")
			return
		}

		next.ServeHTTP(w, r)
	})
}
