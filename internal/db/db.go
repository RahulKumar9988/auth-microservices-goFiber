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
	dns string,
	maxOpen,
	maxIdle int,
	maxLife time.Duration,
) (*gorm.DB, error) {

	// var db *gorm.DB
	var lastErr error
	backoff := time.Second

	// retry for 5 times
	for i := 0; i < 5; i++ {

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		// try to open DB
		db, err := gorm.Open(postgres.Open(dns), &gorm.Config{
			PrepareStmt: true,
		})

		if err != nil {
			lastErr = err
			cancel()
			log.Printf("db open failed (attempt %d): %v", i, err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		sqlDB, err := db.DB()

		if err != nil {
			lastErr = err
			cancel()
			log.Printf("db unwrap failed (attempt %d): %v", i, err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if err = sqlDB.PingContext(ctx); err != nil {
			lastErr = err
			cancel()
			log.Printf("db ping failed (attempt %d): %v", i, err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		cancel()

		sqlDB.SetMaxIdleConns(maxIdle)
		sqlDB.SetMaxOpenConns(maxOpen)
		sqlDB.SetConnMaxLifetime(maxLife)

		log.Println("DB Connected sucessfully")
		return db, nil
	}

	return nil, errors.New("database unreachable after retries: " + lastErr.Error())
}
