package supabase

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

type supabaseRecommendationRepo struct {
	db pgxDB
}

func NewRecommendationRepo(db *pgxpool.Pool) repository.RecommendationRepository {
	return &supabaseRecommendationRepo{db: poolOrNil(db)}
}

func (r *supabaseRecommendationRepo) RecommendStreams(ctx context.Context, userID string, limit int) (streams []entity.Stream, err error) {
	ctx, span := startRepoSpan(ctx, "recommendation", "RecommendationRepository", "RecommendStreams", "listen_history", "SELECT",
		attribute.Int("reco.limit", limit),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return nil, errDatabaseUnavailable
	}
	if limit <= 0 {
		limit = 10
	}

	// Primary: collaborative filtering — live streams popular with other users
	// that the requesting user has not yet joined.
	// Fallback: when no collaborative data exists (new user, solo demo), return
	// all live streams ordered by listener count so the section is always shown.
	const q = `
		WITH collab AS (
		    SELECT s.id, s.title, s.broadcaster_id, s.status,
		           s.started_at, s.ended_at, s.listener_count,
		           COUNT(DISTINCT lh.user_id) AS score
		    FROM listen_history lh
		    JOIN streams s ON s.id = lh.stream_id
		    WHERE lh.event_type = 'join'
		      AND s.status = 'live'
		      AND lh.stream_id NOT IN (
		          SELECT stream_id FROM listen_history
		          WHERE user_id = $1 AND event_type = 'join'
		            AND stream_id IS NOT NULL
		      )
		    GROUP BY s.id, s.title, s.broadcaster_id, s.status,
		             s.started_at, s.ended_at, s.listener_count
		),
		fallback AS (
		    SELECT s.id, s.title, s.broadcaster_id, s.status,
		           s.started_at, s.ended_at, s.listener_count,
		           s.listener_count AS score
		    FROM streams s
		    WHERE s.status = 'live'
		      AND NOT EXISTS (SELECT 1 FROM collab)
		)
		SELECT id, title, broadcaster_id, status, started_at, ended_at, listener_count
		FROM (SELECT * FROM collab UNION ALL SELECT * FROM fallback) combined
		ORDER BY score DESC, started_at DESC
		LIMIT $2`

	rows, queryErr := r.db.Query(ctx, q, userID, limit)
	if queryErr != nil {
		err = fmt.Errorf("recommendation_repo: query: %w", queryErr)
		return nil, err
	}
	defer rows.Close()

	streams = []entity.Stream{}
	for rows.Next() {
		var s entity.Stream
		if scanErr := rows.Scan(
			&s.ID, &s.Title, &s.BroadcasterID, &s.Status, &s.StartedAt, &s.EndedAt, &s.ListenerCount,
		); scanErr != nil {
			err = fmt.Errorf("recommendation_repo: scan: %w", scanErr)
			return nil, err
		}
		streams = append(streams, s)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		err = rowsErr
		return nil, err
	}

	span.SetAttributes(attribute.Int("db.result.row_count", len(streams)))
	return streams, nil
}
