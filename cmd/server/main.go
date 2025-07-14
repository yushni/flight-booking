package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flight-booking/internal/app"
)

// @title Flight Booking Routes API
// @version 1.0
// @description API for aggregating flight route information from multiple providers
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.basic BasicAuth

// @externalDocs.description OpenAPI
// @externalDocs.url https://swagger.io/resources/open-api/

func main() {
	// Create application
	application, err := app.NewApplication()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- application.Run()
	}()

	// Wait for interrupt signal or server error
	select {
	case err := <-serverErrors:
		if err != nil {
			log.Printf("Server error: %v", err)
		}
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
		cancel()
	}

	// Graceful shutdown
	log.Println("Starting graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown the application
	if err := application.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
		os.Exit(1)
	}

	log.Println("Server shutdown complete")
}

// Version information
var (
	version = "1.0.0"
	commit  = "unknown"
	date    = "unknown"
)

// printVersion prints version information
func printVersion() {
	fmt.Printf("Flight Booking API\n")
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Commit: %s\n", commit)
	fmt.Printf("Date: %s\n", date)
}
