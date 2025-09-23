package hub

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	"gotrade/websocket-gateway/internal/client"
	"gotrade/websocket-gateway/internal/domain"
)

// Hub maintains the set of active clients and broadcasts messages to them.
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
		case c := <-h.register:
			h.clients[c] = true
			log.Println("Client registered")
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.Send)
				log.Println("Client unregistered")
			}
		case message := <-h.broadcast:
			var tick domain.Tick
			if err := json.Unmarshal(message, &tick); err != nil {
				log.Printf("Could not unmarshal tick from redis: %v", err)
				continue
			}

			// Iterate over clients and send message only if they are subscribed
			for c := range h.clients {
				if c.IsSubscribed(tick.Symbol) {
					select {
					case c.Send <- message:
					default:
						close(c.Send)
						delete(h.clients, c)
					}
				}
			}
		}
	}
}

func (h *Hub) RegisterClient(conn *websocket.Conn) {
	c := client.NewClient(h, conn)
	h.register <- c
	go c.WritePump()
	go c.ReadPump()
}

// Unregister satisfies the Hub interface for the client.
func (h *Hub) Unregister(client *client.Client) {
	h.unregister <- client
}

func (h *Hub) subscribeToRedis() {
	pubsub := h.redisClient.Subscribe(context.Background(), "orders", "market_data") // Changed channel
	defer pubsub.Close()
	ch := pubsub.Channel()
	for msg := range ch {
		log.Printf("Received message from Redis on channel %s", msg.Channel)
		if msg.Channel == "market_data" {
			h.broadcast <- []byte(msg.Payload)
		} else {
			log.Printf("[WARNING] Idk what to do here")
		}
	}
}
