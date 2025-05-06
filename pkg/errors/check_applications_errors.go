package errors

import "net/http"

var(
	ErrAuditRequestNotFound=&AppError{
		Message:"Audit request not found",
		Status:http.StatusNotFound,
	}

	ErrAuditRequestCreateFailed=&AppError{
		Message:"Audit request create failed",
		Status:http.StatusInternalServerError,
	}

	ErrAuditRequestAlreadyExists=&AppError{
		Message:"Audit request already exists",
		Status:http.StatusConflict,
	}

	ErrAuditRequestUpdateFailed=&AppError{
		Message:"Audit request update failed",
		Status:http.StatusInternalServerError,
	}
)