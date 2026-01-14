package repositories

import (
	"errors"

	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.UserModel) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByEmail(email string) (*models.UserModel, error) {
	var user models.UserModel

	err := r.db.Where("email = ?", email).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetAllUsers() ([]models.UserModel, error) {
	var users []models.UserModel
	err := r.db.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetAllAdmins() ([]models.UserModel, error) {
	var admins []models.UserModel
	err := r.db.
		Where("role = ?", "admin").
		Find(&admins).Error
	if err != nil {
		return nil, err
	}
	return admins, nil
}
