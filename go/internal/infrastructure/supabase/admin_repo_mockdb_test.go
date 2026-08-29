package supabase

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"

	"github.com/Ghost-15/streaming/internal/entity"
)

func TestAdminRepo_ListUsers_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &adminRepo{db: mock}

	mock.ExpectQuery("SELECT COUNT").WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(2))
	userRows := pgxmock.NewRows([]string{"id", "email", "password_hash", "first_name", "last_name", "role", "created_at", "suspended_at"}).
		AddRow("u1", "a@b.c", "h", "A", "B", "user", time.Now(), nil).
		AddRow("u2", "c@d.e", "h", "C", "D", "diffuseur", time.Now(), nil)
	mock.ExpectQuery("ORDER BY created_at").WithArgs(anyArgs(2)...).WillReturnRows(userRows)

	users, total, err := r.ListUsers(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("ListUsers err = %v", err)
	}
	if total != 2 || len(users) != 2 {
		t.Errorf("ListUsers total=%d len=%d, want 2/2", total, len(users))
	}
	if users[0].PasswordHash != "" {
		t.Error("ListUsers must redact password hash")
	}
}

func TestAdminRepo_GetUser_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &adminRepo{db: mock}

	rows := pgxmock.NewRows([]string{"id", "email", "password_hash", "first_name", "last_name", "role", "created_at", "suspended_at"}).
		AddRow("u1", "a@b.c", "h", "A", "B", "admin", time.Now(), nil)
	mock.ExpectQuery("FROM users").WithArgs("u1").WillReturnRows(rows)

	u, err := r.GetUser(context.Background(), "u1")
	if err != nil || u == nil || u.Role != entity.RoleAdmin {
		t.Fatalf("GetUser u=%v err=%v", u, err)
	}

	mock.ExpectQuery("FROM users").WithArgs("x").WillReturnError(pgx.ErrNoRows)
	u, err = r.GetUser(context.Background(), "x")
	if err != nil || u != nil {
		t.Fatalf("GetUser not-found u=%v err=%v", u, err)
	}
}

func TestAdminRepo_UpdateUserRole_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &adminRepo{db: mock}

	mock.ExpectExec("UPDATE users SET role").WithArgs(anyArgs(2)...).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := r.UpdateUserRole(context.Background(), "u1", entity.RoleDiffuseur); err != nil {
		t.Fatalf("UpdateUserRole err = %v", err)
	}

	mock.ExpectExec("UPDATE users SET role").WithArgs(anyArgs(2)...).WillReturnResult(pgxmock.NewResult("UPDATE", 0))
	if err := r.UpdateUserRole(context.Background(), "x", entity.RoleUser); err == nil {
		t.Error("UpdateUserRole missing: expected error")
	}
}

func TestAdminRepo_SuspendUser_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &adminRepo{db: mock}

	mock.ExpectExec("UPDATE").WithArgs(anyArgs(2)...).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := r.SuspendUser(context.Background(), "u1", true); err != nil {
		t.Fatalf("SuspendUser err = %v", err)
	}
}

func TestAdminRepo_GetStats_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &adminRepo{db: mock}

	mock.ExpectQuery("SELECT COUNT").WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(42))
	roleRows := pgxmock.NewRows([]string{"role", "count"}).
		AddRow("user", 38).
		AddRow("diffuseur", 3).
		AddRow("admin", 1)
	mock.ExpectQuery("GROUP BY role").WillReturnRows(roleRows)

	stats, err := r.GetStats(context.Background())
	if err != nil {
		t.Fatalf("GetStats err = %v", err)
	}
	if stats.TotalUsers != 42 || stats.ByRole["user"] != 38 {
		t.Errorf("GetStats = %+v", stats)
	}
}

func TestAdminRepo_DeleteUser_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &adminRepo{db: mock}

	mock.ExpectExec("DELETE FROM users").WithArgs("u1").WillReturnResult(pgxmock.NewResult("DELETE", 1))
	if err := r.DeleteUser(context.Background(), "u1"); err != nil {
		t.Fatalf("DeleteUser err = %v", err)
	}

	mock.ExpectExec("DELETE FROM users").WithArgs("ghost").WillReturnResult(pgxmock.NewResult("DELETE", 0))
	if err := r.DeleteUser(context.Background(), "ghost"); err == nil {
		t.Error("DeleteUser on a missing user: expected an error")
	}
}

func TestAdminRepo_DeleteUser_NilDB(t *testing.T) {
	r := &adminRepo{db: nil}
	if err := r.DeleteUser(context.Background(), "u1"); err == nil {
		t.Error("DeleteUser with a nil pool: expected an error")
	}
}
