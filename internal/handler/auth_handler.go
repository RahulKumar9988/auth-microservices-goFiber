package handler

import (
	"errors"
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

type userRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req userRequest

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

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req userRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	user, err := h.authService.Login(req.Email, req.Password)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidInput):
			return c.Status(400).JSON(fiber.Map{"error": "invalid input"})
		case errors.Is(err, services.ErrInvalidCredentials):
			return c.Status(401).JSON(fiber.Map{"error": "invalid email or password"})
		default:
			return c.Status(500).JSON(fiber.Map{"error": "internal server error"})
		}
	}

	return c.Status(200).JSON(fiber.Map{
		"userId": user.ID,
		"email":  user.Email,
		"role":   user.Role,
	})
}

func (h *AuthHandler) UserList(c *fiber.Ctx) error {
	users, err := h.authService.GetAllUsers()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to retrieve users"})
	}

	return c.JSON(fiber.Map{
		"users": users,
	})
}
