package entity

import "time"

type Favorite struct {
	UserID    string    `json:"user_id"`
	TrackID   string    `json:"track_id"`
	CreatedAt time.Time `json:"created_at"`
}
