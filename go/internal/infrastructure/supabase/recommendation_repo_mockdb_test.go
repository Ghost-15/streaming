package supabase

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
)

func TestRecommendationRepo_RecommendStreams_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseRecommendationRepo{db: mock}

	rows := pgxmock.NewRows([]string{"id", "title", "broadcaster_id", "status", "started_at", "ended_at", "listener_count"}).
		AddRow("s1", "Popular Live", "bc-1", "live", time.Now(), nil, 12).
		AddRow("s2", "Second", "bc-2", "live", time.Now(), nil, 8)
	mock.ExpectQuery("FROM listen_history").WithArgs(anyArgs(2)...).WillReturnRows(rows)

	streams, err := r.RecommendStreams(context.Background(), "u1", 10)
	if err != nil {
		t.Fatalf("RecommendStreams err = %v", err)
	}
	if len(streams) != 2 {
		t.Errorf("RecommendStreams len = %d, want 2", len(streams))
	}
}

func TestRecommendationRepo_NilDB(t *testing.T) {
	r := NewRecommendationRepo(nil)
	if _, err := r.RecommendStreams(context.Background(), "u1", 10); err == nil {
		t.Error("RecommendStreams nil db: expected error")
	}
}
