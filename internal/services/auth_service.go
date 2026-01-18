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
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Custom errors
var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

const (
	maxLoginAttemts = 5
	failWindow      = 10 * time.Minute
	locakDuration   = 15 * time.Minute
)

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	RefreshTTL   time.Duration
}

type AuthService struct {
	userRepo *repositories.UserRepository
	jwtCfg   config.JWTConfig
	//tokenRepo *repositories.RefreshTokenRepository
	auditRepo   *repositories.AuditRepo
	sessionRepo *repositories.SessionRepository
}

func NewAuthService(repo *repositories.UserRepository, jwtCfg config.JWTConfig, sessionRepo *repositories.SessionRepository, auditRepo *repositories.AuditRepo) *AuthService {
	return &AuthService{
		userRepo:    repo,
		jwtCfg:      jwtCfg,
		sessionRepo: sessionRepo,
		auditRepo:   auditRepo,
	}
}

func (s *AuthService) Register(email string, password string, role string) error {
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
		Role:     models.UserRole(role),
	}

	return s.userRepo.Create(user)
}

func (s *AuthService) Login(email, password, ip, ua string) (*TokenPair, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" || password == "" {
		return nil, ErrInvalidInput
	}

	ctx := context.Background()

	if s.IsLocking(ctx, email) {
		s.auditRepo.Log("ACCOUNT_LOCKED", nil, ip, ua)
		return nil, errors.New("account temporarily locked")
	}

	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		s.auditRepo.Log("LOGIN_FAILED", nil, ip, ua)
		return nil, err
	}

	if user == nil {
		s.auditRepo.Log("LOGIN_FAILED", nil, ip, ua)
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		_ = s.RecordFailedLogin(ctx, email)
		s.auditRepo.Log(
			"LOGIN_FAILED",
			nil,
			ip,
			ua,
		)
		return nil, ErrInvalidCredentials
	}

	s.ClearFailLogin(ctx, email)
	s.auditRepo.Log(
		"LOGIN_SUCCESS",
		&user.ID,
		ip,
		ua,
	)

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

	sessionID := uuid.NewString()

	refreshToken, err := security.GenerateRefreshToken(
		user.ID,
		sessionID,
		s.jwtCfg.RefreshSecret,
		s.jwtCfg.RefreshTTL,
	)

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := s.sessionRepo.Create(
		ctx,
		sessionID,
		user.ID,
		s.jwtCfg.RefreshTTL,
	); err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtCfg.AccessTTL.Seconds()),
		RefreshTTL:   s.jwtCfg.RefreshTTL,
	}, nil
}

func (s *AuthService) GetAllUsers() ([]models.UserModel, error) {
	return s.userRepo.GetAllUsers()
}

func (s *AuthService) GetAllAdmins() ([]models.UserModel, error) {
	return s.userRepo.GetAllAdmins()
}

func (s *AuthService) Refresh(refreshToken string, ip, ua string) (*TokenPair, error) {

	claims, err := security.ParseRefreshToken(
		refreshToken,
		s.jwtCfg.RefreshSecret,
	)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	ctx := context.Background()

	userID, err := s.sessionRepo.GetUserID(ctx, claims.SessionID)
	if err != nil {
		s.auditRepo.Log("REFRESH_TOKEN_REUSE_DETECTED", &claims.UserID, ip, ua)
		return nil, ErrInvalidCredentials
	}

	// rotate session
	_ = s.sessionRepo.Delete(ctx, claims.SessionID, userID)

	newSessionID := uuid.NewString()

	newRefreshToken, err := security.GenerateRefreshToken(
		userID,
		newSessionID,
		s.jwtCfg.RefreshSecret,
		s.jwtCfg.RefreshTTL,
	)
	if err != nil {
		return nil, err
	}

	_ = s.sessionRepo.Create(
		ctx,
		newSessionID,
		userID,
		s.jwtCfg.RefreshTTL,
	)

	accessToken, err := security.GenerateAccessToken(
		userID,
		"", "",
		s.jwtCfg.AccessSecret,
		s.jwtCfg.AccessTTL,
	)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.jwtCfg.AccessTTL.Seconds()),
		RefreshTTL:   s.jwtCfg.RefreshTTL,
	}, nil
}

func (s *AuthService) Logout(refreshToken, ip, ua string) error {
	claims, err := security.ParseRefreshToken(
		refreshToken,
		s.jwtCfg.RefreshSecret,
	)
	if err != nil {
		return ErrInvalidCredentials
	}
	s.auditRepo.Log(
		"LOGOUT",
		&claims.UserID,
		ip,
		ua,
	)
	return s.sessionRepo.Delete(
		context.Background(),
		claims.SessionID,
		claims.UserID,
	)
}

func (s *AuthService) ListSessions(userID uint) ([]repositories.SessionInfo, error) {
	return s.sessionRepo.ListByUsers(context.Background(), userID)
}

func (s *AuthService) IsLocking(ctx context.Context, email string) bool {
	key := "login_lock:" + email
	exists, _ := s.sessionRepo.Redis().Exists(ctx, key).Result()
	return exists == 1
}

func (s *AuthService) RecordFailedLogin(ctx context.Context, email string) error {
	failKey := "login_fail:" + email
	lockKey := "login_lock:" + email

	count, err := s.sessionRepo.Redis().Incr(ctx, failKey).Result()
	if err != nil {
		return err
	}

	if count == 1 {
		s.sessionRepo.Redis().Expire(ctx, failKey, failWindow)
	}
	if count >= maxLoginAttemts {
		s.sessionRepo.Redis().Set(ctx, lockKey, "1", locakDuration)
	}

	return nil
}

func (s *AuthService) ClearFailLogin(ctx context.Context, email string) {
	s.sessionRepo.Redis().Del(ctx,
		"login_fail:"+email,
		"login_lock:"+email,
	)
}

// func (s *AuthService) Create() error {

// }
