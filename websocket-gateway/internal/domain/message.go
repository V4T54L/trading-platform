package domain

// WsMessage defines the structure for messages sent from the client.
type WsMessage struct {
	Action  string   `json:"action"` // Should be "subscribe" or "unsubscribe"
	Symbols []string `json:"symbols"`
}

// Tick defines the structure for a market data price update.
type Tick struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}
