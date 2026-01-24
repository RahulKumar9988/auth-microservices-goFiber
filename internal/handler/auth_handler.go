package handler

import (
	"errors"
	"net/mail"
	"os"

	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/services"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(asv *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: asv}
}

func isProd() bool {
	return os.Getenv("ENV") == "production"
}

type userRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
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

	err := h.authService.Register(req.Email, req.Password, req.Role)

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

	ip := c.IP()
	ua := c.Get("User-Agent")

	tokens, err := h.authService.Login(req.Email, req.Password, ip, ua)

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
	ip := c.IP()
	ua := c.Get("User-Agent")

	if refreshToken == "" {
		return c.Status(401).JSON(fiber.Map{"error": "refresh token missing"})
	}

	tokens, err := h.authService.Refresh(refreshToken, ip, ua)
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

	ip := c.IP()
	ua := c.Get("User-Agent")

	_ = h.authService.Logout(refreshToken, ip, ua)

	// Refresh token
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/auth",
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   isProd(),
		SameSite: fiber.CookieSameSiteLaxMode,
	})

	// CSRF token
	c.Cookie(&fiber.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   isProd(),
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteStrictMode,
	})

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

func (h *AuthHandler) LogoutSession(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	sessionID := c.Params("sessionID")
	ip := c.IP()
	ua := c.Get("User-Agent")

	if sessionID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "sessionID missing",
		})
	}

	err := h.authService.LogoutSession(userID, sessionID, ip, ua)
	if err != nil {
		return c.Status(403).JSON(fiber.Map{
			"error": "cannot revoke session",
		})
	}

	return c.JSON(fiber.Map{
		"message": "session logout successfully",
	})
}

func (h *AuthHandler) LogoutAllSession(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	ip := c.IP()
	ua := c.Get("User-Agent")

	refreshToken := c.Cookies("refresh_token")

	if err := h.authService.LogoutAllSessions(
		userID,
		refreshToken,
		ip,
		ua,
	); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "failed to logout sessions",
		})
	}

	// Refresh token
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/auth",
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   isProd(),
		SameSite: fiber.CookieSameSiteLaxMode,
	})

	// CSRF token
	c.Cookie(&fiber.Cookie{
		Name:     "csrf_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   isProd(),
		SameSite: fiber.CookieSameSiteStrictMode,
	})

	return c.Status(200).JSON(fiber.Map{
		"message": "logout all sessions",
	})
}

func (h *AuthHandler) PasswordReset(c *fiber.Ctx) error {

}
