package repositories

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type PasswordResetRepository struct {
	rdb *redis.Client
}

func NewResetPasswordRepository(rdb *redis.Client) *PasswordResetRepository {
	return &PasswordResetRepository{rdb: rdb}
}

func hashToken(token string) string {
	h := sha256.Sum224([]byte(token))
	return hex.EncodeToString(h[:])
}

func (r *PasswordResetRepository) Store(ctx context.Context, rowToken string, userID uint, ttl time.Duration) error {
	key := fmt.Sprintf("pwd_reset:%s", hashToken(rowToken))
	return r.rdb.Set(ctx, key, userID, ttl).Err()
}

func (r *PasswordResetRepository) Get(ctx context.Context, rowToken string) (uint, error) {
	key := fmt.Sprintf("pwd_reset:%s", hashToken(rowToken))
	uid64, err := r.rdb.Get(ctx, key).Uint64()
	if err != nil {
		return 0, err
	}

	return uint(uid64), nil
}

func (r *PasswordResetRepository) Delete(ctx context.Context, rowToken string) error {
	key := fmt.Sprintf("pwd_reset:%s", hashToken(rowToken))
	return r.rdb.Del(ctx, key).Err()
}
