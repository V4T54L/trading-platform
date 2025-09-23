package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"gotrade/websocket-gateway/internal/domain"

	"github.com/redis/go-redis/v9"
)

// Instrument is a simplified struct to decode data from the instrument-service.
type Instrument struct {
	Symbol string `json:"symbol"`
}

// Simulator generates and publishes mock market data.
type Simulator struct {
	redisClient          *redis.Client
	instrumentServiceURL string
}

// NewSimulator creates a new market data simulator.
func NewSimulator(redisAddr, instrumentServiceURL string) (*Simulator, error) {
	opts, err := redis.ParseURL(redisAddr)
	if err != nil {
		return nil, fmt.Errorf("could not parse redis URL: %w", err)
	}
	rdb := redis.NewClient(opts)
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("could not connect to redis: %w", err)
	}
	return &Simulator{
		redisClient:          rdb,
		instrumentServiceURL: instrumentServiceURL,
	}, nil
}

// Start runs the simulator in a loop.
func (s *Simulator) Start(ctx context.Context) {
	log.Println("Starting market data feed simulator...")
	instruments, err := s.fetchInstruments()
	if err != nil || len(instruments) == 0 {
		log.Fatalf("Could not fetch instruments to start simulator. Is instrument-service running? Error: %v", err)
	}
	log.Printf("Loaded %d instruments for simulation.", len(instruments))

	prices := make(map[string]float64)
	for _, inst := range instruments {
		prices[inst.Symbol] = 100 + rand.Float64()*2000 // Assign a random initial price
	}

	ticker := time.NewTicker(time.Second) // Generate a new tick every 250ms
	defer ticker.Stop()

LOOP:
	for range ticker.C {
		select {
		case <-ctx.Done():
			break LOOP
		default:
		}

		now := time.Now()
		formatted := now.Format("15:04:05 02/01") // HH:MM:SS dd/mm

		for _, inst := range instruments {
			change := (rand.Float64() - 0.5) * 0.5 // small random change
			prices[inst.Symbol] += change
			if prices[inst.Symbol] < 0 {
				prices[inst.Symbol] = 0.01
			}

			tick := domain.Tick{
				Symbol:    inst.Symbol,
				Price:     prices[inst.Symbol],
				Timestamp: formatted,
			}

			payload, err := json.Marshal(tick)
			if err != nil {
				log.Printf("Simulator: Error marshalling tick: %v", err)
				continue
			}

			// Publish to the 'market_data' channel
			err = s.redisClient.Publish(context.Background(), "market_data", payload).Err()
			if err != nil {
				log.Printf("Simulator: Error publishing to Redis: %v", err)
			}
		}
	}
}

// fetchInstruments gets the list of tradable instruments.
func (s *Simulator) fetchInstruments() ([]Instrument, error) {
	resp, err := http.Get(s.instrumentServiceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to call instrument service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("instrument service returned non-200 status: %d", resp.StatusCode)
	}

	var instruments []Instrument
	if err := json.NewDecoder(resp.Body).Decode(&instruments); err != nil {
		return nil, fmt.Errorf("failed to decode instruments: %w", err)
	}
	return instruments, nil
}
