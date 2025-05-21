package errors

import "net/http"

var (
	ErrUserNotFound = &AppError{
		Message: "用户不存在",
		Status:  http.StatusNotFound,
	}

	ErrUserAlreadyExists = &AppError{
		Message: "用户已存在",
		Status:  http.StatusConflict,
	}

	ErrInvalidPassword = &AppError{
		Message: "密码错误",
		Status:  http.StatusUnauthorized,
	}

	ErrPasswordEncryption = &AppError{
		Message: "密码加密失败",
		Status:  http.StatusInternalServerError,
	}

	ErrUserCreationFailed = &AppError{
		Message: "用户创建失败",
		Status:  http.StatusInternalServerError,
	}

	ErrPasswordUpdateFailed = &AppError{
		Message: "密码更新失败",
		Status:  http.StatusInternalServerError,
	}

	ErrInvalidVerificationCode = &AppError{
		Message: "邮箱验证码错误",
		Status:  http.StatusUnauthorized,
	}
		
)
