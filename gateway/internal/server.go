package internal

import (
	"context"
	authv1 "gateway/gen/auth/v1"
	"gateway/internal/cache"
	"gateway/internal/configs"
	grpc_client "gateway/internal/grpc"
	"gateway/internal/middlewares"
	"gateway/internal/routes"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Listen() error {
	// Initialize context
	ctx := context.Background()

	// Initialize background loggers
	logger := slog.Default()

	conn, err := grpc.NewClient("localhost:5000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Failed to initialize grpc client", "error", err)
		os.Exit(1)
	}
	gatewayClient := authv1.NewAuthServiceClient(conn)

	gClientStruct := grpc_client.GrpcClient{
		Client: gatewayClient,
	}
	// Load configurations
	env := configs.LoadEnv(logger)
	cfg := configs.LoadServiceConfig(logger)

	// connect to services
	c := cache.InitializeRedis(ctx, env, logger)

	router := gin.New()
	router.Use(gin.Recovery())
	router.SetTrustedProxies(nil)

	if env.PRODUCTION {
		gin.SetMode(gin.ReleaseMode)
	}

	// General middlewares
	router.Use(middlewares.GenerateRequestID())
	router.Use(middlewares.AttachScopedLogger(env))
	router.Use(middlewares.RateLimit(cfg, c))
	router.Use(middlewares.AuthenticatePrivateRoutes(env, c, gClientStruct))

	router.Use(routes.Proxy(cfg))

	return router.Run(env.SERVER_ADDR)
}
