package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/config"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/middlewares/security"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/models"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

// Custom errors
var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type AuthService struct {
	userRepo  *repositories.UserRepository
	jwtCfg    config.JWTConfig
	tokenRepo *repositories.RefreshTokenRepository
}

func NewAuthService(repo *repositories.UserRepository, jwtCfg config.JWTConfig, tokenRepo *repositories.RefreshTokenRepository) *AuthService {
	return &AuthService{
		userRepo:  repo,
		jwtCfg:    jwtCfg,
		tokenRepo: tokenRepo,
	}
}

func (s *AuthService) Register(email string, password string) error {
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" || password == "" {
		return ErrInvalidInput
	}

	existing, err := s.userRepo.FindByEmail(email)

	if err != nil {
		return err
	}

	if existing != nil {
		return errors.New("user already existed")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	user := &models.UserModel{
		Email:    email,
		Password: string(hash),
		Role:     models.User,
	}

	return s.userRepo.Create(user)
}

func (s *AuthService) Login(email string, password string) (*TokenPair, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" || password == "" {
		return nil, ErrInvalidInput
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := security.GenerateAccessToken(
		user.ID,
		user.Email,
		string(user.Role),
		s.jwtCfg.AccessSecret,
		s.jwtCfg.AccessTTL,
	)

	if err != nil {
		return nil, err
	}

	refreshToken, err := security.GenerateRefreshToken(
		user.ID,
		s.jwtCfg.RefreshSecret,
		s.jwtCfg.RefreshTTL,
	)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.tokenRepo.Store(
		ctx,
		refreshToken,
		user.ID,
		s.jwtCfg.RefreshTTL,
	); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtCfg.AccessTTL.Seconds()),
	}, nil
}

func (s *AuthService) GetAllUsers() ([]models.UserModel, error) {
	return s.userRepo.GetAllUsers()
}

func (s *AuthService) Refresh(refreshToken string) (*TokenPair, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Check Redis
	userID, err := s.tokenRepo.Get(ctx, refreshToken)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Invalidate old token (ROTATION)
	_ = s.tokenRepo.Delete(ctx, refreshToken)

	// Generate new tokens
	accessToken, err := security.GenerateAccessToken(
		userID,
		"", "", // email & role optional here (can fetch if needed)
		s.jwtCfg.AccessSecret,
		s.jwtCfg.AccessTTL,
	)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := security.GenerateRefreshToken(
		userID,
		s.jwtCfg.RefreshSecret,
		s.jwtCfg.RefreshTTL,
	)
	if err != nil {
		return nil, err
	}

	// Store new refresh token
	if err := s.tokenRepo.Store(
		ctx,
		newRefreshToken,
		userID,
		s.jwtCfg.RefreshTTL,
	); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) Logout(refeshToken string) error {
	ctx := context.Background()

	return s.tokenRepo.Delete(ctx, refeshToken)
}
