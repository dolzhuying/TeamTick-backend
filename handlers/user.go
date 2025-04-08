package handlers

import (
	"TeamTickBackend/gen"
	"TeamTickBackend/services"
	"context"
	"errors"
)

type UserHandler struct{
	userService service.UserService
}

func NewUserHandler() gen.UsersServerInterface {
	userService:=service.UserService{}
	handler:=&UserHandler{
		userService: userService,
	}
	return gen.NewUsersStrictHandler(handler, nil)
}


func (h *UserHandler) GetUsersMe(ctx context.Context, request gen.GetUsersMeRequestObject) (gen.GetUsersMeResponseObject, error) {
	userID, ok := ctx.Value("userID").(int)
	if !ok{
		return nil,errors.New("用户未授权")
	}
	user,err:=h.userService.GetUserMe(ctx,userID)
	if err!=nil{
		return nil,err
	}
	
	userId:=int64(user.UserID)
	genUser:=gen.User{
		UserId: &userId,
		Username: &user.Username,
	}
	
	return &gen.GetUsersMe200JSONResponse{
		Code: "200",
		Data: genUser,
	},nil
}
