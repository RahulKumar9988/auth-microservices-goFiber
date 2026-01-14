package security

import "github.com/gofiber/fiber/v2"

func RequiredRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")

		if role == "" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "role not found",
			})
		}

		for _, r := range roles {
			if role == r {
				return c.Next()
			}
		}

		return c.Status(401).JSON(fiber.Map{
			"message": "forbidden",
		})
	}
}
