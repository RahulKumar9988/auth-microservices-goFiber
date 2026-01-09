package services

import (
	"errors"
	"strings"

	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/models"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidInput = errors.New("invalid input")

type AuthService struct {
	userRepo *repositories.UserRepository
}

func NewAuthService(repo *repositories.UserRepository) *AuthService {
	return &AuthService{userRepo: repo}
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

var ErrInvalidCredentials = errors.New("invalid credentials")

func (s *AuthService) Login(email string, password string) (*models.UserModel, error) {
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

	return user, nil
}

func (s *AuthService) GetAllUsers() ([]models.UserModel, error) {
	return s.userRepo.GetAllUsers()
}
