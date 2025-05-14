package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	appErrors "TeamTickBackend/pkg/errors"
	"TeamTickBackend/pkg/logger"
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserService struct {
	userDao            dao.UserDAO
	transactionManager dao.TransactionManager
}

func NewUserService(
	userDao dao.UserDAO,
	transactionManager dao.TransactionManager,
) *UserService {
	return &UserService{
		userDao:            userDao,
		transactionManager: transactionManager,
	}
}

func (s *UserService) GetUserMe(ctx context.Context, userID int) (*models.User, error) {
	var existUser models.User
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		user, err := s.userDao.GetByID(ctx, userID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("获取用户信息失败：用户不存在",
					zap.Int("userID", userID),
					zap.String("operation", "GetByID"),
					zap.Error(err),
				)
				return appErrors.ErrUserNotFound
			}
			logger.Error("获取用户信息失败：数据库操作错误",
				zap.Int("userID", userID),
				zap.String("operation", "GetByID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		existUser = *user
		logger.Info("成功获取用户信息",
			zap.Int("userID", userID),
			zap.String("username", existUser.Username),
			zap.Time("createTime", existUser.CreatedAt),
			zap.Time("updateTime", existUser.UpdatedAt),
			zap.String("operation", "GetUserMe"),
		)
		return nil
	})
	if err != nil {
		logger.Error("获取用户信息事务失败",
			zap.Int("userID", userID),
			zap.String("operation", "GetUserMeTransaction"),
			zap.Error(err),
		)
		return nil, err
	}
	return &existUser, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context,username string)(*models.User,error){
	var user *models.User
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		User, err := s.userDao.GetByUsername(ctx, username, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrUserNotFound
			}
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		user=User
		return nil
	})
	if err!=nil{
		return nil,err
	}
	return user,nil
}
