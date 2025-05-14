package main

import (
	"TeamTickBackend/app"
	"TeamTickBackend/pkg/cleaner"
	"TeamTickBackend/pkg/logger"
	"TeamTickBackend/router"
	"time"

	"go.uber.org/zap"
)

func main() {
	// 启动定时清理任务（只保留 export_files 目录下总大小不超过128MB的xlsx文件，每7天清理一次）
	cleaner.StartCleaner("export_files", 128, 7*24*time.Hour)

	// 初始化日志系统
	if err := logger.InitLogger(); err != nil {
		panic(err)
	}
	defer logger.Sync()

	// 初始化应用容器
	container := app.NewAppContainer()

	// 设置路由
	engine := router.SetupRouter(container)

	// 启动服务器
	if err := router.RunServer(engine, ":8080"); err != nil {
		logger.Error("服务器运行失败", zap.Error(err))
	}
}
