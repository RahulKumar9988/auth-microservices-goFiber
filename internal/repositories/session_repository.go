package repositories

import (
	"context"
	"fmt"
	"strconv"
)

type SessionInfo struct {
	SessionID string `json:"session_id"`
	CreatedAt int64  `json:"created_id"`
}

func (r *SessionRepository) ListByUsers(
	ctx context.Context,
	userID uint,
) ([]SessionInfo, error) {
	sessionIDs, err := r.rdb.SMembers(
		ctx,
		fmt.Sprintf("user_session:%d", userID),
	).Result()
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
			CreatedAt: createdAt,
		})
	}

	return sessions, nil

}
