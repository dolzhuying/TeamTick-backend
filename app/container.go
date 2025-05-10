package app

import (
	db "TeamTickBackend/dal"
	"TeamTickBackend/dal/dao"
	"TeamTickBackend/dal/redis" 
	"TeamTickBackend/pkg"

	"gorm.io/gorm"
)

type AppContainer struct {
	Db         *gorm.DB
	DaoFactory *dao.DAOFactory
	JwtHandler pkg.JwtHandler
}

func NewAppContainer() *AppContainer {
	db := db.InitDB()
	if db == nil {
		panic("Failed to initialize database")
	}

	redisClient := redis.InitRedis()
	if redisClient == nil {
		panic("Failed to initialize redis")
	}

	daoFactory := dao.NewDAOFactory(db, redisClient)

	jwtHandler, err := pkg.NewJwtHandler()
	if err != nil {
		panic("Failed to initialize JWT handler")
	}
	return &AppContainer{
		Db:         db,
		DaoFactory: daoFactory,
		JwtHandler: jwtHandler,
	}
}
