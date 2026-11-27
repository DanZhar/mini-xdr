package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FeedStore struct {
	pool *pgxpool.Pool
}

func NewFeedStore(pool *pgxpool.Pool) *FeedStore {
	return &FeedStore{pool: pool}
}

// RecordFetch records a completed attempt. A nil count marks a failed import.
func (s *FeedStore) RecordFetch(ctx context.Context, name string, finishedAt time.Time, count *int) error {
	status := "ok"
	if count == nil {
		status = "error"
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO feeds (name, last_fetch_at, last_status, last_count)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name) DO UPDATE SET
			last_fetch_at = EXCLUDED.last_fetch_at,
			last_status = EXCLUDED.last_status,
			last_count = EXCLUDED.last_count
		WHERE feeds.last_fetch_at IS NULL OR feeds.last_fetch_at <= EXCLUDED.last_fetch_at
	`, name, finishedAt, status, count)
	if err != nil {
		return fmt.Errorf("record fetch for %q: %w", name, err)
	}
	return nil
}
