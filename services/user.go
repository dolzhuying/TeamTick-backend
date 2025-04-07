package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	"context"
	"errors"

	"gorm.io/gorm"
)

type UserService struct{}

func(s*UserService) GetUserMe(ctx context.Context,userID int)(*models.User,error){
	var existUser models.User

	err:=dao.WithTransaction(ctx,func(tx *gorm.DB)error{
		user,err:=dao.DAOInstance.UserDAO.GetByID(ctx,userID,tx)
		if err!=nil{
			return errors.New("用户不存在或其他错误")
		}
		existUser=*user
		return nil
	})
	if err!=nil{
		return nil,err
	}
	return &existUser,nil
}