package server

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Start(app *fiber.App, port string) {

	go func() {
		if err := app.Listen(":" + port); err != nil {
			// Fiber returns error on shutdown → this is NORMAL
			log.Println("server stopped:", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	log.Println("server exited cleanly")
}
