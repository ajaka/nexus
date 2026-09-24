package repositories

import (
	"auth/internal/errs"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) RevokeSession(ctx context.Context, sessionId, userId uuid.UUID) (string, error) {
	var session string
	query := `
		DELETE FROM sessions
		WHERE id = $1 AND user_id =$2
		RETURNING session_id
	`

	err := r.pool.QueryRow(ctx, query, sessionId, userId).Scan(&session)
	if err != nil {
		return "", err
	}

	return session, nil
}

func (r *Repository) SetUserOffline(ctx context.Context, sessionId string, userId uuid.UUID) error {
	query := `
		DELETE FROM sessions
		WHERE session_id = $1 AND user_id = $2
	`

	tag, err := r.pool.Exec(ctx, query, sessionId, userId)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errs.ERR_SESSION_NOT_FOUND
	}

	return nil
}

func (r *Repository) RevokeAllSessions(ctx context.Context, sessionId string, userId uuid.UUID) ([]string, error) {

	query := `
	DELETE FROM sessions
	WHERE user_id = $1 AND session_id != $2
	RETURNING session_id
	`
	rows, err := r.pool.Query(ctx, query, userId, sessionId)
	if err != nil {
		return nil, err
	}
	sessions, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[string])
	if err != nil {
		return nil, err
	}
	return sessions, nil
}
