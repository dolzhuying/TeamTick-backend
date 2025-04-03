package impl

import (
	"TeamTickBackend/dal/models"
	"TeamTickBackend/global"
	"context"

	"gorm.io/gorm"
)

type MySQLGroupMemberDAOImpl struct{}

// Create 创建组员
func (dao *MySQLGroupMemberDAOImpl) Create(ctx context.Context, member *models.GroupMember, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Create(member).Error
}

// GetMembersByGroupID 通过group_id查询组中的所有成员信息
func (dao *MySQLGroupMemberDAOImpl) GetMembersByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) ([]*models.GroupMember, error) {
	var members []*models.GroupMember
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("group_id = ?", groupID).Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

// GetMemberByGroupIDAndUserID 通过group_id和user_id查询特定组员信息
func (dao *MySQLGroupMemberDAOImpl) GetMemberByGroupIDAndUserID(ctx context.Context, groupID int, userID int, tx ...*gorm.DB) (*models.GroupMember, error) {
	var member models.GroupMember
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// Delete 删除组员
func (dao *MySQLGroupMemberDAOImpl) Delete(ctx context.Context, groupID int, userID int, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&models.GroupMember{}).Error
}
