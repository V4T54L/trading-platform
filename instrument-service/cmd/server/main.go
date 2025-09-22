package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gotrade/instrument-service/internal/handler"
	"gotrade/instrument-service/internal/platform/database"
	"gotrade/instrument-service/internal/repository"
	"gotrade/instrument-service/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8082"
	}

	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		log.Fatal("POSTGRES_URL environment variable is not set")
	}

	dbPool, err := database.NewPostgresPool(postgresURL)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer dbPool.Close()

	if err := database.InitDatabase(dbPool); err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	instrumentRepo := repository.NewPostgresInstrumentRepository(dbPool)
	instrumentUsecase := usecase.NewInstrumentUsecase(instrumentRepo)
	instrumentHandler := handler.NewInstrumentHandler(instrumentUsecase)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Metrics placeholder"))
	})

	r.Route("/instruments", func(r chi.Router) {
		r.Get("/", instrumentHandler.GetAllInstruments)
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", port, err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}

