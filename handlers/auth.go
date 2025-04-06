package handlers

import (
	"TeamTickBackend/gen"
	"context"
)

type AuthHandler struct{}

func NewAuthHandler() gen.AuthServerInterface {
	handler:=&AuthHandler{}
	return gen.NewAuthStrictHandler(handler, nil)
}

func (h *AuthHandler) PostAuthLogin(ctx context.Context, request gen.PostAuthLoginRequestObject) (gen.PostAuthLoginResponseObject, error){

}


func (h*AuthHandler) PostAuthRegister(ctx context.Context, request gen.PostAuthRegisterRequestObject) (gen.PostAuthRegisterResponseObject, error){
	
}
