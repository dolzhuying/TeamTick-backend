package models

import (
	"time"
)

type JoinApplication struct {
	RequestID    int       `gorm:"primaryKey;column:request_id;type:int;not null;autoIncrement" json:"request_id"`
	GroupID      int       `gorm:"column:group_id;type:int;not null;index:idx_groupid_status;uniqueIndex:idx_groupid_userid;comment:用户组id" json:"group_id"`
	UserID       int       `gorm:"column:user_id;type:int;not null;uniqueIndex:idx_groupid_userid;comment:申请用户id" json:"user_id"`
	Username     string    `gorm:"column:username;type:varchar(50);not null;comment:申请用户名" json:"username"`
	Reason       string    `gorm:"column:reason;type:varchar(512);not null;comment:申请理由" json:"reason"`
	RejectReason string    `gorm:"column:reject_reason;type:varchar(512);comment:拒绝理由" json:"reject_reason"`
	Status       string    `gorm:"column:status;type:enum('pending','accepted','rejected');default:pending;index:idx_groupid_status;comment:审核状态" json:"status"`
	CreatedAt    time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:申请时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
}

func (JoinApplication) TableName() string {
	return "join_application"
}
