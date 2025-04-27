package models

import (
	"time"
)

type TaskRecord struct {
	RecordID   int       `gorm:"primaryKey;column:record_id;type:int;not null;autoIncrement" json:"record_id"`
	TaskID     int       `gorm:"column:task_id;type:int;not null;index:idx_task_user_id,priority:1;comment:签到任务id" json:"task_id"`
	TaskName   string    `gorm:"column:task_name;type:varchar(50);not null;comment:任务名称" json:"task_name"`
	GroupID    int       `gorm:"column:group_id;type:int;not null;comment:任务对应用户组id" json:"group_id"`
	GroupName  string    `gorm:"column:group_name;type:varchar(50);not null;comment:用户组名称" json:"group_name"`
	UserID     int       `gorm:"column:user_id;type:int;not null;index:idx_task_user_id,priority:2;index:idx_userid;comment:签到用户id" json:"user_id"`
	Username   string    `gorm:"column:username;type:varchar(50);not null;comment:签到用户名" json:"username"`
	SignedTime time.Time `gorm:"column:signed_time;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:签到时间" json:"signed_time"`
	Latitude   float64   `gorm:"column:latitude;type:float;comment:签到地点（纬度）" json:"latitude"`
	Longitude  float64   `gorm:"column:longitude;type:float;comment:签到地点（经度）" json:"longitude"`
	FaceData   string    `gorm:"column:face_data;type:mediumtext;comment:人脸识别数据" json:"face_data"`
	SSID       string    `gorm:"column:ssid;type:varchar(50);comment:wifi名称" json:"ssid"`
	BSSID      string    `gorm:"column:bssid;type:varchar(50);comment:wifi mac地址" json:"bssid"`
	TagID      string    `gorm:"column:tagid;type:varchar(50);comment:nfc标签id" json:"tagid"`
	TagName    string    `gorm:"column:tagname;type:varchar(50);comment:nfc标签名称" json:"tagname"`
	Status     int       `gorm:"column:status;type:int;not null;default:1;comment:签到状态，1表示正常校验成功，2表示人工审核通过" json:"status"`
	CreatedAt  time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
}

func (TaskRecord) TableName() string {
	return "tasks_record"
}
