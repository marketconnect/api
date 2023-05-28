package auth_service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	pb "mc_api/pkg/api"

	"github.com/golang-jwt/jwt"
	"github.com/i-b8o/logging"
	"github.com/jackc/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO remove to config
var (
	sampleSecretKey = []byte("SecretYouShouldHide")
)

type AuthStorage interface {
	RegisterUser(ctx context.Context, email, password string) (uint64, error)
	LoginUser(ctx context.Context, email, password string) (uint64, error)
}

type AuthService struct {
	storage AuthStorage
	logging logging.Logger
	pb.UnimplementedAPIServer
}

func NewAuthService(storage AuthStorage, logging logging.Logger) *AuthService {
	return &AuthService{
		storage: storage,
		logging: logging,
	}
}

func (s *AuthService) RegisterUser(ctx context.Context, user *pb.User) (*pb.TokenMessage, error) {
	// // Anything linked to this variable will fetch request headers.
	// md, ok := metadata.FromIncomingContext(ctx)

	// if !ok {
	// 	return nil, status.Errorf(codes.DataLoss, "failed to get metadata")
	// }
	// // authorization := md["Authorization"][0]
	// fmt.Printf("Authorization: %v", md["authorization"][0])
	err := user.Validate()
	if err != nil {
		return nil, err
	}
	email := user.GetEmail()
	pswd := user.GetPassword()
	id, err := s.storage.RegisterUser(ctx, email, pswd)
	if err != nil {
		return nil, errHandler(err)
	}
	token, err := createToken(id)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")

	}

	return &pb.TokenMessage{Token: token}, nil
}

func (s *AuthService) LoginUser(ctx context.Context, user *pb.User) (*pb.TokenMessage, error) {
	email := user.GetEmail()
	pswd := user.GetPassword()
	id, err := s.storage.LoginUser(ctx, email, pswd)
	if err != nil {
		s.logging.Errorf("Error login: %v", err)
		return nil, err
	}
	token, err := createToken(id)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &pb.TokenMessage{Token: token}, nil
}

func errHandler(err error) error {
	fmt.Println(err)
	if err != sql.ErrNoRows {
		return status.Error(codes.AlreadyExists, "no rowxcxbbvvxcbcvs in result set")
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		fmt.Println(pgErr.Code)
		if pgErr.Code == "23505" {
			return status.Error(codes.AlreadyExists, "email already exists")
		}

	}
	return err
}

func createToken(id uint64) (string, error) {
	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"exp": time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	})
	return token.SignedString([]byte("YOUR_SECRET_KEY")) // Replace with your own secret key
}

// func verifyJWT(endpointHandler func(writer http.ResponseWriter, request *http.Request)) http.HandlerFunc {
// 	return httprouter.Handle(func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 		authHeader := r.Header.Get("Authorization")
// 		fmt.Printf("Authorization %s", authHeader)
// 		if authHeader != "" {
// 			if strings.HasPrefix(authHeader, "Bearer") {
// 				tokenString := strings.TrimPrefix(authHeader, "Bearer ")
// 				token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
// 						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
// 					}
// 					return []byte("YOUR_SECRET_KEY"), nil
// 				})
// 				if err != nil {
// 					http.Error(w, err.Error(), http.StatusUnauthorized)
// 					return
// 				}
// 				if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
// 					ctx := context.WithValue(r.Context(), "username", claims["username"])
// 					next.ServeHTTP(w, r.WithContext(ctx))
// 				} else {
// 					http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
// 				}
// 			} else {
// 				http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
// 			}
// 		} else {
// 			http.Error(w, "Authorization header required", http.StatusUnauthorized)
// 		}
// 	})

// }
