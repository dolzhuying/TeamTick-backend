package handlers

import (
	"TeamTickBackend/gen"
	"context"
)

type GroupsHandler struct{}

func NewGroupsHandler() gen.GroupsServerInterface {
	handler := &GroupsHandler{}
	return gen.NewGroupsStrictHandler(handler, nil)
}

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
