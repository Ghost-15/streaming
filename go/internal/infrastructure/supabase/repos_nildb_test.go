package supabase_test

import (
	"context"
	"testing"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/infrastructure/supabase"
)

// These tests exercise the repository guard paths (span setup + nil-db guard)
// without a live database. They ensure every method fails safely when the pool
// is unavailable instead of panicking.

func TestUserRepo_NilDB(t *testing.T) {
	r := supabase.NewUserRepo(nil)
	ctx := context.Background()
	if _, err := r.FindByEmail(ctx, "a@b.c"); err == nil {
		t.Error("FindByEmail nil db: expected error")
	}
	if _, err := r.FindByID(ctx, "id"); err == nil {
		t.Error("FindByID nil db: expected error")
	}
	if err := r.Create(ctx, &entity.User{}); err == nil {
		t.Error("Create nil db: expected error")
	}
	if err := r.Update(ctx, &entity.User{}); err == nil {
		t.Error("Update nil db: expected error")
	}
	if err := r.Delete(ctx, "id"); err == nil {
		t.Error("Delete nil db: expected error")
	}
}

func TestStreamRepo_NilDB(t *testing.T) {
	r := supabase.NewStreamRepo(nil)
	ctx := context.Background()
	if _, err := r.FindByID(ctx, "id"); err == nil {
		t.Error("FindByID nil db: expected error")
	}
	if _, err := r.ListActive(ctx); err == nil {
		t.Error("ListActive nil db: expected error")
	}
	if err := r.Create(ctx, &entity.Stream{}); err == nil {
		t.Error("Create nil db: expected error")
	}
	if err := r.UpdateStatus(ctx, "id", entity.StreamStatusEnded); err == nil {
		t.Error("UpdateStatus nil db: expected error")
	}
	if err := r.IncrementListeners(ctx, "id", 1); err == nil {
		t.Error("IncrementListeners nil db: expected error")
	}
}

func TestPlaylistRepo_NilDB(t *testing.T) {
	r := supabase.NewPlaylistRepo(nil)
	ctx := context.Background()
	if _, err := r.FindByID(ctx, "id"); err == nil {
		t.Error("FindByID nil db: expected error")
	}
	if _, err := r.ListByOwner(ctx, "o"); err == nil {
		t.Error("ListByOwner nil db: expected error")
	}
	if err := r.Create(ctx, &entity.Playlist{}); err == nil {
		t.Error("Create nil db: expected error")
	}
	if err := r.Update(ctx, &entity.Playlist{}); err == nil {
		t.Error("Update nil db: expected error")
	}
	if err := r.Delete(ctx, "id"); err == nil {
		t.Error("Delete nil db: expected error")
	}
	if err := r.AddTrack(ctx, &entity.Track{}); err == nil {
		t.Error("AddTrack nil db: expected error")
	}
	if err := r.RemoveTrack(ctx, "p", "t"); err == nil {
		t.Error("RemoveTrack nil db: expected error")
	}
	if err := r.ReorderTracks(ctx, "p", []string{"t"}); err == nil {
		t.Error("ReorderTracks nil db: expected error")
	}
}

func TestAdminRepo_NilDB(t *testing.T) {
	r := supabase.NewAdminRepo(nil)
	ctx := context.Background()
	if _, _, err := r.ListUsers(ctx, 1, 20); err == nil {
		t.Error("ListUsers nil db: expected error")
	}
	if _, err := r.GetUser(ctx, "id"); err == nil {
		t.Error("GetUser nil db: expected error")
	}
	if err := r.UpdateUserRole(ctx, "id", entity.RoleUser); err == nil {
		t.Error("UpdateUserRole nil db: expected error")
	}
	if err := r.SuspendUser(ctx, "id", true); err == nil {
		t.Error("SuspendUser nil db: expected error")
	}
	if _, err := r.GetStats(ctx); err == nil {
		t.Error("GetStats nil db: expected error")
	}
}

func TestFavoriteRepo_NilDB(t *testing.T) {
	r := supabase.NewFavoriteRepo(nil)
	ctx := context.Background()
	if err := r.Add(ctx, "u", "t"); err == nil {
		t.Error("Add nil db: expected error")
	}
	if err := r.Remove(ctx, "u", "t"); err == nil {
		t.Error("Remove nil db: expected error")
	}
	if _, err := r.ListByUser(ctx, "u"); err == nil {
		t.Error("ListByUser nil db: expected error")
	}
}

func TestListenHistoryRepo_NilDB(t *testing.T) {
	r := supabase.NewListenHistoryRepo(nil)
	ctx := context.Background()
	if err := r.Record(ctx, &entity.ListenHistory{}); err == nil {
		t.Error("Record nil db: expected error")
	}
	if _, err := r.ListByUser(ctx, "u"); err == nil {
		t.Error("ListByUser nil db: expected error")
	}
}
