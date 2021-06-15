package repository

import (
	"context"
	"encoding/json"

	"github.com/backium/backend/handler"
	"github.com/go-redis/redis/v8"
)

type redisRepository struct {
	client *redis.Client
}

func NewSessionRepository(addr string) *redisRepository {
	return &redisRepository{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: "",
			DB:       0,
		}),
	}
}

func (r *redisRepository) Set(ctx context.Context, s handler.Session) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if err := r.client.Set(ctx, s.ID, string(b), 0).Err(); err != nil {
		return err
	}
	return nil
}

func (r *redisRepository) Get(ctx context.Context, id string) (handler.Session, error) {
	s := handler.Session{}
	bs, err := r.client.Get(ctx, id).Result()
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal([]byte(bs), &s); err != nil {
		return s, err
	}

	return s, err
}

func (r *redisRepository) Delete(ctx context.Context, id string) error {
	return r.client.Del(ctx, id).Err()
}
