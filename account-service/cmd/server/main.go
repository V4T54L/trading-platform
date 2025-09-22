package main

import (
	"context"
	"errors"
	"fmt"
	"gotrade/account-service/internal/handler"
	"gotrade/account-service/internal/platform/database"
	"gotrade/account-service/internal/repository"
	"gotrade/account-service/internal/usecase"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load() // Load .env file if it exists

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8084"
	}

	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		log.Fatal("POSTGRES_URL environment variable is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	// Initialize database
	dbPool, err := database.NewPostgresPool(postgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer dbPool.Close()

	if err := database.InitDatabase(dbPool); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Setup dependencies
	accountRepo := repository.NewPostgresAccountRepository(dbPool)
	accountUsecase := usecase.NewAccountUsecase(accountRepo)
	accountHandler := handler.NewAccountHandler(accountUsecase)

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	} )
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	} )

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(handler.JWTAuthMiddleware(jwtSecret))
		r.Get("/account", accountHandler.GetAccount)
		r.Post("/account/transfer", accountHandler.FundTransfer)
	} )

	// Server setup
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on port %s\n", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Could not listen on %s: %v\n", port, err)
		}
	} ()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

