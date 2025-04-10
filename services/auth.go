package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	"TeamTickBackend/pkg"
	apperrors "TeamTickBackend/pkg/errors"
	"context"
	"errors"

	"gorm.io/gorm"
)

type AuthService struct {
	userDao dao.UserDAO
	transactionManager dao.TransactionManager
	jwtHandler pkg.JwtHandler
}

func NewAuthService(
	userDao dao.UserDAO,
	transactionManager dao.TransactionManager,
	jwtHandler pkg.JwtHandler,
) *AuthService {
	return &AuthService{
		userDao: userDao,
		transactionManager: transactionManager,
		jwtHandler: jwtHandler,
	}
}

func (s *AuthService) AuthRegister(ctx context.Context, username, password string) (*models.User, error) {
	var createdUser models.User

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查用户是否已存在
		user, err := s.userDao.GetByUsername(ctx, username, tx)
		if err == nil && user != nil {
			return apperrors.ErrUserAlreadyExists
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.ErrDatabaseOperation.WithError(err)
		}
		//加密密码
		hashedPassword, err := pkg.GenerateFromPassword(password)
		if err != nil {
			return apperrors.ErrPasswordEncryption.WithError(err)
		}
		newUser := models.User{
			Username: username,
			Password: hashedPassword,
		}
		//创建用户
		if err := s.userDao.Create(ctx, &newUser, tx); err != nil {
			return apperrors.ErrUserCreationFailed.WithError(err)
		}
		createdUser = newUser
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &createdUser, nil

}

func (s *AuthService) AuthLogin(ctx context.Context, username, password string) (*models.User, string, error) {
	var existUser models.User
	var userToken string

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查用户是否存在
		user, err := s.userDao.GetByUsername(ctx, username, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrUserNotFound
			}
			return apperrors.ErrDatabaseOperation.WithError(err)
		}
		//检查密码是否正确
		if !pkg.CheckPassword(user.Password, password) {
			return apperrors.ErrInvalidPassword
		}
		existUser = *user
		//生成token
		token, err := s.jwtHandler.GenerateJWTToken(user.Username, user.UserID)
		if err != nil {
			return apperrors.ErrTokenGenerationFailed.WithError(err)
		}
		userToken = token

		return nil
	})
	if err != nil {
		return nil, "", err
	}
	return &existUser, userToken, nil
}
