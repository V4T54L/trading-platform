package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gotrade/order-service/internal/handler"
	"gotrade/order-service/internal/platform/database"
	"gotrade/order-service/internal/platform/messaging"
	"gotrade/order-service/internal/repository"
	"gotrade/order-service/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8083"
	}
	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		log.Fatal("POSTGRES_URL environment variable is not set")
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL environment variable is not set")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	dbPool, err := database.NewPostgresPool(postgresURL)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer dbPool.Close()

	if err := database.InitDatabase(dbPool); err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	redisPublisher, err := messaging.NewRedisPublisher(redisURL)
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	defer redisPublisher.Close()

	orderRepo := repository.NewPostgresOrderRepository(dbPool)
	orderUsecase := usecase.NewOrderUsecase(orderRepo, redisPublisher, dbPool)
	orderHandler := handler.NewOrderHandler(orderUsecase)

	// Start the outbox poller in a separate goroutine
	pollerCtx, cancelPoller := context.WithCancel(context.Background())
	defer cancelPoller()
	go orderUsecase.StartOutboxPolling(pollerCtx)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Metrics placeholder"))
	})

	r.Route("/orders", func(r chi.Router) {
		r.Use(handler.JWTAuthMiddleware(jwtSecret))
		r.Post("/", orderHandler.PlaceOrder)
		r.Delete("/{orderID}", orderHandler.CancelOrder)
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
	cancelPoller() // Signal the poller to stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}

