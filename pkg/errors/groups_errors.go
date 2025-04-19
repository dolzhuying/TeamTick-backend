package errors

import "net/http"

var (
	ErrGroupCreationFailed = &AppError{
		Message: "用户组创建失败",
		Status:  http.StatusInternalServerError,
	}

	ErrGroupNotFound = &AppError{
		Message: "用户组不存在",
		Status:  http.StatusNotFound,
	}

	ErrGroupUpdateFailed = &AppError{
		Message: "用户组更新失败",
		Status:  http.StatusInternalServerError,
	}

	ErrGroupMemberAlreadyExists = &AppError{
		Message: "用户组成员已存在",
		Status:  http.StatusConflict,
	}

	ErrGroupMemberCreationFailed = &AppError{
		Message: "用户组成员创建失败",
		Status:  http.StatusInternalServerError,
	}

	ErrGroupMemberDeletionFailed = &AppError{
		Message: "用户组成员删除失败",
		Status:  http.StatusInternalServerError,
	}

	ErrJoinApplicationCreationFailed = &AppError{
		Message: "用户申请加入记录创建失败",
		Status:  http.StatusInternalServerError,
	}

	ErrRolePermissionDenied = &AppError{
		Message: "权限不足",
		Status:  http.StatusForbidden,
	}

	ErrGroupMemberNotFound = &AppError{
		Message: "用户组成员不存在",
		Status:  http.StatusNotFound,
	}

	ErrJoinApplicationNotFound = &AppError{
		Message: "用户申请加入记录不存在",
		Status:  http.StatusNotFound,
	}

	ErrJoinApplicationUpdateFailed = &AppError{
		Message: "用户申请加入记录更新失败",
		Status:  http.StatusInternalServerError,
	}

	ErrGroupDeletionFailed = &AppError{
		Message: "用户组删除失败",
		Status:  http.StatusInternalServerError,
	}

	//待完善
)
