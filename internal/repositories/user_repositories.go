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
