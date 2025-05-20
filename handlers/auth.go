package handlers

import (
	"TeamTickBackend/app"
	"TeamTickBackend/gen"
	appErrors "TeamTickBackend/pkg/errors"
	service "TeamTickBackend/services"
	"context"
	"errors"
)

type AuthHandler struct {
	authService   *service.AuthService
	groupsService *service.GroupsService
}

func NewAuthHandler(container *app.AppContainer) gen.AuthServerInterface {
	authService := service.NewAuthService(
		container.DaoFactory.UserDAO,
		container.DaoFactory.TransactionManager,
		container.JwtHandler,
		container.DaoFactory.EmailRedisDAO,
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
	handler := &AuthHandler{
		authService:   authService,
		groupsService: groupsService,
	}
	return gen.NewAuthStrictHandler(handler, nil)
}

// 用户登录
func (h *AuthHandler) PostAuthLogin(ctx context.Context, request gen.PostAuthLoginRequestObject) (gen.PostAuthLoginResponseObject, error) {
	username := request.Body.Username
	password := request.Body.Password

	user, token, err := h.authService.AuthLogin(ctx, username, password)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			return &gen.PostAuthLogin401JSONResponse{
				Code:    "1",
				Message: "用户不存在",
			}, nil
		}
		if errors.Is(err, appErrors.ErrInvalidPassword) {
			return &gen.PostAuthLogin401JSONResponse{
				Code:    "1",
				Message: "用户名或密码错误",
			}, nil
		}
		return nil, err
	}

	username = user.Username
	userId := user.UserID

	data := struct {
		Token    string `json:"token,omitempty"`
		UserId   int    `json:"userId,omitempty"`
		Username string `json:"username,omitempty"`
	}{
		Token:    token,
		UserId:   userId,
		Username: username,
	}

	return gen.PostAuthLogin200JSONResponse{
		Code: "0",
		Data: data,
	}, nil
}

func (h *AuthHandler) PostAuthRegister(ctx context.Context, request gen.PostAuthRegisterRequestObject) (gen.PostAuthRegisterResponseObject, error) {
	username := request.Body.Username
	password := request.Body.Password
	email := string(request.Body.Email)
	verificationCode := request.Body.VerificationCode

	user, err := h.authService.AuthRegister(ctx, username, password, email, verificationCode)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserAlreadyExists) {
			return &gen.PostAuthRegister409JSONResponse{
				Code:    "1",
				Message: "用户名已存在",
			}, nil
		}

		return nil, err
	}

	userId := user.UserID
	return &gen.PostAuthRegister201JSONResponse{
		Code: "0",
		Data: gen.User{
			UserId:   userId,
			Username: user.Username,
		},
	}, nil
}

// PostAuthAdminLogin 管理员登录
func (h *AuthHandler) PostAuthAdminLogin(ctx context.Context, request gen.PostAuthAdminLoginRequestObject) (gen.PostAuthAdminLoginResponseObject, error) {
	username := request.Body.Username
	password := request.Body.Password

	// 先验证用户名和密码
	user, token, err := h.authService.AuthLogin(ctx, username, password)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			return &gen.PostAuthAdminLogin401JSONResponse{
				Code:    "1",
				Message: "用户不存在",
			}, nil
		}
		if errors.Is(err, appErrors.ErrInvalidPassword) {
			return &gen.PostAuthAdminLogin401JSONResponse{
				Code:    "1",
				Message: "用户名或密码错误",
			}, nil
		}
		return nil, err
	}

	// 查询用户作为管理员的组
	groups, err := h.groupsService.GetGroupsByUserID(ctx, user.UserID, "created")
	if err != nil {
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.PostAuthAdminLogin403JSONResponse{
				Code:    "1",
				Message: "非管理员用户",
			}, nil
		}
		return nil, err
	}

	// 如果没有管理员的组，返回403错误
	if len(groups) == 0 {
		return &gen.PostAuthAdminLogin403JSONResponse{
			Code:    "1",
			Message: "非管理员用户",
		}, nil
	}

	// 构建返回数据
	managedGroups := make([]struct {
		GroupId   int    `json:"groupId,omitempty"`
		GroupName string `json:"groupName,omitempty"`
	}, len(groups))
	for i, group := range groups {
		managedGroups[i] = struct {
			GroupId   int    `json:"groupId,omitempty"`
			GroupName string `json:"groupName,omitempty"`
		}{
			GroupId:   group.GroupID,
			GroupName: group.GroupName,
		}
	}

	return &gen.PostAuthAdminLogin200JSONResponse{
		Code: "0",
		Data: struct {
			ManagedGroups *[]struct {
				GroupId   int    `json:"groupId,omitempty"`
				GroupName string `json:"groupName,omitempty"`
			} `json:"managedGroups,omitempty"`
			Token    string `json:"token,omitempty"`
			UserId   int    `json:"userId,omitempty"`
			Username string `json:"username,omitempty"`
		}{
			ManagedGroups: &managedGroups,
			Token:         token,
			UserId:        user.UserID,
			Username:      user.Username,
		},
	}, nil
}

func (h *AuthHandler) PostAuthResetPassword(ctx context.Context, request gen.PostAuthResetPasswordRequestObject) (gen.PostAuthResetPasswordResponseObject, error) {

}

func (h *AuthHandler) PostAuthSendVerificationCode(ctx context.Context, request gen.PostAuthSendVerificationCodeRequestObject) (gen.PostAuthSendVerificationCodeResponseObject, error) {
	
}
