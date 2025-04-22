package models

import (
	"time"
)

type Task struct {
	TaskID      int       `gorm:"primaryKey;column:task_id;type:int;not null;comment:签到任务id" json:"task_id"`
	TaskName    string    `gorm:"column:task_name;type:varchar(50);not null;comment:任务名称" json:"task_name"`
	Description string    `gorm:"column:description;type:varchar(512);comment:任务描述" json:"description"`
	GroupID     int       `gorm:"column:group_id;type:int;not null;index:idx_group_id;comment:任务对应用户组id" json:"group_id"`
	StartTime   time.Time `gorm:"column:start_time;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:签到开始时间" json:"start_time"`
	EndTime     time.Time `gorm:"column:end_time;type:datetime;not null;comment:签到结束时间" json:"end_time"`
	Latitude    float64   `gorm:"column:latitude;type:float;comment:任务地点（纬度）" json:"latitude"`
	Longitude   float64   `gorm:"column:longitude;type:float;comment:任务地点（经度）" json:"longitude"`
	Radius      int       `gorm:"column:radius;type:int;not null;default:50;comment:有效半径" json:"radius"`
	SSID        string    `gorm:"column:ssid;type:varchar(50);comment:wifi名称" json:"ssid"`
	BSSID       string    `gorm:"column:bssid;type:varchar(50);comment:wifi mac地址" json:"bssid"`
	TagID       string    `gorm:"column:tagid;type:varchar(50);comment:nfc标签id" json:"tagid"`
	TagName     string    `gorm:"column:tagname;type:varchar(50);comment:nfc标签名称" json:"tagname"`
	GPS         bool      `gorm:"column:gps;type:boolean;default:false;comment:gps策略" json:"gps"`
	Face        bool      `gorm:"column:face;type:boolean;default:false;comment:face策略" json:"face"`
	WiFi        bool      `gorm:"column:wifi;type:boolean;default:false;comment:wifi策略" json:"wifi"`
	NFC         bool      `gorm:"column:nfc;type:boolean;default:false;comment:nfc策略" json:"nfc"`
	CreatedAt   time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`
}

func (Task) TableName() string {
	return "tasks"
}
