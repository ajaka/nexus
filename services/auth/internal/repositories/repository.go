package repositories

import (
	"auth/internal/errs"
	"auth/internal/models"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	pool *pgxpool.Pool
}

const passwordResetCooldown = 10 * 24 * time.Hour

func InitRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	query := `
	SELECT id, email, password, verified
	FROM users
	WHERE email = $1
	`
	rows, err := r.pool.Query(ctx, query, email)
	if err != nil {
		return nil, err
	}
	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[models.User])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ERR_EMAIL_NO_EXISTS
		}
		return nil, err
	}

	return &user, nil
}

func (r *Repository) VerifyUser(ctx context.Context, email string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.pool.Exec(ctx, `
		UPDATE users
		SET verified = TRUE, active = TRUE
		WHERE email = $1 AND verified = FALSE AND active = FALSE
	`, email)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errs.ERR_EMAIL_NO_EXISTS
	}

	return nil
}

func (r *Repository) ResetPassword(ctx context.Context, email string, request *models.ResetPasswordRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var user models.User
	query := `
	SELECT id, password, password_updated_at
	FROM users
	WHERE email = $1
	FOR UPDATE
	`
	if err := tx.QueryRow(ctx, query, email).Scan(
		&user.Id,
		&user.Password,
		&user.PasswordUpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ERR_EMAIL_NO_EXISTS
		}
		return err
	}
	if time.Since(user.PasswordUpdatedAt) < passwordResetCooldown {
		return errs.ERR_PASSWORD_RESET_COOLDOWN
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.NewPassword)); err == nil {
		return errs.ERR_PASSWORD_REUSED
	} else if err != bcrypt.ErrMismatchedHashAndPassword {
		return err
	}

	historyRows, err := tx.Query(ctx, `
		SELECT password
		FROM password_history
		WHERE user_id = $1
	`, user.Id)
	if err != nil {
		return err
	}
	defer historyRows.Close()

	for historyRows.Next() {
		var passwordHash string
		if err := historyRows.Scan(&passwordHash); err != nil {
			return err
		}
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(request.NewPassword)); err == nil {
			return errs.ERR_PASSWORD_REUSED
		} else if err != bcrypt.ErrMismatchedHashAndPassword {
			return err
		}
	}
	if err := historyRows.Err(); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO password_history (user_id, password, created_at)
		VALUES ($1, $2, $3)
	`, user.Id, user.Password, user.PasswordUpdatedAt)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE users
		SET password = $1, password_updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, string(hashedPassword), user.Id)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) CreateUser(ctx context.Context, user *models.RegisterRequest, payload []byte, event string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var newUser models.User
	query := `
	INSERT INTO users (full_name,email,password)
	VALUES ($1, $2, $3)
	RETURNING id
	`
	err = tx.QueryRow(ctx, query, user.FullName, user.Email, user.Password).Scan(&newUser.Id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return errs.ERR_DUPLICATE_EMAIL
		}
		return err
	}
	err = r.CreateNewOutboxEvent(tx, ctx, payload, event, newUser.Id)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) CreateNewOutboxEvent(tx pgx.Tx, ctx context.Context, payload []byte, event string, trig_by uuid.UUID) error {
	query := `
		INSERT INTO outbox (payload,event,triggered_by)
		VALUES ($1,$2,$3)
	`
	_, err := tx.Exec(ctx, query, payload, event, trig_by)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) ActivateEmailRecovery(ctx context.Context, req *models.ForgotPasswordRequest, payload []byte, event string, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = r.CreateNewOutboxEvent(tx, ctx, payload, event, id)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) CheckIfUserExists(ctx context.Context, userId uuid.UUID, email string) error {
	var user models.User

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
	SELECT FROM users
	WHERE id = $1 AND email = $2 AND verified = true
	`

	err := r.pool.QueryRow(ctx, query, userId, email).Scan(&user)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ERR_EMAIL_NO_EXISTS
		}
		return err
	}
	return nil
}

func (r *Repository) SetUserOnline(ctx context.Context, session *models.Sessions) ([]string, error) {
	var rotatedToken pgtype.Text
	var evictedToken pgtype.Text

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `WITH
  purged AS (
    DELETE FROM sessions
    WHERE user_id = $1 AND expires_at < CURRENT_TIMESTAMP
    RETURNING id
  ),
  rotated AS (
    DELETE FROM sessions
    WHERE user_id = $1 AND os = $2 AND browser = $3
    RETURNING session_id AS old_token
  ),
  remaining AS (
    SELECT id, session_id
    FROM sessions
    WHERE user_id = $1
    ORDER BY expires_at ASC
  ),

  evicted AS (
    DELETE FROM sessions
    WHERE id IN (
      SELECT id FROM remaining
      WHERE (SELECT COUNT(*) FROM remaining) >= 5
      LIMIT 1
    )
    RETURNING session_id AS evicted_token
  )

	INSERT INTO user_sessions (user_id, session_id, os, browser, expires_at)
	VALUES ($1, $4, $2, $3, $5)
	RETURNING
  (SELECT old_token FROM rotated LIMIT 1) AS rotated_token,
  (SELECT evicted_token FROM evicted LIMIT 1) AS evicted_token;
`

	err = tx.QueryRow(ctx, query, session.UserId, session.Os, session.Browser, session.SessionId, session.ExpiresAt).Scan(&rotatedToken, &evictedToken)
	var result []string
	if rotatedToken.Valid && rotatedToken.String != "" {
		result = append(result, rotatedToken.String)
	}
	if evictedToken.Valid && evictedToken.String != "" {
		result = append(result, evictedToken.String)
	}
	err = tx.Commit(ctx)

	if err != nil {
		return nil, err
	}
	return result, nil
}
