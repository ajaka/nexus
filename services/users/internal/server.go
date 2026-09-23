package internal

import (
	"log/slog"
	"net/http"
	"os"
	authv1 "users/gen/auth/v1"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Listen() error {
	logger := slog.Default()
	conn, err := grpc.NewClient("localhost:5000", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		logger.Info("Could not initialize grpc client", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	authClient := authv1.NewAuthServiceClient(conn)

	dStruct := &DemoStruct{UserClient: authClient}

	router := gin.New()
	router.Use(gin.Recovery())
	router.SetTrustedProxies(nil)

	router.GET("/users/stuff", DemoHandler(dStruct, logger))
	return router.Run(":6000")
}

type DemoStruct struct {
	UserClient authv1.AuthServiceClient
}

type DemoRequest struct {
	Id    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

func DemoHandler(demoStruct *DemoStruct, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		var d DemoRequest

		if err := c.ShouldBindJSON(&d); err != nil {
			logger.Error("Failed to bind data", "err", err)
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Bad request"})
			return
		}

		res, err := demoStruct.UserClient.VerifyUserExists(c.Request.Context(), &authv1.VerifyUserExistsRequest{
			Email: d.Email,
			Id:    d.Id.String(),
		})
		if err != nil {
			logger.Error("Auth returned err", "err", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "it worked. exists or not, it did", "value": res.Exists})
	}
}
