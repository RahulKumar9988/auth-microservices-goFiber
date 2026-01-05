package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Register(app *fiber.App, db *gorm.DB) {
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
}
