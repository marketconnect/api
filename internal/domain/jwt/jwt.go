package mc_jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO remove to config
var (
	sampleSecretKey = []byte("SecretYouShouldHide")
)

func CreateToken(id uint64) (string, error) {
	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"exp": time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	})
	return token.SignedString([]byte(sampleSecretKey)) // Replace with your own secret key
}

func GetIdFromToken(tokenString string) (uint64, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return sampleSecretKey, nil
	})
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, "invalid token 1")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, status.Error(codes.Unauthenticated, "invalid token")
	}
	id, ok := claims["id"].(float64)
	if !ok {
		return 0, status.Error(codes.Unknown, "could not parse user ID from token")
	}

	return uint64(id), nil
}
