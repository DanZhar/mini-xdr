package storage

import (
	"context"
	"fmt"

	"github.com/DanZhar/mini-xdr/internal/ti/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IndicatorStore struct {
	pool *pgxpool.Pool
}

func NewIndicatorStore(pool *pgxpool.Pool) *IndicatorStore {
	indicatorStore := IndicatorStore{
		pool: pool,
	}

	return &indicatorStore
}

func (s *IndicatorStore) Upsert(ctx context.Context, source string, inds []domain.Indicator) error {
	tx, beginErr := s.pool.Begin(ctx)

	if beginErr != nil {
		return fmt.Errorf("begin pool error: %s", beginErr)
	}

	defer tx.Rollback(ctx)

	query := `INSERT INTO indicators (type, value, first_seen, last_seen)
			  VALUES ($1, $2, $3, $4)
			  ON CONFLICT (type, value) DO UPDATE
			  SET first_seen = LEAST(indicators.first_seen, EXCLUDED.first_seen),
    			last_seen  = GREATEST(indicators.last_seen, EXCLUDED.last_seen)
			  RETURNING id;`

	sourceQuery := `INSERT INTO indicator_sources (indicator_id, source, seen_at)
					VALUES ($1, $2, $3)
					ON CONFLICT (indicator_id, source) DO UPDATE
					SET seen_at = GREATEST(indicator_sources.seen_at, EXCLUDED.seen_at)`

	for _, ind := range inds {
		var id int64
		err := tx.QueryRow(ctx, query,
			ind.Type,
			ind.Value,
			ind.FirstSeen,
			ind.LastSeen,
		).Scan(&id)

		if err != nil {
			return fmt.Errorf("upsert indicator: %w", err)
		}

		_, execErr := tx.Exec(ctx, sourceQuery, id, source, ind.LastSeen)
		if execErr != nil {
			return fmt.Errorf("upsert indicator: %w", execErr)
		}
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		return fmt.Errorf("commit transaction error: %s", commitErr)
	}

	return nil
}
