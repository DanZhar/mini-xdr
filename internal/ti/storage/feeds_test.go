package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestFeedStoreRecordFetch(t *testing.T) {
	dsn := os.Getenv("TI_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TI_TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	store := NewFeedStore(pool)
	name := fmt.Sprintf("test-feed-%d", time.Now().UnixNano())
	first := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	second := first.Add(time.Minute)
	count := 12

	attempts := []struct {
		label      string
		finishedAt time.Time
		count      *int
		wantTime   time.Time
		wantStatus string
		wantCount  sql.NullInt32
	}{
		{
			label:      "insert successful attempt",
			finishedAt: first,
			count:      &count,
			wantTime:   first,
			wantStatus: "ok",
			wantCount:  sql.NullInt32{Int32: 12, Valid: true},
		},
		{
			label:      "replace with failed attempt",
			finishedAt: second,
			wantTime:   second,
			wantStatus: "error",
		},
		{
			label:      "ignore an older attempt recorded late",
			finishedAt: first,
			count:      &count,
			wantTime:   second,
			wantStatus: "error",
		},
	}

	for _, attempt := range attempts {
		if err := store.RecordFetch(ctx, name, attempt.finishedAt, attempt.count); err != nil {
			t.Fatalf("%s: %v", attempt.label, err)
		}

		var gotTime time.Time
		var gotStatus string
		var gotCount sql.NullInt32
		err := pool.QueryRow(ctx, `
			SELECT last_fetch_at, last_status, last_count
			FROM feeds WHERE name = $1
		`, name).Scan(&gotTime, &gotStatus, &gotCount)
		if err != nil {
			t.Fatalf("%s: %v", attempt.label, err)
		}
		if !gotTime.Equal(attempt.wantTime) || gotStatus != attempt.wantStatus || gotCount != attempt.wantCount {
			t.Fatalf("%s: got (%v, %q, %+v), want (%v, %q, %+v)",
				attempt.label, gotTime, gotStatus, gotCount,
				attempt.wantTime, attempt.wantStatus, attempt.wantCount)
		}
	}
}
