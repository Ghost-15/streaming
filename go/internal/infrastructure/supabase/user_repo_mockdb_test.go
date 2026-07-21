package supabase

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"

	"github.com/Ghost-15/streaming/internal/entity"
)

func anyArgs(n int) []any {
	a := make([]any, n)
	for i := range a {
		a[i] = pgxmock.AnyArg()
	}
	return a
}

func TestUserRepo_FindByEmail_Mock(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("new mock: %v", err)
	}
	defer mock.Close()
	r := &supabaseUserRepo{db: mock}

	t.Run("found", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "email", "password_hash", "first_name", "last_name", "role", "created_at", "suspended_at"}).
			AddRow("u1", "a@b.c", "hash", "A", "B", "user", time.Now(), nil)
		mock.ExpectQuery("FROM users WHERE email").WithArgs("a@b.c").WillReturnRows(rows)

		u, err := r.FindByEmail(context.Background(), "a@b.c")
		if err != nil {
			t.Fatalf("FindByEmail err = %v", err)
		}
		if u == nil || u.ID != "u1" {
			t.Fatalf("FindByEmail u = %v", u)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("FROM users WHERE email").WithArgs("x@y.z").WillReturnError(pgx.ErrNoRows)
		u, err := r.FindByEmail(context.Background(), "x@y.z")
		if err != nil || u != nil {
			t.Fatalf("FindByEmail not-found u=%v err=%v", u, err)
		}
	})
}

func TestUserRepo_FindByID_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseUserRepo{db: mock}

	rows := pgxmock.NewRows([]string{"id", "email", "password_hash", "first_name", "last_name", "role", "created_at", "suspended_at"}).
		AddRow("u1", "a@b.c", "hash", "A", "B", "admin", time.Now(), nil)
	mock.ExpectQuery("FROM users WHERE id").WithArgs("u1").WillReturnRows(rows)

	u, err := r.FindByID(context.Background(), "u1")
	if err != nil || u == nil || u.Role != entity.RoleAdmin {
		t.Fatalf("FindByID u=%v err=%v", u, err)
	}
}

func TestUserRepo_Create_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseUserRepo{db: mock}

	rows := pgxmock.NewRows([]string{"id", "created_at"}).AddRow("new-id", time.Now())
	mock.ExpectQuery("INSERT INTO users").WithArgs(anyArgs(5)...).WillReturnRows(rows)

	u := &entity.User{Email: "a@b.c", Role: entity.RoleUser}
	if err := r.Create(context.Background(), u); err != nil {
		t.Fatalf("Create err = %v", err)
	}
	if u.ID != "new-id" {
		t.Errorf("Create id = %q, want new-id", u.ID)
	}
}

func TestUserRepo_Update_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseUserRepo{db: mock}

	mock.ExpectExec("UPDATE users SET role").WithArgs(anyArgs(4)...).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := r.Update(context.Background(), &entity.User{ID: "u1", Role: entity.RoleUser}); err != nil {
		t.Fatalf("Update err = %v", err)
	}
}

func TestUserRepo_Delete_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	defer mock.Close()
	r := &supabaseUserRepo{db: mock}

	mock.ExpectExec("DELETE FROM users").WithArgs("u1").WillReturnResult(pgxmock.NewResult("DELETE", 1))
	if err := r.Delete(context.Background(), "u1"); err != nil {
		t.Fatalf("Delete err = %v", err)
	}

	mock.ExpectExec("DELETE FROM users").WithArgs("missing").WillReturnResult(pgxmock.NewResult("DELETE", 0))
	if err := r.Delete(context.Background(), "missing"); err == nil {
		t.Error("Delete missing: expected error")
	}
}
