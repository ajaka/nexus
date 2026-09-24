package routes

import (
	"auth/internal/cache"
	"auth/internal/configs"
	"auth/internal/handlers"
	"auth/internal/middlewares"
	"auth/internal/repositories"

	"github.com/gin-gonic/gin"
)

func MountRoutes(rg *gin.RouterGroup, repo *repositories.Repository, c *cache.Cache, env *configs.Env) {
	rg.POST("/register", middlewares.ValidateRegisterRequest(), handlers.HandleRegister(repo, env))
	rg.POST("/login", middlewares.ValidateLoginRequest(), handlers.HandleLogin(repo, c, env))
	rg.POST("/password/forgot", middlewares.ValidateForgotPasswordRequest(), handlers.HandleForgotPassword(repo, env))
	rg.GET("/password/reset", handlers.HandleVerifyPasswordReset(env, c))
	rg.POST("/password/reset", middlewares.ValidateResetPasswordRequest(), handlers.HandleResetPassword(repo, env))
	rg.GET("/verify", handlers.HandleVerifyUser(repo, env, c))
	rg.POST("/logout", middlewares.AuthenticationMiddlewares(c), handlers.HandleLogout(env, c))
}
