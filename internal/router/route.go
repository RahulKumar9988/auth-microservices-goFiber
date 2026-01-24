package router

import (
	"time"

	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/config"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/handler"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/middlewares/security"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/repositories"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/services"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Register(app *fiber.App, db *gorm.DB, jwtCfg config.JWTConfig, sessionRepo *repositories.SessionRepository, rateLimiter *security.Ratelimiter, auditRepo *repositories.AuditRepo) {
	app.Get("/health", func(c *fiber.Ctx) error {
		sqlDB, _ := db.DB()
		if err := sqlDB.Ping(); err != nil {
			return c.Status(503).JSON(fiber.Map{
				"status": "db-down",
			})
		}
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	userRepo := repositories.NewUserRepository(db)
	userService := services.NewAuthService(userRepo, jwtCfg, sessionRepo, auditRepo)
	authHandler := handler.NewAuthHandler(userService)

	auth := app.Group("/auth")

	auth.Post("/register", rateLimiter.Limit("register", 5, time.Minute, func(ip, ua string) {
		auditRepo.Log("REGISTER_RATE_LIMIT", nil, ip, ua)
	}), authHandler.Register)
	auth.Post("/login", rateLimiter.Limit("login", 5, time.Minute, func(ip, ua string) {
		auditRepo.Log("RATE_LIMIT_HIT", nil, ip, ua)
	}), authHandler.Login)
	auth.Post("/refresh", rateLimiter.Limit("refresh", 5, time.Minute, func(ip, ua string) {
		auditRepo.Log("REFRESH_RATE_LIMIT", nil, ip, ua)
	}), authHandler.Refresh)
	auth.Post("/logout", authHandler.Logout)

	protected := auth.Group("/", security.JWT(jwtCfg.AccessSecret))
	protected.Get("/userlist", authHandler.UserList)
	protected.Get("/sessions", authHandler.ListSessions)
	protected.Delete("/sessions/:sessionID", authHandler.LogoutSession)
	protected.Post("/logout-all", authHandler.LogoutAllSession)
	protected.Patch("/reset-password", authHandler.PasswordReset)

	admin := protected.Group("/admin", security.RequiredRole("admin"))
	admin.Get("/adminlist", authHandler.AdminUserList)

}
