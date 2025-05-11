package redisImpl

import (
	"TeamTickBackend/dal/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
)

const (
	groupKeyPrefix = "group:"
)

type GroupRedisDAOImpl struct {
	Client *redis.Client
}

func buildGroupKey(groupID int) string {
	return fmt.Sprintf("%s%d", groupKeyPrefix, groupID)
}

// GetByGroupID 通过groupID查询group
func (dao *GroupRedisDAOImpl) GetByGroupID(ctx context.Context, groupID int, tx ...*redis.Client) (*models.Group, error) {
	client := dao.Client
	if len(tx) > 0 && tx[0] != nil {
		client = tx[0]
	}

	key := buildGroupKey(groupID)
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var group models.Group
	if err := json.Unmarshal(data, &group); err != nil {
		return nil, err
	}
	return &group, nil
}

// SetByGroupID 通过groupID设置group缓存
func (dao *GroupRedisDAOImpl) SetByGroupID(ctx context.Context, groupID int, group *models.Group) error {
	client := dao.Client

	key := buildGroupKey(groupID)
	data, err := json.Marshal(group)
	if err != nil {
		return err
	}

	return client.Set(ctx, key, data, DefaultExpireTime).Err()
}

// DeleteCacheByGroupID 通过groupID删除group缓存
func (dao *GroupRedisDAOImpl) DeleteCacheByGroupID(ctx context.Context, groupID int) error {
	client := dao.Client
	key := buildGroupKey(groupID)
	return client.Del(ctx, key).Err()
}
