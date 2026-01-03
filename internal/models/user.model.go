package models

import "time"

type UserRole string

const (
	Admin UserRole = "admin"
	User  UserRole = "user"
)

type UserModel struct {
	ID         uint      `json:"id" gorm:"Primerykey"`
	Email      string    `json:"email" gorm:"text;not null"`
	Role       UserRole  `json:"role" gorm:"text;not null"`
	Password   string    `json:"-"`
	Created_At time.Time `json:"created_at"`
	Updated_At time.Time `json:"updated_at"`
}
