package handlers

import (
	"TeamTickBackend/app"
	"TeamTickBackend/gen"
	appErrors "TeamTickBackend/pkg/errors"
	service "TeamTickBackend/services"
	"context"
	"errors"
	"regexp"

	"gorm.io/gorm"
)

type AuthHandler struct {
	authService   *service.AuthService
	groupsService *service.GroupsService
}

func NewAuthHandler(container *app.AppContainer) gen.AuthServerInterface {
	emailService := service.NewEmailService(
		container.DaoFactory.EmailRedisDAO,
		"smtp.qq.com",
		465,
		"3087918372@qq.com",
		"miyspxbsjwctdeda",
	)
	authService := service.NewAuthService(
		container.DaoFactory.UserDAO,
		container.DaoFactory.TransactionManager,
		container.JwtHandler,
		container.DaoFactory.EmailRedisDAO,
		emailService,
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

	// 检查邮箱格式
	if !isValidEmail(email) {
		return &gen.PostAuthRegister400JSONResponse{
			Code:    "1",
			Message: "邮箱格式不正确",
		}, nil
	}

	// 检查用户名格式
	if !isValidUsername(username) {
		return &gen.PostAuthRegister400JSONResponse{
			Code:    "1",
			Message: "用户名不能包含特殊字符（如 . 或 @）",
		}, nil
	}

	// 检查邮箱是否已注册
	existingUser, err := h.authService.GetUserByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return &gen.PostAuthRegister409JSONResponse{
			Code:    "1",
			Message: "该邮箱已注册",
		}, nil
	}

	user, err := h.authService.AuthRegister(ctx, username, password, email, verificationCode)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserAlreadyExists) {
			return &gen.PostAuthRegister409JSONResponse{
				Code:    "1",
				Message: "用户名已存在",
			}, nil
		}
		if errors.Is(err, appErrors.ErrInvalidVerificationCode) {
			return &gen.PostAuthRegister401JSONResponse{
				Code:    "1",
				Message: "验证码错误",
			}, nil
		}
		if errors.Is(err, appErrors.ErrVerificationCodeExpiredOrNotFound) {
			return &gen.PostAuthRegister410JSONResponse{
				Code:    "1",
				Message: "验证码已过期或不存在",
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
	email := string(request.Body.Email)
	newPassword := request.Body.NewPassword
	verificationCode := request.Body.VerificationCode
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return &gen.PostAuthResetPassword401JSONResponse{
			Code:    "1",
			Message: "无权限操作",
		}, nil
	}
	// 检查邮箱格式
	if !isValidEmail(email) {
		return &gen.PostAuthResetPassword400JSONResponse{
			Code:    "1",
			Message: "邮箱格式不正确",
		}, nil
	}

	// 获取用户信息
	user, err := h.authService.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			return &gen.PostAuthResetPassword404JSONResponse{
				Code:    "1",
				Message: "该邮箱未注册",
			}, nil
		}
		return nil, err
	}

	if user.UserID != userID {
		return &gen.PostAuthResetPassword401JSONResponse{
			Code:    "1",
			Message: "无权限操作",
		}, nil
	}

	// 更新密码
	err = h.authService.AuthUpdatePassword(ctx, user.UserID, newPassword, verificationCode, email)
	if err != nil {
		if errors.Is(err, appErrors.ErrInvalidVerificationCode) {
			return &gen.PostAuthResetPassword401JSONResponse{
				Code:    "1",
				Message: "验证码错误",
			}, nil
		}
		if errors.Is(err, appErrors.ErrVerificationCodeExpiredOrNotFound) {
			return &gen.PostAuthResetPassword410JSONResponse{
				Code:    "1",
				Message: "验证码已过期或不存在",
			}, nil
		}
		if errors.Is(err, appErrors.ErrUserNotFound) {
			return &gen.PostAuthResetPassword404JSONResponse{
				Code:    "1",
				Message: "用户不存在",
			}, nil
		}
		return nil, err
	}

	message := "密码重置成功"
	return &gen.PostAuthResetPassword200JSONResponse{
		Code: "0",
		Data: &map[string]interface{}{
			"message": message,
		},
	}, nil
}

func (h *AuthHandler) PostAuthSendVerificationCode(ctx context.Context, request gen.PostAuthSendVerificationCodeRequestObject) (gen.PostAuthSendVerificationCodeResponseObject, error) {
	email := string(request.Body.Email)
	scene := request.Body.Scene

	// 检查邮箱格式
	if !isValidEmail(email) {
		return &gen.PostAuthSendVerificationCode400JSONResponse{
			Code:    "1",
			Message: "邮箱格式不正确",
		}, nil
	}

	// 检查邮箱是否已注册
	user, err := h.authService.GetUserByEmail(ctx, email)
	if err == nil && user != nil {
		// 如果是注册场景，邮箱已存在则返回错误
		if scene == "register" {
			return &gen.PostAuthSendVerificationCode409JSONResponse{
				Code:    "1",
				Message: "该邮箱已注册",
			}, nil
		}
	} else if scene == "reset_password" {
		// 如果是重置密码场景，邮箱不存在则返回错误
		if errors.Is(err, gorm.ErrRecordNotFound) || user == nil {
			return &gen.PostAuthSendVerificationCode404JSONResponse{
				Code:    "1",
				Message: "该邮箱未注册",
			}, nil
		}
		if err != nil {
			return nil, err
		}
	}

	// 生成验证码
	code, err := h.authService.GenerateVerificationCode(6)
	if err != nil {
		return nil, err
	}

	// 发送验证码邮件
	err = h.authService.SendVerificationEmail(ctx, email, code)
	if err != nil {
		if errors.Is(err, appErrors.ErrTooManyRequests) {
			return &gen.PostAuthSendVerificationCode429JSONResponse{
				Code:    "1",
				Message: "请求过于频繁，请稍后再试",
			}, nil
		}
		return nil, err
	}

	return &gen.PostAuthSendVerificationCode200JSONResponse{
		Code: "0",
		Data: &map[string]interface{}{
			"message": "验证码已发送",
		},
	}, nil
}

// isValidEmail 验证邮箱格式
func isValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(emailRegex, email)
	return match
}

// isValidUsername 验证用户名格式
func isValidUsername(username string) bool {
	// 用户名不能包含 . 或 @
	invalidChars := regexp.MustCompile(`[.@]`)
	return !invalidChars.MatchString(username)
}
