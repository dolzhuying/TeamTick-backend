package redisImpl

import (
	"TeamTickBackend/dal/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
)

const (
	TaskRecordKeyPrefix     = "task_record:"
	TaskRecordUserKeyPrefix = "task_record:user:"
	TaskRecordTaskKeyPrefix = "task_record:task:"
)

type TaskRecordDAORedisImpl struct {
	Client *redis.Client
}

// buildTaskRecordKey 构建任务记录缓存key
func buildTaskRecordKey(taskID, userID int) string {
	return fmt.Sprintf("%s%d:%d", TaskRecordKeyPrefix, taskID, userID)
}

// buildTaskRecordsTaskKey 构建任务所有记录缓存key
func buildTaskRecordsTaskKey(taskID int) string {
	return fmt.Sprintf("%s%d", TaskRecordTaskKeyPrefix, taskID)
}

// buildTaskRecordsUserKey 构建用户所有记录缓存key
func buildTaskRecordsUserKey(userID int) string {
	return fmt.Sprintf("%s%d", TaskRecordUserKeyPrefix, userID)
}

// GetByTaskID 通过task_id查询组内成员签到记录
func (dao *TaskRecordDAORedisImpl) GetByTaskID(ctx context.Context, taskID int, tx ...*redis.Client) ([]*models.TaskRecord, error) {
	client := dao.Client
	if len(tx) > 0 && tx[0] != nil {
		client = tx[0]
	}

	key := buildTaskRecordsTaskKey(taskID)
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, err
	}

	var records []*models.TaskRecord
	err = json.Unmarshal(data, &records)
	if err != nil {
		return nil, err
	}
	return records, nil
}

// SetByTaskID 缓存任务相关的所有记录
func (dao *TaskRecordDAORedisImpl) SetByTaskID(ctx context.Context, taskID int, records []*models.TaskRecord) error {
	data, err := json.Marshal(records)
	if err != nil {
		return err
	}

	key := buildTaskRecordsTaskKey(taskID)
	return dao.Client.Set(ctx, key, data, DefaultExpireTime).Err()
}

// GetByUserID 通过user_id查询个人所有签到记录
func (dao *TaskRecordDAORedisImpl) GetByUserID(ctx context.Context, userID int, tx ...*redis.Client) ([]*models.TaskRecord, error) {
	client := dao.Client
	if len(tx) > 0 && tx[0] != nil {
		client = tx[0]
	}

	key := buildTaskRecordsUserKey(userID)
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, err
	}

	var records []*models.TaskRecord
	err = json.Unmarshal(data, &records)
	if err != nil {
		return nil, err
	}
	return records, nil
}

// SetByUserID 缓存用户相关的所有记录
func (dao *TaskRecordDAORedisImpl) SetByUserID(ctx context.Context, userID int, records []*models.TaskRecord) error {
	data, err := json.Marshal(records)
	if err != nil {
		return err
	}

	key := buildTaskRecordsUserKey(userID)
	return dao.Client.Set(ctx, key, data, DefaultExpireTime).Err()
}

// GetByTaskIDAndUserID 通过task_id和user_id查询指定签到记录
func (dao *TaskRecordDAORedisImpl) GetByTaskIDAndUserID(ctx context.Context, taskID, userID int, tx ...*redis.Client) (*models.TaskRecord, error) {
	client := dao.Client
	if len(tx) > 0 && tx[0] != nil {
		client = tx[0]
	}

	key := buildTaskRecordKey(taskID, userID)
	data, err := client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, err
	}

	var record models.TaskRecord
	err = json.Unmarshal(data, &record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// SetTaskIDAndUserID 缓存指定的签到记录
func (dao *TaskRecordDAORedisImpl) SetTaskIDAndUserID(ctx context.Context, record *models.TaskRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	key := buildTaskRecordKey(record.TaskID, record.UserID)
	return dao.Client.Set(ctx, key, data, DefaultExpireTime).Err()
}

// DeleteCache 删除所有相关缓存
func (dao *TaskRecordDAORedisImpl) DeleteCache(ctx context.Context, taskID, userID int) error {
	key := buildTaskRecordKey(taskID, userID)
	taskKey := buildTaskRecordsTaskKey(taskID)
	userKey := buildTaskRecordsUserKey(userID)

	pipe := dao.Client.Pipeline()
	pipe.Del(ctx, key)
	pipe.Del(ctx, taskKey)
	pipe.Del(ctx, userKey)
	_, err := pipe.Exec(ctx)
	return err
}
