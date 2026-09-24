package middlewares

import (
	authv1 "gateway/gen/auth/v1"
	"gateway/internal/cache"
	"gateway/internal/common"
	"gateway/internal/configs"
	"gateway/internal/domain"
	grpc_client "gateway/internal/grpc"
	"net/http"
	"strings"

	"github.com/ajaka/nexus-shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuthenticatePrivateRoutes(env *configs.Env, r *cache.Redis, gClient grpc_client.GrpcClient) gin.HandlerFunc {

	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		if checkIfPathIsAllowed(c.Request.URL.Path) {
			c.Next()
			return
		}
		id, err := c.Cookie("sessionId")
		if err != nil {
			logger.Warn("Request sessionId is either missing or invalid")
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
			c.Abort()
			return
		}
		authResponse, err := gClient.Client.AuthenticateUser(c.Request.Context(), &authv1.AuthenticateUserRequest{
			SessionID: id,
		})
		if err != nil {
			logger.Error("Could not verify user authentication", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			c.Abort()
			return
		}
		if !authResponse.Allow {
			logger.Warn("Request participant is not authorized")
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
			c.Abort()
			return
		}
		userId, err := uuid.Parse(authResponse.User.UserId)
		if err != nil {
			logger.Error("Could not standardize returned id back to uuid", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			c.Abort()
			return
		}
		user := domain.MinimalUserStruct{
			UserId: userId,
			Email:  authResponse.User.Email,
		}
		logger.Info("Successfully authenticated the request")
		stringifiedUser, err := utils.Stringify(user)

		c.Request.Header.Set("X-REQUEST-IDENTITY", stringifiedUser)
		c.Set("user", &user)
		c.Next()
	}
}

func checkIfPathIsAllowed(path string) bool {
	allowedPaths := []string{
		"/api/auth",
	}

	for _, allowedPath := range allowedPaths {
		if strings.HasPrefix(path, allowedPath) {
			return true
		}
	}
	return false
}
