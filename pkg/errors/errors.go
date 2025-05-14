package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Message string
	Status  int
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) WithError(err error) *AppError {
	return &AppError{
		Message: e.Message,
		Status:  e.Status,
		Err:     err,
	}
}

var (
	// 400 Bad Request - 请求参数错误
	ErrInvalidRequest = &AppError{
		Message: "无效的请求参数",
		Status:  http.StatusBadRequest,
	}

	// 401 Unauthorized - 未登录
	ErrUnauthorized = &AppError{
		Message: "请先登录",
		Status:  http.StatusUnauthorized,
	}

	// 403 Forbidden - 无权限
	ErrForbidden = &AppError{
		Message: "没有操作权限",
		Status:  http.StatusForbidden,
	}

	// 404 Not Found - 资源不存在
	ErrResourceNotFound = &AppError{
		Message: "资源不存在",
		Status:  http.StatusNotFound,
	}

	// 500 Internal Server Error - 服务器内部错误
	ErrInternalServer = &AppError{
		Message: "服务器内部错误",
		Status:  http.StatusInternalServerError,
	}

	// 数据库相关错误
	ErrDatabaseOperation = &AppError{
		Message: "数据库操作失败",
		Status:  http.StatusInternalServerError,
	}
)
