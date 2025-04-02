package dao

import (
	"TeamTickBackend/dal/models"
	"TeamTickBackend/global"

	"gorm.io/gorm"
)

type GroupDAOImpl struct{}

func NewGroupDAO() GroupDAO {
	return &GroupDAOImpl{}
}

// Create 创建组
func (dao *GroupDAOImpl) Create(group *models.Group, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.Create(group).Error
}

// GetByID 通过group_id查询组信息
func (dao *GroupDAOImpl) GetByID(id int, tx ...*gorm.DB) (*models.Group, error) {
	var group models.Group
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.Where("group_id = ?", id).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetByUserID 获取用户所在的所有用户组
func (dao *GroupDAOImpl) GetByUserID(userID int, tx ...*gorm.DB) ([]*models.Group, error) {
	var groups []*models.Group
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.Table("groups g").
		Select("g.*").
		Joins("JOIN group_member gm ON g.group_id = gm.group_id").
		Where("gm.user_id = ?", userID).
		Find(&groups).Error
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// UpdateMessage 更新组信息
func (dao *GroupDAOImpl) UpdateMessage(groupID int, groupName, description string, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.
		Model(&models.Group{}).
		Where("group_id = ?", groupID).
		Update("group_name", groupName).
		Update("description", description).Error
}

// UpdateMemberNum 更新组成员数量
func (dao *GroupDAOImpl) UpdateMemberNum(groupID int, increment bool, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}

	expr := "member_num + 1"
	if !increment {
		expr = "member_num - 1"
	}

	return db.
		Model(&models.Group{}).
		Where("group_id = ?", groupID).
		Update("member_num", gorm.Expr(expr)).Error
}

