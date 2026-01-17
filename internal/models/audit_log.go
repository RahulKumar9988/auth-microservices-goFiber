package models

import "time"

type AuditLog struct {
	ID        uint      `gorm:"primeryKey"`
	UserID    *uint     `gorm:"index"`
	Event     string    `gorm:"type:varchar(50);index"`
	IP        string    `gorm:"type:varchar(45)"`
	UserAgent string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}	
