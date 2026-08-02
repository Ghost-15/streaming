package repository

import (
	"context"

	"github.com/Ghost-15/streaming/internal/entity"
)

// RecommendationRepository derives stream suggestions from listen history.
type RecommendationRepository interface {
	// RecommendStreams returns live streams popular among other listeners that
	// the given user has not joined yet, ranked by distinct audience.
	RecommendStreams(ctx context.Context, userID string, limit int) ([]entity.Stream, error)
}
