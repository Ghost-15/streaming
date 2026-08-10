package supabase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"

	"github.com/Ghost-15/streaming/internal/entity"
)

var errStreamMockErr = errors.New("stream repo mock error")

const mockSessionID = "11111111-1111-4111-8111-111111111111"

func streamRows() *pgxmock.Rows {
	return pgxmock.NewRows([]string{"id", "title", "broadcaster_id", "status", "active_session_id", "started_at", "ended_at", "listener_count"}).
		AddRow("s1", "Live", "bc-1", "live", mockSessionID, time.Now(), nil, 3)
}

func TestStreamRepo_ListByBroadcaster_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseStreamRepo{db: mock}

	mock.ExpectQuery("FROM streams WHERE broadcaster_id").WithArgs("bc-1").WillReturnRows(streamRows())
	streams, err := r.ListByBroadcaster(context.Background(), "bc-1")
	if err != nil || len(streams) != 1 {
		t.Fatalf("ListByBroadcaster streams=%v err=%v", streams, err)
	}
}

func TestStreamRepo_FindByID_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseStreamRepo{db: mock}

	mock.ExpectQuery("FROM streams WHERE id").WithArgs("s1").WillReturnRows(streamRows())
	s, err := r.FindByID(context.Background(), "s1")
	if err != nil || s == nil || s.ID != "s1" {
		t.Fatalf("FindByID s=%v err=%v", s, err)
	}

	mock.ExpectQuery("FROM streams WHERE id").WithArgs("x").WillReturnError(pgx.ErrNoRows)
	s, err = r.FindByID(context.Background(), "x")
	if err != nil || s != nil {
		t.Fatalf("FindByID not-found s=%v err=%v", s, err)
	}
}

func TestStreamRepo_ListActive_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseStreamRepo{db: mock}

	mock.ExpectQuery("FROM streams WHERE status").WillReturnRows(streamRows())
	streams, err := r.ListActive(context.Background())
	if err != nil {
		t.Fatalf("ListActive err = %v", err)
	}
	if len(streams) != 1 {
		t.Errorf("ListActive len = %d, want 1", len(streams))
	}
}

func TestStreamRepo_Create_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseStreamRepo{db: mock}

	rows := pgxmock.NewRows([]string{"id", "started_at", "ended_at", "listener_count"}).
		AddRow("s1", time.Now(), nil, 0)
	mock.ExpectQuery("INSERT INTO streams").WithArgs(anyArgs(4)...).WillReturnRows(rows)

	sessionID := mockSessionID
	s := &entity.Stream{Title: "Live", BroadcasterID: "bc-1", Status: entity.StreamStatusLive, ActiveSessionID: &sessionID}
	if err := r.Create(context.Background(), s); err != nil {
		t.Fatalf("Create err = %v", err)
	}
	if s.ID != "s1" {
		t.Errorf("Create id = %q, want s1", s.ID)
	}
}

func TestStreamRepo_ActivateAndDelete_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseStreamRepo{db: mock}

	mock.ExpectExec("UPDATE streams").WithArgs("s1", mockSessionID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := r.Activate(context.Background(), "s1", mockSessionID); err != nil {
		t.Fatalf("Activate err = %v", err)
	}
	mock.ExpectExec("UPDATE streams").WithArgs("s1", mockSessionID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	deactivated, err := r.Deactivate(context.Background(), "s1", mockSessionID)
	if err != nil || !deactivated {
		t.Fatalf("Deactivate = (%v, %v), want (true, nil)", deactivated, err)
	}
	mock.ExpectExec("DELETE FROM streams").WithArgs("s1").WillReturnResult(pgxmock.NewResult("DELETE", 1))
	if err := r.Delete(context.Background(), "s1"); err != nil {
		t.Fatalf("Delete err = %v", err)
	}
}

func TestStreamRepo_UpdateStatus_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseStreamRepo{db: mock}

	mock.ExpectExec("UPDATE streams").WithArgs(anyArgs(3)...).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := r.UpdateStatus(context.Background(), "s1", entity.StreamStatusEnded); err != nil {
		t.Fatalf("UpdateStatus ended err = %v", err)
	}

	mock.ExpectExec("UPDATE streams").WithArgs(anyArgs(3)...).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := r.UpdateStatus(context.Background(), "s1", entity.StreamStatusLive); err != nil {
		t.Fatalf("UpdateStatus live err = %v", err)
	}

	mock.ExpectExec("UPDATE streams").WithArgs(anyArgs(3)...).WillReturnError(errStreamMockErr)
	if err := r.UpdateStatus(context.Background(), "s1", entity.StreamStatusEnded); err == nil {
		t.Error("UpdateStatus exec error: expected error")
	}
}

func TestStreamRepo_IncrementListeners_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseStreamRepo{db: mock}

	mock.ExpectExec("UPDATE streams").WithArgs(anyArgs(2)...).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := r.IncrementListeners(context.Background(), "s1", 1); err != nil {
		t.Fatalf("IncrementListeners err = %v", err)
	}
}
