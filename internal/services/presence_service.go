package services

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const onlineKey = "presence:online"

type PresenceService struct {
	rdb *redis.Client
}

type UserOnlineStatus struct {
	UserID uint `json:"user_id"`
	Online bool `json:"online"`
}

func NewPresenceService(rdb *redis.Client) *PresenceService {
	return &PresenceService{rdb: rdb}
}

func (s *PresenceService) MarkOnline(ctx context.Context, userID uint) error {
	return s.rdb.HSet(ctx, onlineKey, strconv.FormatUint(uint64(userID), 10), time.Now().Unix()).Err()
}

func (s *PresenceService) GetOnlineCount(ctx context.Context) (int, error) {
	all, err := s.rdb.HGetAll(ctx, onlineKey).Result()
	if err != nil {
		return 0, err
	}

	now := time.Now().Unix()
	count := 0
	var stale []string

	for userID, tsStr := range all {
		ts, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			stale = append(stale, userID)
			continue
		}

		if now-ts <= 70 {
			count++
		} else {
			stale = append(stale, userID)
		}
	}

	if len(stale) > 0 {
		_ = s.rdb.HDel(ctx, onlineKey, stale...).Err()
	}

	return count, nil
}

func (s *PresenceService) GetUsersStatus(ctx context.Context, userIDs []uint) ([]UserOnlineStatus, error) {
	now := time.Now().Unix()
	result := make([]UserOnlineStatus, 0, len(userIDs))

	for _, userID := range userIDs {
		val, err := s.rdb.HGet(ctx, onlineKey, strconv.FormatUint(uint64(userID), 10)).Result()
		if err == redis.Nil {
			result = append(result, UserOnlineStatus{
				UserID: userID,
				Online: false,
			})
			continue
		}
		if err != nil {
			return nil, err
		}

		ts, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			result = append(result, UserOnlineStatus{
				UserID: userID,
				Online: false,
			})
			continue
		}

		isOnline := now-ts <= 70
		if !isOnline {
			_ = s.rdb.HDel(ctx, onlineKey, strconv.FormatUint(uint64(userID), 10)).Err()
		}

		result = append(result, UserOnlineStatus{
			UserID: userID,
			Online: isOnline,
		})
	}

	return result, nil
}
