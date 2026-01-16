package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	rdb *redis.Client
}

func NewSessionRepository(rdb *redis.Client) *SessionRepository {
	return &SessionRepository{rdb: rdb}
}

func (r *SessionRepository) Create(
	ctx context.Context,
	sessionID string,
	userID uint,
	ttl time.Duration,
) error {
	now := time.Now().Unix()
	pipe := r.rdb.TxPipeline()

	sessionKey := fmt.Sprintf("session:%s", sessionID)
	userSessionsKey := fmt.Sprintf("user_sessions:%d", userID)

	pipe.HSet(
		ctx, sessionKey,
		"user_id", userID,
		"created_at", now,
	)

	pipe.Expire(ctx, "session:"+sessionID, ttl)

	pipe.SAdd(ctx, userSessionsKey, sessionID)
	pipe.Expire(ctx, userSessionsKey, ttl)

	_, err := pipe.Exec(ctx)
	return err

}

func (r *SessionRepository) GetUserID(
	ctx context.Context,
	sessionID string,
) (uint, error) {

	sessionKey := fmt.Sprintf("session:%s", sessionID)

	userID, err := r.rdb.HGet(ctx, sessionKey, "user_id").Uint64()
	if err != nil {
		return 0, err
	}

	return uint(userID), nil
}

func (r *SessionRepository) Delete(
	ctx context.Context,
	sessionID string,
	userID uint,
) error {

	sessionKey := fmt.Sprintf("session:%s", sessionID)
	userSessionsKey := fmt.Sprintf("user_sessions:%d", userID)

	pipe := r.rdb.TxPipeline()

	pipe.Del(ctx, sessionKey)
	pipe.SRem(ctx, userSessionsKey, sessionID)

	_, err := pipe.Exec(ctx)
	return err
}
