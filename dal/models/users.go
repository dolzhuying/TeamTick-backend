package models

import (
	"time"
)

type User struct {
	UserID        int       `gorm:"primaryKey;column:user_id;type:int;not null;autoIncrement" json:"user_id"`
	Username  string    `gorm:"column:username;type:varchar(50);not null;uniqueIndex;comment:用户名" json:"username"`
	Password  string    `gorm:"column:password;type:varchar(128);not null;comment:密码，加密存储" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
