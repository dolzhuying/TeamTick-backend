package errors

import "net/http"

var (
	ErrGroupCreationFailed = &AppError{
		Code:    "group_creation_failed",
		Message: "用户组创建失败",
		Status:  http.StatusInternalServerError,
	}

	ErrGroupNotFound = &AppError{
		Code:    "group_not_found",
		Message: "用户组不存在",
		Status:  http.StatusNotFound,
	}

	ErrGroupUpdateFailed = &AppError{
		Code:    "group_update_failed",
		Message: "用户组更新失败",
		Status:  http.StatusInternalServerError,
	}

	ErrGroupMemberAlreadyExists = &AppError{
		Code:    "group_member_already_exists",
		Message: "用户组成员已存在",
		Status:  http.StatusConflict,
	}

	ErrGroupMemberCreationFailed = &AppError{
		Code:    "group_member_creation_failed",
		Message: "用户组成员创建失败",
		Status:  http.StatusInternalServerError,
	}

	ErrGroupMemberDeletionFailed = &AppError{
		Code:    "group_member_deletion_failed",
		Message: "用户组成员删除失败",
		Status:  http.StatusInternalServerError,
	}

	ErrJoinApplicationCreationFailed = &AppError{
		Code:    "join_application_creation_failed",
		Message: "用户申请加入记录创建失败",
		Status:  http.StatusInternalServerError,
	}

	ErrRolePermissionDenied = &AppError{
		Code:    "role_permission_denied",
		Message: "权限不足",
		Status:  http.StatusForbidden,
	}

	ErrGroupMemberNotFound = &AppError{
		Code:    "group_member_not_found",
		Message: "用户组成员不存在",
		Status:  http.StatusNotFound,
	}

	ErrJoinApplicationNotFound = &AppError{
		Code:    "join_application_not_found",
		Message: "用户申请加入记录不存在",
		Status:  http.StatusNotFound,
	}

	ErrJoinApplicationUpdateFailed = &AppError{
		Code:    "join_application_update_failed",
		Message: "用户申请加入记录更新失败",
		Status:  http.StatusInternalServerError,
	}

	ErrGroupDeletionFailed = &AppError{
		Code:    "group_deletion_failed",
		Message: "用户组删除失败",
		Status:  http.StatusInternalServerError,
	}

	//待完善
)
