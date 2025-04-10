package errors

import (
	"net/http"
)

var (
	ErrTokenConfigMissing = &AppError{
		Code:    "token_config_missing",
		Message: "JWT配置缺失",
		Status:  http.StatusInternalServerError,
	}

	ErrTokenGenerationFailed = &AppError{
		Code:    "token_generation_failed",
		Message: "令牌生成失败",
		Status:  http.StatusInternalServerError,
	}
)