package handlers

import (
	"auth/internal/common"
	"auth/internal/models"
	"auth/internal/store"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleRevokeSession(s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		user, ok := common.GetFromContext[*models.MinimalUserStruct](c, "user")
		if !ok {
			logger.Error("Could not retrieve user data from context")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}
		req, ok := common.GetFromContext[*models.RevokeSession](c, "request")
		if !ok {
			logger.Error("Could not retrieve req body from context")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		if err := s.RevokeSession(c.Request.Context(), req.SessionID, user.UserId); err != nil {
			logger.Error("Could not revoke user session", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}
		logger.Info("Successfully revoked user session")
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Session Revoked successfully"})
	}
}

func HandleRevokeAllOtherSessions(s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		user, ok := common.GetFromContext[*models.MinimalUserStruct](c, "user")
		if !ok {
			logger.Error("Could not fetch user data from context")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}
		session := c.GetString("sessionId")
		if session == "" {
			logger.Error("Could not retrieve sessionId from context")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}
		if err := s.RevokeAllOtherSessions(c.Request.Context(), session, user.UserId); err != nil {
			logger.Error("Could not revoke user sessions", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}
		logger.Info("Successfully revoked user sessions")
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Session Revoked successfully"})
	}
}
