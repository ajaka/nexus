package cache

import (
	"auth/internal/models"
	"context"
	"time"
)

func (c *Cache) SetUserOnline(ctx context.Context, sessionId string, u *models.MinimalUserStruct, exps ...string) (time.Time, error) {
	var keys []string
	for _, s := range exps {
		key := REDISSESSIONPREFIX + s
		keys = append(keys, key)
	}
	newKey := REDISSESSIONPREFIX + sessionId
	exp := time.Hour * 24
	r := time.Now().Add(exp)

	m := structToInterface(u)
	pipe := c.db.Pipeline()

	pipe.HSet(ctx, newKey, m)
	pipe.Expire(ctx, newKey, exp)
	if len(exps) > 0 {
		pipe.Del(ctx, keys...)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return time.Time{}, err
	}
	return r, nil
}

func (c *Cache) SetUserOffline(ctx context.Context, sessionIds ...string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if len(sessionIds) == 0 {
		return nil
	}
	var keys []string
	for _, s := range sessionIds {
		key := REDISSESSIONPREFIX + s
		keys = append(keys, key)
	}
	err := c.db.Del(ctx, keys...).Err()
	return err
}

func (c *Cache) GetUser(ctx context.Context, sessionId string) (*models.MinimalUserStruct, bool) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var user models.MinimalUserStruct

	key := REDISSESSIONPREFIX + sessionId
	s := c.db.HGetAll(ctx, key)

	u, err := s.Result()
	if err != nil || len(u) == 0 {
		return nil, false
	}
	err = s.Scan(&user)

	if err != nil {
		return nil, false
	}
	return &user, true
}
