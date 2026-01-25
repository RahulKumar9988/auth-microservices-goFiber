package security

import "github.com/gofiber/fiber/v2"

func CSRF() fiber.Handler {
	return func(c *fiber.Ctx) error {

		switch c.Method() {
		case fiber.MethodGet, fiber.MethodHead, fiber.MethodOptions:
			return c.Next()
		}

		csrfCookie := c.Cookies("csrf_token")
		csrfHeader := c.Get("csrf_token")

		if csrfCookie == "" || csrfHeader == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "CSRF token missing",
			})
		}

		if csrfCookie != csrfHeader {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "CSRF token mismatch",
			})
		}

		return c.Next()

	}
}
