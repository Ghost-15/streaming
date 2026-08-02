package supabase

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"

	"github.com/Ghost-15/streaming/internal/entity"
)

func playlistRow() *pgxmock.Rows {
	return pgxmock.NewRows([]string{"id", "owner_id", "title", "is_queue", "track_count", "created_at"}).
		AddRow("pl-1", "owner-1", "My Playlist", true, 2, time.Now())
}

func trackJoinRows() *pgxmock.Rows {
	return pgxmock.NewRows([]string{"id", "title", "artist", "duration", "file_url", "uploaded_by", "created_at", "position", "added_at"}).
		AddRow("t1", "Song1", "A", 180, "http://x/1.mp3", "u1", time.Now(), 0, time.Now()).
		AddRow("t2", "Song2", "A", 200, "http://x/2.mp3", "u1", time.Now(), 1, time.Now())
}

func TestPlaylistRepo_FindByID_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabasePlaylistRepo{db: mock}

	mock.ExpectQuery("FROM playlists WHERE id").WithArgs("pl-1").WillReturnRows(playlistRow())
	mock.ExpectQuery("FROM playlist_tracks").WithArgs("pl-1").WillReturnRows(trackJoinRows())

	p, err := r.FindByID(context.Background(), "pl-1")
	if err != nil {
		t.Fatalf("FindByID err = %v", err)
	}
	if p == nil || len(p.Tracks) != 2 {
		t.Fatalf("FindByID p=%v", p)
	}

	mock.ExpectQuery("FROM playlists WHERE id").WithArgs("x").WillReturnError(pgx.ErrNoRows)
	p, err = r.FindByID(context.Background(), "x")
	if err != nil || p != nil {
		t.Fatalf("FindByID not-found p=%v err=%v", p, err)
	}
}

func TestPlaylistRepo_ListByOwner_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabasePlaylistRepo{db: mock}

	mock.ExpectQuery("WHERE owner_id").WithArgs("owner-1").WillReturnRows(playlistRow())
	pls, err := r.ListByOwner(context.Background(), "owner-1")
	if err != nil || len(pls) != 1 {
		t.Fatalf("ListByOwner pls=%v err=%v", pls, err)
	}
}

func TestPlaylistRepo_Create_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabasePlaylistRepo{db: mock}

	rows := pgxmock.NewRows([]string{"id", "track_count", "created_at"}).AddRow("pl-1", 0, time.Now())
	mock.ExpectQuery("INSERT INTO playlists").WithArgs(anyArgs(3)...).WillReturnRows(rows)

	p := &entity.Playlist{OwnerID: "owner-1", Title: "X"}
	if err := r.Create(context.Background(), p); err != nil {
		t.Fatalf("Create err = %v", err)
	}
	if p.ID != "pl-1" {
		t.Errorf("Create id = %q", p.ID)
	}
}

func TestPlaylistRepo_UpdateDelete_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabasePlaylistRepo{db: mock}

	mock.ExpectExec("UPDATE playlists SET title").WithArgs(anyArgs(2)...).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := r.Update(context.Background(), &entity.Playlist{ID: "pl-1", Title: "New"}); err != nil {
		t.Fatalf("Update err = %v", err)
	}

	mock.ExpectExec("DELETE FROM playlists").WithArgs("pl-1").WillReturnResult(pgxmock.NewResult("DELETE", 1))
	if err := r.Delete(context.Background(), "pl-1"); err != nil {
		t.Fatalf("Delete err = %v", err)
	}
}

func TestPlaylistRepo_AddRemoveTrack_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabasePlaylistRepo{db: mock}

	mock.ExpectExec("INSERT INTO playlist_tracks").WithArgs(anyArgs(2)...).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	if err := r.AddTrack(context.Background(), &entity.Track{ID: "t1", PlaylistID: "pl-1"}); err != nil {
		t.Fatalf("AddTrack err = %v", err)
	}

	mock.ExpectExec("DELETE FROM playlist_tracks").WithArgs(anyArgs(2)...).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	if err := r.RemoveTrack(context.Background(), "pl-1", "t1"); err != nil {
		t.Fatalf("RemoveTrack err = %v", err)
	}
}

func TestPlaylistRepo_ReorderTracks_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabasePlaylistRepo{db: mock}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE playlist_tracks SET position = position").WithArgs("pl-1").WillReturnResult(pgxmock.NewResult("UPDATE", 2))
	mock.ExpectExec("UPDATE playlist_tracks SET position").WithArgs(anyArgs(3)...).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec("UPDATE playlist_tracks SET position").WithArgs(anyArgs(3)...).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectCommit()

	if err := r.ReorderTracks(context.Background(), "pl-1", []string{"t2", "t1"}); err != nil {
		t.Fatalf("ReorderTracks err = %v", err)
	}
}
