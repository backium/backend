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

func NewSessionRepository(addr string, password string) core.SessionStorage {
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
	if err := r.client.Set(ctx, string(sess.ID), string(b), 0).Err(); err != nil {
		return err
	}
	return nil
}

func (r *redisRepository) Get(ctx context.Context, id core.ID) (core.Session, error) {
	sess := core.Session{}
	bs, err := r.client.Get(ctx, string(id)).Result()
	if err != nil {
		return sess, err
	}
	if err := json.Unmarshal([]byte(bs), &sess); err != nil {
		return sess, err
	}
	return sess, err
}

func (r *redisRepository) Delete(ctx context.Context, id core.ID) error {
	return r.client.Del(ctx, string(id)).Err()
}
