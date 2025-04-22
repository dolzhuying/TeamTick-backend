package models

import (
	"time"
)

type Group struct {
	GroupID     int       `gorm:"primaryKey;column:group_id;type:int;not null;autoIncrement;comment:用户组ID" json:"group_id"`
	GroupName   string    `gorm:"column:group_name;type:varchar(50);not null;comment:用户组名称" json:"group_name"`
	Description string    `gorm:"column:description;type:varchar(1024);comment:用户组描述" json:"description"`
	CreatorID   int       `gorm:"column:creator_id;type:int;not null;index:idx_creatorid;comment:创建者用户ID" json:"creator_id"`
	CreatorName string    `gorm:"column:creator_name;type:varchar(50);not null;comment:创建者用户名" json:"creator_name"`
	MemberNum   int       `gorm:"column:member_num;type:int;not null;default:1;comment:成员数量" json:"member_num"`
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
}

func (Group) TableName() string {
	return "groups"
}
