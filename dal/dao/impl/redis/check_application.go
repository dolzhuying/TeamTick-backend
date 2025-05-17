package redisImpl

import (
	"TeamTickBackend/dal/models"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type CheckApplicationRedisDAOImpl struct {
	Client *redis.Client
}

func buildGroupIDKey(groupID int) string {
	return fmt.Sprintf("check_application:group:%d", groupID)
}

func buildUserIDKey(userID int) string {
	return fmt.Sprintf("check_application:user:%d", userID)
}

func buildGroupIDAndStatusKey(groupID int, status string) string {
	return fmt.Sprintf("check_application:group:%d:status:%s", groupID, status)
}

func (dao *CheckApplicationRedisDAOImpl) GetByGroupID(ctx context.Context, groupID int, tx ...*redis.Client) ([]*models.CheckApplication, error) {
	client := dao.Client
	if len(tx) > 0 {
		client = tx[0]
	}
	key := buildGroupIDKey(groupID)
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var applications []*models.CheckApplication
	err = json.Unmarshal(data, &applications)
	if err != nil {
		return nil, err
	}
	return applications, nil
}

func (dao *CheckApplicationRedisDAOImpl) SetByGroupID(ctx context.Context, groupID int, applications []*models.CheckApplication) error {
	client := dao.Client
	key := buildGroupIDKey(groupID)
	data, err := json.Marshal(applications)
	if err != nil {
		return err
	}
	err = client.Set(ctx, key, data, DefaultExpireTime).Err()
	if err != nil {
		return err
	}
	return nil
}

func (dao *CheckApplicationRedisDAOImpl) DeleteCacheByGroupID(ctx context.Context, groupID int) error {
	client := dao.Client
	key := buildGroupIDKey(groupID)
	return client.Del(ctx, key).Err()
}

func (dao *CheckApplicationRedisDAOImpl) GetByUserID(ctx context.Context, userID int, tx ...*redis.Client) ([]*models.CheckApplication, error) {
	client := dao.Client
	if len(tx) > 0 {
		client = tx[0]
	}
	key := buildUserIDKey(userID)
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var applications []*models.CheckApplication
	err = json.Unmarshal(data, &applications)
	if err != nil {
		return nil, err
	}
	return applications, nil
}

func (dao *CheckApplicationRedisDAOImpl) SetByUserID(ctx context.Context, userID int, applications []*models.CheckApplication) error {
	client := dao.Client
	key := buildUserIDKey(userID)
	data, err := json.Marshal(applications)
	if err != nil {
		return err
	}
	err = client.Set(ctx, key, data, DefaultExpireTime).Err()
	if err != nil {
		return err
	}
	return nil
}

func (dao *CheckApplicationRedisDAOImpl) DeleteCacheByUserID(ctx context.Context, userID int) error {
	client := dao.Client
	key := buildUserIDKey(userID)
	return client.Del(ctx, key).Err()
}

func (dao *CheckApplicationRedisDAOImpl) GetByGroupIDAndStatus(ctx context.Context, groupID int, status string, tx ...*redis.Client) ([]*models.CheckApplication, error) {
	client := dao.Client
	if len(tx) > 0 {
		client = tx[0]
	}
	key := buildGroupIDAndStatusKey(groupID, status)
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var applications []*models.CheckApplication
	err = json.Unmarshal(data, &applications)
	if err != nil {
		return nil, err
	}
	return applications, nil
}

func (dao *CheckApplicationRedisDAOImpl) SetByGroupIDAndStatus(ctx context.Context, groupID int, status string, applications []*models.CheckApplication) error {
	client := dao.Client
	key := buildGroupIDAndStatusKey(groupID, status)
	data, err := json.Marshal(applications)
	if err != nil {
		return err
	}
	err = client.Set(ctx, key, data, DefaultExpireTime).Err()
	if err != nil {
		return err
	}
	return nil
}

func (dao *CheckApplicationRedisDAOImpl) DeleteCacheByGroupIDAndStatus(ctx context.Context, groupID int, status string) error {
	client := dao.Client
	key := buildGroupIDAndStatusKey(groupID, status)
	return client.Del(ctx, key).Err()
}


	
