package entity

import "time"

// StreamStatus represents the live status of a stream.
type StreamStatus string

const (
	StreamStatusLive  StreamStatus = "live"
	StreamStatusEnded StreamStatus = "ended"
)

// Stream represents a live audio broadcast session.
type Stream struct {
	ID            string       `json:"id"`
	Title         string       `json:"title"`
	BroadcasterID string       `json:"broadcaster_id"`
	StreamURL     string       `json:"stream_url"`
	Status        StreamStatus `json:"status"`
	StartedAt     time.Time    `json:"started_at"`
	EndedAt       *time.Time   `json:"ended_at,omitempty"`
	ListenerCount int          `json:"listener_count"`
}

func (s *Stream) IsLive() bool {
	return s.Status == StreamStatusLive
}
