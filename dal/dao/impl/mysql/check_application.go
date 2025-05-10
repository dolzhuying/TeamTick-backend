package mysqlImpl

import (
	"TeamTickBackend/dal/models"
	"context"

	"gorm.io/gorm"
)

type CheckApplicationDAOMySQLImpl struct{
	DB *gorm.DB
}

// Create 创建签到申请
func (dao *CheckApplicationDAOMySQLImpl) Create(ctx context.Context, application *models.CheckApplication, tx ...*gorm.DB) error {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Create(application).Error
}

// GetByID 通过id查询签到申请
func (dao *CheckApplicationDAOMySQLImpl) GetByID(ctx context.Context, id int, tx ...*gorm.DB) (*models.CheckApplication, error) {
	var application models.CheckApplication
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("id = ?", id).First(&application).Error
	if err != nil {
		return nil, err
	}
	return &application, nil
}


// 通过group_id查询当前组的所有任务签到申请
func (dao *CheckApplicationDAOMySQLImpl) GetByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.CheckApplication, error) {
	var checkApplications []*models.CheckApplication
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("group_id = ?", groupID).Find(&checkApplications).Error
	if err != nil {
		return nil, err
	}
	return checkApplications, nil
}

// GetByUserID 通过user_id查询签到申请
func (dao *CheckApplicationDAOMySQLImpl) GetByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.CheckApplication, error) {
	var applications []*models.CheckApplication
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("user_id = ?", userID).Find(&applications).Error
	if err != nil {
		return nil, err
	}
	return applications, nil
}

// Update 更新签到申请（管理员审批）
func (dao *CheckApplicationDAOMySQLImpl) Update(ctx context.Context, status string, requestID int, tx ...*gorm.DB) error {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).
		Model(&models.CheckApplication{}).
		Where("id = ?", requestID).
		Update("status", status).Error
}

func (dao *CheckApplicationDAOMySQLImpl) GetByTaskIDAndUserID(ctx context.Context, taskID int, userID int, tx ...*gorm.DB) (*models.CheckApplication, error) {
	var application models.CheckApplication
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("task_id = ? AND user_id = ?", taskID, userID).First(&application).Error
	if err != nil {
		return nil, err
	}
	return &application, nil
}
