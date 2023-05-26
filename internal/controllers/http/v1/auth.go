package v1

import (
	"context"
	"encoding/json"

	"net/http"

	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-playground/validator/v10"

	"api/internal/domain/entity"

	"github.com/julienschmidt/httprouter"
)

type AuthResponse struct {
	Token string `json:"token"`
}

var validate *validator.Validate

const (
	login    = "/login"
	register = "/register"
)

type AuthUsecase interface {
	Login(ctx context.Context, user entity.User) (uint64, error)
	Register(ctx context.Context, user entity.User) (uint64, error)
}

type authHandler struct {
	authUsecase AuthUsecase
}

func NewAuthHandler(authUsecase AuthUsecase) *authHandler {
	validate = validator.New()
	return &authHandler{authUsecase: authUsecase}
}

func (h *authHandler) Register(router *httprouter.Router) {
	router.POST(register, h.RegisterUser)
	router.POST(login, h.Login)
}

func (h *authHandler) RegisterUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var user entity.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Validation
	err = validate.Struct(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Register
	id, err := h.authUsecase.Register(context.Background(), user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"exp": time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	})
	tokenString, err := token.SignedString([]byte("YOUR_SECRET_KEY")) // Replace with your own secret key
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send JWT token as a response
	resp := AuthResponse{tokenString}
	json.NewEncoder(w).Encode(resp)
}

func (h *authHandler) Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var user entity.User
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validation
	err = validate.Struct(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	login := user.Username
	password := user.Password

	id, err := h.authUsecase.Login(r.Context(), entity.User{Username: login, Password: password})
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
	}

	// Create a JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"exp": time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	})
	tokenString, err := token.SignedString([]byte("YOUR_SECRET_KEY"))
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	// Send the token as a response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}
