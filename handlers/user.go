package handlers

import(
	"TeamTickBackend/gen"
	"context"
)

type UserHandler struct{}

func NewUserHandler() gen.UsersServerInterface {
	handler := &UserHandler{}
	return gen.NewUsersStrictHandler(handler, nil)
}


func (h *UserHandler) GetUsersMe(ctx context.Context, request gen.GetUsersMeRequestObject) (gen.GetUsersMeResponseObject, error) {

}
