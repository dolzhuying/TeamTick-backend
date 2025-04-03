package models

import (
	"time"
)

type CheckApplication struct {
	ID          int       `gorm:"primaryKey;column:id;type:int;not null;autoIncrement" json:"id"`
	TaskID      int       `gorm:"column:task_id;type:int;not null;index:idx_taskid_status;comment:签到任务id" json:"task_id"`
	TaskName    string    `gorm:"column:task_name;type:varchar(50);not null;comment:任务名称" json:"task_name"`
	UserID      int       `gorm:"column:user_id;type:int;not null;comment:申请用户id" json:"user_id"`
	Username    string    `gorm:"column:username;type:varchar(50);not null;comment:申请用户名" json:"username"`
	AppliedTime time.Time `gorm:"column:applied_time;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:申请时间" json:"applied_time"`
	Status      string    `gorm:"column:status;type:enum('pending','passed','rejected');default:pending;index:idx_taskid_status;comment:审核状态" json:"status"`
	Description string    `gorm:"column:description;type:varchar(1024);not null;comment:申请人工审核原因" json:"description"`
	Image       string    `gorm:"column:image;type:mediumtext;comment:辅佐照片材料" json:"image"`
}

func (CheckApplication) TableName() string {
	return "check_application"
}
