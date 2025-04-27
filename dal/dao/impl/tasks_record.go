package impl

import (
	"TeamTickBackend/dal/models"
	"context"

	"gorm.io/gorm"
)

type TaskRecordDAOMySQLImpl struct {
	DB *gorm.DB
}

// Create 创建签到记录
func (dao *TaskRecordDAOMySQLImpl) Create(ctx context.Context, record *models.TaskRecord, tx ...*gorm.DB) error {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Create(record).Error
}

// GetByTaskID 通过task_id查询组内成员签到记录
func (dao *TaskRecordDAOMySQLImpl) GetByTaskID(ctx context.Context, taskID int, tx ...*gorm.DB) ([]*models.TaskRecord, error) {
	var records []*models.TaskRecord
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("task_id = ?", taskID).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// GetByUserID 通过user_id查询个人所有签到记录
func (dao *TaskRecordDAOMySQLImpl) GetByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.TaskRecord, error) {
	var records []*models.TaskRecord
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("user_id = ?", userID).Find(&records).Error
	if err != nil {
		return nil, err
	}
	return records, nil
}

// GetByTaskIDAndUserID 通过task_id和user_id查询指定签到记录
func (dao*TaskRecordDAOMySQLImpl) GetByTaskIDAndUserID(ctx context.Context,taskID,userID int,tx ...*gorm.DB) (*models.TaskRecord,error){
	var record models.TaskRecord
	db:=dao.DB
	if len(tx)>0&&tx[0]!=nil{
		db=tx[0]
	}
	err:=db.WithContext(ctx).Where("task_id=? AND user_id=?",taskID,userID).First(&record).Error
	if err!=nil{
		return nil,err
	}
	return &record,nil
}
