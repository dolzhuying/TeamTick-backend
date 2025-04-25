package handlers

import (
	"TeamTickBackend/app"
	"TeamTickBackend/gen"
	"TeamTickBackend/services"
	"context"
	"TeamTickBackend/middlewares"
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
	return gen.NewAuthStrictHandler(handler, []gen.AuthStrictMiddlewareFunc{
		middlewares.AuthRecoveryMiddleware(),
	})
}

//这里gen结构体的指针类型需要修改(md赶紧改@ych)（最好int64类型也换掉,省的多一步类型转换）
//还有gen生成的响应结构能不能别这么恶心，匿名结构体还得先在这里定义

// 实际需要对不同返回错误构造不同的返回响应体，这里先简单返回err
func (h *AuthHandler) PostAuthLogin(ctx context.Context, request gen.PostAuthLoginRequestObject) (gen.PostAuthLoginResponseObject, error) {
	username := request.Body.Username
	password := request.Body.Password

	user, token, err := h.authService.AuthLogin(ctx, username, *password)
	if err != nil {
		return nil, err
	}

	username = user.Username
	userId := int64(user.UserID)
	data := struct {
		Token    *string `json:"token,omitempty"`
		UserId   *int64  `json:"userId,omitempty"`
		Username *string `json:"username,omitempty"`
	}{
		Token:    &token,
		UserId:   &userId,
		Username: &username,
	}

	return &gen.PostAuthLogin200JSONResponse{
		Code: "200",
		Data: data,
	}, nil

}

func (h *AuthHandler) PostAuthRegister(ctx context.Context, request gen.PostAuthRegisterRequestObject) (gen.PostAuthRegisterResponseObject, error) {
	username := request.Body.Username
	password := request.Body.Password

	user, err := h.authService.AuthRegister(ctx, username, *password)
	if err != nil {
		return nil, err
	}

	userId := int64(user.UserID)
	return &gen.PostAuthRegister201JSONResponse{
		Code: "201",
		Data: gen.User{
			UserId:   &userId,
			Username: &user.Username,
		},
	}, nil
}
