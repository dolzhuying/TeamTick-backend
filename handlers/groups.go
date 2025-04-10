package handlers

import (
	"TeamTickBackend/app"
	"TeamTickBackend/gen"
	"TeamTickBackend/services"
	"context"
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

// 会议中提及过部分接口需要修改（查询用户组信息是否需要该用户组成员）
// 时序图部分逻辑存疑

func (h *GroupsHandler) GetGroups(ctx context.Context, request gen.GetGroupsRequestObject) (gen.GetGroupsResponseObject, error) {

}

func (h *GroupsHandler) PostGroups(ctx context.Context, request gen.PostGroupsRequestObject) (gen.PostGroupsResponseObject, error) {

}

func (h *GroupsHandler) GetGroupsGroupId(ctx context.Context, request gen.GetGroupsGroupIdRequestObject) (gen.GetGroupsGroupIdResponseObject, error) {

}

func (h *GroupsHandler) PutGroupsGroupId(ctx context.Context, request gen.PutGroupsGroupIdRequestObject) (gen.PutGroupsGroupIdResponseObject, error) {

}

func (h *GroupsHandler) GetGroupsGroupIdJoinRequests(ctx context.Context, request gen.GetGroupsGroupIdJoinRequestsRequestObject) (gen.GetGroupsGroupIdJoinRequestsResponseObject, error) {

}

func (h *GroupsHandler) PostGroupsGroupIdJoinRequests(ctx context.Context, request gen.PostGroupsGroupIdJoinRequestsRequestObject) (gen.PostGroupsGroupIdJoinRequestsResponseObject, error) {

}

func (h *GroupsHandler) PutGroupsGroupIdJoinRequestsRequestId(ctx context.Context, request gen.PutGroupsGroupIdJoinRequestsRequestIdRequestObject) (gen.PutGroupsGroupIdJoinRequestsRequestIdResponseObject, error) {

}

func (h *GroupsHandler) GetGroupsGroupIdMembers(ctx context.Context, request gen.GetGroupsGroupIdMembersRequestObject) (gen.GetGroupsGroupIdMembersResponseObject, error) {

}

func (h *GroupsHandler) DeleteGroupsGroupIdMembersUserId(ctx context.Context, request gen.DeleteGroupsGroupIdMembersUserIdRequestObject) (gen.DeleteGroupsGroupIdMembersUserIdResponseObject, error) {

}
