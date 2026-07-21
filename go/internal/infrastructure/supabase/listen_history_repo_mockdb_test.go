package supabase

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"

	"github.com/Ghost-15/streaming/internal/entity"
)

func TestListenHistoryRepo_Record_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseListenHistoryRepo{db: mock}

	rows := pgxmock.NewRows([]string{"id", "listened_at"}).AddRow("h1", time.Now())
	mock.ExpectQuery("INSERT INTO listen_history").WithArgs(anyArgs(5)...).WillReturnRows(rows)

	sid := "s1"
	entry := &entity.ListenHistory{UserID: "u1", StreamID: &sid, Event: entity.ListenEventJoin}
	if err := r.Record(context.Background(), entry); err != nil {
		t.Fatalf("Record err = %v", err)
	}
	if entry.ID != "h1" {
		t.Errorf("Record id = %q, want h1", entry.ID)
	}
}

func TestListenHistoryRepo_ListByUser_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseListenHistoryRepo{db: mock}

	sid := "s1"
	rows := pgxmock.NewRows([]string{"id", "user_id", "track_id", "stream_id", "event_type", "listened_at", "duration_sec"}).
		AddRow("h1", "u1", "", &sid, "join", time.Now(), 0).
		AddRow("h2", "u1", "", &sid, "leave", time.Now(), 120)
	mock.ExpectQuery("FROM listen_history").WithArgs("u1").WillReturnRows(rows)

	history, err := r.ListByUser(context.Background(), "u1")
	if err != nil {
		t.Fatalf("ListByUser err = %v", err)
	}
	if len(history) != 2 {
		t.Errorf("ListByUser len = %d, want 2", len(history))
	}
}
