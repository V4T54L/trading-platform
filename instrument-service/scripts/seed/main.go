package main

import (
	"context"
	"log"
	"os"

	"gotrade/instrument-service/internal/domain"
	"gotrade/instrument-service/internal/platform/database"
	"gotrade/instrument-service/internal/repository"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		log.Fatal("POSTGRES_URL environment variable is not set")
	}

	dbPool, err := database.NewPostgresPool(postgresURL)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer dbPool.Close()

	// Ensure table exists
	if err := database.InitDatabase(dbPool); err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	instrumentRepo := repository.NewPostgresInstrumentRepository(dbPool)

	instruments := []domain.Instrument{
		{Symbol: "RELIANCE", Name: "Reliance Industries", Type: domain.Equity, Exchange: "NSE", TickSize: 0.05, LotSize: 1},
		{Symbol: "TCS", Name: "Tata Consultancy Services", Type: domain.Equity, Exchange: "NSE", TickSize: 0.05, LotSize: 1},
		{Symbol: "HDFCBANK", Name: "HDFC Bank", Type: domain.Equity, Exchange: "NSE", TickSize: 0.05, LotSize: 1},
		{Symbol: "INFY", Name: "Infosys", Type: domain.Equity, Exchange: "NSE", TickSize: 0.05, LotSize: 1},
		{Symbol: "NIFTY24JULFUT", Name: "Nifty July 2024 Future", Type: domain.Future, Exchange: "NFO", TickSize: 0.05, LotSize: 50},
		{Symbol: "BANKNIFTY24JULFUT", Name: "Bank Nifty July 2024 Future", Type: domain.Future, Exchange: "NFO", TickSize: 0.05, LotSize: 15},
		{Symbol: "NIFTY24JUL23000CE", Name: "Nifty 25 JUL 2024 23000 CE", Type: domain.Option, Exchange: "NFO", TickSize: 0.05, LotSize: 50},
		{Symbol: "NIFTY24JUL22000PE", Name: "Nifty 25 JUL 2024 22000 PE", Type: domain.Option, Exchange: "NFO", TickSize: 0.05, LotSize: 50},
	}

	log.Println("Seeding instruments...")
	for _, inst := range instruments {
		err := instrumentRepo.Create(context.Background(), &inst)
		if err != nil {
			// We assume UNIQUE constraint on symbol will cause errors if already seeded, which is fine.
			log.Printf("Could not create instrument %s: %v", inst.Symbol, err)
		} else {
			log.Printf("Created instrument: %s", inst.Symbol)
		}
	}
	log.Println("Seeding complete.")
}

