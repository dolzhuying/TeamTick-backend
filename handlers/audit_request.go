package handlers

import (
	"TeamTickBackend/app"
	"TeamTickBackend/gen"
	appErrors "TeamTickBackend/pkg/errors"
	service "TeamTickBackend/services"
	"context"
	"errors"
)

type AuditRequestHandler struct {
	auditRequestService *service.AuditRequestService
	groupsService       *service.GroupsService
}

func NewAuditRequestHandler(container *app.AppContainer) gen.AuditRequestsServerInterface {
	auditRequestService := service.NewAuditRequestService(
		container.DaoFactory.TransactionManager,
		container.DaoFactory.CheckApplicationDAO,
		container.DaoFactory.TaskRecordDAO,
		container.DaoFactory.TaskDAO,
		container.DaoFactory.GroupDAO,
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
	handler := &AuditRequestHandler{
		auditRequestService: auditRequestService,
		groupsService:       groupsService,
	}
	return gen.NewAuditRequestsStrictHandler(handler, nil)
}

// 获取当前用户提交的所有审核请求列表
func (h *AuditRequestHandler) GetUsersMeAuditRequests(ctx context.Context, request gen.GetUsersMeAuditRequestsRequestObject) (gen.GetUsersMeAuditRequestsResponseObject, error) {
	// 从上下文中获取用户ID
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}

	// 调用服务获取审核请求列表
	requests, err := h.auditRequestService.GetAuditRequestListByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, appErrors.ErrAuditRequestNotFound) {
			return &gen.GetUsersMeAuditRequests200JSONResponse{
				Code: "0",
				Data: []gen.AuditRequest{},
			}, nil
		}
		return nil, err
	}

	// 转换为API响应格式
	var response []gen.AuditRequest
	for _, req := range requests {
		status := gen.AuditRequestStatus(req.Status)
		response = append(response, gen.AuditRequest{
			AuditRequestId: req.ID,
			TaskId:         req.TaskID,
			TaskName:       req.TaskName,
			UserId:         req.UserID,
			Username:       req.Username,
			Reason:         req.Reason,
			ProofImageUrls: req.Image,
			Status:         status,
			RequestedAt:    int(req.RequestAt.Unix()),
			AdminId:        req.AdminID,
			AdminUsername:  req.AdminUsername,
		})
	}

	if len(response) == 0 {
		return &gen.GetUsersMeAuditRequests200JSONResponse{
			Code: "0",
			Data: []gen.AuditRequest{},
		}, nil
	}

	return &gen.GetUsersMeAuditRequests200JSONResponse{
		Code: "0",
		Data: response,
	}, nil
}

// 管理员查看用户组内的所有签到异常审核申请
func (h *AuditRequestHandler) GetGroupsGroupIdAuditRequests(ctx context.Context, request gen.GetGroupsGroupIdAuditRequestsRequestObject) (gen.GetGroupsGroupIdAuditRequestsResponseObject, error) {
	// 从上下文中获取用户ID
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}

	// 检查用户是否是组的管理员
	if err := h.groupsService.CheckMemberPermission(ctx, request.GroupId, userID); err != nil {
		if errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return &gen.GetGroupsGroupIdAuditRequests403JSONResponse{
				Code:    "1",
				Message: "权限不足",
			}, nil
		}
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.GetGroupsGroupIdAuditRequests404JSONResponse{
				Code:    "1",
				Message: "用户组不存在",
			}, nil
		}
		return nil, err
	}

	// 获取状态参数
	var status string
	status = string(*request.Params.Status)
	if len(status) == 0 {
		status = string(gen.RequestQueryStatusAll)
	}

	// 调用服务获取审核请求列表
	requests, err := h.auditRequestService.GetAuditRequestByGroupIDWithStatus(ctx, request.GroupId, status)
	if err != nil {
		if errors.Is(err, appErrors.ErrAuditRequestNotFound) {
			return &gen.GetGroupsGroupIdAuditRequests200JSONResponse{
				Code: "0",
				Data: []gen.AuditRequest{},
			}, nil
		}
		return nil, err
	}

	// 转换为API响应格式
	response := make([]gen.AuditRequest, 0, len(requests))
	for _, req := range requests {
		status := gen.AuditRequestStatus(req.Status)
		response = append(response, gen.AuditRequest{
			AuditRequestId: req.ID,
			TaskId:         req.TaskID,
			TaskName:       req.TaskName,
			UserId:         req.UserID,
			Username:       req.Username,
			Reason:         req.Reason,
			ProofImageUrls: req.Image,
			Status:         status,
			RequestedAt:    int(req.RequestAt.Unix()),
			AdminId:        req.AdminID,
			AdminUsername:  req.AdminUsername,
		})
	}

	return &gen.GetGroupsGroupIdAuditRequests200JSONResponse{
		Code: "0",
		Data: response,
	}, nil
}

// 当用户无法通过常规方式成功签到时，可以提交异常情况说明和证明，申请人工审核。审核通过后才会生成签到记录
func (h *AuditRequestHandler) PostCheckinTasksTaskIdAuditRequests(ctx context.Context, request gen.PostCheckinTasksTaskIdAuditRequestsRequestObject) (gen.PostCheckinTasksTaskIdAuditRequestsResponseObject, error) {
	// 从上下文中获取用户ID和用户名
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	username, ok := ctx.Value("username").(string)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}

	// 处理可能为nil的ProofImageUrls
	var proofImageUrls string
	if request.Body.ProofImageUrls != nil {
		proofImageUrls = *request.Body.ProofImageUrls
	}
	if len(proofImageUrls) == 0 {
		proofImageUrls = ""
	}

	// 调用服务创建审核请求
	auditRequest, err := h.auditRequestService.CreateAuditRequest(ctx, request.TaskId, userID, username, request.Body.Reason, proofImageUrls)
	if err != nil {
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.PostCheckinTasksTaskIdAuditRequests404JSONResponse{
				Code:    "1",
				Message: "用户组不存在",
			}, nil
		}
		if errors.Is(err, appErrors.ErrTaskNotFound) {
			return &gen.PostCheckinTasksTaskIdAuditRequests404JSONResponse{
				Code:    "1",
				Message: "任务不存在",
			}, nil
		}
		if errors.Is(err, appErrors.ErrAuditRequestAlreadyExists) {
			return &gen.PostCheckinTasksTaskIdAuditRequests409JSONResponse{
				Code:    "1",
				Message: "审核请求已存在",
			}, nil
		}
		return nil, err
	}

	// 转换为API响应格式
	status := gen.AuditRequestStatus(auditRequest.Status)
	return &gen.PostCheckinTasksTaskIdAuditRequests201JSONResponse{
		Code: "0",
		Data: gen.AuditRequest{
			AuditRequestId: auditRequest.ID,
			TaskId:         auditRequest.TaskID,
			TaskName:       auditRequest.TaskName,
			UserId:         auditRequest.UserID,
			Username:       auditRequest.Username,
			Reason:         auditRequest.Reason,
			ProofImageUrls: auditRequest.Image,
			Status:         status,
			RequestedAt:    int(auditRequest.RequestAt.Unix()),
		},
	}, nil
}

// 用户组管理员批准或拒绝签到审核申请。批准后将创建成功的签到记录
func (h *AuditRequestHandler) PutAuditRequestsAuditRequestId(ctx context.Context, request gen.PutAuditRequestsAuditRequestIdRequestObject) (gen.PutAuditRequestsAuditRequestIdResponseObject, error) {
	// 从上下文中获取用户ID
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	
	// 获取审核请求对应的组ID
	groupID, err := h.auditRequestService.GetGroupIDByAuditRequestID(ctx, request.AuditRequestId)
	if err != nil {
		if errors.Is(err, appErrors.ErrAuditRequestNotFound) {
			return &gen.PutAuditRequestsAuditRequestId404JSONResponse{
				Code:    "1", 
				Message: "未找到审核请求",
			}, nil
		}
		return nil, err
	}

	// 检查用户是否有权限处理审核请求
	if err := h.groupsService.CheckMemberPermission(ctx, groupID, userID); err != nil {
		if errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return &gen.PutAuditRequestsAuditRequestId403JSONResponse{
				Code:    "1",
				Message: "权限不足",
			}, nil
		}
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.PutAuditRequestsAuditRequestId404JSONResponse{
				Code:    "1",
				Message: "用户组不存在",
			}, nil
		}
		return nil, err
	}
	
	
	// 调用服务更新审核请求状态
	err = h.auditRequestService.UpdateAuditRequest(ctx, request.AuditRequestId, string(request.Body.Action))
	if err != nil {
		if errors.Is(err, appErrors.ErrAuditRequestNotFound) {
			return &gen.PutAuditRequestsAuditRequestId404JSONResponse{
				Code:    "1",
				Message: "未找到审核请求",
			}, nil
		}
		return nil, err
	}

	return &gen.PutAuditRequestsAuditRequestId200JSONResponse{
		Code: "0",
		Data: &map[string]interface{}{},
	}, nil
}
