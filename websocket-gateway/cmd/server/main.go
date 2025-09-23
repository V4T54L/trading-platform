package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"gotrade/websocket-gateway/internal/feed" // Import simulator
	"gotrade/websocket-gateway/internal/hub"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for simplicity
	},
}

func serveWs(hub *hub.Hub, w http.ResponseWriter, r *http.Request) {
	log.Println("[Ws Service] Upgrade Request")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	hub.RegisterClient(conn)
}

func main() {
	_ = godotenv.Load()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8085"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL environment variable is not set")
	}

	// New: Get instrument service URL from environment
	instrumentServiceURL := os.Getenv("INSTRUMENT_SERVICE_URL")
	if instrumentServiceURL == "" {
		log.Fatal("INSTRUMENT_SERVICE_URL environment variable is not set")
	}

	// New: Create and start the data simulator
	simulator, err := feed.NewSimulator(redisURL, instrumentServiceURL)
	if err != nil {
		log.Fatalf("Failed to create data feed simulator: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go simulator.Start(ctx)

	// Create and run the WebSocket hub
	hub, err := hub.NewHub(redisURL)
	if err != nil {
		log.Fatalf("Failed to create hub: %v", err)
	}
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	log.Printf("WebSocket server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
