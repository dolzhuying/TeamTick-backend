package impl

import (
	"TeamTickBackend/dal/models"
	"TeamTickBackend/global"
	"context"

	"gorm.io/gorm"
)

type MySQLJoinApplicationDAOImpl struct{}

// Create 创建加入申请
func (dao *MySQLJoinApplicationDAOImpl) Create(ctx context.Context, application *models.JoinApplication, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Create(application).Error
}

// GetByGroupIDAndStatus 通过group_id和status查询加入申请
func (dao *MySQLJoinApplicationDAOImpl) GetByGroupIDAndStatus(ctx context.Context, groupID int, status string, tx ...*gorm.DB) ([]*models.JoinApplication, error) {
	var applications []*models.JoinApplication
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("group_id = ? AND status = ?", groupID, status).Find(&applications).Error
	if err != nil {
		return nil, err
	}
	return applications, nil
}

// GetByUserID 通过user_id查询加入申请
func (dao *MySQLJoinApplicationDAOImpl) GetByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.JoinApplication, error) {
	var applications []*models.JoinApplication
	db := global.DB
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
func (dao *MySQLJoinApplicationDAOImpl) UpdateStatus(ctx context.Context, requestID int, status string, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Model(&models.JoinApplication{}).
		Where("request_id=?", requestID).
		Update("status", status).Error
}
