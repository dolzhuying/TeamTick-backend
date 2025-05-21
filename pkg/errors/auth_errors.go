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

	ErrVerificationCodeExpiredOrNotFound = &AppError{
		Message: "验证码已过期或不存在",
		Status:  http.StatusGone,
	}

	ErrInvalidEmail = &AppError{
		Message: "邮箱格式不合法",
		Status:  http.StatusBadRequest,
	}

	ErrEmailNotRegistered = &AppError{
		Message: "该邮箱未注册",
		Status:  http.StatusNotFound,
	}

	ErrEmailAlreadyRegistered = &AppError{
		Message: "该邮箱已注册",
		Status:  http.StatusConflict,
	}

	ErrTooManyRequests = &AppError{
		Message: "请求过于频繁，请稍后再试",
		Status:  http.StatusTooManyRequests,
	}
)
