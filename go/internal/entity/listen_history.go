package entity

import "time"

type ListenEvent string

const (
	ListenEventJoin  ListenEvent = "join"
	ListenEventLeave ListenEvent = "leave"
)

// ListenHistory records what a user listened to and for how long.
// Used by the recommendation engine (US-025) and analytics.
type ListenHistory struct {
	ID          string      `json:"id"`
	UserID      string      `json:"user_id"`
	TrackID     string      `json:"track_id,omitempty"`
	StreamID    *string     `json:"stream_id,omitempty"`
	Event       ListenEvent `json:"event,omitempty"`
	ListenedAt  time.Time   `json:"listened_at"`
	DurationSec int         `json:"duration_sec"`
}
