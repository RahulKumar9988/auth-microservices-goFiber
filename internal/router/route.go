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

	auth.Post("/register", rateLimiter.Limit("register", 5, time.Minute), authHandler.Register)
	auth.Post("/login", rateLimiter.Limit("register", 5, time.Minute), authHandler.Login)
	auth.Post("/refresh", rateLimiter.Limit("register", 5, time.Minute), authHandler.Refresh)
	auth.Post("/logout", authHandler.Logout)

	protected := auth.Group("/", security.JWT(jwtCfg.AccessSecret))
	protected.Get("/userlist", authHandler.UserList)
	protected.Get("/sessions", authHandler.ListSessions)

	admin := protected.Group("/admin", security.RequiredRole("admin"))
	admin.Get("/adminlist", authHandler.AdminUserList)

}
