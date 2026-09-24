package handlers

import (
	"auth/internal/cache"
	"auth/internal/common"
	"auth/internal/configs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleLogout(env *configs.Env, cc *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)

		if err := common.HandleLogoutActivity(c, cc, env); err != nil {
			logger.Error("Failed to log out user", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		logger.Info("Successfully logged user out")
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Logout successful"})
	}
}
