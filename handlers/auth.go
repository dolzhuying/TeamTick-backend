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
	authService service.AuthService
}

func NewAuthHandler(container *app.AppContainer) gen.AuthServerInterface {
	authService := service.NewAuthService(
		container.DaoFactory.UserDAO,
		container.DaoFactory.TransactionManager,
		container.JwtHandler,
	)
	handler := &AuthHandler{
		authService: *authService,
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

	user, err := h.authService.AuthRegister(ctx, username, password)
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
