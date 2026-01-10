package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RefreshTokenRepository struct {
	rdb *redis.Client
}

func NewRefreshTokenRepository(rdb *redis.Client) *RefreshTokenRepository {
	return &RefreshTokenRepository{rdb: rdb}
}

func (r *RefreshTokenRepository) Store(ctx context.Context, token string, userID uint, ttl time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return r.rdb.Set(ctx, key, userID, ttl).Err()
}

func (r *RefreshTokenRepository) Get(ctx context.Context, token string) (uint, error) {
	key := fmt.Sprintf("refresh_token:%s", token)
	val, err := r.rdb.Get(ctx, key).Result()

	if err != nil {
		return 0, err
	}

	var userID uint
	_, err = fmt.Sscanf(val, "%d", &userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *RefreshTokenRepository) Delete(ctx context.Context, token string) error {
	key := fmt.Sprintf("refresh_token:%s", token)
	return r.rdb.Del(ctx, key).Err()
}
