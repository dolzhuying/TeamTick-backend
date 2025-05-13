package models

type AbsentRecord struct {
    GroupMember
    TaskID int64 `gorm:"column:task_id"`
}