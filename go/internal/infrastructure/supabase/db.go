package supabase

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgxDB is the subset of *pgxpool.Pool used by the repositories.
// It lets tests inject a mock database (pgxmock) in place of a real pool.
type pgxDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

// poolOrNil converts a *pgxpool.Pool into a pgxDB, preserving a true nil
// interface when the pool is nil (avoids the typed-nil interface trap so the
// repository nil-db guards keep working).
func poolOrNil(p *pgxpool.Pool) pgxDB {
	if p == nil {
		return nil
	}
	return p
}
