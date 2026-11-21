package storage

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DanZhar/mini-xdr/internal/ti/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestIndicatorStoreUpsert(t *testing.T) {
	dsn := os.Getenv("TI_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TI_TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	store := NewIndicatorStore(pool)
	older := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := older.Add(time.Hour)
	url := fmt.Sprintf("https://test.invalid/%d", time.Now().UnixNano())

	ind := domain.Indicator{
		Type:      domain.TypeURL,
		Value:     url,
		FirstSeen: newer,
		LastSeen:  newer,
	}
	if err := store.Upsert(ctx, "urlhaus", []domain.Indicator{ind}); err != nil {
		t.Fatalf("insert indicator: %v", err)
	}

	ind.FirstSeen, ind.LastSeen = older, older
	if err := store.Upsert(ctx, "urlhaus", []domain.Indicator{ind}); err != nil {
		t.Fatalf("repeat observation: %v", err)
	}

	var firstSeen, lastSeen, seenAt time.Time
	err = pool.QueryRow(ctx, `
		SELECT i.first_seen, i.last_seen, s.seen_at
		FROM indicators i
		JOIN indicator_sources s ON s.indicator_id = i.id
		WHERE i.type = $1 AND i.value = $2 AND s.source = $3
	`, ind.Type, ind.Value, "urlhaus").Scan(&firstSeen, &lastSeen, &seenAt)
	if err != nil {
		t.Fatal(err)
	}
	if !firstSeen.Equal(older) || !lastSeen.Equal(newer) || !seenAt.Equal(newer) {
		t.Fatalf("unexpected dates: first_seen=%v last_seen=%v seen_at=%v",
			firstSeen, lastSeen, seenAt)
	}

	if err := store.Upsert(ctx, "another-feed", []domain.Indicator{ind}); err != nil {
		t.Fatalf("insert second source: %v", err)
	}

	var sourceCount int
	err = pool.QueryRow(ctx, `
		SELECT count(*)
		FROM indicators i
		JOIN indicator_sources s ON s.indicator_id = i.id
		WHERE i.type = $1 AND i.value = $2
	`, ind.Type, ind.Value).Scan(&sourceCount)
	if err != nil {
		t.Fatal(err)
	}
	if sourceCount != 2 {
		t.Fatalf("got %d sources, want 2", sourceCount)
	}

	valid := ind
	valid.Value = url + "/rollback"
	invalid := valid
	invalid.Type = domain.IndicatorType("invalid")
	if err := store.Upsert(ctx, "urlhaus", []domain.Indicator{valid, invalid}); err == nil {
		t.Fatal("expected invalid indicator type to fail")
	}

	var exists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM indicators WHERE type = $1 AND value = $2)
	`, valid.Type, valid.Value).Scan(&exists)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("first indicator survived a failed batch")
	}
}
