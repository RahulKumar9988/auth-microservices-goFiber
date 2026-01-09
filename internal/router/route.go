package router

import (
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/config"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/handler"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/repositories"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/services"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Register(app *fiber.App, db *gorm.DB, jwtCfg config.JWTConfig) {
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
	userService := services.NewAuthService(userRepo, jwtCfg)
	authHandler := handler.NewAuthHandler(userService)

	app.Post("/auth/register", authHandler.Register)
	app.Post("/auth/login", authHandler.Login)
	app.Get("/auth/userlist", authHandler.UserList)

}
