package store

import (
	"auth/internal/cache"
	"auth/internal/errs"
	"auth/internal/models"
	"auth/internal/repositories"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	repo  *repositories.Repository
	cache *cache.Cache
}

func InitStore(repo *repositories.Repository, cache *cache.Cache) *Store {
	return &Store{
		repo: repo, cache: cache,
	}
}

func (s *Store) SetUserOnline(ctx context.Context, payload *models.MinimalUserStruct, session *models.Sessions) (time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tokens, err := s.repo.SetUserOnline(ctx, session)
	if err != nil {
		return time.Time{}, err
	}
	exp, err := s.cache.SetUserOnline(ctx, session.SessionId, payload, tokens...)
	if err != nil {
		return time.Time{}, err
	}
	return exp, nil
}

func (s *Store) SetUserOffline(ctx context.Context, sessionId string, userId uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.repo.SetUserOffline(ctx, sessionId, userId)
	if err != nil && !errors.Is(err, errs.ERR_SESSION_NOT_FOUND) {
		return err
	}
	err = s.cache.SetUserOffline(ctx, sessionId)
	return err
}
