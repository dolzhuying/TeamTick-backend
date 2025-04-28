package impl

import (
	"TeamTickBackend/dal/models"
	"context"

	"gorm.io/gorm"
)

type GroupDAOMySQLImpl struct {
	DB *gorm.DB
}

// Create 创建组
func (dao *GroupDAOMySQLImpl) Create(ctx context.Context, group *models.Group, tx ...*gorm.DB) error {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Create(group).Error
}

// GetByGroupID 通过group_id查询组信息
func (dao *GroupDAOMySQLImpl) GetByGroupID(ctx context.Context, groupID int, tx ...*gorm.DB) (*models.Group, error) {
	var group models.Group
	db := dao.DB
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
func (dao *GroupDAOMySQLImpl) GetGroupsByUserID(ctx context.Context, userID int, tx ...*gorm.DB) ([]*models.Group, error) {
	var groups []*models.Group
	db := dao.DB
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

// GetGroupsByUserIDAndfilter 通过user_id和filter获取用户所在的所有用户组
func (dao *GroupDAOMySQLImpl) GetGroupsByUserIDAndfilter(ctx context.Context, userID int, filter string, tx ...*gorm.DB) ([]*models.Group, error) {
	var groups []*models.Group
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	if filter == "created" {
		role := "admin"
		err := db.WithContext(ctx).Table("`groups` g").
			Select("g.*").
			Joins("JOIN group_member gm ON g.group_id = gm.group_id").
			Where("gm.user_id = ? AND gm.role = ?", userID, role).
			Find(&groups).Error
		if err != nil {
			return nil, err
		}
	} else if filter == "joined" {
		role := "member"
		err := db.WithContext(ctx).Table("`groups` g").
			Select("g.*").
			Joins("JOIN group_member gm ON g.group_id = gm.group_id").
			Where("gm.user_id = ? AND gm.role = ?", userID, role).
			Find(&groups).Error
		if err != nil {
			return nil, err
		}
	}
	return groups, nil
}

// UpdateMessage 更新组信息
func (dao *GroupDAOMySQLImpl) UpdateMessage(ctx context.Context, groupID int, groupName, description string, tx ...*gorm.DB) error {
	db := dao.DB
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
func (dao *GroupDAOMySQLImpl) UpdateMemberNum(ctx context.Context, groupID int, increment bool, tx ...*gorm.DB) error {
	db := dao.DB
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

// 删除用户组
func (dao *GroupDAOMySQLImpl) Delete(ctx context.Context, groupID int, tx ...*gorm.DB) error {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Where("group_id = ?", groupID).Delete(&models.Group{}).Error
}
