package entity

import "time"

type Favorite struct {
	UserID    string    `json:"user_id"`
	StreamID   string    `json:"stream_id"`
	CreatedAt time.Time `json:"created_at"`
}
