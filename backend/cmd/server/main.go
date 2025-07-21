package main

import (
	"context"
	"fmt"
	"log"

	"github.com/email-sorting-app/internal/adapters/ai"
	"github.com/email-sorting-app/internal/adapters/database/postgres"
	"github.com/email-sorting-app/internal/adapters/gmail"
	"github.com/email-sorting-app/internal/adapters/http"
	"github.com/email-sorting-app/internal/adapters/http/handlers"
	"github.com/email-sorting-app/internal/adapters/unsubscribe"
	"github.com/email-sorting-app/internal/config"
	"github.com/email-sorting-app/internal/usecases"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	db, err := initDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize repositories
	accountRepo := postgres.NewAccountRepository(db)
	emailRepo := postgres.NewEmailRepository(db)
	categoryRepo := postgres.NewCategoryRepository(db)

	// Initialize OAuth config
	oauthConfig := cfg.OAuthConfig()

	// Initialize AI service first
	aiService, err := ai.NewGeminiService(cfg.GeminiAPIKey)
	if err != nil {
		log.Fatal("Failed to initialize AI service:", err)
	}
	defer aiService.Close()

	// Initialize external services
	gmailService := gmail.NewGmailService(oauthConfig, aiService)

	// Initialize unsubscribe service
	unsubscribeService := unsubscribe.NewWebAutomationService(aiService)
	defer func() {
		if err := unsubscribeService.Close(); err != nil {
			log.Printf("Failed to close unsubscribe service: %v", err)
		}
	}()

	// Initialize use cases
	authUsecase := usecases.NewAuthUsecase(accountRepo, oauthConfig)
	accountUsecase := usecases.NewAccountUsecase(accountRepo)
	categoryUsecase := usecases.NewCategoryUsecase(categoryRepo, accountRepo, gmailService)
	emailUsecase := usecases.NewEmailUsecase(emailRepo, accountRepo, categoryRepo, gmailService, aiService, unsubscribeService)

	// Initialize HTTP handlers
	authHandler := handlers.NewAuthHandler(authUsecase)
	accountHandler := handlers.NewAccountHandler(accountUsecase)
	categoryHandler := handlers.NewCategoryHandler(categoryUsecase)
	emailHandler := handlers.NewEmailHandler(emailUsecase)

	// Setup routes
	r := http.SetupRoutes(authHandler, accountHandler, categoryHandler, emailHandler)

	// Start server
	fmt.Printf("Server starting on port %s\n", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}

func initDB(databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
