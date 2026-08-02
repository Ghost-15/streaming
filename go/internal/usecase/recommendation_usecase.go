package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

// ErrRecommendationInvalid is returned when the requester is unknown.
var ErrRecommendationInvalid = errors.New("recommendation: invalid input")

const defaultRecommendationLimit = 10

// RecommendationUseCase produces stream suggestions for a user.
type RecommendationUseCase interface {
	Recommend(ctx context.Context, userID string) ([]entity.Stream, error)
}

type recommendationUseCase struct {
	repo repository.RecommendationRepository
}

// NewRecommendationUseCase creates a new RecommendationUseCase.
func NewRecommendationUseCase(repo repository.RecommendationRepository) RecommendationUseCase {
	return &recommendationUseCase{repo: repo}
}

func (uc *recommendationUseCase) Recommend(ctx context.Context, userID string) ([]entity.Stream, error) {
	if userID == "" {
		return nil, ErrRecommendationInvalid
	}
	streams, err := uc.repo.RecommendStreams(ctx, userID, defaultRecommendationLimit)
	if err != nil {
		return nil, fmt.Errorf("recommendation: %w", err)
	}
	return streams, nil
}
