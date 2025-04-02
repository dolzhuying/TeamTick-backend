package dao

import (
	"TeamTickBackend/dal/models"
	"TeamTickBackend/global"

	"gorm.io/gorm"
)

// JoinApplicationDAOImpl 实现JoinApplicationDAO接口
type JoinApplicationDAOImpl struct{}

// NewJoinApplicationDAO 创建JoinApplicationDAO实例
func NewJoinApplicationDAO() JoinApplicationDAO {
	return &JoinApplicationDAOImpl{}
}

// Create 创建加入申请
func (dao *JoinApplicationDAOImpl) Create(application *models.JoinApplication, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.Create(application).Error
}

// GetByGroupID 通过group_id查询加入申请
func (dao *JoinApplicationDAOImpl) GetByGroupIDAndStatus(groupID int, status string, tx ...*gorm.DB) ([]*models.JoinApplication, error) {
	var applications []*models.JoinApplication
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.Where("group_id = ? AND status = ?", groupID, status).Find(&applications).Error
	if err != nil {
		return nil, err
	}
	return applications, nil
}

// GetByUserID 通过user_id查询加入申请
func (dao *JoinApplicationDAOImpl) GetByUserID(userID int, tx ...*gorm.DB) ([]*models.JoinApplication, error) {
	var applications []*models.JoinApplication
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.Where("user_id = ?", userID).Find(&applications).Error
	if err != nil {
		return nil, err
	}
	return applications, nil
}

// 更新加入申请状态(管理员审批)
func (dao *JoinApplicationDAOImpl) UpdateStatus(requestID int, status string, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.Model(&models.JoinApplication{}).
		Where("request_id=?", requestID).
		Update("status", status).Error
}
