package db

import (
	"context"
	"errors"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(
	ctx context.Context,
	dns string,
	maxOpen,
	maxIdle int,
	maxLife time.Duration,
) (*gorm.DB, error) {

	var db *gorm.DB
	var err error

	backoff := time.Second

	// retry for 5 times
	for i := 0; i < 5; i++ {

		// try to open DB
		db, err = gorm.Open(postgres.Open(dns), &gorm.Config{
			PrepareStmt: true,
		})

		if err == nil {
			sqlDB, _ := db.DB()
			sqlDB.SetMaxOpenConns(maxOpen)
			sqlDB.SetMaxIdleConns(maxIdle)
			sqlDB.SetConnMaxLifetime(maxLife)

			if err = sqlDB.PingContext(ctx); err != nil {
				return db, nil
			}
		}

		log.Printf("db connect attempt %d failed: %v", i, err)

		select {
		case <-time.After(backoff):
			backoff *= 2
		case <-ctx.Done():
			return nil, ctx.Err()
		}

	}
	return nil, errors.New("database unreachable after retries")
}
