package dao

import (
	"TeamTickBackend/dal/models"
	"TeamTickBackend/global"

	"gorm.io/gorm"
)

type TaskRecordDAOImpl struct{}

func NewTaskRecordDAO() TaskRecordDAO {
	return &TaskRecordDAOImpl{}
}

// Create 创建签到记录
func (dao *TaskRecordDAOImpl) Create(record *models.TaskRecord, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.Create(record).Error
}

// GetByTaskID 通过task_id查询组内成员签到记录
func (dao *TaskRecordDAOImpl) GetByTaskID(taskID int, tx ...*gorm.DB) ([]*models.TaskRecord, error) {
	var records []*models.TaskRecord
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.Where("task_id = ?", taskID).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// GetByUserID 通过user_id查询个人所有签到记录
func (dao *TaskRecordDAOImpl) GetByUserID(userID int, tx ...*gorm.DB) ([]*models.TaskRecord, error) {
	var records []*models.TaskRecord
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.Where("user_id = ?", userID).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}


