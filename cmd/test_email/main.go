package main

import(
	"TeamTickBackend/services"
	"TeamTickBackend/app"
	"context"
	"log"
)

func main() {
	container := app.NewAppContainer()

	e:= service.NewEmailService(container.DaoFactory.EmailRedisDAO, "smtp.qq.com", 465, "3087918372@qq.com", "miyspxbsjwctdeda")
	code,err:=e.GenerateVerificationCode(6)
	if err!=nil{
		log.Fatalf("生成验证码失败: %v", err)
	}

	err=e.SendVerificationEmail(context.Background(), "1240800466@qq.com", code)
	if err!=nil{
		log.Fatalf("发送验证码失败: %v", err)
	}
}