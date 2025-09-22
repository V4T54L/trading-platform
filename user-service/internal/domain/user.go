package domain

import "time"

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Omit from JSON responses
	CreatedAt time.Time `json:"created_at"`
}

