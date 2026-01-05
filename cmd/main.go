package main

import (
	"log"

	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/config"
	"github.com/RahulKumar9988/auth-microservices-goFiber/internal/db"
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
		log.Fatalf("startup failed: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName: "auth-service",
	})

	router.Register(app, dbConn)
	server.Start(app, cfg.AppPort)

}
