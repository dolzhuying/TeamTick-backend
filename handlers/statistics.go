package handlers

import (
	"TeamTickBackend/app"
	"TeamTickBackend/gen"
	appErrors "TeamTickBackend/pkg/errors"
	service "TeamTickBackend/services"
	"context"
	"errors"
	"time"
)

// StatisticsHandler 处理统计相关的请求
type StatisticsHandler struct {
	statisticsService *service.StatisticsService
	groupsService     *service.GroupsService
}

// NewStatisticsHandler 创建StatisticsHandler实例
func NewStatisticsHandler(container *app.AppContainer) gen.StatisticsServerInterface {
	statisticsService := service.NewStatisticsService(
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
		container.DaoFactory.TaskRedisDAO,
	)

	handler := &StatisticsHandler{
		statisticsService: statisticsService,
		groupsService:     groupsService,
	}
	return gen.NewStatisticsStrictHandler(handler, nil)
}

// GetStatisticsGroups 获取用户组签到统计数据
func (h *StatisticsHandler) GetStatisticsGroups(ctx context.Context, request gen.GetStatisticsGroupsRequestObject) (gen.GetStatisticsGroupsResponseObject, error) {

	startDate := time.Unix(int64(*request.Params.StartDate), 0)
	endDate := time.Unix(int64(*request.Params.EndDate), 0)

	// 获取所有用户组
	groups, err := h.statisticsService.GetAllGroups(ctx)
	if err != nil {
		if errors.Is(err, appErrors.ErrStatisticsInvalidTimeRange) {
			return &gen.GetStatisticsGroups400JSONResponse{
				Code:    "1",
				Message: "时间参数错误",
			}, nil
		}

		return nil, err
	}

	data := make([]struct {
		Absent    int    `json:"absent"`
		Exception int    `json:"exception"`
		GroupId   int    `json:"groupId"`
		GroupName string `json:"groupName"`
		Success   int    `json:"success"`
	}, 0, len(groups))

	for _, group := range groups {
		statistics, err := h.statisticsService.GetGroupSignInStatistics(ctx, group.GroupID, startDate, endDate)
		if err != nil {
			if errors.Is(err, appErrors.ErrStatisticsGroupNotFound) {
				continue // 跳过不存在的组
			}
			return nil, err
		}
		data = append(data, struct {
			Absent    int    `json:"absent"`
			Exception int    `json:"exception"`
			GroupId   int    `json:"groupId"`
			GroupName string `json:"groupName"`
			Success   int    `json:"success"`
		}{
			Absent:    len(statistics.AbsentRecords),
			Exception: len(statistics.ExecptionRecords),
			GroupId:   group.GroupID,
			GroupName: group.GroupName,
			Success:   len(statistics.SuccessRecords),
		})
	}

	return &gen.GetStatisticsGroups200JSONResponse{
		Code: "0",
		Data: data,
	}, nil
}

// GetStatisticsUsers 获取用户签到统计数据
func (h *StatisticsHandler) GetStatisticsUsers(ctx context.Context, request gen.GetStatisticsUsersRequestObject) (gen.GetStatisticsUsersResponseObject, error) {
	groupId := *request.Params.GroupId
	startDate := time.Unix(int64(*request.Params.StartDate), 0)
	endDate := time.Unix(int64(*request.Params.EndDate), 0)
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	if err := h.groupsService.CheckMemberPermission(ctx, groupId, userID); err != nil {
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.GetStatisticsUsers400JSONResponse{
				Code:    "1",
				Message: "未找到用户组",
			}, nil
		}
		if errors.Is(err, appErrors.ErrGroupMemberNotFound) {
			return &gen.GetStatisticsUsers403JSONResponse{
				Code:    "1",
				Message: "无权限访问该组数据",
			}, nil
		}
		if errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return &gen.GetStatisticsUsers403JSONResponse{
				Code:    "1",
				Message: "无权限访问该组数据",
			}, nil
		}
		return nil, err
	}
	// 使用 GroupsService 获取组内所有成员
	members, err := h.groupsService.GetMembersByGroupID(ctx, groupId)
	if err != nil {
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.GetStatisticsUsers400JSONResponse{
				Code:    "1",
				Message: "未找到用户组",
			}, nil
		}
		return nil, err
	}

	// 使用 GroupsService 获取组信息
	group, err := h.groupsService.GetGroupByGroupID(ctx, groupId)
	if err != nil {
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.GetStatisticsUsers400JSONResponse{
				Code:    "1",
				Message: "未找到用户组",
			}, nil
		}
		return nil, err
	}

	data := make([]struct {
		Absent    int    `json:"absent"`
		Exception int    `json:"exception"`
		GroupId   int    `json:"groupId"`
		GroupName string `json:"groupName"`
		Success   int    `json:"success"`
		UserId    int    `json:"userId"`
		Username  string `json:"username"`
	}, 0, len(members))

	for _, member := range members {
		statistics, err := h.statisticsService.GetGroupMemberSignInStatistics(ctx, groupId, member.UserID, startDate, endDate)
		if err != nil {
			if errors.Is(err, appErrors.ErrStatisticsGroupNotFound) {
				continue // 跳过不存在的成员
			}
			if errors.Is(err, appErrors.ErrStatisticsInvalidTimeRange) {
				return &gen.GetStatisticsUsers400JSONResponse{
					Code:    "1",
					Message: "时间参数错误",
				}, nil
			}
			return nil, err
		}
		data = append(data, struct {
			Absent    int    `json:"absent"`
			Exception int    `json:"exception"`
			GroupId   int    `json:"groupId"`
			GroupName string `json:"groupName"`
			Success   int    `json:"success"`
			UserId    int    `json:"userId"`
			Username  string `json:"username"`
		}{
			Absent:    statistics.AbsentNum,
			Exception: statistics.ExceptionNum,
			GroupId:   groupId,
			GroupName: group.GroupName,
			Success:   statistics.SuccessNum,
			UserId:    member.UserID,
			Username:  member.Username,
		})
	}

	return &gen.GetStatisticsUsers200JSONResponse{
		Code: "0",
		Data: data,
	}, nil
}
