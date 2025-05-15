package dao

import (
	"TeamTickBackend/dal/models"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// UserDAO 用户数据访问接口
type UserDAO interface {
	Create(ctx context.Context, user *models.User, tx ...*gorm.DB) error
	GetByUsername(ctx context.Context, username string, tx ...*gorm.DB) (*models.User, error)
	GetByID(ctx context.Context, id int, tx ...*gorm.DB) (*models.User, error)
}

// TaskDAO 任务数据访问接口
type TaskDAO interface {
	Create(ctx context.Context, task *models.Task, tx ...*gorm.DB) error
	GetByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.Task, error)
	GetByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.Task, error)
	GetActiveTasksByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.Task, error)
	GetEndedTasksByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.Task, error)
	GetActiveTasksByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.Task, error)
	GetEndedTasksByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.Task, error)
	GetByTaskID(ctx context.Context, taskID int, tx ...*gorm.DB) (*models.Task, error)
	UpdateTask(ctx context.Context, taskID int, newTask *models.Task, tx ...*gorm.DB) error
	Delete(ctx context.Context, taskID int, tx ...*gorm.DB) error
}

// GroupDAO 用户组数据访问接口
type GroupDAO interface {
	Create(ctx context.Context, group *models.Group, tx ...*gorm.DB) error
	GetByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) (*models.Group, error)
	GetGroupsByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.Group, error)
	UpdateMessage(ctx context.Context, groupID int, groupName, description string, tx ...*gorm.DB) error
	UpdateMemberNum(ctx context.Context, groupID int, increment bool, tx ...*gorm.DB) error
	GetGroupsByUserIDAndfilter(ctx context.Context, userID int, filter string, tx ...*gorm.DB) ([]*models.Group, error)
	Delete(ctx context.Context, groupID int, tx ...*gorm.DB) error
}

// GroupMemberDAO 组成员数据访问接口
type GroupMemberDAO interface {
	Create(ctx context.Context, member *models.GroupMember, tx ...*gorm.DB) error
	GetMembersByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.GroupMember, error)
	GetMemberByGroupIDAndUserID(ctx context.Context, groupID int, userID int, tx ...*gorm.DB) (*models.GroupMember, error)
	Delete(ctx context.Context, groupID int, userID int, tx ...*gorm.DB) error
}

// TaskRecordDAO 签到记录数据访问接口
type TaskRecordDAO interface {
	Create(ctx context.Context, record *models.TaskRecord, tx ...*gorm.DB) error
	GetByTaskID(ctx context.Context, taskID int, tx ...*gorm.DB) ([]*models.TaskRecord, error)
	GetByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.TaskRecord, error)
	GetByTaskIDAndUserID(ctx context.Context, taskID, userID int, tx ...*gorm.DB) (*models.TaskRecord, error)
}

// JoinApplicationDAO 加入申请数据访问接口
type JoinApplicationDAO interface {
	Create(ctx context.Context, application *models.JoinApplication, tx ...*gorm.DB) error
	GetByGroupIDAndStatus(ctx context.Context, groupID int, status string, tx ...*gorm.DB) ([]*models.JoinApplication, error)
	GetByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.JoinApplication, error)
	UpdateStatus(ctx context.Context, requestID int, status string, tx ...*gorm.DB) error
	UpdateRejectReason(ctx context.Context, requestID int, rejectReason string, tx ...*gorm.DB) error
	GetByGroupIDAndUserID(ctx context.Context, groupID int, userID int, tx ...*gorm.DB) (*models.JoinApplication, error)
	GetByRequestID(ctx context.Context, requestID int, tx ...*gorm.DB) (*models.JoinApplication, error)
	GetByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.JoinApplication, error)
}

// CheckApplicationDAO 签到申请数据访问接口
type CheckApplicationDAO interface {
	Create(ctx context.Context, application *models.CheckApplication, tx ...*gorm.DB) error
	GetByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.CheckApplication, error)
	GetByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.CheckApplication, error)
	Update(ctx context.Context, status string, requestID int, tx ...*gorm.DB) error
	GetByTaskIDAndUserID(ctx context.Context, taskID int, userID int, tx ...*gorm.DB) (*models.CheckApplication, error)
	GetByID(ctx context.Context, id int, tx ...*gorm.DB) (*models.CheckApplication, error)
}


// TaskRecordRedisDAO Redis缓存实现的签到记录访问接口
type TaskRecordRedisDAO interface {
	GetByTaskID(ctx context.Context, taskID int, tx ...*redis.Client) ([]*models.TaskRecord, error)
	SetByTaskID(ctx context.Context, taskID int, records []*models.TaskRecord) error
	GetByUserID(ctx context.Context, userID int, tx ...*redis.Client) ([]*models.TaskRecord, error)
	SetByUserID(ctx context.Context, userID int, records []*models.TaskRecord) error
	GetByTaskIDAndUserID(ctx context.Context, taskID, userID int, tx ...*redis.Client) (*models.TaskRecord, error)
	SetTaskIDAndUserID(ctx context.Context, record *models.TaskRecord) error
	DeleteCache(ctx context.Context, taskID, userID int) error
}

// TaskRedisDAO Redis缓存实现的签到任务访问接口
type TaskRedisDAO interface {
	GetByGroupID(ctx context.Context,groupID int,tx ...*redis.Client) ([]*models.Task,error)
	SetByGroupID(ctx context.Context,groupID int,tasks []*models.Task) error
	GetByTaskID(ctx context.Context,taskID int,tx ...*redis.Client) (*models.Task,error)
	SetByTaskID(ctx context.Context,taskID int,task *models.Task) error
	DeleteCacheByTaskID(ctx context.Context,taskID int) error
	DeleteCacheByGroupID(ctx context.Context,groupID int) error
	GetByUserID(ctx context.Context,userID int,tx ...*redis.Client) ([]*models.Task,error)
	SetByUserID(ctx context.Context,userID int,tasks []*models.Task) error
	DeleteCacheByUserID(ctx context.Context,userID int) error
}

// GroupRedisDAO Redis缓存实现的组访问接口
type GroupRedisDAO interface{
	GetByGroupID(ctx context.Context,groupID int,tx ...*redis.Client) (*models.Group,error)
	SetByGroupID(ctx context.Context,groupID int,group *models.Group) error
	DeleteCacheByGroupID(ctx context.Context,groupID int) error
}

// GroupMemberRedisDAO Redis缓存实现的组成员访问接口
type GroupMemberRedisDAO interface{
	GetMembersByGroupID(ctx context.Context,groupID int,tx ...*redis.Client) ([]*models.GroupMember,error)
	SetMembersByGroupID(ctx context.Context,groupID int,members []*models.GroupMember) error
	DeleteCacheByGroupID(ctx context.Context,groupID int) error
	GetMemberByGroupIDAndUserID(ctx context.Context,groupID,userID int,tx ...*redis.Client) (*models.GroupMember,error)
	SetMemberByGroupIDAndUserID(ctx context.Context,member *models.GroupMember) error
	DeleteCacheByGroupIDAndUserID(ctx context.Context,groupID,userID int) error
}

// JoinApplicationRedisDAO Redis缓存实现的加入申请访问接口
type JoinApplicationRedisDAO interface{
	GetByGroupID(ctx context.Context,groupID int,tx ...*redis.Client) ([]*models.JoinApplication,error)
	SetByGroupID(ctx context.Context,groupID int,applications []*models.JoinApplication) error
	DeleteCacheByGroupID(ctx context.Context,groupID int) error
	GetByGroupIDAndStatus(ctx context.Context,groupID int,status string,tx ...*redis.Client) ([]*models.JoinApplication,error)
	SetByGroupIDAndStatus(ctx context.Context,groupID int,status string,applications []*models.JoinApplication) error
	DeleteCacheByGroupIDAndStatus(ctx context.Context,groupID int,status string) error
	SetByGroupIDAndUserID(ctx context.Context,application *models.JoinApplication) error
	DeleteCacheByGroupIDAndUserID(ctx context.Context,groupID,userID int) error
	GetByGroupIDAndUserID(ctx context.Context, groupID int, userID int, tx ...*redis.Client) (*models.JoinApplication, error)
}

// StatisticsDAO 统计数据访问接口
type StatisticsDAO interface {
	GetAllGroups(ctx context.Context,tx ...*gorm.DB) ([]*models.Group,error)
	GetGroupSignInSuccess(ctx context.Context,groupID int,startTime, endTime time.Time,tx ...*gorm.DB) ([]*models.TaskRecord,error)
	GetGroupSignInException(ctx context.Context, groupID int, startTime, endTime time.Time, tx ...*gorm.DB) ([]*models.CheckApplication, error)
	GetGroupSignInAbsent(ctx context.Context, groupID int, startTime, endTime time.Time, tx ...*gorm.DB) ([]*models.AbsentRecord, error)
	GetMemberSignInSuccessNum(ctx context.Context,groupID,userID int,startTime, endTime time.Time,tx ...*gorm.DB) (int,error)
	GetMemberSignInExceptionNum(ctx context.Context,groupID,userID int,startTime, endTime time.Time,tx ...*gorm.DB) (int,error)
	GetMemberSignInAbsentNum(ctx context.Context,groupID,userID int,startTime, endTime time.Time,tx ...*gorm.DB) (int,error)
}

