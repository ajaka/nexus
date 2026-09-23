package internal

import (
	authv1 "auth/gen/auth/v1"
	"auth/internal/cache"
	"auth/internal/configs"
	"auth/internal/grpc_server"
	"auth/internal/middlewares"
	"auth/internal/repositories"
	"auth/internal/routes"
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

func Listen() error {

	ctx := context.Background()
	logger := slog.Default()

	env := configs.LoadEnv(logger)
	pool := configs.ConnectDB(ctx, logger, env.DATABASE_URL)

	repo := repositories.InitRepository(pool)
	cache := cache.Initcache(ctx, env, logger)

	authServer := grpc_server.BuildAuthGrpcServer(repo)

	// Initialize kafka outbox
	// configs.InitializeKafkaOutbox(ctx, repo, env, logger)

	lis, err := net.Listen("tcp", env.SERVER_ADDR)
	if err != nil {
		logger.Error("Failed tp bind to port", "error", err)
		os.Exit(1)
	}
	mux := cmux.New(lis)

	grpcListener := mux.Match(cmux.HTTP2())

	restListener := mux.Match(cmux.HTTP1Fast())

	grpcServer := grpc.NewServer()
	authv1.RegisterAuthServiceServer(grpcServer, authServer)

	router := gin.New()
	router.Use(gin.Recovery())
	router.SetTrustedProxies(nil)

	router.Use(middlewares.AttachScopedLogger(env))

	g := router.Group("/auth")

	routes.MountRoutes(g, repo, cache, env)

	restHandler := http.Server{
		Handler:      router,
		TLSNextProto: nil,
	}

	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			logger.Error("An error occured while serving the grpc listener")
			os.Exit(1)
		}

	}()
	go func() {
		if err := restHandler.Serve(restListener); err != nil {
			logger.Error("An error occured while serving the rest listener")
			os.Exit(1)
		}
	}()

	logger.Info(fmt.Sprintf("Listening to both rest and grpc on port %s", env.SERVER_ADDR))
	return mux.Serve()
	// return router.Run(env.SERVER_ADDR)
}
