package repositories

import (
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/models"
	"gorm.io/gorm"
)

type AuditRepo struct {
	db *gorm.DB
}

func NewAuditRepo(db *gorm.DB) *AuditRepo {
	return &AuditRepo{db: db}
}

func (r *AuditRepo) Log(event string, userID *uint, ip, ua string) {
	log := models.AuditLog{
		UserID:    userID,
		Event:     event,
		IP:        ip,
		UserAgent: ua,
	}

	go r.db.Create(&log)
}
