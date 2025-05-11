package mysqlImpl

import (
	"TeamTickBackend/dal/models"
	"context"

	"gorm.io/gorm"
)

type JoinApplicationDAOMySQLImpl struct{
	DB *gorm.DB
}

// Create 创建加入申请
func (dao *JoinApplicationDAOMySQLImpl) Create(ctx context.Context, application *models.JoinApplication, tx ...*gorm.DB) error {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Create(application).Error
}

// GetByGroupIDAndStatus 通过group_id和status查询加入申请
func (dao *JoinApplicationDAOMySQLImpl) GetByGroupIDAndStatus(ctx context.Context, groupID int, status string, tx ...*gorm.DB) ([]*models.JoinApplication, error) {
	var applications []*models.JoinApplication
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("group_id = ? AND status = ?", groupID, status).Find(&applications).Error
	if err != nil {
		return nil, err
	}
	return applications, nil
}

// GetByGroupID 通过group_id查询所有加入申请
func (dao *JoinApplicationDAOMySQLImpl) GetByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.JoinApplication, error) {
	var applications []*models.JoinApplication
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("group_id = ?", groupID).Find(&applications).Error
	if err != nil {
		return nil, err
	}
	return applications, nil
}

// GetByUserID 通过user_id查询加入申请
func (dao *JoinApplicationDAOMySQLImpl) GetByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.JoinApplication, error) {
	var applications []*models.JoinApplication
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

// UpdateStatus 更新加入申请状态(管理员审批)
func (dao *JoinApplicationDAOMySQLImpl) UpdateStatus(ctx context.Context, requestID int, status string, tx ...*gorm.DB) error {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Model(&models.JoinApplication{}).
		Where("request_id=?", requestID).
		Update("status", status).Error
}

// UpdateRejectReason 更新拒绝理由
func (dao *JoinApplicationDAOMySQLImpl) UpdateRejectReason(ctx context.Context, requestID int, rejectReason string, tx ...*gorm.DB) error {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Model(&models.JoinApplication{}).
		Where("request_id=?", requestID).
		Update("reject_reason", rejectReason).Error
}

// GetByGroupIDAndUserID 通过group_id和user_id查询加入申请
func (dao *JoinApplicationDAOMySQLImpl) GetByGroupIDAndUserID(ctx context.Context, groupID int, userID int, tx ...*gorm.DB) (*models.JoinApplication, error) {
	var application models.JoinApplication
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("group_id = ? AND user_id = ?", groupID, userID).First(&application).Error
	if err != nil {
		return nil, err
	}
	return &application, nil
}

// GetByRequestID 通过request_id查询加入申请
func (dao *JoinApplicationDAOMySQLImpl) GetByRequestID(ctx context.Context, requestID int, tx ...*gorm.DB) (*models.JoinApplication, error) {
	var application models.JoinApplication
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("request_id = ?", requestID).First(&application).Error
	if err != nil {
		return nil, err
	}
	return &application, nil
}
