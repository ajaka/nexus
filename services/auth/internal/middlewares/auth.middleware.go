package middlewares

import (
	"auth/internal/cache"
	"auth/internal/common"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthenticationMiddlewares(ca *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)

		sessionId, err := c.Cookie("sessionId")
		if err != nil || sessionId == "" {
			logger.Warn("Request sessionId is either missing or invalid")
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
			c.Abort()
			return
		}
		user, loggedIn := ca.GetUser(c.Request.Context(), sessionId)
		if !loggedIn {
			logger.Warn("Request participant is not logged in")
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("sessionId", sessionId)
		c.Next()
	}
}
