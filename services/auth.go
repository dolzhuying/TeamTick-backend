package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	"TeamTickBackend/pkg"
	appErrors "TeamTickBackend/pkg/errors"
	"TeamTickBackend/pkg/logger"
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthService struct {
	userDao            dao.UserDAO
	transactionManager dao.TransactionManager
	jwtHandler         pkg.JwtHandler
	emailRedisDAO      dao.EmailRedisDAO
}

func NewAuthService(
	userDao dao.UserDAO,
	transactionManager dao.TransactionManager,
	jwtHandler pkg.JwtHandler,
	emailRedisDAO dao.EmailRedisDAO,
) *AuthService {
	return &AuthService{
		userDao:            userDao,
		transactionManager: transactionManager,
		jwtHandler:         jwtHandler,
		emailRedisDAO:      emailRedisDAO,
	}
}

func (s *AuthService) AuthRegister(ctx context.Context, username, password, email, verificationCode string) (*models.User, error) {
	var createdUser models.User

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查用户是否已存在
		user, err := s.userDao.GetByUsername(ctx, username, tx)
		if err == nil && user != nil {
			logger.Error("用户注册失败：用户已存在",
				zap.String("username", username),
				zap.Int("existingUserID", user.UserID),
				zap.Time("existingUserCreateTime", user.CreatedAt),
				zap.Time("existingUserUpdateTime", user.UpdatedAt),
				zap.Error(err),
			)
			return appErrors.ErrUserAlreadyExists
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Error("用户注册失败：数据库操作错误",
				zap.String("username", username),
				zap.String("operation", "GetByUsername"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		// 检查邮箱验证码
		code, err := s.emailRedisDAO.GetVerificationCodeByEmail(ctx, email)
		if err == nil && len(code) > 0 {
			if code != verificationCode {
				logger.Error("用户注册失败：邮箱验证码错误",
					zap.String("email", email),
					zap.String("operation", "GetVerificationCodeByEmail"),
					zap.Error(err),
				)
				return appErrors.ErrInvalidVerificationCode
			}
		} else if err != nil {
			logger.Error("用户注册失败：获取邮箱验证码错误",
				zap.String("email", email),
				zap.String("operation", "GetVerificationCodeByEmail"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		//加密密码
		hashedPassword, err := pkg.GenerateFromPassword(password)
		if err != nil {
			logger.Error("用户注册失败：密码加密错误",
				zap.String("username", username),
				zap.String("operation", "GenerateFromPassword"),
				zap.Error(err),
			)
			return appErrors.ErrPasswordEncryption.WithError(err)
		}
		newUser := models.User{
			Username: username,
			Password: hashedPassword,
			Mail:     email,
		}
		//创建用户
		if err := s.userDao.Create(ctx, &newUser, tx); err != nil {
			logger.Error("用户注册失败：创建用户错误",
				zap.String("username", username),
				zap.String("operation", "Create"),
				zap.Error(err),
			)
			return appErrors.ErrUserCreationFailed.WithError(err)
		}
		createdUser = newUser
		logger.Info("用户注册成功",
			zap.String("username", username),
			zap.Int("userID", createdUser.UserID),
			zap.Time("createTime", createdUser.CreatedAt),
			zap.Time("updateTime", createdUser.UpdatedAt),
			zap.String("operation", "Register"),
		)
		return nil
	})
	if err != nil {
		logger.Error("用户注册事务失败",
			zap.String("username", username),
			zap.String("operation", "RegisterTransaction"),
			zap.Error(err),
		)
		return nil, err
	}
	if err := s.emailRedisDAO.DeleteCacheByEmail(ctx, email); err != nil {
		logger.Error("删除邮箱验证码缓存错误",
			zap.String("email", email),
			zap.String("operation", "DeleteCacheByEmail"),
			zap.Error(err),
		)
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
				logger.Error("用户登录失败：用户不存在",
					zap.String("username", username),
					zap.String("operation", "GetByUsername"),
					zap.Error(err),
				)
				return appErrors.ErrUserNotFound
			}
			logger.Error("用户登录失败：数据库操作错误",
				zap.String("username", username),
				zap.String("operation", "GetByUsername"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		//检查密码是否正确
		if !pkg.CheckPassword(user.Password, password) {
			logger.Error("用户登录失败：密码错误",
				zap.String("username", username),
				zap.Int("userID", user.UserID),
				zap.Time("lastLoginTime", user.UpdatedAt),
				zap.String("operation", "CheckPassword"),
				zap.Error(err),
			)
			return appErrors.ErrInvalidPassword
		}
		existUser = *user
		//生成token
		token, err := s.jwtHandler.GenerateJWTToken(user.Username, user.UserID)
		if err != nil {
			logger.Error("用户登录失败：Token生成错误",
				zap.String("username", username),
				zap.Int("userID", user.UserID),
				zap.Time("lastLoginTime", user.UpdatedAt),
				zap.String("operation", "GenerateJWTToken"),
				zap.Error(err),
			)
			return appErrors.ErrTokenGenerationFailed.WithError(err)
		}
		userToken = token

		logger.Info("用户登录成功",
			zap.String("username", username),
			zap.Int("userID", existUser.UserID),
			zap.Time("lastLoginTime", existUser.UpdatedAt),
			zap.Time("createTime", existUser.CreatedAt),
			zap.String("operation", "Login"),
		)
		return nil
	})
	if err != nil {
		logger.Error("用户登录事务失败",
			zap.String("username", username),
			zap.String("operation", "LoginTransaction"),
			zap.Error(err),
		)
		return nil, "", err
	}
	return &existUser, userToken, nil
}

// 更新密码(内部检查邮箱验证码作为校验)
func (s *AuthService) AuthUpdatePassword(ctx context.Context, userID int, password, verificationCode, email string) error {
	//检查邮箱验证码
	code, err := s.emailRedisDAO.GetVerificationCodeByEmail(ctx, email)
	if err == nil && len(code) > 0 {
		if code != verificationCode {
			logger.Error("更新密码失败：邮箱验证码错误",
				zap.String("email", email),
				zap.String("operation", "GetVerificationCodeByEmail"),
				zap.Error(err),
			)
			return appErrors.ErrInvalidVerificationCode
		}
	} else if err != nil {
		logger.Error("更新密码失败：获取邮箱验证码错误",
			zap.String("email", email),
			zap.String("operation", "GetVerificationCodeByEmail"),
			zap.Error(err),
		)
		return appErrors.ErrDatabaseOperation.WithError(err)
	}
	//更新密码
	err = s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查用户是否存在
		_, err := s.userDao.GetByID(ctx, userID, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("更新密码失败：用户不存在",
					zap.Int("userID", userID),
					zap.String("operation", "GetByID"),
					zap.Error(err),
				)
				return appErrors.ErrUserNotFound
			}
			logger.Error("更新密码失败：数据库操作错误",
				zap.Int("userID", userID),
				zap.String("operation", "GetByID"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}

		//加密密码
		hashedPassword, err := pkg.GenerateFromPassword(password)
		if err != nil {
			logger.Error("更新密码失败：密码加密错误",
				zap.Int("userID", userID),
				zap.String("operation", "GenerateFromPassword"),
				zap.Error(err),
			)
			return appErrors.ErrPasswordEncryption.WithError(err)
		}
		//更新密码
		err = s.userDao.UpdatePassword(ctx, userID, hashedPassword, tx)
		if err != nil {
			logger.Error("更新密码失败：更新密码错误",
				zap.Int("userID", userID),
				zap.String("operation", "UpdatePassword"),
				zap.Error(err),
			)
			return appErrors.ErrPasswordUpdateFailed.WithError(err)
		}
		return nil
	})
	if err != nil {
		logger.Error("更新密码事务失败",
			zap.Int("userID", userID),
			zap.String("operation", "UpdatePasswordTransaction"),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (s *AuthService) AuthLoginByEmail(ctx context.Context, email, password string) (*models.User, string, error) {
	var existUser models.User
	var userToken string

	err := s.transactionManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		//检查用户是否存在
		user, err := s.userDao.GetByEmail(ctx, email, tx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("用户登录失败：用户不存在",
					zap.String("email", email),
					zap.String("operation", "GetByEmail"),
					zap.Error(err),
				)
				return appErrors.ErrUserNotFound
			}
			logger.Error("用户登录失败：数据库操作错误",
				zap.String("email", email),
				zap.String("operation", "GetByEmail"),
				zap.Error(err),
			)
			return appErrors.ErrDatabaseOperation.WithError(err)
		}
		//检查密码是否正确
		if !pkg.CheckPassword(user.Password, password) {
			logger.Error("用户登录失败：密码错误",
				zap.String("email", email),
				zap.Int("userID", user.UserID),
				zap.Time("lastLoginTime", user.UpdatedAt),
				zap.String("operation", "CheckPassword"),
				zap.Error(err),
			)
			return appErrors.ErrInvalidPassword
		}
		existUser = *user
		//生成token
		token, err := s.jwtHandler.GenerateJWTToken(user.Username, user.UserID)
		if err != nil {
			logger.Error("用户登录失败：Token生成错误",
				zap.String("email", email),
				zap.Int("userID", user.UserID),
				zap.Time("lastLoginTime", user.UpdatedAt),
				zap.String("operation", "GenerateJWTToken"),
				zap.Error(err),
			)
			return appErrors.ErrTokenGenerationFailed.WithError(err)
		}
		userToken = token

		logger.Info("用户登录成功",
			zap.String("email", email),
			zap.Int("userID", existUser.UserID),
			zap.Time("lastLoginTime", existUser.UpdatedAt),
			zap.Time("createTime", existUser.CreatedAt),
			zap.String("operation", "Login"),
		)
		return nil
	})
	if err != nil {
		logger.Error("用户登录事务失败",
			zap.String("email", email),
			zap.String("operation", "LoginTransaction"),
			zap.Error(err),
		)
		return nil, "", err
	}
	return &existUser, userToken, nil

}
