package repositories

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	rdb *redis.Client
}

type SessionInfo struct {
	SessionID string `json:"session_id"`
	IP        string `json:"ip"`
	Browser   string `json:"browser"`
	OS        string `json:"os"`
	CreatedAt int64  `json:"created_at"`
}

func NewSessionRepository(rdb *redis.Client) *SessionRepository {
	return &SessionRepository{rdb: rdb}
}

func parseDevice(ua string) (browser, os string) {
	switch {
	case strings.Contains(ua, "Chrome"):
		browser = "Chrome"
	case strings.Contains(ua, "Firefox"):
		browser = "Firefox"
	case strings.Contains(ua, "Safari"):
		browser = "Safari"
	default:
		browser = "Unknown"
	}

	switch {
	case strings.Contains(ua, "Windows"):
		os = "Windows"
	case strings.Contains(ua, "Mac"):
		os = "macOS"
	case strings.Contains(ua, "Linux"):
		os = "Linux"
	case strings.Contains(ua, "Android"):
		os = "Android"
	case strings.Contains(ua, "iPhone"):
		os = "iOS"
	default:
		os = "Unknown"
	}

	return
}

/* ============================
   Session Create
============================ */

func (r *SessionRepository) Create(
	ctx context.Context,
	sessionID string,
	userID uint,
	ip string,
	userAgent string,
	ttl time.Duration,
) error {

	now := time.Now().Unix()
	browser, os := parseDevice(userAgent)
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	userSessionsKey := fmt.Sprintf("user_session:%d", userID)

	pipe := r.rdb.TxPipeline()

	pipe.HSet(
		ctx,
		sessionKey,
		"ip", ip,
		"user_agent", userAgent,
		"os", os,
		"brwoser", browser,
		"user_id", userID,
		"created_at", now,
	)

	pipe.Expire(ctx, sessionKey, ttl)

	pipe.SAdd(ctx, userSessionsKey, sessionID)
	pipe.Expire(ctx, userSessionsKey, ttl)

	_, err := pipe.Exec(ctx)
	return err
}

/* ============================
   Get userID from session
============================ */

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

/* ============================
   Delete session
============================ */

func (r *SessionRepository) Delete(
	ctx context.Context,
	sessionID string,
	userID uint,
) error {

	sessionKey := fmt.Sprintf("session:%s", sessionID)
	userSessionsKey := fmt.Sprintf("user_session:%d", userID)

	pipe := r.rdb.TxPipeline()

	pipe.Del(ctx, sessionKey)
	pipe.SRem(ctx, userSessionsKey, sessionID)

	_, err := pipe.Exec(ctx)
	return err
}

func (r *SessionRepository) DeleteByUser(
	ctx context.Context,
	sessionID string,
	userID uint,
) error {
	sessionKey := fmt.Sprintf("session:%s", sessionID)
	userSessionKey := fmt.Sprintf("user_session:%d", userID)

	pipe := r.rdb.TxPipeline()
	pipe.Del(ctx, sessionKey)
	pipe.SRem(ctx, userSessionKey, sessionID)

	_, err := pipe.Exec(ctx)
	return err
}

func (r *SessionRepository) DeleteAll(
	ctx context.Context,
	userID uint,
	expectedSessionID string,
) error {
	userSessionKey := fmt.Sprintf("user_session:%d", userID)
	sessionIDs, err := r.rdb.SMembers(ctx, userSessionKey).Result()

	if err != nil {
		return err
	}

	pipe := r.rdb.TxPipeline()
	for _, sid := range sessionIDs {
		if expectedSessionID != "" && sid == expectedSessionID {
			continue
		}
		pipe.Del(ctx, fmt.Sprintf("session:%s", sid))
		pipe.SRem(ctx, userSessionKey, sid)
	}
	_, err = pipe.Exec(ctx)
	return err
}

/* ============================
   List user sessions
============================ */

func (r *SessionRepository) ListByUsers(
	ctx context.Context,
	userID uint,
) ([]SessionInfo, error) {

	userSessionsKey := fmt.Sprintf("user_session:%d", userID)

	sessionIDs, err := r.rdb.SMembers(ctx, userSessionsKey).Result()
	if err != nil {
		return nil, err
	}

	sessions := make([]SessionInfo, 0)

	for _, sid := range sessionIDs {
		data, err := r.rdb.HGetAll(ctx, "session:"+sid).Result()
		if err != nil || len(data) == 0 {
			continue
		}

		createdAt, _ := strconv.ParseInt(data["created_at"], 10, 64)

		sessions = append(sessions, SessionInfo{
			SessionID: sid,
			IP:        data["ip"],
			Browser:   data["browser"],
			OS:        data["os"],
			CreatedAt: createdAt,
		})
	}

	return sessions, nil
}

/* ============================
   Redis access (for rate-limit / locks)
============================ */

func (r *SessionRepository) Redis() *redis.Client {
	return r.rdb
}
