package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	apperrors "TeamTickBackend/pkg/errors"
	"context"
	"errors"

	"gorm.io/gorm"
)

type UserService struct {
	userDao dao.UserDAO
	transactionManager dao.TransactionManager
}

func NewUserService(
	userDao dao.UserDAO,
	transactionManager dao.TransactionManager,
) *UserService {
	return &UserService{
		userDao: userDao,
		transactionManager: transactionManager,
	}
}

func (s *UserService) GetUserMe(ctx context.Context, userID int) (*models.User, error) {
	var existUser models.User
	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		user, err := s.userDao.GetByID(ctx, userID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrUserNotFound
			}
			return apperrors.ErrDatabaseOperation.WithError(err)
		}
		existUser = *user
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &existUser, nil
}
