package router

import (
	"context"

	"appsechub/internal/application/dto"
	"appsechub/internal/config"
	"appsechub/internal/infras/ratelimit"
	"appsechub/internal/interfaces/http/handler"
	"appsechub/internal/interfaces/http/middleware"
	"appsechub/pkg/logger"

	"github.com/gin-gonic/gin"
)

func registerAuthRoutes(v1 *gin.RouterGroup, userHandler *handler.UserHandler, cfg *config.Config, authMiddleware ...gin.HandlerFunc) {
	auth := v1.Group("/auth")
	auth.POST("/register", middleware.ValidateJSON[dto.CreateUserRequest]("req", cfg.HTTP.MaxBodyBytes), userHandler.Register)

	if cfg.HTTP.LoginRateLimitRPS > 0 && cfg.HTTP.LoginRateLimitBurst > 0 {
		if cfg.Env == "prod" {
			logger.L().Warn("login_rate_limit_in_memory", "note", "in-memory limiter is per-instance; consider Redis for multi-instance", "env", cfg.Env)
		}
		login := auth.Group("")
		// Per-IP limiter at group-level (does not need DTO)
		login.Use(middleware.RateLimitForPath("/v1/auth/login", cfg.HTTP.LoginRateLimitRPS, cfg.HTTP.LoginRateLimitBurst))
		if cfg.RedisAddr != "" {
			rl := ratelimit.NewRedisLimiter(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB).WithFailClosed(cfg.HTTP.LoginRateLimitFailClosed)
			extract := func(c *gin.Context) string {
				v, exists := c.Get("req")
				if !exists {
					return ""
				}
				if req, ok := v.(dto.LoginRequest); ok {
					return req.Email
				}
				return ""
			}
			// Order matters: bind/validate JSON first → extract email → apply per-email limiter → handler
			login.POST("/login",
				middleware.ValidateJSON[dto.LoginRequest]("req", cfg.HTTP.MaxBodyBytes),
				rl.LimitEmail(cfg.HTTP.LoginRateLimitRPS, cfg.HTTP.LoginRateLimitBurst, extract),
				userHandler.Login,
			)
		} else {
			login.POST("/login", middleware.ValidateJSON[dto.LoginRequest]("req", cfg.HTTP.MaxBodyBytes), userHandler.Login)
		}
	} else {
		auth.POST("/login", middleware.ValidateJSON[dto.LoginRequest]("req", cfg.HTTP.MaxBodyBytes), userHandler.Login)
	}

	if cfg.Security.RefreshEnabled {
		auth.POST("/refresh", userHandler.Refresh)
		auth.POST("/logout", userHandler.Logout)
	}

	// Entra ID (Azure AD) OIDC login endpoints (optional wiring if configured)
	if cfg.Entra.ClientID != "" && cfg.Entra.RedirectURL != "" && cfg.Entra.TenantID != "" {
		eh, err := handler.NewEntraHandler(context.Background(), cfg, userHandler.UC(), nil)
		if err != nil {
			logger.L().Warn("entra_handler_init_failed", "error", err)
		} else {
			auth.GET("/entra/login", eh.Login)
			auth.GET("/entra/callback", eh.Callback)
		}
	}

	if len(authMiddleware) > 0 {
		protected := auth.Group("")
		protected.Use(authMiddleware...)
		protected.GET("/me", userHandler.GetMe)
		protected.POST("/change-password", middleware.ValidateJSON[dto.ChangePasswordRequest]("req", cfg.HTTP.MaxBodyBytes), userHandler.ChangePassword)
		return
	}
	auth.GET("/me", userHandler.GetMe)
	auth.POST("/change-password", middleware.ValidateJSON[dto.ChangePasswordRequest]("req", cfg.HTTP.MaxBodyBytes), userHandler.ChangePassword)
}
