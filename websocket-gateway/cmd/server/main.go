package main

import (
	"gotrade/websocket-gateway/internal/hub"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections for simplicity
		return true
	},
}

func serveWs(hub *hub.Hub, w http.ResponseWriter, r *http.Request) {
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

	hub, err := hub.NewHub(redisURL)
	if err != nil {
		log.Fatalf("Failed to create hub: %v", err)
	}
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	log.Printf("WebSocket server starting on port %s", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

