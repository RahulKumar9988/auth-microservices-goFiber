package repositories

import (
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/models"
	"gorm.io/gorm"
)

type UserRepositories struct {
	db *gorm.DB
}

func NewUserRepositories(db *gorm.DB) *UserRepositories {
	return &UserRepositories{db: db}
}

func (r *UserRepositories) Create(user *models.UserModel) error {
	return r.db.Create(user).Error
}

func (r *UserRepositories) FindByEmail(email string) (*models.UserModel, error) {
	var user models.UserModel
	err := r.db.Where("email = ?", user.Email).Find(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
