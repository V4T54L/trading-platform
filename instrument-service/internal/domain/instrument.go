package domain

import "time"

type InstrumentType string

const (
	Equity InstrumentType = "EQUITY"
	Future InstrumentType = "FUTURE"
	Option InstrumentType = "OPTION"
)

type Instrument struct {
	ID        int64          `json:"id"`
	Symbol    string         `json:"symbol"`
	Name      string         `json:"name"`
	Type      InstrumentType `json:"type"`
	Exchange  string         `json:"exchange"`
	TickSize  float64        `json:"tick_size"`
	LotSize   int            `json:"lot_size"`
	CreatedAt time.Time      `json:"created_at"`
}

