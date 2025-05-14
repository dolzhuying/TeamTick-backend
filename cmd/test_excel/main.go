package main

import (
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/pkg/logger"
	service "TeamTickBackend/services"
	"context"
	"log"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 初始化全局 logger
	logger.InitLogger()
	defer logger.Sync()

	// 初始化数据库连接
	dsn := "root:root@tcp(localhost:3306)/teamtick?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 初始化 DAO Factory
	daoFactory := dao.NewDAOFactory(db, nil)

	// 创建 StatisticsService 实例
	statisticsService := service.NewStatisticsService(
		daoFactory.StatisticsDAO,
		daoFactory.GroupDAO,
		daoFactory.TransactionManager,
		daoFactory.GroupMemberDAO,
	)

	// 创建上下文
	ctx := context.Background()

	// 设置测试参数
	groupIDs := []int{1,3}                      // 替换为实际存在的组ID
	startTime := time.Now().AddDate(0, -1, 0) // 一个月前
	endTime := time.Now()
	statuses := []string{"success", "absent", "exception"}

	// 调用函数生成Excel
	filePath, err := statisticsService.GenerateGroupSignInStatisticsExcel(ctx, groupIDs, startTime, endTime, statuses)
	if err != nil {
		logger.Error("生成Excel文件失败",
			zap.Error(err),
		)
		log.Fatalf("Failed to generate Excel: %v", err)
	}

	logger.Info("成功生成Excel文件",
		zap.String("filePath", filePath),
	)
}
