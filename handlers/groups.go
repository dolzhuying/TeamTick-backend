package handlers

import (
	"TeamTickBackend/app"
	"TeamTickBackend/dal/models"
	"TeamTickBackend/gen"
	appErrors "TeamTickBackend/pkg/errors"
	service "TeamTickBackend/services"
	"context"
	"errors"
)

type GroupsHandler struct {
	groupsService service.GroupsService
}

func NewGroupsHandler(container *app.AppContainer) gen.GroupsServerInterface {
	GroupsService := service.NewGroupsService(
		container.DaoFactory.GroupDAO,
		container.DaoFactory.GroupMemberDAO,
		container.DaoFactory.JoinApplicationDAO,
		container.DaoFactory.TransactionManager,
	)
	handler := &GroupsHandler{
		groupsService: *GroupsService,
	}
	return gen.NewGroupsStrictHandler(handler, nil)
}

// 获取当前用户创建的或加入的用户组列表
func (h *GroupsHandler) GetGroups(ctx context.Context, request gen.GetGroupsRequestObject) (gen.GetGroupsResponseObject, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}

	var groups []*models.Group
	var err error

	if request.Params.Filter != nil {
		filter := string(*request.Params.Filter)
		if len(filter) > 0 {
			groups, err = h.groupsService.GetGroupsByUserID(ctx, userID, filter)
		} else {
			groups, err = h.groupsService.GetGroupsByUserID(ctx, userID)
		}
	} else {
		groups, err = h.groupsService.GetGroupsByUserID(ctx, userID)
	}

	if err != nil {
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.GetGroups200JSONResponse{
				Code: "0",
				Data: []struct {
					CreatedAt   int           `json:"createdAt,omitempty"`
					CreatorId   int           `json:"creatorId,omitempty"`
					CreatorName string        `json:"creatorName,omitempty"`
					Description string        `json:"description,omitempty"`
					GroupId     int           `json:"groupId,omitempty"`
					GroupName   string        `json:"groupName,omitempty"`
					MemberCount int           `json:"memberCount,omitempty"`
					RoleInGroup gen.GroupRole `json:"roleInGroup,omitempty"`
				}{},
			}, nil
		}
		return nil, err
	}

	genGroups := make([]struct {
		CreatedAt   int           `json:"createdAt,omitempty"`
		CreatorId   int           `json:"creatorId,omitempty"`
		CreatorName string        `json:"creatorName,omitempty"`
		Description string        `json:"description,omitempty"`
		GroupId     int           `json:"groupId,omitempty"`
		GroupName   string        `json:"groupName,omitempty"`
		MemberCount int           `json:"memberCount,omitempty"`
		RoleInGroup gen.GroupRole `json:"roleInGroup,omitempty"`
	}, len(groups))

	for i, group := range groups {
		if group.CreatorID == userID {
			genGroups[i] = struct {
				CreatedAt   int           `json:"createdAt,omitempty"`
				CreatorId   int           `json:"creatorId,omitempty"`
				CreatorName string        `json:"creatorName,omitempty"`
				Description string        `json:"description,omitempty"`
				GroupId     int           `json:"groupId,omitempty"`
				GroupName   string        `json:"groupName,omitempty"`
				MemberCount int           `json:"memberCount,omitempty"`
				RoleInGroup gen.GroupRole `json:"roleInGroup,omitempty"`
			}{
				CreatedAt:   int(group.CreatedAt.Unix()),
				CreatorId:   group.CreatorID,
				CreatorName: group.CreatorName,
				Description: group.Description,
				GroupId:     group.GroupID,
				GroupName:   group.GroupName,
				MemberCount: group.MemberNum,
				RoleInGroup: "admin",
			}
		} else {
			genGroups[i] = struct {
				CreatedAt   int           `json:"createdAt,omitempty"`
				CreatorId   int           `json:"creatorId,omitempty"`
				CreatorName string        `json:"creatorName,omitempty"`
				Description string        `json:"description,omitempty"`
				GroupId     int           `json:"groupId,omitempty"`
				GroupName   string        `json:"groupName,omitempty"`
				MemberCount int           `json:"memberCount,omitempty"`
				RoleInGroup gen.GroupRole `json:"roleInGroup,omitempty"`
			}{
				CreatedAt:   int(group.CreatedAt.Unix()),
				CreatorId:   group.CreatorID,
				CreatorName: group.CreatorName,
				Description: group.Description,
				GroupId:     group.GroupID,
				GroupName:   group.GroupName,
				MemberCount: group.MemberNum,
				RoleInGroup: "member",
			}
		}
	}

	if len(genGroups) == 0 {
		return &gen.GetGroups200JSONResponse{
			Code: "0",
			Data: []struct {
				CreatedAt   int           `json:"createdAt,omitempty"`
				CreatorId   int           `json:"creatorId,omitempty"`
				CreatorName string        `json:"creatorName,omitempty"`
				Description string        `json:"description,omitempty"`
				GroupId     int           `json:"groupId,omitempty"`
				GroupName   string        `json:"groupName,omitempty"`
				MemberCount int           `json:"memberCount,omitempty"`
				RoleInGroup gen.GroupRole `json:"roleInGroup,omitempty"`
			}{},
		}, nil
	}

	return &gen.GetGroups200JSONResponse{
		Code: "0",
		Data: genGroups,
	}, nil
}

// 任何登录用户可以创建用户组，并自动成为该组管理员
func (h *GroupsHandler) PostGroups(ctx context.Context, request gen.PostGroupsRequestObject) (gen.PostGroupsResponseObject, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	username, ok := ctx.Value("username").(string)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	groupName := request.Body.GroupName
	description := request.Body.Description

	group, err := h.groupsService.CreateGroup(ctx, groupName, description, username, userID)
	if err != nil {
		return nil, err
	}

	return &gen.PostGroups201JSONResponse{
		Code: "0",
		Data: struct {
			CreatedAt   int    `json:"createdAt,omitempty"`
			CreatorId   int    `json:"creatorId,omitempty"`
			CreatorName string `json:"creatorName,omitempty"`
			Description string `json:"description,omitempty"`
			GroupId     int    `json:"groupId,omitempty"`
			GroupName   string `json:"groupName,omitempty"`
			MemberCount int    `json:"memberCount,omitempty"`
		}{
			CreatedAt:   int(group.CreatedAt.Unix()),
			CreatorId:   group.CreatorID,
			CreatorName: group.CreatorName,
			Description: group.Description,
			GroupId:     group.GroupID,
			GroupName:   group.GroupName,
			MemberCount: group.MemberNum,
		},
	}, nil
}

// 获取指定用户组的详细信息 (组名、描述、成员数量等)。不需要是该组成员
func (h *GroupsHandler) GetGroupsGroupId(ctx context.Context, request gen.GetGroupsGroupIdRequestObject) (gen.GetGroupsGroupIdResponseObject, error) {
	groupID := request.GroupId

	group, err := h.groupsService.GetGroupByGroupID(ctx, groupID)
	if err != nil {
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.GetGroupsGroupId404JSONResponse{
				Code:    "1",
				Message: "用户组不存在",
			}, nil
		}
		return nil, err
	}

	return &gen.GetGroupsGroupId200JSONResponse{
		Code: "0",
		Data: gen.Group{
			GroupId:     group.GroupID,
			GroupName:   group.GroupName,
			Description: group.Description,
			CreatorId:   group.CreatorID,
			CreatorName: group.CreatorName,
			CreatedAt:   int(group.CreatedAt.Unix()),
			MemberCount: group.MemberNum,
		},
	}, nil
}

// 更新用户组修改用户组的名称或描述。需要是该组管理员
func (h *GroupsHandler) PutGroupsGroupId(ctx context.Context, request gen.PutGroupsGroupIdRequestObject) (gen.PutGroupsGroupIdResponseObject, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	groupID := request.GroupId
	groupName := request.Body.GroupName
	description := request.Body.Description

	_, err := h.groupsService.GetGroupByGroupID(ctx, groupID)
	if err != nil {
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.PutGroupsGroupId404JSONResponse{
				Code:    "1",
				Message: "用户组不存在",
			}, nil
		}
		return nil, err
	}

	// 检查用户是否是群组管理员
	if err := h.groupsService.CheckMemberPermission(ctx, groupID, userID); err != nil {
		if errors.Is(err, appErrors.ErrGroupMemberNotFound) {
			return &gen.PutGroupsGroupId404JSONResponse{
				Code:    "1",
				Message: "您不是该组成员",
			}, nil

		}
		if errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return nil, err
		}
	}
	group, err := h.groupsService.UpdateGroup(ctx, groupID, userID, groupName, description)
	if err != nil {
		if errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return &gen.PutGroupsGroupId403JSONResponse{
				Code:    "1",
				Message: "权限不足",
			}, nil
		}
		return nil, err

	}

	return &gen.PutGroupsGroupId200JSONResponse{
		Code: "0",
		Data: gen.Group{
			GroupId:     group.GroupID,
			GroupName:   group.GroupName,
			Description: group.Description,
			CreatorId:   group.CreatorID,
			CreatorName: group.CreatorName,
			CreatedAt:   int(group.CreatedAt.Unix()),
			MemberCount: group.MemberNum,
		},
	}, nil

}

// 删除指定的用户组及其关联数据。需要是该组的创建者
func (h *GroupsHandler) DeleteGroupsGroupId(ctx context.Context, request gen.DeleteGroupsGroupIdRequestObject) (gen.DeleteGroupsGroupIdResponseObject, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	groupID := request.GroupId

	// 检查用户组是否存在
	group, err := h.groupsService.GetGroupByGroupID(ctx, groupID)
	if err != nil {
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.DeleteGroupsGroupId404JSONResponse{
				Code:    "1",
				Message: "用户组不存在",
			}, nil
		}
		return nil, err
	}

	// 检查用户是否是群组成员
	if err := h.groupsService.CheckMemberPermission(ctx, groupID, userID); err != nil {
		if errors.Is(err, appErrors.ErrGroupMemberNotFound) {
			return &gen.DeleteGroupsGroupId404JSONResponse{
				Code:    "1",
				Message: "您不是该组成员",
			}, nil
		}
		if !errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return &gen.DeleteGroupsGroupId403JSONResponse{
				Code:    "1",
				Message: "权限不足",
			}, nil
		}
	}

	// 检查是否是群组创建者
	if userID != group.CreatorID {
		return &gen.DeleteGroupsGroupId403JSONResponse{
			Code:    "1",
			Message: "只有群组创建者可以删除群组",
		}, nil
	}

	err = h.groupsService.DeleteGroup(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}

	return &gen.DeleteGroupsGroupId200JSONResponse{
		Code: "0",
		Data: &map[string]interface{}{},
	}, nil
}

// 用户组管理员查看待处理的加入申请
func (h *GroupsHandler) GetGroupsGroupIdJoinRequests(ctx context.Context, request gen.GetGroupsGroupIdJoinRequestsRequestObject) (gen.GetGroupsGroupIdJoinRequestsResponseObject, error) {
	// 从上下文中获取用户ID
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}

	// 检查用户是否是组的管理员
	if err := h.groupsService.CheckMemberPermission(ctx, request.GroupId, userID); err != nil {
		if errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return &gen.GetGroupsGroupIdJoinRequests403JSONResponse{
				Code:    "1",
				Message: "权限不足",
			}, nil
		}
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.GetGroupsGroupIdJoinRequests404JSONResponse{
				Code:    "1",
				Message: "用户组不存在",
			}, nil
		}
		return nil, err
	}

	// 根据status参数决定是否传入filter
	var applications []*models.JoinApplication
	var err error
	status := string(*request.Params.Status)
	if len(status) > 0 {
		applications, err = h.groupsService.GetJoinApplicationsByGroupID(ctx, request.GroupId, userID, status)
	} else {
		applications, err = h.groupsService.GetJoinApplicationsByGroupID(ctx, request.GroupId, userID)
	}

	if err != nil {
		if errors.Is(err, appErrors.ErrJoinApplicationNotFound) {
			return &gen.GetGroupsGroupIdJoinRequests200JSONResponse{
				Code: "0",
				Data: []gen.JoinRequest{},
			}, nil
		}
		if errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return &gen.GetGroupsGroupIdJoinRequests403JSONResponse{
				Code:    "1",
				Message: "权限不足",
			}, nil
		}
		return nil, err
	}

	// 转换为API响应格式
	response := make([]gen.JoinRequest, 0, len(applications))
	for _, app := range applications {
		response = append(response, gen.JoinRequest{
			RequestId:   app.RequestID,
			GroupId:     app.GroupID,
			UserId:      app.UserID,
			Username:    app.Username,
			Status:      gen.JoinRequestStatus(app.Status),
			RequestedAt: int(app.CreatedAt.Unix()),
		})
	}

	if len(response) == 0 {
		return &gen.GetGroupsGroupIdJoinRequests200JSONResponse{
			Code: "0",
			Data: []gen.JoinRequest{},
		}, nil
	}

	return &gen.GetGroupsGroupIdJoinRequests200JSONResponse{
		Code: "0",
		Data: response,
	}, nil
}

// 用户向指定用户组提交加入申请
func (h *GroupsHandler) PostGroupsGroupIdJoinRequests(ctx context.Context, request gen.PostGroupsGroupIdJoinRequestsRequestObject) (gen.PostGroupsGroupIdJoinRequestsResponseObject, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	username, ok := ctx.Value("username").(string)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	groupID := request.GroupId
	reason := request.Body.Reason

	application, err := h.groupsService.CreateJoinApplication(ctx, groupID, userID, username, reason)
	if err != nil {
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.PostGroupsGroupIdJoinRequests404JSONResponse{
				Code:    "1",
				Message: "用户组不存在",
			}, nil
		}
		if errors.Is(err, appErrors.ErrGroupMemberAlreadyExists) {
			return &gen.PostGroupsGroupIdJoinRequests409JSONResponse{
				Code:    "1",
				Message: "您已经是该组成员",
			}, nil
		}
		if errors.Is(err, appErrors.ErrJoinApplicationAlreadyExists) {
			return &gen.PostGroupsGroupIdJoinRequests409JSONResponse{
				Code:    "1",
				Message: "您已经提交过加入申请",
			}, nil
		}
		return nil, err
	}

	return &gen.PostGroupsGroupIdJoinRequests201JSONResponse{
		Code: "0",
		Data: gen.JoinRequest{
			RequestId:   application.RequestID,
			GroupId:     application.GroupID,
			UserId:      application.UserID,
			Username:    application.Username,
			Status:      gen.JoinRequestStatus(application.Status),
			RequestedAt: int(application.CreatedAt.Unix()),
		},
	}, nil
}

// 用户组管理员批准或拒绝用户的加入申请
func (h *GroupsHandler) PutGroupsGroupIdJoinRequestsRequestId(ctx context.Context, request gen.PutGroupsGroupIdJoinRequestsRequestIdRequestObject) (gen.PutGroupsGroupIdJoinRequestsRequestIdResponseObject, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	_, ok = ctx.Value("username").(string)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	groupID := request.GroupId
	requestID := request.RequestId
	action := request.Body.Action

	// 检查用户是否在用户组中
	if err := h.groupsService.CheckMemberPermission(ctx, groupID, userID); err != nil {
		if errors.Is(err, appErrors.ErrGroupMemberNotFound) {
			return &gen.PutGroupsGroupIdJoinRequestsRequestId403JSONResponse{
				Code:    "1",
				Message: "您不是该组成员",
			}, nil
		}
		if errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return &gen.PutGroupsGroupIdJoinRequestsRequestId403JSONResponse{
				Code:    "1",
				Message: "权限不足",
			}, nil
		}
		return nil, err
	}

	applications, err := h.groupsService.GetJoinApplicationsByGroupID(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.PutGroupsGroupIdJoinRequestsRequestId404JSONResponse{
				Code:    "1",
				Message: "用户组不存在",
			}, nil
		}
		if errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return &gen.PutGroupsGroupIdJoinRequestsRequestId403JSONResponse{
				Code:    "1",
				Message: "权限不足",
			}, nil
		}
		if errors.Is(err, appErrors.ErrJoinApplicationNotFound) {
			return &gen.PutGroupsGroupIdJoinRequestsRequestId404JSONResponse{
				Code:    "1",
				Message: "申请记录不存在",
			}, nil
		}
		return nil, err
	}

	var targetApplication *models.JoinApplication
	for _, app := range applications {
		if app.RequestID == requestID {
			targetApplication = app
			break
		}
	}
	if targetApplication == nil {
		return &gen.PutGroupsGroupIdJoinRequestsRequestId404JSONResponse{
			Code:    "1",
			Message: "申请记录不存在",
		}, nil
	}

	var processErr error
	if action == "approve" {
		processErr = h.groupsService.ApproveJoinApplication(ctx, groupID, targetApplication.UserID, userID, requestID, targetApplication.Username)
	} else if action == "reject" {
		processErr = h.groupsService.RejectJoinApplication(ctx, groupID, targetApplication.UserID, userID, requestID, targetApplication.Username, request.Body.RejectReason)
	} else {
		return &gen.PutGroupsGroupIdJoinRequestsRequestId400JSONResponse{
			Code:    "1",
			Message: "无效的操作类型",
		}, nil
	}

	if processErr != nil {
		if errors.Is(processErr, appErrors.ErrGroupNotFound) {
			return &gen.PutGroupsGroupIdJoinRequestsRequestId404JSONResponse{
				Code:    "1",
				Message: "用户组不存在",
			}, nil
		}
		if errors.Is(processErr, appErrors.ErrRolePermissionDenied) {
			return &gen.PutGroupsGroupIdJoinRequestsRequestId403JSONResponse{
				Code:    "1",
				Message: "权限不足",
			}, nil
		}
		return nil, processErr
	}

	return &gen.PutGroupsGroupIdJoinRequestsRequestId200JSONResponse{
		Code: "0",
		Data: &map[string]interface{}{},
	}, nil
}

// 查看用户组的所有成员及其角色。需要是该组成员
func (h *GroupsHandler) GetGroupsGroupIdMembers(ctx context.Context, request gen.GetGroupsGroupIdMembersRequestObject) (gen.GetGroupsGroupIdMembersResponseObject, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	groupID := request.GroupId

	members, err := h.groupsService.GetMembersByGroupID(ctx, groupID)
	if err != nil {
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.GetGroupsGroupIdMembers404JSONResponse{
				Code:    "1",
				Message: "用户组不存在",
			}, nil
		}
		return nil, err
	}

	// 检查当前用户是否是群组成员
	if err := h.groupsService.CheckUserExistInGroup(ctx, groupID, userID); err != nil {
		if errors.Is(err, appErrors.ErrGroupMemberNotFound) {
			return &gen.GetGroupsGroupIdMembers403JSONResponse{
				Code:    "1",
				Message: "您不是该组成员",
			}, nil
		}
		if !errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return nil, err
		}
	}

	genMembers := make([]gen.GroupMember, len(members))
	for i, m := range members {
		genMembers[i] = gen.GroupMember{
			UserId:   m.UserID,
			Username: m.Username,
			Role:     m.Role,
			JoinedAt: int(m.CreatedAt.Unix()),
		}
	}

	return &gen.GetGroupsGroupIdMembers200JSONResponse{
		Code: "0",
		Data: genMembers,
	}, nil
}

// 用户组管理员移除指定成员
func (h *GroupsHandler) DeleteGroupsGroupIdMembersUserId(ctx context.Context, request gen.DeleteGroupsGroupIdMembersUserIdRequestObject) (gen.DeleteGroupsGroupIdMembersUserIdResponseObject, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	groupID := request.GroupId
	targetUserID := request.UserId

	// 检查用户组是否存在
	group, err := h.groupsService.GetGroupByGroupID(ctx, groupID)
	if err != nil {
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.DeleteGroupsGroupIdMembersUserId404JSONResponse{
				Code:    "1",
				Message: "用户组不存在",
			}, nil
		}
		return nil, err
	}

	// 检查目标用户是否是群组成员
	if err := h.groupsService.CheckMemberPermission(ctx, groupID, targetUserID); err != nil {
		if errors.Is(err, appErrors.ErrGroupMemberNotFound) {
			return &gen.DeleteGroupsGroupIdMembersUserId404JSONResponse{
				Code:    "1",
				Message: "指定用户不是该组成员",
			}, nil
		}
		if !errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return &gen.DeleteGroupsGroupIdMembersUserId403JSONResponse{
				Code:    "1",
				Message: "权限不足",
			}, nil
		}
	}

	// 检查操作者是否有权限
	if userID != group.CreatorID {
		return &gen.DeleteGroupsGroupIdMembersUserId403JSONResponse{
			Code:    "1",
			Message: "权限不足",
		}, nil
	}

	// 不能删除自己
	if userID == targetUserID {
		return &gen.DeleteGroupsGroupIdMembersUserId403JSONResponse{
			Code:    "1",
			Message: "不能删除自己",
		}, nil
	}

	// 执行删除操作
	err = h.groupsService.RemoveMemberFromGroup(ctx, groupID, targetUserID, userID)
	if err != nil {
		if errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return &gen.DeleteGroupsGroupIdMembersUserId403JSONResponse{
				Code:    "1",
				Message: "权限不足",
			}, nil
		}
		return nil, err
	}

	return &gen.DeleteGroupsGroupIdMembersUserId200JSONResponse{
		Code: "0",
		Data: &map[string]interface{}{},
	}, nil
}

// 查询当前登录用户在指定用户组中的状态，包括未关联、申请中、普通成员、管理员等
func (h *GroupsHandler) GetGroupsGroupIdMyStatus(ctx context.Context, request gen.GetGroupsGroupIdMyStatusRequestObject) (gen.GetGroupsGroupIdMyStatusResponseObject, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	groupID := request.GroupId

	userStatus, requestID, err := h.groupsService.GetUserGroupStatus(ctx, groupID, userID)
	if err != nil {
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.GetGroupsGroupIdMyStatus404JSONResponse{
				Code:    "1",
				Message: "用户组不存在",
			}, nil
		}
		if errors.Is(err, appErrors.ErrGroupMemberNotFound) {
			if userStatus == "none" {
				return &gen.GetGroupsGroupIdMyStatus200JSONResponse{
					Code: "0",
					Data: struct {
						JoinRequestId int                       `json:"joinRequestId,omitempty"`
						Message       string                    `json:"message,omitempty"`
						Status        gen.GroupMembershipStatus `json:"status"`
					}{
						Message: "您未申请加入该用户组",
						Status:  gen.GroupMembershipStatusNone,
					},
				}, nil
			}
		}
		return nil, err
	}
	var status gen.GroupMembershipStatus
	var joinRequestId int
	var message string

	switch userStatus {
	case "pending":
		status = gen.GroupMembershipStatusPending
		joinRequestId = requestID
		message = "您的加入申请正在等待审核"
	case "rejected":
		status = gen.GroupMembershipStatusRejected
		joinRequestId = requestID
		message = "您的加入申请已被拒绝"
	case "admin":
		status = "admin"
		message = "您是该用户组的管理员"
	default:
		status = gen.GroupMembershipStatusMember
		message = "您是该用户组的成员"
	}

	return &gen.GetGroupsGroupIdMyStatus200JSONResponse{
		Code: "0",
		Data: struct {
			JoinRequestId int                       `json:"joinRequestId,omitempty"`
			Message       string                    `json:"message,omitempty"`
			Status        gen.GroupMembershipStatus `json:"status"`
		}{
			JoinRequestId: joinRequestId,
			Message:       message,
			Status:        status,
		},
	}, nil

}
