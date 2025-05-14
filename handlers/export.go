package handlers

import (
	"TeamTickBackend/app"
	"TeamTickBackend/gen"
	appErrors "TeamTickBackend/pkg/errors"
	service "TeamTickBackend/services"
	"bytes"
	"context"
	"errors"
	"os"
	"time"
)

// ExportHandler 处理导出相关的请求
type ExportHandler struct {
	exportService service.StatisticsService
	groupsService service.GroupsService
}

// NewExportHandler 创建ExportHandler实例
func NewExportHandler(container *app.AppContainer) gen.ExportServerInterface {
	exportService := service.NewStatisticsService(
		container.DaoFactory.StatisticsDAO,
		container.DaoFactory.GroupDAO,
		container.DaoFactory.TransactionManager,
		container.DaoFactory.GroupMemberDAO,
	)
	groupsService := service.NewGroupsService(
		container.DaoFactory.GroupDAO,
		container.DaoFactory.GroupMemberDAO,
		container.DaoFactory.JoinApplicationDAO,
		container.DaoFactory.TransactionManager,
		container.DaoFactory.GroupRedisDAO,
		container.DaoFactory.GroupMemberRedisDAO,
		container.DaoFactory.JoinApplicationRedisDAO,
	)
	handler := &ExportHandler{
		exportService: *exportService,
		groupsService: *groupsService,
	}
	return gen.NewExportStrictHandler(handler, nil)
}

// GetExportCheckinsXlsx 导出签到记录为XLSX文件
func (h *ExportHandler) GetExportCheckinsXlsx(ctx context.Context, request gen.GetExportCheckinsXlsxRequestObject) (gen.GetExportCheckinsXlsxResponseObject, error) {
	groupIds := request.Params.GroupIds
	dateStart := time.Unix(int64(request.Params.DateStart), 0)
	dateEnd := time.Unix(int64(request.Params.DateEnd), 0)

	// userID, ok := ctx.Value("userID").(int)
	// if !ok {
	// 	return &gen.GetExportCheckinsXlsx403JSONResponse{
	// 		Code:    "1",
	// 		Message: "未登录或无权限",
	// 	}, nil
	// }

	// 校验所有 groupId 权限
	// for _, groupId := range groupIds {
	// 	if err := h.groupsService.CheckMemberPermission(ctx, groupId, userID); err != nil {
	// 		return &gen.GetExportCheckinsXlsx403JSONResponse{
	// 			Code:    "1",
	// 			Message: "无权限导出部分或全部用户组数据",
	// 		}, nil
	// 	}
	// }

	var statusStrings []string
	if request.Params.Statuses == nil || *request.Params.Statuses == nil {
		statusStrings = []string{"success", "absent", "exception"}
	} else {
		statusStrings = make([]string, len(*request.Params.Statuses))
		for i, status := range *request.Params.Statuses {
			statusStrings[i] = string(status)
		}
		if len(statusStrings) == 0 || statusStrings[0] == "" {
			statusStrings = []string{"success", "absent", "exception"}
		} else {
			for _, status := range statusStrings {
				if status != "success" && status != "absent" && status != "exception" {
					return &gen.GetExportCheckinsXlsx400JSONResponse{
						Code:    "1",
						Message: "状态参数错误",
					}, nil
				}
			}
		}
	}

	filePath, err := h.exportService.GenerateGroupSignInStatisticsExcel(ctx, groupIds, dateStart, dateEnd, statusStrings)
	if err != nil {
		// 针对不同错误类型返回不同响应体
		switch {
		case errors.Is(err, appErrors.ErrStatisticsInvalidTimeRange):
			return &gen.GetExportCheckinsXlsx400JSONResponse{
				Code:    "1",
				Message: "时间范围无效",
			}, nil
		case errors.Is(err, appErrors.ErrStatisticsGroupNotFound):
			return &gen.GetExportCheckinsXlsx400JSONResponse{
				Code:    "1",
				Message: "未找到用户组",
			}, nil
		case errors.Is(err, appErrors.ErrGroupNotFound):
			return &gen.GetExportCheckinsXlsx400JSONResponse{
				Code:    "1",
				Message: "未找到用户组",
			}, nil
		case errors.Is(err, appErrors.ErrStatisticsQueryFailed):
			return &gen.GetExportCheckinsXlsx500JSONResponse{
				Code:    "1",
				Message: "查询失败",
			}, nil
		case errors.Is(err, appErrors.ErrStatisticsFileCreationFailed):
			return &gen.GetExportCheckinsXlsx500JSONResponse{
				Code:    "1",
				Message: "文件生成失败",
			}, nil
		case errors.Is(err, appErrors.ErrStatisticsFileSaveFailed):
			return &gen.GetExportCheckinsXlsx500JSONResponse{
				Code:    "1",
				Message: "文件保存失败",
			}, nil
		default:
			return &gen.GetExportCheckinsXlsx500JSONResponse{
				Code:    "1",
				Message: "服务器内部错误",
			}, nil
		}
	}

	// 读取文件内容
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return &gen.GetExportCheckinsXlsx500JSONResponse{
			Code:    "1",
			Message: "文件读取失败",
		}, nil
	}

	// 将文件内容转换为 io.Reader
	fileReader := bytes.NewReader(fileContent)

	return &gen.GetExportCheckinsXlsx200ApplicationvndOpenxmlformatsOfficedocumentSpreadsheetmlSheetResponse{
		Body:          fileReader,
		ContentLength: int64(len(fileContent)),
	}, nil
}
