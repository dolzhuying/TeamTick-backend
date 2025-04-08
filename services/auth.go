package service

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/models"
	"TeamTickBackend/pkg"
	"context"
	"errors"

	"gorm.io/gorm"
)

type AuthService struct{}

//不同错误处理待考究,应该根据不同情况返回不同错误，这里暂时先简单返回err

func (s*AuthService) AuthRegister(ctx context.Context,username,password string)(*models.User,error){
	var createdUser models.User

	err:=dao.WithTransaction(ctx,func(tx *gorm.DB)error{
		_,err:=dao.DAOInstance.UserDAO.GetByUsername(ctx,username,tx)
		if err!=nil{
			return errors.New("用户已存在或其他错误")
		}
		hashedPassword,err:=pkg.GenerateFromPassword(password)
		if err!=nil{
			return errors.New("密码加密错误")
		}
		newUser:=models.User{
			Username:username,
			Password:hashedPassword,
		}
		if err:=dao.DAOInstance.UserDAO.Create(ctx,&newUser,tx);err!=nil{
			return errors.New("用户创建失败或其他错误")
		}
		createdUser=newUser
		return nil
	})
	if err!=nil{
		return nil,err
	}
	return &createdUser,nil

}

func (s*AuthService) AuthLogin(ctx context.Context,username,password string)(*models.User,string,error){
	var existUser models.User
	var userToken string

	err:=dao.WithTransaction(ctx,func(tx *gorm.DB)error{
		user,err:=dao.DAOInstance.UserDAO.GetByUsername(ctx,username,tx)
		if err!=nil{
			return errors.New("用户不存在或其他错误")
		}
		if !pkg.CheckPassword(user.Password,password){
			return errors.New("密码错误")
		}
		existUser=*user
		token,err:=pkg.JwtTokenInstance.GenerateJWTToken(user.Username,user.UserID)
		if err!=nil{
			return err
		}
		userToken=token

		return nil
	})
	if err!=nil{
		return nil,"",err
	}
	return &existUser,userToken,nil
}
