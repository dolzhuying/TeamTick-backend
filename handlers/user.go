package handlers

import (
	"TeamTickBackend/app"
	"TeamTickBackend/gen"
	appErrors "TeamTickBackend/pkg/errors"
	service "TeamTickBackend/services"
	"context"
	"errors"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(container *app.AppContainer) gen.UsersServerInterface {
	userService := service.NewUserService(
		container.DaoFactory.UserDAO,
		container.DaoFactory.TransactionManager,
	)
	handler := &UserHandler{
		userService: userService,
	}
	return gen.NewUsersStrictHandler(handler, nil)
}

func (h *UserHandler) GetUsersMe(ctx context.Context, request gen.GetUsersMeRequestObject) (gen.GetUsersMeResponseObject, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	user, err := h.userService.GetUserMe(ctx, userID)
	if err != nil {
		if errors.Is(err, appErrors.ErrUserNotFound) {
			return &gen.GetUsersMe401JSONResponse{
				Code:    "1",
				Message: "用户未登录",
			}, nil
		}
		return nil, err
	}

	userId := user.UserID
	genUser := gen.User{
		UserId:   userId,
		Username: user.Username,
	}

	return &gen.GetUsersMe200JSONResponse{
		Code: "0",
		Data: genUser,
	}, nil
}
