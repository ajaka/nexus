package repositories

import (
	"auth/internal/errs"
	"auth/internal/models"
	"context"
	"time"

	"github.com/ajaka/nexus-shared/outbox"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) FetchPending(ctx context.Context, limit int) ([]outbox.Record, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var outboxRecords []*models.Outbox
	var tinyOutBoxRecords []outbox.Record

	query := `
	SELECT FROM outbox
	where status = 'pending'
	LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return []outbox.Record{}, err
	}
	defer rows.Close()
	outboxRecords, err = pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[models.Outbox])
	if err != nil {
		return []outbox.Record{}, nil
	}
	for _, r := range outboxRecords {
		miniRecord := outbox.Record{
			ID:        r.Id,
			Payload:   r.Payload,
			EventName: r.Event,
		}
		tinyOutBoxRecords = append(tinyOutBoxRecords, miniRecord)
	}
	return tinyOutBoxRecords, nil
}

func (r *Repository) MarkAsFailed(ctx context.Context, id int64, reason string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
	UPDATE outbox
	SET status = 'failed', context = $2
	WHERE id = $1 AND status = 'sending'
	`

	pgTag, err := r.pool.Exec(ctx, query, id, reason)

	if pgTag.RowsAffected() == 0 {
		return errs.ERR_ROW_DOES_NOT_EXIST
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) MarkAsProcessed(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
	UPDATE outbox
	SET status = 'completed'
	WHERE id = $1 AND status = 'sending'
	`

	pgTag, err := r.pool.Exec(ctx, query, id)

	if pgTag.RowsAffected() == 0 {
		return errs.ERR_ROW_DOES_NOT_EXIST
	}
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) MarkAsSending(ctx context.Context, ids []int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	batch := pgx.Batch{}

	for _, id := range ids {
		batch.Queue(`
			UPDATE outbox
			SET status = 'sending'
			WHERE id = $1 AND status = 'pending'
			`, id,
		)
	}

	results := r.pool.SendBatch(ctx, &batch)
	return results.Close()
}
