package redisImpl

import (
	"TeamTickBackend/dal/models"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	TaskKeyPrefix = "task:"
	TaskGroupIDKeyPrefix = "task:group:"
	TaskUserIDKeyPrefix = "task:user:"
	DefaultExpireTime = 30 * time.Minute
)

type TaskRedisDAOImpl struct {
	Client *redis.Client
}

// buildTaskKey 构建任务缓存key
func buildTaskKey(taskID int) string {
	return fmt.Sprintf("%s%d", TaskKeyPrefix, taskID)
}

// buildTaskGroupIDKey 构建任务组缓存key
func buildTaskGroupIDKey(groupID int) string {
	return fmt.Sprintf("%s%d", TaskGroupIDKeyPrefix, groupID)
}

// buildTaskUserIDKey 构建用户缓存key
func buildTaskUserIDKey(userID int) string {
	return fmt.Sprintf("%s%d", TaskUserIDKeyPrefix, userID)
}

// GetByGroupID 通过groupID获取任务列表
func (dao *TaskRedisDAOImpl) GetByGroupID(ctx context.Context,groupID int,tx ...*redis.Client) ([]*models.Task,error) {
	client := dao.Client
	if len(tx) > 0 && tx[0] != nil {
		client = tx[0]
	}
	key := buildTaskGroupIDKey(groupID)
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, err
	}
	var tasks []*models.Task
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// SetByGroupID 通过GroupID设置任务列表缓存
func (dao *TaskRedisDAOImpl) SetByGroupID(ctx context.Context,groupID int,tasks []*models.Task) error {
	client := dao.Client
	data, err := json.Marshal(tasks)
	if err != nil {
		return err
	}
	key := buildTaskGroupIDKey(groupID)
	return client.Set(ctx, key, data, DefaultExpireTime).Err()
}

// GetByTaskID 通过taskID获取任务
func (dao *TaskRedisDAOImpl) GetByTaskID(ctx context.Context,taskID int,tx ...*redis.Client) (*models.Task,error) {
	client := dao.Client
	if len(tx) > 0 && tx[0] != nil {
		client = tx[0]
	}
	key := buildTaskKey(taskID)
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, err
	}
	var task models.Task
	err = json.Unmarshal(data, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// SetByTaskID 通过taskID设置任务缓存
func (dao *TaskRedisDAOImpl) SetByTaskID(ctx context.Context,taskID int,task *models.Task) error {
	client := dao.Client
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	key := buildTaskKey(taskID)
	return client.Set(ctx, key, data, DefaultExpireTime).Err()
}

// DeleteCacheByTaskID 通过taskID删除任务缓存
func (dao *TaskRedisDAOImpl) DeleteCacheByTaskID(ctx context.Context,taskID int) error {
	client := dao.Client
	key := buildTaskKey(taskID)
	return client.Del(ctx, key).Err()
}

// DeleteCacheByGroupID 通过groupID删除任务列表缓存
func (dao *TaskRedisDAOImpl) DeleteCacheByGroupID(ctx context.Context,groupID int) error {
	client := dao.Client
	key := buildTaskGroupIDKey(groupID)
	return client.Del(ctx, key).Err()
}

// GetByUserID 通过userID获取任务列表
func (dao *TaskRedisDAOImpl) GetByUserID(ctx context.Context,userID int,tx ...*redis.Client) ([]*models.Task,error) {
	client := dao.Client
	if len(tx) > 0 && tx[0] != nil {
		client = tx[0]
	}
	key := buildTaskUserIDKey(userID)
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, err
	}
	var tasks []*models.Task
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// SetByUserID 通过userID设置任务列表缓存
func (dao *TaskRedisDAOImpl) SetByUserID(ctx context.Context,userID int,tasks []*models.Task) error {
	client := dao.Client
	data,err:=json.Marshal(tasks)
	if err != nil {
		return err
	}
	key := buildTaskUserIDKey(userID)
	return client.Set(ctx, key, data, DefaultExpireTime).Err()
}

// DeleteCacheByUserID 通过userID删除任务列表缓存
func (dao *TaskRedisDAOImpl) DeleteCacheByUserID(ctx context.Context,userID int) error {
	client := dao.Client
	key := buildTaskUserIDKey(userID)
	return client.Del(ctx, key).Err()
}

