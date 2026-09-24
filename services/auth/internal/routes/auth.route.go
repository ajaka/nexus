package routes

import (
	"auth/internal/cache"
	"auth/internal/configs"
	"auth/internal/handlers"
	"auth/internal/middlewares"
	"auth/internal/repositories"
	"auth/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/ua-parser/uap-go/uaparser"
)

func MountRoutes(rg *gin.RouterGroup, repo *repositories.Repository, c *cache.Cache, s *store.Store, env *configs.Env, p *uaparser.Parser) {
	rg.POST("/register", middlewares.ValidateRegisterRequest(), handlers.HandleRegister(repo, env))
	rg.POST("/login", middlewares.ValidateLoginRequest(), handlers.HandleLogin(repo, s, env, p))
	rg.POST("/password/forgot", middlewares.ValidateForgotPasswordRequest(), handlers.HandleForgotPassword(repo, env))
	rg.GET("/password/reset", handlers.HandleVerifyPasswordReset(env, c))
	rg.POST("/password/reset", middlewares.ValidateResetPasswordRequest(), handlers.HandleResetPassword(repo, env))
	rg.GET("/verify", handlers.HandleVerifyUser(repo, env, c))
	rg.Use(middlewares.AuthenticationMiddlewares(c))
	rg.POST("/logout", handlers.HandleLogout(env, s))
	rg.POST("/session", handlers.HandleRevokeAllOtherSessions(s))
	rg.POST("/session/revoke", middlewares.ValidateSessionRevocation(), handlers.HandleRevokeSession(s))
}
