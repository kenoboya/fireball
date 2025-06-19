package model

import "time"

type Location struct {
	LocationID int64     `json:"location_id,omitempty"`
	Latitude   float64   `json:"latitude" db:"latitude"`
	Longitude  float64   `json:"longitude" db:"longitude"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type MessageLocation struct {
	MessageID  int64 `json:"message_id"`
	LocationID int64 `json:"location_id"`
}
