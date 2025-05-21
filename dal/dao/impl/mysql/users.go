package mysqlImpl

import (
	"TeamTickBackend/dal/models"
	"context"

	"gorm.io/gorm"
)

type UserDAOMySQLImpl struct {
	DB *gorm.DB
}

// Create 创建用户
func (dao *UserDAOMySQLImpl) Create(ctx context.Context, user *models.User, tx ...*gorm.DB) error {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Create(user).Error
}

// GetByUsername 通过username查询用户信息
func (dao *UserDAOMySQLImpl) GetByUsername(ctx context.Context, username string, tx ...*gorm.DB) (*models.User, error) {
	var user models.User
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID 通过id查询用户信息
func (dao *UserDAOMySQLImpl) GetByID(ctx context.Context, id int, tx ...*gorm.DB) (*models.User, error) {
	var user models.User
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// 更新密码
func (dao *UserDAOMySQLImpl) UpdatePassword(ctx context.Context, userID int, password string, tx ...*gorm.DB) error {
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.WithContext(ctx).Model(&models.User{}).Where("user_id = ?", userID).Update("password", password).Error
}

// GetByEmail 通过email查询用户信息
func (dao *UserDAOMySQLImpl) GetByEmail(ctx context.Context, email string, tx ...*gorm.DB) (*models.User, error) {
	var user models.User
	db := dao.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.WithContext(ctx).Where("mail = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

