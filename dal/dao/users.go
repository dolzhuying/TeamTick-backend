package dao

import (
	"TeamTickBackend/dal/models"
	"TeamTickBackend/global"

	"gorm.io/gorm"
)

// UserDAOImpl 实现UserDAO接口
type UserDAOImpl struct{}

// NewUserDAO 创建UserDAO实例
func NewUserDAO() UserDAO {
	return &UserDAOImpl{}
}

// Create 创建用户
func (dao *UserDAOImpl) Create(user *models.User, tx ...*gorm.DB) error {
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	return db.Create(user).Error
}

// GetByUsername 通过username查询用户信息
func (dao *UserDAOImpl) GetByUsername(username string, tx ...*gorm.DB) (*models.User, error) {
	var user models.User
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID 通过id查询用户信息
func (dao *UserDAOImpl) GetByID(id int, tx ...*gorm.DB) (*models.User, error) {
	var user models.User
	db := global.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	}
	err := db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
