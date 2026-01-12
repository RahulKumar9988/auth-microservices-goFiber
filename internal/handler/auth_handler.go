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

	tokens, err := h.authService.Login(req.Email, req.Password)

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
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"tokenType":    "Bearer",
		"expiresIn":    tokens.ExpiresIn,
	})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	tokens, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid refresh token"})
	}

	return c.JSON(fiber.Map{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
	})

}

func (h *AuthHandler) UserList(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	email := c.Locals("email").(string)
	role := c.Locals("role").(string)

	return c.JSON(fiber.Map{
		"requested_by": fiber.Map{
			"user_id": userID,
			"email":   email,
			"role":    role,
		},
		"users": "list here",
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.BodyParser(&req); err != nil || req.RefreshToken == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "refresh token missing",
		})
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "invalid refresh token",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "user Loged out successfully",
	})

}
