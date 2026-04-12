package entity

import "time"

// Playlist est une liste ordonnée de tracks appartenant à un user.
// is_queue = true → file de lecture active.
type Playlist struct {
	ID         string    `json:"id"`
	OwnerID    string    `json:"owner_id"`
	Title      string    `json:"title"`
	IsQueue    bool      `json:"is_queue"`
	TrackCount int       `json:"track_count"` // maintenu par trigger BDD (migration 005)
	CreatedAt  time.Time `json:"created_at"`
	Tracks     []Track   `json:"tracks,omitempty"`
}

// Track est un fichier audio référencé dans une playlist.
// Le fichier audio est stocké dans Supabase Storage (bucket: audio).
type Track struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Artist     string    `json:"artist"`
	Duration   int       `json:"duration"` // secondes
	FileURL    string    `json:"file_url"` // URL Supabase Storage
	UploadedBy string    `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`

	// Champs de liaison playlist (remplis quand la track est dans une playlist)
	PlaylistID string    `json:"playlist_id,omitempty"`
	Position   int       `json:"position,omitempty"`
	AddedAt    time.Time `json:"added_at,omitempty"`
}
