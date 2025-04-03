package models

import (
	"time"
)

type TaskRecord struct {
	RecordID   int       `gorm:"primaryKey;column:record_id;type:int;not null;autoIncrement" json:"record_id"`
	TaskID     int       `gorm:"column:task_id;type:int;not null;index:idx_task_user_id,priority:1;comment:签到任务id" json:"task_id"`
	UserID     int       `gorm:"column:user_id;type:int;not null;index:idx_task_user_id,priority:2;comment:签到用户id" json:"user_id"`
	Username   string    `gorm:"column:username;type:varchar(50);not null;comment:签到用户名" json:"username"`
	SignedTime time.Time `gorm:"column:signed_time;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:签到时间" json:"signed_time"`
	Location   string    `gorm:"column:location;type:point;not null;comment:签到地点（经纬度）" json:"location"`
	Status     int       `gorm:"column:status;type:int;not null;default:1;comment:签到状态，1表示正常校验成功，2表示人工审核通过" json:"status"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
}

func (TaskRecord) TableName() string {
	return "tasks_record"
}
