package errors

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Code    string
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
		Code:    e.Code,
		Message: e.Message,
		Status:  e.Status,
		Err:     err,
	}
}

var (
	ErrDatabaseOperation = &AppError{
		Code:    "database_operation_failed",
		Message: "数据库操作失败",
		Status:  http.StatusInternalServerError,
	}
)
