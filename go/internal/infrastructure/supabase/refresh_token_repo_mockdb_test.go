package supabase

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v4"

	"github.com/Ghost-15/streaming/internal/entity"
)

func TestRefreshTokenRepo_Create_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseRefreshTokenRepo{db: mock}

	rows := pgxmock.NewRows([]string{"id", "created_at"}).AddRow("rt1", time.Now())
	mock.ExpectQuery("INSERT INTO refresh_tokens").WithArgs(anyArgs(3)...).WillReturnRows(rows)

	token := &entity.RefreshToken{UserID: "u1", TokenHash: "hash", ExpiresAt: time.Now().Add(time.Hour)}
	if err := r.Create(context.Background(), token); err != nil {
		t.Fatalf("Create err = %v", err)
	}
	if token.ID != "rt1" {
		t.Errorf("Create id = %q, want %q", token.ID, "rt1")
	}
}

func TestRefreshTokenRepo_FindByHash_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseRefreshTokenRepo{db: mock}

	rows := pgxmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at", "revoked_at"}).
		AddRow("rt1", "u1", "hash", time.Now().Add(time.Hour), time.Now(), nil)
	mock.ExpectQuery("FROM refresh_tokens").WithArgs("hash").WillReturnRows(rows)

	token, err := r.FindByHash(context.Background(), "hash")
	if err != nil {
		t.Fatalf("FindByHash err = %v", err)
	}
	if token == nil || token.UserID != "u1" {
		t.Fatalf("FindByHash token = %v, want the row for u1", token)
	}
}

// A missing token is a normal outcome, not an error.
func TestRefreshTokenRepo_FindByHashMissing_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseRefreshTokenRepo{db: mock}

	empty := pgxmock.NewRows([]string{"id", "user_id", "token_hash", "expires_at", "created_at", "revoked_at"})
	mock.ExpectQuery("FROM refresh_tokens").WithArgs("ghost").WillReturnRows(empty)

	token, err := r.FindByHash(context.Background(), "ghost")
	if err != nil {
		t.Fatalf("FindByHash err = %v", err)
	}
	if token != nil {
		t.Errorf("FindByHash token = %v, want nil", token)
	}
}

func TestRefreshTokenRepo_Revoke_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseRefreshTokenRepo{db: mock}

	mock.ExpectExec("UPDATE refresh_tokens").WithArgs("hash").WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := r.Revoke(context.Background(), "hash"); err != nil {
		t.Fatalf("Revoke err = %v", err)
	}

	// Revoking an already revoked token affects no row and must stay successful.
	mock.ExpectExec("UPDATE refresh_tokens").WithArgs("hash").WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	if err := r.Revoke(context.Background(), "hash"); err != nil {
		t.Errorf("Revoke on an already revoked token err = %v, want nil", err)
	}
}

func TestRefreshTokenRepo_RevokeAllForUser_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseRefreshTokenRepo{db: mock}

	mock.ExpectExec("UPDATE refresh_tokens").WithArgs("u1").WillReturnResult(pgxmock.NewResult("UPDATE", 3))
	if err := r.RevokeAllForUser(context.Background(), "u1"); err != nil {
		t.Fatalf("RevokeAllForUser err = %v", err)
	}
}

func TestRefreshTokenRepo_DeleteExpired_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseRefreshTokenRepo{db: mock}

	mock.ExpectExec("DELETE FROM refresh_tokens").WithArgs(anyArgs(1)...).WillReturnResult(pgxmock.NewResult("DELETE", 5))
	deleted, err := r.DeleteExpired(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("DeleteExpired err = %v", err)
	}
	if deleted != 5 {
		t.Errorf("DeleteExpired deleted = %d, want 5", deleted)
	}
}

// Every method must refuse to run without a pool rather than panic.
func TestRefreshTokenRepo_NilDB(t *testing.T) {
	r := &supabaseRefreshTokenRepo{db: nil}
	ctx := context.Background()

	if err := r.Create(ctx, &entity.RefreshToken{}); err == nil {
		t.Error("Create with a nil pool: expected an error")
	}
	if _, err := r.FindByHash(ctx, "hash"); err == nil {
		t.Error("FindByHash with a nil pool: expected an error")
	}
	if err := r.Revoke(ctx, "hash"); err == nil {
		t.Error("Revoke with a nil pool: expected an error")
	}
	if err := r.RevokeAllForUser(ctx, "u1"); err == nil {
		t.Error("RevokeAllForUser with a nil pool: expected an error")
	}
	if _, err := r.DeleteExpired(ctx, time.Now()); err == nil {
		t.Error("DeleteExpired with a nil pool: expected an error")
	}
}
