package impl

import (
	"TeamTickBackend/dal/models"
	"TeamTickBackend/global"
	"context"

	"gorm.io/gorm"
)

type MySQLGroupDAOImpl struct{}

// Create 创建组
func (dao *MySQLGroupDAOImpl) Create(ctx context.Context, group *models.Group, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Create(group).Error
}

// GetByGroupID 通过group_id查询组信息
func (dao *MySQLGroupDAOImpl) GetByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) (*models.Group, error) {
	var group models.Group
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("group_id = ?", groupID).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// GetGroupsByUserID 通过user_id获取用户所在的所有用户组
func (dao *MySQLGroupDAOImpl) GetGroupsByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.Group, error) {
	var groups []*models.Group
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Table("groups g").
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
func (dao *MySQLGroupDAOImpl) UpdateMessage(ctx context.Context, groupID int, groupName, description string, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).
		Model(&models.Group{}).
		Where("group_id = ?", groupID).
		Update("group_name", groupName).
		Update("description", description).Error
}

// UpdateMemberNum 更新组成员数量
func (dao *MySQLGroupDAOImpl) UpdateMemberNum(ctx context.Context, groupID int, increment bool, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}

	expr := "member_num + 1"
	if !increment {
		expr = "member_num - 1"
	}

	return db.WithContext(ctx).
		Model(&models.Group{}).
		Where("group_id = ?", groupID).
		Update("member_num", gorm.Expr(expr)).Error
}
