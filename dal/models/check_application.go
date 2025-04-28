package models

import (
	"time"
)

type CheckApplication struct {
	ID            int        `gorm:"primaryKey;column:id;type:int;not null;autoIncrement" json:"id"`
	GroupID       int        `gorm:"column:group_id;type:int;not null;comment:用户组id" json:"group_id"`
	TaskID        int        `gorm:"column:task_id;type:int;not null;index:idx_taskid_status;comment:签到任务id" json:"task_id"`
	TaskName      string     `gorm:"column:task_name;type:varchar(50);not null;comment:任务名称" json:"task_name"`
	UserID        int        `gorm:"column:user_id;type:int;not null;comment:申请用户id" json:"user_id"`
	Username      string     `gorm:"column:username;type:varchar(50);not null;comment:申请用户名" json:"username"`
	RequestAt     time.Time  `gorm:"column:request_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:申请时间" json:"request_at"`
	Status        string     `gorm:"column:status;type:enum('pending','approved','rejected');default:pending;index:idx_taskid_status;comment:审核状态" json:"status"`
	Reason        string     `gorm:"column:reason;type:varchar(1024);not null;comment:申请人工审核原因" json:"reason"`
	Image         string     `gorm:"column:image;type:mediumtext;comment:辅佐照片材料" json:"image"`
	AdminID       int        `gorm:"column:admin_id;type:int;not null;comment:处理管理员ID" json:"admin_id"`
	AdminUsername string     `gorm:"column:admin_username;type:varchar(50);not null;comment:处理管理员用户名" json:"admin_username"`
	ProcessedAt   *time.Time `gorm:"column:processed_at;type:datetime;comment:处理时间" json:"processed_at"`
	CreatedAt     time.Time  `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
}

func (CheckApplication) TableName() string {
	return "check_application"
}
