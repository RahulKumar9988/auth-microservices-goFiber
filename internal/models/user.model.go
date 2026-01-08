package models

import "time"

type UserRole string

const (
	Admin UserRole = "admin"
	User  UserRole = "user"
)

type UserModel struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Role      UserRole  `json:"role" gorm:"type:text;not null;default:'user'"`
	Password  string    `json:"-" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserModel) TableName() string {
	return "users"
}
