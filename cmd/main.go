package main

import (
	"log"

	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/config"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/db"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/middlewares/security"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/redis"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/repositories"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/router"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/server"
	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.Load()

	dbConn, err := db.Connect(
		cfg.DB.URL,
		cfg.DB.MaxIdleConns,
		cfg.DB.MaxOpenConns,
		cfg.DB.ConnMaxLife,
	)

	if err != nil {
		log.Printf("startup failed: %v", err)
	}

	redisClient, err := redis.Connect(redis.Config{
		Addr:     cfg.RedisURL.Addr,
		Password: cfg.RedisURL.Password,
		DB:       cfg.RedisURL.DB,
	})

	if err != nil {
		log.Printf("redis startup failed: %v", err)
	}
	defer redisClient.Close()

	app := fiber.New(fiber.Config{
		AppName: "auth-service",
	})

	sessionRepo := repositories.NewSessionRepository(redisClient)
	rateLimiter := security.NewRateLimiter(redisClient)
	AuditRepo := repositories.NewAuditRepo(dbConn)
	passwordResetRepo := repositories.NewResetPasswordRepository(redisClient)
	router.Register(app, dbConn, cfg.JWT, sessionRepo, rateLimiter, AuditRepo, passwordResetRepo)
	server.Start(app, cfg.AppPort)

}
