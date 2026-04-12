package entity

import "time"

// ListenHistory records what a user listened to and for how long.
// Used by the recommendation engine (US-025).
type ListenHistory struct {
	UserID      string    `json:"user_id"`
	TrackID     string    `json:"track_id"`
	StreamID    *string   `json:"stream_id,omitempty"`
	ListenedAt  time.Time `json:"listened_at"`
	DurationSec int       `json:"duration_sec"`
}
