package handlers

import (
	"auth/internal/common"
	"auth/internal/configs"
	"auth/internal/errs"
	"auth/internal/models"
	"auth/internal/repositories"
	"auth/internal/store"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ua-parser/uap-go/uaparser"
)

func HandleLogin(repo *repositories.Repository, s *store.Store, env *configs.Env, parser *uaparser.Parser) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		request, exists := common.GetFromContext[models.LoginRequest](c, "loginRequest")
		if !exists {
			logger.Error("Login request was not found in context")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		ua := c.GetHeader("user-agent")

		clientDevice := parser.Parse(ua)

		NewSession := models.Sessions{
			Brand:     clientDevice.Device.Brand,
			Browser:   clientDevice.UserAgent.Family,
			Os:        clientDevice.Os.Family,
			Ip:        c.ClientIP(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		user, err := repo.GetUserByEmail(c.Request.Context(), request.Email)
		if err != nil {
			if errors.Is(err, errs.ERR_EMAIL_NO_EXISTS) {
				logger.Warn("Login attempted with an email that does not exist", "email", request.Email)
				c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid credentials"})
				return
			}
			logger.Error("Failed to retrieve user for login", "email", request.Email, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		if !user.Verified || !user.Active {
			logger.Warn("Login attempted with an unverified or inactive account", "email", request.Email)
			c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "Please verify your email or contact support"})
			return
		}

		matches, err := common.VerifyPassword(request.Password, user.Password)
		if err != nil {
			logger.Error("Failed to verify login password", "email", request.Email, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}
		if !matches {
			logger.Warn("Login request provided an incorrect password", "email", request.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid credentials"})
			return
		}

		payload := models.MinimalUserStruct{
			UserId: user.Id,
			Email:  user.Email,
		}
		NewSession.UserId = user.Id
		if err := common.HandleLoginActivity(
			c,
			&payload,
			s,
			&NewSession,
			env.PRODUCTION,
		); err != nil {
			logger.Error("Failed to create login session", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		logger.Info("Successfully authenticated user", "email", request.Email)
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Login successful"})
	}
}
