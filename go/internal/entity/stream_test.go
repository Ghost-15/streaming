package entity_test

import (
	"testing"

	"github.com/Ghost-15/streaming/internal/entity"
)

func TestStream_IsLive(t *testing.T) {
	tests := []struct {
		name   string
		status entity.StreamStatus
		want   bool
	}{
		{"live stream is live", entity.StreamStatusLive, true},
		{"ended stream is not live", entity.StreamStatusEnded, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &entity.Stream{Status: tt.status}
			if got := s.IsLive(); got != tt.want {
				t.Errorf("IsLive() = %v, want %v", got, tt.want)
			}
		})
	}
}
