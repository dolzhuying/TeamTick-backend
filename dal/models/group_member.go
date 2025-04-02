package models

import (
	"time"
)

type GroupMember struct {
	GroupID   int       `gorm:"column:group_id;type:int;not null;index:idx_groupid_userid_role,priority:1;index:idx_groupid;comment:用户组ID" json:"group_id"`
	UserID    int       `gorm:"column:user_id;type:int;not null;index:idx_groupid_userid_role,priority:2;index:idx_userid;comment:用户ID" json:"user_id"`
	GroupName string    `gorm:"column:group_name;type:varchar(50);not null;comment:用户组名" json:"group_name"`
	Username  string    `gorm:"column:username;type:varchar(50);not null;comment:用户名" json:"username"`
	Role      string    `gorm:"column:role;type:enum('admin','member');not null;default:member;index:idx_groupid_userid_role,priority:3;comment:角色" json:"role"`
	JoinedAt  time.Time `gorm:"column:joined_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:加入时间" json:"joined_at"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
}

func (GroupMember) TableName() string {
	return "group_member"
}
