package hub

import (
	"context"
	"gotrade/websocket-gateway/internal/client"
	"log"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	clients     map[*client.Client]bool
	broadcast   chan []byte
	register    chan *client.Client
	unregister  chan *client.Client
	redisClient *redis.Client
}

func NewHub(redisAddr string) (*Hub, error) {
	opts, err := redis.ParseURL(redisAddr)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(opts)

	// Ping Redis to check connection
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return &Hub{
		broadcast:   make(chan []byte),
		register:    make(chan *client.Client),
		unregister:  make(chan *client.Client),
		clients:     make(map[*client.Client]bool),
		redisClient: rdb,
	}, nil
}

func (h *Hub) Run() {
	go h.subscribeToRedis()

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Println("Client registered")
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				log.Println("Client unregistered")
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) RegisterClient(conn *websocket.Conn) {
	c := &client.Client{Conn: conn, Send: make(chan []byte, 256)}
	h.register <- c

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go c.WritePump()
	go c.ReadPump(func(c *client.Client) {
		h.unregister <- c
	})
}

func (h *Hub) subscribeToRedis() {
	pubsub := h.redisClient.Subscribe(context.Background(), "orders")
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		log.Printf("Received message from Redis on channel %s: %s", msg.Channel, msg.Payload)
		h.broadcast <- []byte(msg.Payload)
	}
}
