package dao

import (
	"TeamTickBackend/dal/models"
	"TeamTickBackend/global"

	"gorm.io/gorm"
)

type GroupMemberDAOImpl struct{}

func NewGroupMemberDAO() GroupMemberDAO {
	return &GroupMemberDAOImpl{}
}

// Create 创建组员
func (dao *GroupMemberDAOImpl) Create(member *models.GroupMember, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.Create(member).Error
}

// GetByGroupID 通过group_id查询组中的所有成员信息
func (dao *GroupMemberDAOImpl) GetMembersByGroupID(groupID int, tx ...*gorm.DB) ([]*models.GroupMember, error) {
	var members []*models.GroupMember
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.Where("group_id = ?", groupID).Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

// GetByUserID 通过user_id查询用户加入的所有组的成员信息
func (dao *GroupMemberDAOImpl) GetGroupsByUserID(userID int, tx ...*gorm.DB) ([]*models.GroupMember, error) {
	var members []*models.GroupMember
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.Where("user_id = ?", userID).Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

// Delete 删除组员
func (dao *GroupMemberDAOImpl) Delete(groupID int, userID int, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&models.GroupMember{}).Error
}

