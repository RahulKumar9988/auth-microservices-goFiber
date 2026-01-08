package handler

import (
	"net/mail"

	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/services"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(asv *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: asv}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req registerRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid body parsed",
		})
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid email",
		})
	}

	err := h.authService.Register(req.Email, req.Password)

	if err != nil {
		return c.Status(405).JSON(fiber.Map{
			"error": err,
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "User Registered",
	})

}
