package supabase_test

import (
	"context"
	"strings"
	"testing"

	"github.com/Ghost-15/streaming/internal/infrastructure/supabase"
)

func TestNewPoolReportsInvalidDSN(t *testing.T) {
	pool, err := supabase.NewPool(context.Background(), "://not-a-postgres-dsn")
	if err == nil {
		if pool != nil {
			pool.Close()
		}
		t.Fatal("NewPool() error = nil, want parse dsn error")
	}
	if !strings.Contains(err.Error(), "supabase: parse dsn") {
		t.Fatalf("NewPool() error = %q, want parse dsn context", err)
	}
}
