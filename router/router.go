package router

import (
	"TeamTickBackend/app"
	"TeamTickBackend/gen"
	"TeamTickBackend/handlers"
	"TeamTickBackend/middlewares"
	"TeamTickBackend/pkg/logger"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRouter(container *app.AppContainer) *gin.Engine {
	router := gin.Default()
	router.Use(middlewares.LoggerMiddleware())
	router.Use(middlewares.RecoverMiddleware())
	router.Use(middlewares.ResponseMiddleware())

	authHandler := handlers.NewAuthHandler(container)
	gen.RegisterAuthHandlers(router, authHandler)

	userHandler := handlers.NewUserHandler(container)
	userRouter := router.Group("")
	userRouter.Use(middlewares.AuthMiddleware(container.JwtHandler))
	gen.RegisterUsersHandlers(userRouter, userHandler)

	groupsHandler := handlers.NewGroupsHandler(container)
	groupsRouter := router.Group("")
	groupsRouter.Use(middlewares.AuthMiddleware(container.JwtHandler))
	gen.RegisterGroupsHandlers(groupsRouter, groupsHandler)

	taskHandler, checkinRecordsHandler := handlers.NewTaskHandler(container)
	taskRouter := router.Group("")
	taskRouter.Use(middlewares.AuthMiddleware(container.JwtHandler))
	gen.RegisterCheckinTasksHandlers(taskRouter, taskHandler)
	gen.RegisterCheckinRecordsHandlers(taskRouter, checkinRecordsHandler)

	// 注册审核请求相关路由
	auditRequestHandler := handlers.NewAuditRequestHandler(container)
	auditRequestRouter := router.Group("")
	auditRequestRouter.Use(middlewares.AuthMiddleware(container.JwtHandler))
	gen.RegisterAuditRequestsHandlers(auditRequestRouter, auditRequestHandler)

	return router
}


func RunServer(router *gin.Engine, addr string) error {
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	
	go func() {
		if err := router.Run(addr); err != nil {
			logger.Error("服务器启动失败", zap.Error(err))
		}
	}()

	
	<-quit
	logger.Info("正在关闭服务器...")
	return nil
}
