package handlers

import (
	"auth/internal/common"
	"auth/internal/configs"
	"auth/internal/models"
	"auth/internal/store"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleLogout(env *configs.Env, s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		session := c.GetString("sessionId")
		if session == "" {
			logger.Error("Could not retrieve sessionId from context")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}
		user, ok := common.GetFromContext[*models.MinimalUserStruct](c, "user")
		if !ok {
			logger.Error("Could not fetch user data from context")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}
		if err := common.HandleLogoutActivity(c, s, env, session, user); err != nil {
			logger.Error("Failed to log out user", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		logger.Info("Successfully logged user out")
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Logout successful"})
	}
}
