package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"mc_api/internal/config"
	postgressql "mc_api/internal/data_provider"
	"mc_api/internal/domain/service/auth_service"
	"mc_api/internal/domain/service/product_service"
	"mc_api/internal/domain/service/rank_service"

	pb "mc_api/pkg/api"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/i-b8o/logging"
	postgresql "github.com/i-b8o/postgresql_client"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type App struct {
	cfg        *config.Config
	grpcServer *grpc.Server
	httpServer *runtime.ServeMux
	logger     logging.Logger
	pgClient   *pgxpool.Pool
}

func NewApp(config *config.Config) (App, error) {
	logger := logging.GetLogger(config.AppConfig.LogLevel)

	logger.Print("Postgres initializing")
	fmt.Printf("%s %s %s %s %s", config.PostgreSQL.Username, config.PostgreSQL.Password,
		config.PostgreSQL.Host, config.PostgreSQL.Port, config.PostgreSQL.Database)
	pgConfig := postgresql.NewPgConfig(
		config.PostgreSQL.Username, config.PostgreSQL.Password,
		config.PostgreSQL.Host, config.PostgreSQL.Port, config.PostgreSQL.Database,
	)

	pgClient, err := postgresql.NewClient(context.Background(), 5, time.Second*5, pgConfig)
	if err != nil {
		logger.Fatal(err)
	}

	// Data Providers
	authDataProvider := postgressql.NewAuthStorage(pgClient)
	phraseDataProvider := postgressql.NewPhraseStorage(pgClient)
	productDataProvider := postgressql.NewProductStorage(pgClient)

	// Services
	authService := auth_service.NewAuthService(authDataProvider, logger)
	rankingService := rank_service.NewRankingService(phraseDataProvider, logger)
	productService := product_service.NewProductService(productDataProvider, logger)

	// Servers
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(unaryInterceptor))
	httpServer := runtime.NewServeMux()

	pb.RegisterAuthServiceHandlerServer(context.Background(), httpServer, authService)
	pb.RegisterAuthServiceServer(grpcServer, authService)

	pb.RegisterRankServiceHandlerServer(context.Background(), httpServer, rankingService)
	pb.RegisterRankServiceServer(grpcServer, rankingService)

	pb.RegisterProductServiceHandlerServer(context.Background(), httpServer, productService)
	pb.RegisterProductServiceServer(grpcServer, productService)

	return App{cfg: config, grpcServer: grpcServer, logger: logger, pgClient: pgClient, httpServer: httpServer}, nil
}

func (a *App) Run(ctx context.Context) error {
	grp, _ := errgroup.WithContext(ctx)
	grp.Go(func() error {
		httpAddress := fmt.Sprintf("%s:%d", a.cfg.HTTP.IP, a.cfg.HTTP.Port)
		a.logger.Printf("started http server on %s", httpAddress)
		return http.ListenAndServe(httpAddress, a.httpServer)
	})

	grp.Go(func() error {
		grpcAddress := fmt.Sprintf("%s:%d", a.cfg.GRPC.IP, a.cfg.GRPC.Port)
		listener, err := net.Listen("tcp", grpcAddress)
		if err != nil {
			return err
		}
		a.logger.Printf("started grpc server on %s", grpcAddress)
		return a.grpcServer.Serve(listener)
	})

	return grp.Wait()

}

func unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	log.Println("--> unary interceptor: ", info.FullMethod)
	return handler(ctx, req)
}
