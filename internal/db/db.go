package db

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
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
	for i := 1; i <= 10; i++ {

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

		// Run migrations
		if err := runMigrations(db.WithContext(context.Background())); err != nil {
			return nil, err
		}

		log.Println("DB Connected sucessfully")
		return db, nil
	}

	return nil, errors.New("database unreachable after retries: " + lastErr.Error())
}

// runMigrations executes all SQL migration files
func runMigrations(db *gorm.DB) error {
	migrationsPath := "migrations"

	// Read migrations directory
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("migrations directory not found, skipping migrations")
			return nil
		}
		return err
	}

	// Execute migration files in order
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		filePath := filepath.Join(migrationsPath, file.Name())
		sqlBytes, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		log.Printf("Running migration: %s", file.Name())
		if err := db.Exec(string(sqlBytes)).Error; err != nil {
			return err
		}
		log.Printf("Migration %s completed", file.Name())
	}

	return nil
}
