package app

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"api/internal/config"
	v1 "api/internal/controllers/http/v1"
	"api/internal/domain/service"

	"github.com/dgrijalva/jwt-go"
	"github.com/i-b8o/logging"
	"github.com/julienschmidt/httprouter"
	pb "github.com/marketconnect/contracts/pb/auth/v1"
	"github.com/rs/cors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	cfg        *config.Config
	router     *httprouter.Router
	httpServer *http.Server

	logger logging.Logger
}

func NewApp(ctx context.Context, config *config.Config) (App, error) {
	logger := logging.GetLogger(config.AppConfig.LogLevel)
	logger.Print("router initializing")
	router := httprouter.New()
	creds := insecure.NewCredentials()
	conn, err := grpc.Dial(
		fmt.Sprintf("%s:%d", config.Service.IP, config.Service.PORT),
		grpc.WithTransportCredentials(creds),
	)

	if err != nil {
		return App{}, err
	}
	authClient := pb.NewAuthServiceClient(conn)

	// Service
	authUsecase := service.NewAuthUsecase(authClient)

	// Handler
	authHandler := v1.NewAuthHandler(authUsecase)
	rankingHandler := v1.NewRankingHandler()

	// Register
	authHandler.Register(router)
	rankingHandler.Register(router)

	return App{cfg: config, logger: logger, router: router}, nil

}

func (a *App) Run(ctx context.Context) error {
	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		return a.startHTTP(ctx)
	})
	// redirect
	// grp.Go(func() error {
	// 	return http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 		http.Redirect(w, r, fmt.Sprintf("https://%s:443", a.cfg.HTTP.IP)+r.RequestURI, http.StatusMovedPermanently)
	// 	}))
	// })
	return grp.Wait()
}

func (a *App) startHTTP(ctx context.Context) error {

	// Define the listener (Unix or TCP)
	// var listener net.Listener

	a.logger.Infof("bind application to host: %s and port: %d", a.cfg.HTTP.IP, a.cfg.HTTP.Port)
	var err error
	// start up a tcp listener
	// listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", a.cfg.HTTP.IP, a.cfg.HTTP.Port))
	if err != nil {
		a.logger.Fatal(err)
	}

	// create a new Cors handler
	c := cors.New(cors.Options{
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	})

	// apply the CORS specification on the request, and add relevant CORS headers
	handler := c.Handler(a.router)

	// define parameters for an HTTP server
	a.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", a.cfg.HTTP.Port),
		Handler:      handler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	a.logger.Println("application initialized and started")

	// accept incoming connections on the listener, creating a new service goroutine for each

	err = http.ListenAndServe(fmt.Sprintf(":%d", a.cfg.HTTP.Port), a.router)
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}

	// if err := a.httpServer.ListenAndServeTLS(curdir+"/.certs/read-only.crt", curdir+"/.certs/read-only.key"); err != nil {
	// 	switch {
	// 	case errors.Is(err, http.ErrServerClosed):
	// 		a.logger.Warn("server shutdown")

	// 	default:
	// 		a.logger.Fatal(err)
	// 	}
	// }
	err = a.httpServer.Shutdown(context.Background())
	if err != nil {
		a.logger.Fatal(err)
	}
	return nil
}

func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString := strings.Split(authHeader, " ")[1]
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
				ctx := context.WithValue(r.Context(), "user", claims["user"])
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
			}
		} else {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
		}
	})
}
