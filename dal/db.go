package db

import (
	"log"

	"TeamTickBackend/dal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 错误处理待完善，数据库配置考虑写到配置文件，后面做修改
func InitDB() *gorm.DB {
	// 连接到Docker中的MySQL
	var dsn = "root:root@tcp(localhost:3306)/teamtick?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Cannot open database: %v", err)
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Cannot configure database: %v", err)
	}
	//最大连接数等配置，仅作参考
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// 自动迁移所有表结构
	err = db.AutoMigrate(
		&models.User{},
		&models.Group{},
		&models.GroupMember{},
		&models.Task{},
		&models.TaskRecord{},
		&models.CheckApplication{},
		&models.JoinApplication{},
		&models.TaskRecord{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}
