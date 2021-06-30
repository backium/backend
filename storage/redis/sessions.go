package redis

import (
	"context"
	"encoding/json"

	"github.com/backium/backend/core"
	"github.com/go-redis/redis/v8"
)

type redisRepository struct {
	client *redis.Client
}

func NewSessionRepository(addr string, password string) *redisRepository {
	return &redisRepository{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       0,
		}),
	}
}

func (r *redisRepository) Set(ctx context.Context, sess core.Session) error {
	b, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	if err := r.client.Set(ctx, sess.ID, string(b), 0).Err(); err != nil {
		return err
	}
	return nil
}

func (r *redisRepository) Get(ctx context.Context, id string) (core.Session, error) {
	sess := core.Session{}
	bs, err := r.client.Get(ctx, id).Result()
	if err != nil {
		return sess, err
	}
	if err := json.Unmarshal([]byte(bs), &sess); err != nil {
		return sess, err
	}
	return sess, err
}

func (r *redisRepository) Delete(ctx context.Context, id string) error {
	return r.client.Del(ctx, id).Err()
}
