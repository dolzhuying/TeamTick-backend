package errors

import "net/http"

var (
	// 400 Bad Request - 统计相关参数错误
	ErrStatisticsInvalidTimeRange = &AppError{
		Message: "无效的时间范围",
		Status:  http.StatusBadRequest,
	}

	ErrStatisticsInvalidGroupID = &AppError{
		Message: "无效的用户组ID",
		Status:  http.StatusBadRequest,
	}

	// 401 Unauthorized - 统计相关认证错误
	ErrStatisticsUnauthorized = &AppError{
		Message: "请先登录",
		Status:  http.StatusUnauthorized,
	}

	ErrStatisticsInvalidToken = &AppError{
		Message: "无效的认证令牌",
		Status:  http.StatusUnauthorized,
	}

	ErrStatisticsTokenExpired = &AppError{
		Message: "认证令牌已过期",
		Status:  http.StatusUnauthorized,
	}

	// 403 Forbidden - 统计相关权限错误
	ErrStatisticsPermissionDenied = &AppError{
		Message: "没有操作权限",
		Status:  http.StatusForbidden,
	}

	ErrStatisticsAdminRequired = &AppError{
		Message: "需要管理员权限",
		Status:  http.StatusForbidden,
	}

	ErrStatisticsNotGroupMember = &AppError{
		Message: "不是群组成员",
		Status:  http.StatusForbidden,
	}

	// 404 Not Found - 统计相关资源不存在
	ErrStatisticsGroupNotFound = &AppError{
		Message: "用户组不存在",
		Status:  http.StatusNotFound,
	}

	// 500 Internal Server Error - 统计相关服务器错误
	ErrStatisticsQueryFailed = &AppError{
		Message: "统计查询失败",
		Status:  http.StatusInternalServerError,
	}

	ErrStatisticsFileCreationFailed = &AppError{
		Message: "创建统计文件失败",
		Status:  http.StatusInternalServerError,
	}

	ErrStatisticsFileSaveFailed = &AppError{
		Message: "保存统计文件失败",
		Status:  http.StatusInternalServerError,
	}
)
