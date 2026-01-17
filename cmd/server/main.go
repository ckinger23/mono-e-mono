package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ckinger23/mono-e-mono/internal/api"
	"github.com/ckinger23/mono-e-mono/internal/auth"
	"github.com/ckinger23/mono-e-mono/internal/db"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	godotenv.Load()

	// Get configuration from environment
	port := getEnv("PORT", "8080")
	databaseURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/mono_e_mono?sslmode=disable")
	jwtSecret := getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production")
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:3000")

	// OAuth configuration
	oauthConfig := &auth.OAuthConfig{
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		CallbackBaseURL:    getEnv("CALLBACK_BASE_URL", "http://localhost:"+port),
	}

	// Connect to database
	ctx := context.Background()
	dbConfig := db.DefaultConfig(databaseURL)
	pool, err := db.Connect(ctx, dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Connected to database")

	// Create API server
	apiConfig := &api.Config{
		FrontendURL: frontendURL,
		JWTSecret:   jwtSecret,
		OAuth:       oauthConfig,
	}
	server := api.NewServer(pool, apiConfig)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      server.Handler(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		fmt.Println("=================================")
		fmt.Println("  MONO-E-MONO Fantasy Platform")
		fmt.Printf("  Server starting on :%s\n", port)
		fmt.Println("=================================")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
