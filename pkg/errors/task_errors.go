package errors

import "net/http"

var (
	ErrTaskCreationFailed = &AppError{
		Message: "Failed to create task",
		Status:  http.StatusInternalServerError,
	}

	ErrTaskNotFound = &AppError{
		Message: "Task not found",
		Status:  http.StatusNotFound,
	}

	ErrTaskRecordCreationFailed = &AppError{
		Message: "Failed to create task record",
		Status:  http.StatusInternalServerError,
	}

	ErrTaskHasEnded = &AppError{
		Message: "Task has ended",
		Status:  http.StatusBadRequest,
	}

	ErrTaskNotInRange = &AppError{
		Message: "Task not in range",
		Status:  http.StatusBadRequest,
	}

	ErrTaskRecordAlreadyExists = &AppError{
		Message: "Task record already exists",
		Status:  http.StatusBadRequest,
	}
	
	
)
