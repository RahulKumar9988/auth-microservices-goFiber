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

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Path:     "/auth",
		MaxAge:   int(tokens.RefreshTTL.Seconds()),
	})

	csrfToken, _ := services.GenerateCSRFToken()
	c.Cookie(&fiber.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		HTTPOnly: false,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(tokens.RefreshTTL.Seconds()),
	})

	return c.Status(200).JSON(fiber.Map{
		"accessToken": tokens.AccessToken,
		"csrfToken":   csrfToken,
		"tokenType":   "Bearer",
		"expiresIn":   tokens.ExpiresIn,
	})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")

	if refreshToken == "" {
		return c.Status(401).JSON(fiber.Map{"error": "refresh token missing"})
	}

	tokens, err := h.authService.Refresh(refreshToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid refresh token"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
		Path:     "/auth",
		MaxAge:   int(tokens.RefreshTTL.Seconds()),
	})

	return c.JSON(fiber.Map{
		"access_token": tokens.AccessToken,
		// "refresh_token": tokens.RefreshToken,
		"expires_in": tokens.ExpiresIn,
	})

}

func (h *AuthHandler) UserList(c *fiber.Ctx) error {
	users, err := h.authService.GetAllUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to fetch admin users",
		})
	}

	return c.JSON(fiber.Map{
		"admin_lists": users,
	})
}

func (h *AuthHandler) AdminUserList(c *fiber.Ctx) error {
	admins, err := h.authService.GetAllAdmins()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to fetch admin users",
		})
	}

	return c.JSON(fiber.Map{
		"admin_lists": admins,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {

	refreshToken := c.Cookies("refresh_token")

	if refreshToken == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "refresh token missing",
		})
	}

	_ = h.authService.Logout(refreshToken)

	c.ClearCookie("refresh_token")
	c.ClearCookie("csrf_tokrn")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "user Loged out successfully",
	})

}

func (h *AuthHandler) ListSessions(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)

	sessions, err := h.authService.ListSessions(userID)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to fetch session",
		})
	}

	return c.JSON(fiber.Map{
		"sessions": sessions,
	})
}
