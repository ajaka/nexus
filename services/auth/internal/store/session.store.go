package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (s *Store) RevokeSession(ctx context.Context, sessionId, userId uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	session, err := s.repo.RevokeSession(ctx, sessionId, userId)
	if err != nil {
		return err
	}
	err = s.cache.SetUserOffline(ctx, session)
	return err
}

func (s *Store) RevokeAllOtherSessions(ctx context.Context, sessionId string, userId uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sessions, err := s.repo.RevokeAllSessions(ctx, sessionId, userId)
	if err != nil {
		return err
	}
	err = s.cache.SetUserOffline(ctx, sessions...)
	return err
}
