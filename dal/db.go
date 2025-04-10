package db

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//错误处理待完善，数据库配置考虑写到配置文件，后面做修改
func InitDB() *gorm.DB {
	//dsn待定是否统一
	var dsn="dol:lsj041219@tcp(localhost:3306)/test?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
		Logger:                 logger.Default.LogMode(logger.Info),
	})
	if err!=nil{
		log.Fatalf("Cannot open databasae %v",err)
		return nil
	}
	sqlDB,err:=db.DB()
	if err!=nil{
		log.Fatalf("Cannot configure database %v",err)
	}
	//最大连接数等配置，仅作参考
	sqlDB.SetConnMaxIdleTime(10)
	sqlDB.SetMaxOpenConns(10)

	//迁移表结构
	//db.AutoMigrate(&models.User{})

	return db
}



