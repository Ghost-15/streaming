package supabase

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"
)

func TestFavoriteRepo_Add_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseFavoriteRepo{db: mock}

	mock.ExpectExec("INSERT INTO favorites").WithArgs(anyArgs(2)...).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	if err := r.Add(context.Background(), "u1", "t1"); err != nil {
		t.Fatalf("Add err = %v", err)
	}
}

func TestFavoriteRepo_Remove_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseFavoriteRepo{db: mock}

	mock.ExpectExec("DELETE FROM favorites").WithArgs(anyArgs(2)...).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	if err := r.Remove(context.Background(), "u1", "t1"); err != nil {
		t.Fatalf("Remove err = %v", err)
	}

	mock.ExpectExec("DELETE FROM favorites").WithArgs(anyArgs(2)...).WillReturnResult(pgxmock.NewResult("DELETE", 0))
	if err := r.Remove(context.Background(), "u1", "missing"); err == nil {
		t.Error("Remove missing: expected error")
	}
}

func TestFavoriteRepo_ListByUser_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseFavoriteRepo{db: mock}

	rows := pgxmock.NewRows([]string{"id", "title", "artist", "duration", "file_url", "uploaded_by", "created_at"}).
		AddRow("t1", "Song", "Artist", 180, "http://x/y.mp3", "u1", time.Now()).
		AddRow("t2", "Song2", "Artist2", 200, "http://x/z.mp3", "u1", time.Now())
	mock.ExpectQuery("FROM favorites").WithArgs("u1").WillReturnRows(rows)

	tracks, err := r.ListByUser(context.Background(), "u1")
	if err != nil {
		t.Fatalf("ListByUser err = %v", err)
	}
	if len(tracks) != 2 {
		t.Errorf("ListByUser len = %d, want 2", len(tracks))
	}
}
