package dao

import (
	"TeamTickBackend/dal/models"

	"gorm.io/gorm"
)

// UserDAO 用户数据访问接口
type UserDAO interface {
	Create(user *models.User, tx ...*gorm.DB) error
	GetByUsername(username string, tx ...*gorm.DB) (*models.User, error)
	GetByID(id int, tx ...*gorm.DB) (*models.User, error)
}

// TaskDAO 任务数据访问接口
type TaskDAO interface {
	Create(task *models.Task, tx ...*gorm.DB) error
	GetByGroupID(groupID int, tx ...*gorm.DB) ([]*models.Task, error)
	GetActiveTasksByUserID(userID int, tx ...*gorm.DB) ([]*models.Task, error)
	GetByTaskID(taskID int, tx ...*gorm.DB) (*models.Task, error)
}

// GroupDAO 用户组数据访问接口
type GroupDAO interface {
	Create(group *models.Group, tx ...*gorm.DB) error
	GetByID(id int, tx ...*gorm.DB) (*models.Group, error)
	UpdateMessage(groupID int, groupName, description string, tx ...*gorm.DB) error
	UpdateMemberNum(groupID int, increment bool, tx ...*gorm.DB) error
}

// GroupMemberDAO 组成员数据访问接口
type GroupMemberDAO interface {
	Create(member *models.GroupMember, tx ...*gorm.DB) error
	GetMembersByGroupID(groupID int, tx ...*gorm.DB) ([]*models.GroupMember, error)
	GetGroupsByUserID(userID int, tx ...*gorm.DB) ([]*models.GroupMember, error)
	Delete(groupID int, userID int, tx ...*gorm.DB) error
}

// TaskRecordDAO 签到记录数据访问接口
type TaskRecordDAO interface {
	Create(record *models.TaskRecord, tx ...*gorm.DB) error
	GetByTaskID(taskID int, tx ...*gorm.DB) ([]*models.TaskRecord, error)
	GetByUserID(userID int, tx ...*gorm.DB) ([]*models.TaskRecord, error)
}

// JoinApplicationDAO 加入申请数据访问接口
type JoinApplicationDAO interface {
	Create(application *models.JoinApplication, tx ...*gorm.DB) error
	GetByGroupIDAndStatus(groupID int, status string, tx ...*gorm.DB) ([]*models.JoinApplication, error)
	GetByUserID(userID int, tx ...*gorm.DB) ([]*models.JoinApplication, error)
	UpdateStatus(requestID int, status string, tx ...*gorm.DB) error
}

// CheckApplicationDAO 签到申请数据访问接口
type CheckApplicationDAO interface {
	Create(application *models.CheckApplication, tx ...*gorm.DB) error
	GetByGroupID(groupID int, tx ...*gorm.DB) ([]*models.CheckApplication, error)
	GetByUserID(userID int, tx ...*gorm.DB) ([]*models.CheckApplication, error)
	Update(status string, requestID int, tx ...*gorm.DB) error
}

