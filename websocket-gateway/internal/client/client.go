package client

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"gotrade/websocket-gateway/internal/domain"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Forward declaration of Hub to avoid circular dependency
type Hub interface {
	Unregister(client *Client)
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub           Hub
	Conn          *websocket.Conn
	Send          chan []byte
	mu            sync.RWMutex
	subscriptions map[string]bool
}

// NewClient creates a new client instance.
func NewClient(hub Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:           hub,
		Conn:          conn,
		Send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}
}

// ReadPump pumps messages from the websocket connection to be processed.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.Unregister(c)
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { _ = c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.handleMessage(message)
	}
}

// handleMessage processes incoming messages from the client.
func (c *Client) handleMessage(message []byte) {
	var msg domain.WsMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("error unmarshalling message: %v", err)
		return
	}

	switch msg.Action {
	case "subscribe":
		c.Subscribe(msg.Symbols)
		log.Printf("Client subscribed to: %v", msg.Symbols)
	case "unsubscribe":
		c.Unsubscribe(msg.Symbols)
		log.Printf("Client unsubscribed from: %v", msg.Symbols)
	default:
		log.Printf("Unknown action received: %s", msg.Action)
	}
}

// Subscribe adds symbols to the client's subscription list.
func (c *Client) Subscribe(symbols []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, s := range symbols {
		c.subscriptions[s] = true
	}
}

// Unsubscribe removes symbols from the client's subscription list.
func (c *Client) Unsubscribe(symbols []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, s := range symbols {
		delete(c.subscriptions, s)
	}
}

// IsSubscribed checks if the client is subscribed to a given symbol.
func (c *Client) IsSubscribed(symbol string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.subscriptions[symbol]
	return ok
}

// WritePump pumps messages from the hub to the websocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
