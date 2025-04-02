package dao

import (
	"TeamTickBackend/dal/models"
	"TeamTickBackend/global"

	"gorm.io/gorm"
)

type CheckApplicationDAOImpl struct{}

func NewCheckApplicationDAO() CheckApplicationDAO {
	return &CheckApplicationDAOImpl{}
}

// Create 创建签到申请
func (dao *CheckApplicationDAOImpl) Create(application *models.CheckApplication, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.Create(application).Error
}

// 通过group_id查询当前组的所有任务签到申请
func (dao *CheckApplicationDAOImpl) GetByGroupID(groupID int, tx ...*gorm.DB) ([]*models.CheckApplication, error) {
	var checkApplications []*models.CheckApplication
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.Where("group_id = ?", groupID).Find(&checkApplications).Error
	if err != nil {
		return nil, err
	}
	return checkApplications, nil
}

// GetByUserID 通过user_id查询签到申请
func (dao *CheckApplicationDAOImpl) GetByUserID(userID int, tx ...*gorm.DB) ([]*models.CheckApplication, error) {
	var applications []*models.CheckApplication
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

// Update 更新签到申请（管理员审批）
func (dao *CheckApplicationDAOImpl) Update(status string, requestID int, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.
		Model(&models.CheckApplication{}).
		Where("id = ?", requestID).
		Update("status", status).Error
}

