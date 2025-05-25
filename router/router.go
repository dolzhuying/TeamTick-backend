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
	router.Use(middlewares.CorsMiddleware())

	// 创建认证处理器
	authHandler := handlers.NewAuthHandler(container)

	// 注册认证相关路由
	authRouter := router.Group("/auth")
	authRouter.Use(func(c *gin.Context) {
		// 不需要认证的路径
		publicPaths := map[string]bool{
			"/auth/login":                  true,
			"/auth/register":               true,
			"/auth/send-verification-code": true,
			"/auth/admin/login":            true,
		}

		// 需要认证的路径
		authPaths := map[string]bool{
			"/auth/reset-password": true,
		}

		path := c.Request.URL.Path
		if publicPaths[path] {
			c.Next()
			return
		}

		if authPaths[path] {
			// 应用认证中间件
			authMiddleware := middlewares.AuthMiddleware(container.JwtHandler)
			authMiddleware(c)
			return
		}

		c.AbortWithStatus(404)
	})
	gen.RegisterAuthHandlers(authRouter, authHandler)

	// 其他需要认证的路由
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

	// 注册导出相关路由
	exportHandler := handlers.NewExportHandler(container)
	exportRouter := router.Group("")
	exportRouter.Use(middlewares.AuthMiddleware(container.JwtHandler))
	gen.RegisterExportHandlers(exportRouter, exportHandler)

	// 注册统计相关路由
	statisticsHandler := handlers.NewStatisticsHandler(container)
	statisticsRouter := router.Group("")
	statisticsRouter.Use(middlewares.AuthMiddleware(container.JwtHandler))
	gen.RegisterStatisticsHandlers(statisticsRouter, statisticsHandler)

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
