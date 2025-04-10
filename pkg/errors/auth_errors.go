package errors

import "net/http"

var (
	ErrUserNotFound = &AppError{
		Code:    "user_not_found",
		Message: "用户不存在",
		Status:  http.StatusNotFound,
	}

	ErrUserAlreadyExists = &AppError{
		Code:    "user_already_exists",
		Message: "用户已存在",
		Status:  http.StatusConflict,
	}

	ErrInvalidPassword = &AppError{
		Code:    "invalid_password",
		Message: "密码错误",
		Status:  http.StatusUnauthorized,
	}

	ErrPasswordEncryption = &AppError{
		Code:    "password_encryption_failed",
		Message: "密码加密失败",
		Status:  http.StatusInternalServerError,
	}

	ErrUserCreationFailed = &AppError{
		Code:    "user_creation_failed",
		Message: "用户创建失败",
		Status:  http.StatusInternalServerError,
	}
)
