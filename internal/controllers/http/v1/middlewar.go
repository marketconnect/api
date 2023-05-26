package v1

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/julienschmidt/httprouter"
)

func handlerFunc(fn func(http.ResponseWriter, *http.Request, httprouter.Params)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, nil)
	}
}

func jwtMiddleware(next http.HandlerFunc) httprouter.Handle {
	return httprouter.Handle(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		authHeader := r.Header.Get("Authorization")
		fmt.Printf("Authorization %s", authHeader)
		if authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer") {
				tokenString := strings.TrimPrefix(authHeader, "Bearer ")
				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					}
					return []byte("YOUR_SECRET_KEY"), nil
				})
				if err != nil {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
				if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
					ctx := context.WithValue(r.Context(), "username", claims["username"])
					next.ServeHTTP(w, r.WithContext(ctx))
				} else {
					http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
				}
			} else {
				http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
			}
		} else {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
		}
	})
}
