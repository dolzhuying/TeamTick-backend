package redisImpl

import (
	"TeamTickBackend/dal/models"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const (
	joinApplicationPrefixWithGroupID = "join_application:group_id:"
	joinApplicationPrefixWithGroupIDAndStatus = "join_application:group_id_status:"
	joinApplicationPrefixWithGroupIDAndUserID = "join_application:group_id_user_id:"
)

type JoinApplicationRedisDAO struct{
	Client *redis.Client
}

func buildJoinApplicationKeyWithGroupID(groupID int) string {
	return fmt.Sprintf("%s%d", joinApplicationPrefixWithGroupID, groupID)
}

func buildJoinApplicationKeyWithGroupIDAndStatus(groupID int, status string) string {
	return fmt.Sprintf("%s%d:%s", joinApplicationPrefixWithGroupIDAndStatus, groupID, status)
}

func buildJoinApplicationKeyWithGroupIDAndUserID(groupID int, userID int) string {
	return fmt.Sprintf("%s%d:%d", joinApplicationPrefixWithGroupIDAndUserID, groupID, userID)
}

// GetByGroupID 通过groupID获取加入申请列表缓存
func (dao *JoinApplicationRedisDAO) GetByGroupID(ctx context.Context, groupID int, tx ...*redis.Client) ([]*models.JoinApplication, error) {
	client := dao.Client
	if len(tx) > 0 && tx[0] != nil {
		client = tx[0]
	}
	key := buildJoinApplicationKeyWithGroupID(groupID)
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var applications []*models.JoinApplication
	err = json.Unmarshal(data, &applications)
	if err != nil {
		return nil, err
	}
	return applications, nil
}

// SetByGroupID 通过groupID设置加入申请列表缓存
func (dao *JoinApplicationRedisDAO) SetByGroupID(ctx context.Context, groupID int, applications []*models.JoinApplication) error {
	client := dao.Client
	data, err := json.Marshal(applications)
	if err != nil {
		return err
	}
	key := buildJoinApplicationKeyWithGroupID(groupID)
	return client.Set(ctx, key, data, DefaultExpireTime).Err()
}

// DeleteCacheByGroupID 删除groupID的加入申请列表缓存
func (dao *JoinApplicationRedisDAO) DeleteCacheByGroupID(ctx context.Context, groupID int) error {
	client := dao.Client
	key := buildJoinApplicationKeyWithGroupID(groupID)
	return client.Del(ctx, key).Err()
}

// GetByGroupIDAndStatus 通过groupID和status获取加入申请列表缓存
func (dao *JoinApplicationRedisDAO) GetByGroupIDAndStatus(ctx context.Context, groupID int, status string, tx ...*redis.Client) ([]*models.JoinApplication, error) {
	client := dao.Client
	if len(tx) > 0 && tx[0] != nil {
		client = tx[0]
	}
	key := buildJoinApplicationKeyWithGroupIDAndStatus(groupID, status)
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var applications []*models.JoinApplication
	err = json.Unmarshal(data, &applications)
	if err != nil {
		return nil, err
	}
	return applications, nil
}

// SetByGroupIDAndStatus 通过groupID和status设置加入申请列表缓存
func (dao *JoinApplicationRedisDAO) SetByGroupIDAndStatus(ctx context.Context, groupID int, status string, applications []*models.JoinApplication) error {
	client := dao.Client
	data, err := json.Marshal(applications)
	if err != nil {
		return err
	}
	key := buildJoinApplicationKeyWithGroupIDAndStatus(groupID, status)
	return client.Set(ctx, key, data, DefaultExpireTime).Err()
}

// DeleteCacheByGroupIDAndStatus 删除groupID和status的加入申请列表缓存
func (dao *JoinApplicationRedisDAO) DeleteCacheByGroupIDAndStatus(ctx context.Context, groupID int, status string) error {
	client := dao.Client
	key := buildJoinApplicationKeyWithGroupIDAndStatus(groupID, status)
	return client.Del(ctx, key).Err()
}

// GetByGroupIDAndUserID 通过groupID和userID获取加入申请缓存
func (dao *JoinApplicationRedisDAO) GetByGroupIDAndUserID(ctx context.Context, groupID int, userID int, tx ...*redis.Client) (*models.JoinApplication, error) {
	client := dao.Client
	if len(tx) > 0 && tx[0] != nil {
		client = tx[0]
	}
	key := buildJoinApplicationKeyWithGroupIDAndUserID(groupID, userID)
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var application models.JoinApplication
	err = json.Unmarshal(data, &application)
	if err != nil {
		return nil, err
	}
	return &application, nil
}

// SetByGroupIDAndUserID 通过groupID和userID设置加入申请缓存
func (dao *JoinApplicationRedisDAO) SetByGroupIDAndUserID(ctx context.Context,application *models.JoinApplication) error {
	client := dao.Client
	data, err := json.Marshal(application)
	if err != nil {
		return err
	}
	key := buildJoinApplicationKeyWithGroupIDAndUserID(application.GroupID, application.UserID)
	return client.Set(ctx, key, data, DefaultExpireTime).Err()
}

// DeleteCacheByGroupIDAndUserID 删除groupID和userID的加入申请缓存
func (dao *JoinApplicationRedisDAO) DeleteCacheByGroupIDAndUserID(ctx context.Context,groupID,userID int) error {
	client := dao.Client
	key := buildJoinApplicationKeyWithGroupIDAndUserID(groupID, userID)
	return client.Del(ctx, key).Err()
}