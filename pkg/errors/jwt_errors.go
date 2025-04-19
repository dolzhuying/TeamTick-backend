package errors

import (
	"net/http"
)

var (
	ErrTokenConfigMissing = &AppError{
		Message: "JWT配置缺失",
		Status:  http.StatusInternalServerError,
	}

	ErrTokenGenerationFailed = &AppError{
		Message: "令牌生成失败",
		Status:  http.StatusInternalServerError,
	}
)